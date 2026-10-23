// Fonts (DejaVu Sans with Vietnamese + IPA, rendered by LVGL Tiny TTF) and colours.
#pragma once

#include "lvgl.h"

extern const lv_font_t *font_small;  // 14
extern const lv_font_t *font_body;   // 16
extern const lv_font_t *font_bold;   // 16 bold
extern const lv_font_t *font_title;  // 20 bold
extern const lv_font_t *font_word;   // 30 bold (the word being studied)
extern const lv_font_t *font_huge;   // 46 bold (pairing code)

#define COL_BG       lv_color_hex(0xF1F5F9)
#define COL_CARD     lv_color_hex(0xFFFFFF)
#define COL_TEXT     lv_color_hex(0x1E293B)
#define COL_MUTED    lv_color_hex(0x64748B)
#define COL_PRIMARY  lv_color_hex(0x4F46E5)
#define COL_HEADER   lv_color_hex(0x1E1B4B)
#define COL_OK       lv_color_hex(0x16A34A)
#define COL_BAD      lv_color_hex(0xDC2626)
#define COL_WARN     lv_color_hex(0xEA580C)
#define COL_INFO     lv_color_hex(0x2563EB)
#define COL_SOFT     lv_color_hex(0xE0E7FF)

void ui_theme_init(void);

/** A button with a text label (font_body unless font given). */
lv_obj_t *ui_button(lv_obj_t *parent, const char *text, lv_color_t bg, const lv_font_t *font);
/** A wrapping label. width < 0 = 100% of the parent. */
lv_obj_t *ui_label(lv_obj_t *parent, const char *text, const lv_font_t *font, lv_color_t color, int width);
