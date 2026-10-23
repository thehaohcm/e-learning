#include "ui_theme.h"

#include <stddef.h>

#include "esp_log.h"

extern const uint8_t ttf_regular_start[] asm("_binary_ui_regular_ttf_start");
extern const uint8_t ttf_regular_end[] asm("_binary_ui_regular_ttf_end");
extern const uint8_t ttf_bold_start[] asm("_binary_ui_bold_ttf_start");
extern const uint8_t ttf_bold_end[] asm("_binary_ui_bold_ttf_end");

const lv_font_t *font_small;
const lv_font_t *font_body;
const lv_font_t *font_bold;
const lv_font_t *font_title;
const lv_font_t *font_word;
const lv_font_t *font_huge;

static const lv_font_t *make(const uint8_t *start, const uint8_t *end, int size)
{
    lv_font_t *f = lv_tiny_ttf_create_data(start, end - start, size);
    if (!f) {
        ESP_LOGE("theme", "font %d px failed, using the built-in one", size);
        return size >= 20 ? &lv_font_montserrat_20 : &lv_font_montserrat_14;
    }
    // LV_SYMBOL_* icons are not in DejaVu: take them from the built-in font
    f->fallback = size >= 20 ? &lv_font_montserrat_20 : &lv_font_montserrat_14;
    return f;
}

void ui_theme_init(void)
{
    font_small = make(ttf_regular_start, ttf_regular_end, 14);
    font_body = make(ttf_regular_start, ttf_regular_end, 16);
    font_bold = make(ttf_bold_start, ttf_bold_end, 16);
    font_title = make(ttf_bold_start, ttf_bold_end, 20);
    font_word = make(ttf_bold_start, ttf_bold_end, 30);
    font_huge = make(ttf_bold_start, ttf_bold_end, 46);

    lv_display_t *d = lv_display_get_default();
    lv_theme_t *th = lv_theme_default_init(d, COL_PRIMARY, COL_WARN, false, font_body);
    lv_display_set_theme(d, th);
}

lv_obj_t *ui_button(lv_obj_t *parent, const char *text, lv_color_t bg, const lv_font_t *font)
{
    lv_obj_t *b = lv_button_create(parent);
    lv_obj_set_style_bg_color(b, bg, 0);
    lv_obj_set_style_radius(b, 10, 0);
    lv_obj_set_style_shadow_width(b, 0, 0);
    lv_obj_set_style_pad_ver(b, 8, 0);
    lv_obj_set_style_pad_hor(b, 10, 0);
    lv_obj_t *l = lv_label_create(b);
    lv_label_set_text(l, text);
    lv_obj_set_style_text_font(l, font ? font : font_body, 0);
    lv_obj_set_style_text_color(l, lv_color_white(), 0);
    lv_obj_center(l);
    return b;
}

lv_obj_t *ui_label(lv_obj_t *parent, const char *text, const lv_font_t *font, lv_color_t color, int width)
{
    lv_obj_t *l = lv_label_create(parent);
    lv_label_set_text(l, text ? text : "");
    lv_obj_set_style_text_font(l, font ? font : font_body, 0);
    lv_obj_set_style_text_color(l, color, 0);
    lv_label_set_long_mode(l, LV_LABEL_LONG_WRAP);
    lv_obj_set_width(l, width < 0 ? lv_pct(100) : width);
    return l;
}
