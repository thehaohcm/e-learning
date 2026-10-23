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
#include "net.h"
#include "store.h"
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
    lv_obj_t *t = lv_label_create(bar);
    lv_label_set_text(t, title ? title : "");
    lv_obj_set_style_text_color(t, lv_color_white(), 0);
    lv_obj_set_style_text_font(t, font_bold, 0);
    lv_obj_align(t, LV_ALIGN_LEFT_MID, back_home ? 26 : 0, 0);

    s_clock = lv_label_create(bar);
    lv_obj_set_style_text_color(s_clock, lv_color_hex(0xC7D2FE), 0);
    lv_obj_set_style_text_font(s_clock, font_small, 0);
    lv_obj_align(s_clock, LV_ALIGN_CENTER, 0, 0);
    s_status = lv_label_create(bar);
    lv_obj_set_style_text_color(s_status, lv_color_hex(0xC7D2FE), 0);
    lv_obj_set_style_text_font(s_status, font_small, 0);
    lv_obj_align(s_status, LV_ALIGN_RIGHT_MID, 0, 0);
    char buf[64];
    clock_text(buf, sizeof(buf));
    lv_label_set_text(s_clock, buf);
    header_text(buf, sizeof(buf));
    lv_label_set_text(s_status, buf);
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
    lv_obj_set_style_pad_all(box, 6, 0);
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

static void sync_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    app_request(APP_REQ_SYNC);
}

static void settings_clicked(lv_event_t *e);

void study_clicked_cb(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    study_start(0, false);
}

static void build_home(void)
{
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
    if (due + fresh > 0) {
        snprintf(lbl, sizeof(lbl), LV_SYMBOL_PLAY "  Study %d card%s", due + fresh, due + fresh > 1 ? "s" : "");
        lv_obj_t *sb = ui_button(b, lbl, COL_PRIMARY, font_title);
        lv_obj_set_width(sb, lv_pct(100));
        lv_obj_set_height(sb, 52);
        lv_obj_add_event_cb(sb, study_clicked_cb, LV_EVENT_CLICKED, NULL);
    } else {
        lv_obj_t *m = ui_label(b, d ? "All done for now! New cards come back when they are due."
                                    : "No cards yet: press Sync.", font_body, COL_OK, -1);
        lv_obj_set_style_text_align(m, LV_TEXT_ALIGN_CENTER, 0);
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

static void slider_released(lv_event_t *e)
{
    cfg_save();
    audio_fx(AUDIO_FX_TAP);
}

static void confirm_cb(lv_event_t *e)
{
    app_req_t r = (app_req_t)(intptr_t)lv_event_get_user_data(e);
    lv_obj_t *mbox = lv_obj_get_parent(lv_obj_get_parent(lv_event_get_target(e)));
    lv_msgbox_close(mbox);
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

    char info[400];
    const esp_app_desc_t *app = esp_app_get_description();
    deck_t *d = deck_get();
    snprintf(info, sizeof(info), "Wi-Fi: %s %s\nServer: %s\nPop-up every %d min, %s–%s\nFirmware %s · %d answers to send",
             cfg()->ssid, net_connected() ? net_ip() : "(offline)", cfg()->server,
             d ? d->reminder_every_min : 0, d ? d->active_from : "--", d ? d->active_to : "--",
             app->version, reviews_count());
    deck_unref(d);
    ui_label(b, info, font_small, COL_MUTED, -1);
    ui_label(b, "The pop-up interval, active hours and card mode are set on the website (Settings).", font_small,
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
// screen sleeps (backlight off, panel in sleep mode, LVGL paused, CPU 80 MHz).
// A touch / the BOOT key (polled) or a study reminder wakes it. The touch
// that wakes the screen is swallowed by a transparent layer, so it does not
// press a button by accident.

static void drop_layer_cb(lv_timer_t *t)
{
    if (s_sleep_layer) {
        lv_obj_delete(s_sleep_layer);
        s_sleep_layer = NULL;
    }
}

static void sleep_layer_pressed(lv_event_t *e)
{
    drop_layer_cb(NULL); // the waking touch ends here
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
    study_start(1, true);
}

void ui_wait(int ms)
{
    // asleep: check the touch panel and the BOOT key every 100 ms
    for (int t = 0; t < ms; t += 100) {
        if (s_sleeping && (board_touch_pressed() || board_boot_pressed())) {
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
        if (!s_home && !study_active()) build_home();
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
    if (s_clock && lv_obj_is_valid(s_clock) && s_status && lv_obj_is_valid(s_status)) {
        char buf[64];
        clock_text(buf, sizeof(buf));
        lv_label_set_text(s_clock, buf);
        header_text(buf, sizeof(buf));
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
