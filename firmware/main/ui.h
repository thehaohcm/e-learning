// Screens. Every function locks LVGL itself (call from any task except the LVGL task).
#pragma once

#include <stdbool.h>
#include "lvgl.h"

void ui_init(lv_display_t *disp);

void ui_show_boot(const char *msg);
/** Grey veil + spinner over the screen while waiting for the server; ui_busy_end removes it. */
void ui_busy(const char *msg);
void ui_busy_end(void);
/** The same, for code that already holds the LVGL lock (buttons, timers). */
void ui_busy_locked(const char *msg);
void ui_busy_end_locked(void);
void ui_busy_text_locked(const char *msg);
void ui_show_setup(const char *ap_ssid);
void ui_show_pair(const char *code);
void ui_show_offline(const char *msg);
void ui_show_home(void);

/** Called every second by the app: clock/header, screen dimming and sleep. */
void ui_tick(void);
/** Waits up to ms; returns early (and wakes the screen) on a touch / BOOT key while asleep. */
void ui_wait(int ms);
bool ui_sleeping(void);
/** Last touch, in seconds ago. */
int ui_idle_seconds(void);
/** Study reminder: wake the screen with a chime and show one card. */
void ui_reminder(void);
void ui_wake(void);
/** The home screen is visible (safe to replace the deck). */
bool ui_on_home(void);

// ---- used by ui_study.c (LVGL lock already held)
lv_obj_t *ui_screen_with_header(const char *title, bool back_home);
void ui_go_home_locked(void);
