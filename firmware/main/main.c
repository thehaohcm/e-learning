// E-Learn study device (ES3C28P): vocabulary / PMP cards with spaced repetition,
// synced with the E-Learning server. See README.md.
#include <stdio.h>
#include <string.h>
#include <time.h>

#include "api.h"
#include "app.h"
#include "audio.h"
#include "board.h"
#include "esp_log.h"
#include "esp_system.h"
#include "freertos/FreeRTOS.h"
#include "freertos/queue.h"
#include "freertos/task.h"
#include "net.h"
#include "store.h"
#include "ui.h"
#include "ui_study.h"

static const char *TAG = "app";

#define SYNC_EVERY_S (10 * 60)  // fresh batch from the server
#define PUSH_EVERY_S 60         // pending answers
#define PAIR_POLL_S 5

static QueueHandle_t s_req;
static app_status_t s_st;
static time_t s_last_reminder;

void app_request(app_req_t r) { xQueueSend(s_req, &r, 0); }
const app_status_t *app_status(void) { return &s_st; }

/** Waits ms, handling UI requests and waking the screen meanwhile. Returns a request or -1. */
static int wait_request(int ms)
{
    app_req_t r;
    int64_t until = xTaskGetTickCount() + pdMS_TO_TICKS(ms);
    while ((int64_t)xTaskGetTickCount() < until) {
        if (xQueueReceive(s_req, &r, 0) == pdTRUE) return r;
        ui_wait(200);
    }
    return -1;
}

static void wifi_setup_forever(void)
{
    net_start_setup_ap();
    ui_show_setup(net_ap_ssid());
    for (;;) {          // the setup page restarts the device after saving
        ui_tick();
        ui_wait(1000);
    }
}

static bool have_cards(void)
{
    deck_t *d = deck_get();
    bool ok = false;
    for (int i = 0; d && i < d->n_cards; i++) ok |= !d->cards[i].done;
    deck_unref(d);
    return ok;
}

/** Minutes since midnight, user's time zone. -1 if the clock is unknown. */
static int local_minutes(const deck_t *d)
{
    time_t t = now_utc();
    if (!t) return -1;
    t += d ? d->utc_offset_sec : 7 * 3600;
    struct tm tm;
    gmtime_r(&t, &tm);
    return tm.tm_hour * 60 + tm.tm_min;
}

static int hhmm(const char *s)
{
    int h = 0, m = 0;
    return sscanf(s, "%d:%d", &h, &m) == 2 ? h * 60 + m : -1;
}

/** Device pop-up: inside the active hours, every N minutes of not being used. */
static void maybe_remind(void)
{
    if (!ui_sleeping() || study_active()) return;
    deck_t *d = deck_get();
    bool has = false;
    for (int i = 0; d && i < d->n_cards; i++) has |= !d->cards[i].done;
    if (!has || d->reminder_every_min <= 0) {
        deck_unref(d);
        return;
    }
    int now = local_minutes(d), from = hhmm(d->active_from), to = hhmm(d->active_to);
    bool in_hours = now >= 0 && from >= 0 && to >= 0 && (from <= to ? (now >= from && now < to) : (now >= from || now < to));
    time_t t = now_utc();
    int every = d->reminder_every_min * 60;
    deck_unref(d);
    if (!in_hours || !t) return;
    if (ui_idle_seconds() >= every && (!s_last_reminder || t - s_last_reminder >= every)) {
        s_last_reminder = t;
        ESP_LOGI(TAG, "study reminder");
        ui_reminder();
    }
}

static void do_push(void)
{
    if (!reviews_count()) return;
    api_push_reviews();
    s_st.pending = reviews_count();
}

/** Sends answers and downloads the next batch. Returns false if the device must be linked again. */
static bool do_sync(void)
{
    s_st.syncing = true;
    if (ui_on_home()) ui_show_home();
    do_push();
    esp_err_t err = study_active() ? ESP_OK : api_sync(); // never swap cards in the middle of a round
    s_st.syncing = false;
    s_st.pending = reviews_count();
    if (err == ESP_ERR_INVALID_STATE || err == ESP_ERR_NOT_FOUND) {
        return false; // token unknown or unlinked on the website
    }
    strlcpy(s_st.message, err == ESP_OK ? "" : "Could not sync: using the saved cards.", sizeof(s_st.message));
    if (ui_on_home() && !ui_sleeping()) ui_show_home();
    return true;
}

/** Registers (if needed) and waits until the user links the device on the website. */
static void link_device(void)
{
    for (;;) {
        if (!cfg()->token[0] && api_register() != ESP_OK) {
            ui_show_offline("Cannot reach the server. Check the server address in Wi-Fi setup.");
            int r = wait_request(15000);
            if (r == APP_REQ_WIFI_SETUP) {
                cfg_forget_wifi();
                esp_restart();
            }
            continue;
        }
        api_status_t st;
        esp_err_t err = api_status(&st);
        if (err == ESP_ERR_INVALID_STATE) {
            cfg_forget_token(); // removed on the website: register again
            continue;
        }
        if (err == ESP_OK && st.paired) {
            strlcpy(s_st.user, st.user_name, sizeof(s_st.user));
            return;
        }
        if (err == ESP_OK) ui_show_pair(st.pair_code);
        for (int i = 0; i < PAIR_POLL_S; i++) {
            ui_tick();
            int r = wait_request(1000);
            if (r == APP_REQ_WIFI_SETUP) {
                cfg_forget_wifi();
                esp_restart();
            }
        }
    }
}

static void handle(int r)
{
    switch (r) {
    case APP_REQ_SYNC:
    case APP_REQ_RETRY:
        if (!do_sync()) {
            cfg_forget_token();
            esp_restart();
        }
        break;
    case APP_REQ_PUSH:
        do_push();
        break;
    case APP_REQ_WIFI_SETUP:
        cfg_forget_wifi();
        esp_restart();
        break;
    case APP_REQ_UNPAIR:
        cfg_forget_token();
        esp_restart();
        break;
    }
}

void app_main(void)
{
    ESP_ERROR_CHECK(store_init());
    ESP_ERROR_CHECK(board_init());
    s_req = xQueueCreate(8, sizeof(app_req_t));
    lv_display_t *disp = board_display_init();
    ui_init(disp);
    ui_show_boot("Starting…");
    board_backlight(cfg()->brightness);
    if (audio_init() == ESP_OK) audio_set_volume(cfg()->volume);

    // cards saved last time: the device works offline right away
    deck_t *cached = deck_cache_load();
    if (cached) deck_set(cached);
    s_st.pending = reviews_count();

    net_init();
    if (!cfg()->ssid[0] || board_boot_pressed()) {
        wifi_setup_forever(); // first start, or BOOT held at power-on
    }
    ui_show_boot("Connecting to Wi-Fi…");
    while (net_connect(20000) != ESP_OK) {
        s_st.online = false;
        if (have_cards()) {
            // offline but with saved cards: study now, Wi-Fi keeps retrying in the background
            break;
        }
        ui_show_offline("Cannot connect to the Wi-Fi. It keeps trying in the background.");
        int r = wait_request(20000);
        if (r == APP_REQ_WIFI_SETUP) {
            cfg_forget_wifi();
            esp_restart();
        }
        if (net_connected()) break;
    }
    s_st.online = net_connected();
    if (s_st.online) {
        net_start_time();
        ui_show_boot("Connecting to the server…");
        link_device();
        if (!do_sync()) {
            cfg_forget_token();
            esp_restart();
        }
    }
    ui_show_home();

    int since_sync = 0, since_push = 0;
    bool was_online = s_st.online;
    for (;;) {
        int r = wait_request(1000);
        if (r >= 0) handle(r);
        ui_tick();
        s_st.online = net_connected();
        if (s_st.online && !was_online) { // Wi-Fi is back: send what was answered offline
            net_start_time();
            since_sync = SYNC_EVERY_S;
        }
        was_online = s_st.online;
        since_sync++;
        since_push++;
        if (s_st.online && since_push >= PUSH_EVERY_S) {
            since_push = 0;
            do_push();
        }
        if (s_st.online && since_sync >= SYNC_EVERY_S && !study_active()) {
            since_sync = 0;
            if (!do_sync()) {
                cfg_forget_token();
                esp_restart();
            }
        }
        maybe_remind();
    }
}
