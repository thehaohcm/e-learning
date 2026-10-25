#include "ui.h"

#include <stdio.h>
#include <string.h>
#include <time.h>

#include "app.h"
#include "audio.h"
#include "board.h"
#include "esp_app_desc.h"
#include "esp_lvgl_port.h"
#include "freertos/FreeRTOS.h"
#include "freertos/semphr.h"
#include "freertos/task.h"
#include "esp_system.h"
#include "net.h"
#include "store.h"
#include "ota.h"
#include "ui_study.h"
#include "ui_theme.h"

#define DIM_AFTER_S 45
#define SLEEP_AFTER_S 90

static lv_obj_t *s_clock, *s_status; // header of the current screen
static lv_obj_t *s_sleep_layer;
static bool s_sleeping, s_dimmed, s_home;

#define LOCK() if (lvgl_port_lock(0)) {
#define UNLOCK() lvgl_port_unlock(); }

// ---------------------------------------------------------------- common pieces

static void base_screen(lv_obj_t *scr)
{
    lv_obj_set_style_bg_color(scr, COL_BG, 0);
    lv_obj_set_style_pad_all(scr, 0, 0);
    lv_obj_remove_flag(scr, LV_OBJ_FLAG_SCROLLABLE);
}

static void header_text(char *buf, size_t len)
{
    const app_status_t *st = app_status();
    int bat = board_battery_percent();
    char batt[24] = "";
    if (bat >= 0) {
        const char *sym = bat > 80 ? LV_SYMBOL_BATTERY_FULL : bat > 55 ? LV_SYMBOL_BATTERY_3
                        : bat > 30 ? LV_SYMBOL_BATTERY_2 : bat > 10 ? LV_SYMBOL_BATTERY_1 : LV_SYMBOL_BATTERY_EMPTY;
        snprintf(batt, sizeof(batt), "  %s %d%%", sym, bat);
    }
    char pend[24] = "";
    if (st->pending > 0) snprintf(pend, sizeof(pend), LV_SYMBOL_UPLOAD " %d  ", st->pending);
    snprintf(buf, len, "%s%s%s%s", pend, st->syncing ? LV_SYMBOL_REFRESH "  " : "",
             st->online ? LV_SYMBOL_WIFI : LV_SYMBOL_CLOSE, batt);
}

static void clock_text(char *buf, size_t len)
{
    time_t t = now_utc();
    if (!t) {
        strlcpy(buf, "--:--", len);
        return;
    }
    deck_t *d = deck_get();
    int off = d ? d->utc_offset_sec : 7 * 3600;
    deck_unref(d);
    t += off;
    struct tm tm;
    gmtime_r(&t, &tm);
    snprintf(buf, len, "%02d:%02d", tm.tm_hour, tm.tm_min);
}

static void home_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    ui_go_home_locked();
}

/** "12:10  <wifi> <battery>" for the right of the header. */
static void status_line(char *buf, size_t len)
{
    char c[16], h[64];
    clock_text(c, sizeof(c));
    header_text(h, sizeof(h));
    snprintf(buf, len, "%s%s%s", c, c[0] ? "  " : "", h);
}

lv_obj_t *ui_screen_with_header(const char *title, bool back_home)
{
    lv_obj_t *scr = lv_obj_create(NULL);
    base_screen(scr);
    lv_obj_t *bar = lv_obj_create(scr);
    lv_obj_set_size(bar, lv_pct(100), 28);
    lv_obj_set_style_bg_color(bar, COL_HEADER, 0);
    lv_obj_set_style_radius(bar, 0, 0);
    lv_obj_set_style_border_width(bar, 0, 0);
    lv_obj_set_style_pad_hor(bar, 8, 0);
    lv_obj_set_style_pad_ver(bar, 0, 0);
    lv_obj_remove_flag(bar, LV_OBJ_FLAG_SCROLLABLE);

    if (back_home) {
        lv_obj_t *b = lv_label_create(bar);
        lv_label_set_text(b, LV_SYMBOL_HOME);
        lv_obj_set_style_text_color(b, lv_color_white(), 0);
        lv_obj_set_style_text_font(b, &lv_font_montserrat_14, 0);
        lv_obj_align(b, LV_ALIGN_LEFT_MID, 0, 0);
        lv_obj_add_flag(b, LV_OBJ_FLAG_CLICKABLE);
        lv_obj_set_ext_click_area(b, 14);
        lv_obj_add_event_cb(b, home_clicked, LV_EVENT_CLICKED, NULL);
    }
    // title on the left (cut with "…" if long), "12:10  wifi  94%" on the right: never overlapping
    s_clock = NULL; // the clock is part of the status text
    s_status = lv_label_create(bar);
    lv_obj_set_style_text_color(s_status, lv_color_hex(0xC7D2FE), 0);
    lv_obj_set_style_text_font(s_status, font_small, 0);
    lv_obj_align(s_status, LV_ALIGN_RIGHT_MID, 0, 0);
    char buf[96];
    status_line(buf, sizeof(buf));
    lv_label_set_text(s_status, buf);

    // the title gets what is left, on ONE line ("…" when too long)
    lv_point_t sz;
    lv_text_get_size(&sz, buf, font_small, 0, 0, LV_COORD_MAX, LV_TEXT_FLAG_NONE);
    int left = back_home ? 26 : 0;
    int w = lv_display_get_horizontal_resolution(NULL) - 16 - left - sz.x - 12 - 24; // 24: room for "⬆ n"
    lv_obj_t *t = lv_label_create(bar);
    lv_label_set_text(t, title ? title : "");
    lv_obj_set_style_text_color(t, lv_color_white(), 0);
    lv_obj_set_style_text_font(t, font_bold, 0);
    lv_label_set_long_mode(t, LV_LABEL_LONG_DOT);
    lv_obj_set_size(t, w > 60 ? w : 60, lv_font_get_line_height(font_bold));
    lv_obj_align(t, LV_ALIGN_LEFT_MID, left, 0);
    s_home = false;
    return scr;
}

static lv_obj_t *body(lv_obj_t *scr)
{
    lv_obj_t *b = lv_obj_create(scr);
    lv_obj_set_size(b, lv_pct(100), lv_display_get_vertical_resolution(NULL) - 28);
    lv_obj_align(b, LV_ALIGN_BOTTOM_MID, 0, 0);
    lv_obj_set_style_bg_opa(b, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(b, 0, 0);
    lv_obj_set_style_pad_all(b, 10, 0);
    lv_obj_set_style_pad_row(b, 8, 0);
    lv_obj_set_flex_flow(b, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    return b;
}

static void load(lv_obj_t *scr)
{
    lv_screen_load_anim(scr, LV_SCR_LOAD_ANIM_FADE_IN, 150, 0, true);
}

// ---------------------------------------------------------------- simple screens

void ui_show_boot(const char *msg)
{
    LOCK();
    lv_obj_t *scr = lv_obj_create(NULL);
    base_screen(scr);
    lv_obj_set_style_bg_color(scr, COL_HEADER, 0);
    lv_obj_t *t = lv_label_create(scr);
    lv_label_set_text(t, "E-Learning");
    lv_obj_set_style_text_font(t, font_word, 0);
    lv_obj_set_style_text_color(t, lv_color_white(), 0);
    lv_obj_align(t, LV_ALIGN_CENTER, 0, -30);
    lv_obj_t *s = lv_label_create(scr);
    lv_label_set_text(s, "English · PMP");
    lv_obj_set_style_text_font(s, font_small, 0);
    lv_obj_set_style_text_color(s, lv_color_hex(0xFDBA74), 0);
    lv_obj_align(s, LV_ALIGN_CENTER, 0, 4);
    lv_obj_t *sp = lv_spinner_create(scr);
    lv_obj_set_size(sp, 28, 28);
    lv_obj_align(sp, LV_ALIGN_CENTER, 0, 46);
    lv_obj_t *m = ui_label(scr, msg, font_small, lv_color_hex(0xC7D2FE), lv_pct(90));
    lv_obj_set_style_text_align(m, LV_TEXT_ALIGN_CENTER, 0);
    lv_obj_align(m, LV_ALIGN_BOTTOM_MID, 0, -14);
    s_clock = s_status = NULL;
    s_home = false;
    load(scr);
    UNLOCK();
}

void ui_show_setup(const char *ap_ssid)
{
    LOCK();
    lv_obj_t *scr = ui_screen_with_header("Wi-Fi setup", false);
    lv_obj_t *row = lv_obj_create(scr);
    lv_obj_set_size(row, lv_pct(100), lv_display_get_vertical_resolution(NULL) - 28);
    lv_obj_align(row, LV_ALIGN_BOTTOM_MID, 0, 0);
    lv_obj_set_style_bg_opa(row, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(row, 0, 0);
    lv_obj_set_style_pad_all(row, 10, 0);
    lv_obj_set_flex_flow(row, LV_FLEX_FLOW_ROW);
    lv_obj_set_flex_align(row, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_set_style_pad_column(row, 12, 0);

    // QR code: joins the open setup network on most phones
    char wifi[64];
    snprintf(wifi, sizeof(wifi), "WIFI:T:nopass;S:%s;;", ap_ssid);
    lv_obj_t *qr = lv_qrcode_create(row);
    lv_qrcode_set_size(qr, 120);
    lv_qrcode_set_dark_color(qr, COL_TEXT);
    lv_qrcode_set_light_color(qr, lv_color_white());
    lv_qrcode_update(qr, wifi, strlen(wifi));
    lv_obj_set_style_border_color(qr, lv_color_white(), 0);
    lv_obj_set_style_border_width(qr, 5, 0);

    lv_obj_t *col = lv_obj_create(row);
    lv_obj_set_style_bg_opa(col, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(col, 0, 0);
    lv_obj_set_style_pad_all(col, 0, 0);
    lv_obj_set_flex_grow(col, 1);
    lv_obj_set_height(col, LV_SIZE_CONTENT);
    lv_obj_set_flex_flow(col, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_style_pad_row(col, 6, 0);
    ui_label(col, "1. On your phone, join the Wi-Fi:", font_small, COL_MUTED, -1);
    ui_label(col, ap_ssid, font_title, COL_PRIMARY, -1);
    ui_label(col, "2. The setup page opens by itself (or go to 192.168.4.1).", font_small, COL_MUTED, -1);
    ui_label(col, "3. Choose your Wi-Fi and press Save.", font_small, COL_MUTED, -1);
    load(scr);
    UNLOCK();
}

void ui_show_pair(const char *code)
{
    LOCK();
    lv_obj_t *scr = ui_screen_with_header("Link this device", false);
    lv_obj_t *b = body(scr);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    ui_label(b, "On the website open Settings → Link a device and type this code:", font_small, COL_MUTED, -1);
    char spaced[16];
    snprintf(spaced, sizeof(spaced), "%.3s %.3s", code, code + (strlen(code) > 3 ? 3 : 0));
    lv_obj_t *c = lv_label_create(b);
    lv_label_set_text(c, spaced);
    lv_obj_set_style_text_font(c, font_huge, 0);
    lv_obj_set_style_text_color(c, COL_PRIMARY, 0);
    char host[96];
    const char *s = strstr(cfg()->server, "://");
    strlcpy(host, s ? s + 3 : cfg()->server, sizeof(host));
    ui_label(b, host, font_bold, COL_TEXT, LV_SIZE_CONTENT);
    ui_label(b, "The code changes every 15 minutes. This screen moves on by itself.", font_small, COL_MUTED, -1);
    load(scr);
    UNLOCK();
}

static void retry_clicked(lv_event_t *e) { app_request(APP_REQ_RETRY); }
static void wifi_clicked(lv_event_t *e) { app_request(APP_REQ_WIFI_SETUP); }

void ui_show_offline(const char *msg)
{
    LOCK();
    lv_obj_t *scr = ui_screen_with_header("No connection", false);
    lv_obj_t *b = body(scr);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_t *i = lv_label_create(b);
    lv_label_set_text(i, LV_SYMBOL_WIFI);
    lv_obj_set_style_text_font(i, &lv_font_montserrat_20, 0);
    lv_obj_set_style_text_color(i, COL_BAD, 0);
    lv_obj_t *m = ui_label(b, msg, font_body, COL_TEXT, -1);
    lv_obj_set_style_text_align(m, LV_TEXT_ALIGN_CENTER, 0);
    lv_obj_t *row = lv_obj_create(b);
    lv_obj_set_size(row, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(row, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(row, 0, 0);
    lv_obj_set_flex_flow(row, LV_FLEX_FLOW_ROW);
    lv_obj_set_flex_align(row, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_add_event_cb(ui_button(row, LV_SYMBOL_REFRESH " Try again", COL_PRIMARY, NULL), retry_clicked, LV_EVENT_CLICKED, NULL);
    lv_obj_add_event_cb(ui_button(row, LV_SYMBOL_WIFI " Wi-Fi setup", COL_MUTED, NULL), wifi_clicked, LV_EVENT_CLICKED, NULL);
    deck_t *d = deck_get();
    if (d && d->n_cards) {
        ui_label(b, "You can still study the cards saved on the device: they are sent when the Wi-Fi is back.",
                 font_small, COL_MUTED, -1);
        lv_obj_t *sb = ui_button(b, LV_SYMBOL_PLAY " Study offline", COL_OK, NULL);
        lv_obj_add_event_cb(sb, study_clicked_cb, LV_EVENT_CLICKED, NULL);
    }
    deck_unref(d);
    load(scr);
    UNLOCK();
}

// ---------------------------------------------------------------- home

static lv_obj_t *stat_box(lv_obj_t *parent, const char *value, const char *label, lv_color_t color)
{
    lv_obj_t *box = lv_obj_create(parent);
    lv_obj_set_style_bg_color(box, COL_CARD, 0);
    lv_obj_set_style_radius(box, 12, 0);
    lv_obj_set_style_border_width(box, 0, 0);
    lv_obj_set_style_pad_all(box, 3, 0);
    lv_obj_set_style_pad_row(box, 0, 0);
    lv_obj_set_flex_grow(box, 1);
    lv_obj_set_height(box, LV_SIZE_CONTENT);
    lv_obj_set_flex_flow(box, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(box, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_remove_flag(box, LV_OBJ_FLAG_SCROLLABLE);
    lv_obj_t *v = lv_label_create(box);
    lv_label_set_text(v, value);
    lv_obj_set_style_text_font(v, font_word, 0);
    lv_obj_set_style_text_color(v, color, 0);
    lv_obj_t *l = lv_label_create(box);
    lv_label_set_text(l, label);
    lv_obj_set_style_text_font(l, font_small, 0);
    lv_obj_set_style_text_color(l, COL_MUTED, 0);
    return box;
}

// ---------------------------------------------------------------- busy overlay
// A grey veil over the whole screen with a spinner: shown while the device
// talks to the server after a tap (+5 new words, Sync), so nothing can be
// tapped twice. Removed by ui_busy_end(), or after 60 s at the latest.

static lv_obj_t *s_busy;
static lv_obj_t *s_busy_label;
static lv_timer_t *s_busy_timer;

static void busy_end_locked(void)
{
    if (s_busy_timer) {
        lv_timer_delete(s_busy_timer);
        s_busy_timer = NULL;
    }
    if (s_busy) {
        lv_obj_delete(s_busy);
        s_busy = NULL;
        s_busy_label = NULL;
    }
}

static void busy_timeout_cb(lv_timer_t *t)
{
    s_busy_timer = NULL; // one-shot: LVGL deletes it
    if (s_busy) {
        lv_obj_delete(s_busy);
        s_busy = NULL;
        s_busy_label = NULL;
    }
}

/** Call with the LVGL lock held (e.g. from a button). */
static void busy_locked(const char *msg)
{
    busy_end_locked();
    s_busy = lv_obj_create(lv_layer_top());
    lv_obj_remove_style_all(s_busy);
    lv_obj_set_size(s_busy, lv_pct(100), lv_pct(100));
    lv_obj_set_style_bg_color(s_busy, lv_color_hex(0x111827), 0);
    lv_obj_set_style_bg_opa(s_busy, LV_OPA_60, 0);
    lv_obj_add_flag(s_busy, LV_OBJ_FLAG_CLICKABLE); // swallows taps underneath
    lv_obj_set_flex_flow(s_busy, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(s_busy, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_set_style_pad_row(s_busy, 12, 0);
    lv_obj_t *sp = lv_spinner_create(s_busy);
    lv_spinner_set_anim_params(sp, 1000, 270);
    lv_obj_set_size(sp, 56, 56);
    lv_obj_set_style_arc_width(sp, 6, LV_PART_MAIN);
    lv_obj_set_style_arc_width(sp, 6, LV_PART_INDICATOR);
    lv_obj_set_style_arc_color(sp, lv_color_white(), LV_PART_INDICATOR);
    lv_obj_set_style_arc_opa(sp, LV_OPA_30, LV_PART_MAIN);
    s_busy_label = ui_label(s_busy, msg ? msg : "", font_body, lv_color_white(), LV_SIZE_CONTENT);
    lv_obj_set_style_text_align(s_busy_label, LV_TEXT_ALIGN_CENTER, 0);
    s_busy_timer = lv_timer_create(busy_timeout_cb, 60000, NULL);
    lv_timer_set_repeat_count(s_busy_timer, 1);
}

void ui_busy_locked(const char *msg) { busy_locked(msg); }
void ui_busy_end_locked(void) { busy_end_locked(); }
void ui_busy_text_locked(const char *msg)
{
    if (s_busy_label) lv_label_set_text(s_busy_label, msg);
}

void ui_busy(const char *msg)
{
    LOCK();
    busy_locked(msg);
    UNLOCK();
}

void ui_busy_end(void)
{
    LOCK();
    busy_end_locked();
    UNLOCK();
}

static void sync_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    busy_locked("Syncing…");
    app_request(APP_REQ_SYNC);
}

static void settings_clicked(lv_event_t *e);

void study_clicked_cb(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    study_start(0, false);
}

static void quiz_clicked_cb(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    app_request(APP_REQ_PREP_FORMS); // word-form questions for the next quizzes (AI, on the server)
    quiz_start(10);
}

static void extra_new_cb(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    busy_locked("Getting 5 new words…");
    app_request(APP_REQ_EXTRA_NEW);
}

static void build_home(void)
{
    study_stop(); // leaving a round half-way must not block syncing and pop-ups
    deck_t *d = deck_get();
    const app_status_t *st = app_status();
    char title[64];
    snprintf(title, sizeof(title), "Hi%s%s", st->user[0] ? ", " : "", st->user);
    lv_obj_t *scr = ui_screen_with_header(title, false);
    lv_obj_t *b = body(scr);

    lv_obj_t *row = lv_obj_create(b);
    lv_obj_set_size(row, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(row, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(row, 0, 0);
    lv_obj_set_style_pad_all(row, 0, 0);
    lv_obj_set_style_pad_column(row, 8, 0);
    lv_obj_set_flex_flow(row, LV_FLEX_FLOW_ROW);
    lv_obj_remove_flag(row, LV_OBJ_FLAG_SCROLLABLE);
    char v[16];
    int due = 0, fresh = 0;
    if (d) {
        for (int i = 0; i < d->n_cards; i++) {
            if (d->cards[i].done) continue;
            if (d->cards[i].is_new) fresh++;
            else due++;
        }
    }
    snprintf(v, sizeof(v), "%d", due);
    stat_box(row, v, "to review", COL_PRIMARY);
    snprintf(v, sizeof(v), "%d", fresh);
    stat_box(row, v, "new words", COL_OK);
    snprintf(v, sizeof(v), "%d", d ? d->streak_days : 0);
    stat_box(row, v, "day streak", COL_WARN);

    char lbl[48];
    bool can_quiz = d && d->n_quiz > 0;
    if (due + fresh == 0) {
        lv_obj_t *m = ui_label(b, d ? "All done for now! New cards come back when they are due."
                                    : "No cards yet: press Sync.", font_body, COL_OK, -1);
        lv_obj_set_style_text_align(m, LV_TEXT_ALIGN_CENTER, 0);
    }
    bool can_extra = due + fresh == 0 && d && d->extra_new_left > 0 && st->online;
    if (due + fresh > 0 || can_quiz || can_extra) {
        // [ Study N cards ] [ Quiz ]
        lv_obj_t *row1 = lv_obj_create(b);
        lv_obj_set_size(row1, lv_pct(100), LV_SIZE_CONTENT);
        lv_obj_set_style_bg_opa(row1, LV_OPA_TRANSP, 0);
        lv_obj_set_style_border_width(row1, 0, 0);
        lv_obj_set_style_pad_all(row1, 0, 0);
        lv_obj_set_style_pad_column(row1, 8, 0);
        lv_obj_set_flex_flow(row1, LV_FLEX_FLOW_ROW);
        lv_obj_remove_flag(row1, LV_OBJ_FLAG_SCROLLABLE);
        if (due + fresh > 0) {
            snprintf(lbl, sizeof(lbl), LV_SYMBOL_PLAY "  Study %d", due + fresh);
            lv_obj_t *sb = ui_button(row1, lbl, COL_PRIMARY, font_title);
            lv_obj_set_flex_grow(sb, 2);
            lv_obj_set_height(sb, 46);
            lv_obj_add_event_cb(sb, study_clicked_cb, LV_EVENT_CLICKED, NULL);
        }
        if (can_quiz) {
            lv_obj_t *qb = ui_button(row1, LV_SYMBOL_SHUFFLE "  Quiz", COL_OK, font_title);
            lv_obj_set_flex_grow(qb, 1);
            lv_obj_set_height(qb, 46);
            lv_obj_add_event_cb(qb, quiz_clicked_cb, LV_EVENT_CLICKED, NULL);
        }
        if (can_extra) {
            lv_obj_t *xb = ui_button(row1, LV_SYMBOL_PLUS " 5 new", COL_INFO, font_title);
            lv_obj_set_flex_grow(xb, 1);
            lv_obj_set_height(xb, 46);
            lv_obj_add_event_cb(xb, extra_new_cb, LV_EVENT_CLICKED, NULL);
        }
    }

    lv_obj_t *row2 = lv_obj_create(b);
    lv_obj_set_size(row2, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(row2, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(row2, 0, 0);
    lv_obj_set_style_pad_all(row2, 0, 0);
    lv_obj_set_style_pad_column(row2, 8, 0);
    lv_obj_set_flex_flow(row2, LV_FLEX_FLOW_ROW);
    lv_obj_remove_flag(row2, LV_OBJ_FLAG_SCROLLABLE);
    lv_obj_t *sy = ui_button(row2, st->syncing ? LV_SYMBOL_REFRESH " Syncing…" : LV_SYMBOL_REFRESH " Sync", COL_INFO, NULL);
    lv_obj_set_flex_grow(sy, 1);
    lv_obj_add_event_cb(sy, sync_clicked, LV_EVENT_CLICKED, NULL);
    lv_obj_t *se = ui_button(row2, LV_SYMBOL_SETTINGS " Settings", COL_MUTED, NULL);
    lv_obj_set_flex_grow(se, 1);
    lv_obj_add_event_cb(se, settings_clicked, LV_EVENT_CLICKED, NULL);

    if (st->message[0]) {
        ui_label(b, st->message, font_small, COL_BAD, -1);
    } else if (d && ota_is_newer(d->fw_version)) {
        char m[64];
        snprintf(m, sizeof(m), LV_SYMBOL_DOWNLOAD " New firmware %s: Settings → Update", d->fw_version);
        ui_label(b, m, font_small, COL_OK, -1);
    } else if (d && d->reviews_today > 0) {
        char m[48];
        snprintf(m, sizeof(m), "%d review%s today. Keep going!", d->reviews_today, d->reviews_today > 1 ? "s" : "");
        ui_label(b, m, font_small, COL_MUTED, -1);
    }
    deck_unref(d);
    load(scr);
    s_home = true;
}

void ui_show_home(void)
{
    LOCK();
    build_home();
    UNLOCK();
}

void ui_go_home_locked(void) { build_home(); }

bool ui_on_home(void) { return s_home; }

// ---------------------------------------------------------------- settings

static void bright_changed(lv_event_t *e)
{
    lv_obj_t *s = lv_event_get_target(e);
    cfg()->brightness = lv_slider_get_value(s);
    board_backlight(cfg()->brightness);
}

static void volume_changed(lv_event_t *e)
{
    lv_obj_t *s = lv_event_get_target(e);
    cfg()->volume = lv_slider_get_value(s);
    audio_set_volume(cfg()->volume);
}

static void speak_changed(lv_event_t *e)
{
    lv_obj_t *sw = lv_event_get_target(e);
    cfg()->speak = lv_obj_has_state(sw, LV_STATE_CHECKED);
    cfg_save();
    audio_fx(AUDIO_FX_TAP);
}

static void recall_changed(lv_event_t *e)
{
    lv_obj_t *sw = lv_event_get_target(e);
    cfg()->recall_first = lv_obj_has_state(sw, LV_STATE_CHECKED);
    cfg_save();
    audio_fx(AUDIO_FX_TAP);
}

static void slider_released(lv_event_t *e)
{
    cfg_save();
    audio_fx(AUDIO_FX_TAP);
}

static void confirm_cb(lv_event_t *e)
{
    app_req_t r = (app_req_t)(intptr_t)lv_event_get_user_data(e);
    lv_obj_t *mbox = lv_obj_get_parent(lv_obj_get_parent(lv_event_get_target(e)));
    lv_msgbox_close_async(mbox); // not while its own button is still handling the tap
    if (r == APP_REQ_WIFI_SETUP || r == APP_REQ_UNPAIR) {
        // done right here (the app task may be busy syncing): forget, show it, restart
        if (r == APP_REQ_WIFI_SETUP) cfg_forget_wifi();
        else cfg_forget_token();
        lv_obj_t *scr = lv_obj_create(NULL);
        base_screen(scr);
        lv_obj_t *l = ui_label(scr, "Restarting…", font_title, COL_TEXT, LV_SIZE_CONTENT);
        lv_obj_center(l);
        lv_screen_load(scr);
        lv_refr_now(NULL);
        vTaskDelay(pdMS_TO_TICKS(300));
        esp_restart();
    }
    app_request(r);
}

static void ask(const char *text, app_req_t r)
{
    lv_obj_t *mb = lv_msgbox_create(NULL);
    lv_msgbox_add_text(mb, text);
    lv_obj_t *yes = lv_msgbox_add_footer_button(mb, "Yes");
    lv_obj_add_event_cb(yes, confirm_cb, LV_EVENT_CLICKED, (void *)(intptr_t)r);
    lv_msgbox_add_close_button(mb);
}

static void wifi_reset_clicked(lv_event_t *e) { ask("Forget this Wi-Fi and open the setup page?", APP_REQ_WIFI_SETUP); }
static void unpair_clicked(lv_event_t *e) { ask("Unlink from your account? You will need a new code.", APP_REQ_UNPAIR); }

static lv_obj_t *setting_row(lv_obj_t *parent, const char *icon, const char *name)
{
    lv_obj_t *r = lv_obj_create(parent);
    lv_obj_set_size(r, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_color(r, COL_CARD, 0);
    lv_obj_set_style_border_width(r, 0, 0);
    lv_obj_set_style_radius(r, 10, 0);
    lv_obj_set_style_pad_all(r, 8, 0);
    lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
    lv_obj_set_flex_align(r, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_set_style_pad_column(r, 10, 0);
    lv_obj_remove_flag(r, LV_OBJ_FLAG_SCROLLABLE);
    lv_obj_t *i = lv_label_create(r);
    lv_label_set_text(i, icon);
    lv_obj_set_style_text_color(i, COL_PRIMARY, 0);
    lv_obj_t *n = lv_label_create(r);
    lv_label_set_text(n, name);
    lv_obj_set_width(n, 84);
    return r;
}

static void update_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    app_request(APP_REQ_UPDATE);
}

static void settings_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    lv_obj_t *scr = ui_screen_with_header("Settings", true);
    lv_obj_t *b = body(scr);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_START);
    lv_obj_add_flag(b, LV_OBJ_FLAG_SCROLLABLE);

    lv_obj_t *r = setting_row(b, LV_SYMBOL_EYE_OPEN, "Brightness");
    lv_obj_t *s = lv_slider_create(r);
    lv_slider_set_range(s, 10, 100);
    lv_slider_set_value(s, cfg()->brightness, LV_ANIM_OFF);
    lv_obj_set_flex_grow(s, 1);
    lv_obj_add_event_cb(s, bright_changed, LV_EVENT_VALUE_CHANGED, NULL);
    lv_obj_add_event_cb(s, slider_released, LV_EVENT_RELEASED, NULL);

    r = setting_row(b, LV_SYMBOL_VOLUME_MAX, "Volume");
    s = lv_slider_create(r);
    lv_slider_set_range(s, 0, 100);
    lv_slider_set_value(s, cfg()->volume, LV_ANIM_OFF);
    lv_obj_set_flex_grow(s, 1);
    lv_obj_add_event_cb(s, volume_changed, LV_EVENT_VALUE_CHANGED, NULL);
    lv_obj_add_event_cb(s, slider_released, LV_EVENT_RELEASED, NULL);

    r = setting_row(b, LV_SYMBOL_AUDIO, "Read words");
    lv_obj_t *sw = lv_switch_create(r);
    if (cfg()->speak) lv_obj_add_state(sw, LV_STATE_CHECKED);
    lv_obj_add_event_cb(sw, speak_changed, LV_EVENT_VALUE_CHANGED, NULL);

    // multiple choice: try to remember before seeing the options
    r = setting_row(b, LV_SYMBOL_EYE_CLOSE, "Recall first");
    sw = lv_switch_create(r);
    if (cfg()->recall_first) lv_obj_add_state(sw, LV_STATE_CHECKED);
    lv_obj_add_event_cb(sw, recall_changed, LV_EVENT_VALUE_CHANGED, NULL);

    char info[400];
    const esp_app_desc_t *app = esp_app_get_description();
    deck_t *d = deck_get();
    snprintf(info, sizeof(info), "Wi-Fi: %s %s\nServer: %s\nPop-up every %d min, %s–%s\nFirmware %s · %d answers to send",
             cfg()->ssid, net_connected() ? net_ip() : "(offline)", cfg()->server,
             d ? d->reminder_every_min : 0, d ? d->active_from : "--", d ? d->active_to : "--",
             app->version, reviews_count());
    char newer[16] = "";
    if (d && ota_is_newer(d->fw_version)) strlcpy(newer, d->fw_version, sizeof(newer));
    deck_unref(d);
    if (newer[0]) {
        char t[48];
        snprintf(t, sizeof(t), LV_SYMBOL_DOWNLOAD "  Update to %s", newer);
        lv_obj_t *u = ui_button(b, t, COL_OK, font_bold);
        lv_obj_set_width(u, lv_pct(100));
        lv_obj_add_event_cb(u, update_clicked, LV_EVENT_CLICKED, NULL);
    }
    ui_label(b, info, font_small, COL_MUTED, -1);
    ui_label(b, "Topics, pop-ups and card mode are set on the website (Devices → Settings).", font_small,
             COL_MUTED, -1);

    lv_obj_t *row = lv_obj_create(b);
    lv_obj_set_size(row, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(row, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(row, 0, 0);
    lv_obj_set_style_pad_all(row, 0, 0);
    lv_obj_set_style_pad_column(row, 8, 0);
    lv_obj_set_flex_flow(row, LV_FLEX_FLOW_ROW);
    lv_obj_t *w = ui_button(row, LV_SYMBOL_WIFI " Change Wi-Fi", COL_INFO, font_small);
    lv_obj_set_flex_grow(w, 1);
    lv_obj_add_event_cb(w, wifi_reset_clicked, LV_EVENT_CLICKED, NULL);
    lv_obj_t *u = ui_button(row, LV_SYMBOL_CLOSE " Unlink", COL_BAD, font_small);
    lv_obj_set_flex_grow(u, 1);
    lv_obj_add_event_cb(u, unpair_clicked, LV_EVENT_CLICKED, NULL);
    load(scr);
}

// ---------------------------------------------------------------- sleep / wake
// After DIM_AFTER_S without a touch the backlight dims; after SLEEP_AFTER_S the
// backlight goes off. LVGL keeps running, so a touch (on a transparent layer
// over the screen), the BOOT key or a study reminder wakes it. The touch
// that wakes the screen is swallowed by a transparent layer, so it does not
// press a button by accident.

static void drop_layer_cb(lv_timer_t *t)
{
    if (s_sleep_layer) {
        lv_obj_delete(s_sleep_layer);
        s_sleep_layer = NULL;
    }
}

static volatile bool s_wake_req;

static void sleep_layer_pressed(lv_event_t *e)
{
    drop_layer_cb(NULL); // the waking touch ends here
    s_wake_req = true;   // the main task turns the screen on (ui_wait)
}

static void wake(bool by_touch)
{
    if (!s_sleeping && !s_dimmed) return;
    if (s_sleeping) board_display_sleep(false);
    LOCK();
    s_sleeping = s_dimmed = false;
    lv_display_trigger_activity(NULL);
    if (s_sleep_layer) {
        if (by_touch) {
            lv_timer_t *t = lv_timer_create(drop_layer_cb, 800, NULL); // in case the finger is already up
            lv_timer_set_repeat_count(t, 1);
        } else {
            drop_layer_cb(NULL);
        }
    }
    board_backlight(cfg()->brightness);
    UNLOCK();
}

void ui_wake(void) { wake(false); }

bool ui_sleeping(void) { return s_sleeping; }

int ui_idle_seconds(void)
{
    int s = 0;
    LOCK();
    s = lv_display_get_inactive_time(NULL) / 1000;
    UNLOCK();
    return s;
}

void ui_reminder(void)
{
    wake(false);
    audio_fx(AUDIO_FX_CHIME);
    deck_t *d = deck_get();
    int n = d && d->popup_cards > 0 ? d->popup_cards : 10;
    bool cards = false;
    for (int i = 0; d && i < d->n_cards; i++) cards |= !d->cards[i].done;
    deck_unref(d);
    if (cards) study_start(n, true);
    else quiz_start(5); // nothing due and no new words left: 5 words already learned
}

void ui_wait(int ms)
{
    // asleep: check the touch panel and the BOOT key every 100 ms
    for (int t = 0; t < ms; t += 100) {
        if (s_sleeping && (s_wake_req || board_boot_pressed())) {
            s_wake_req = false;
            wake(true);
            return;
        }
        vTaskDelay(pdMS_TO_TICKS(ms - t < 100 ? ms - t : 100));
    }
}

void ui_tick(void)
{
    if (s_sleeping) return; // nothing to draw; ui_wait() wakes on touch / BOOT
    bool go_sleep = false;
    LOCK();
    uint32_t idle = lv_display_get_inactive_time(NULL) / 1000;
    if (idle >= SLEEP_AFTER_S) {
        go_sleep = true;
        s_sleeping = true;
        s_wake_req = false;
        if (!s_home) build_home(); // also ends a round left unanswered, so syncs and pop-ups go on
        s_sleep_layer = lv_obj_create(lv_layer_top());
        lv_obj_remove_style_all(s_sleep_layer);
        lv_obj_set_size(s_sleep_layer, lv_pct(100), lv_pct(100));
        lv_obj_add_flag(s_sleep_layer, LV_OBJ_FLAG_CLICKABLE);
        lv_obj_add_event_cb(s_sleep_layer, sleep_layer_pressed, LV_EVENT_PRESSED, NULL);
        lv_refr_now(NULL);
    } else if (idle >= DIM_AFTER_S && !s_dimmed) {
        s_dimmed = true;
        board_backlight(cfg()->brightness > 15 ? 15 : cfg()->brightness);
    } else if (idle < DIM_AFTER_S && s_dimmed) {
        s_dimmed = false;
        board_backlight(cfg()->brightness);
    }
    if (s_status && lv_obj_is_valid(s_status)) {
        char buf[96];
        status_line(buf, sizeof(buf));
        lv_label_set_text(s_status, buf);
    }
    UNLOCK();
    if (go_sleep) board_display_sleep(true);
}

void ui_init(lv_display_t *disp)
{
    LOCK();
    ui_theme_init();
    UNLOCK();
}
