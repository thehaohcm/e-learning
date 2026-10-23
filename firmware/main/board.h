// ES3C28P board: ESP32-S3 + 2.8" ILI9341V (SPI) + FT6336G touch + ES8311 audio.
// Pins from the LCDwiki "2.8inch ESP32-S3 Display" page.
#pragma once

#include <stdbool.h>
#include "esp_err.h"
#include "driver/i2c_master.h"
#include "lvgl.h"

// LCD (SPI)
#define PIN_LCD_SCLK 12
#define PIN_LCD_MOSI 11
#define PIN_LCD_MISO 13
#define PIN_LCD_CS   10
#define PIN_LCD_DC   46
#define PIN_LCD_BL   45   // high = backlight on (PWM)
#define LCD_H_RES    240  // native portrait
#define LCD_V_RES    320

// I2C: touch (FT6336G, 0x38) and audio codec (ES8311) share the bus
#define PIN_I2C_SDA   16
#define PIN_I2C_SCL   15
#define PIN_TOUCH_RST 18
#define PIN_TOUCH_INT 17

// I2S to the ES8311 codec
#define PIN_I2S_MCLK 4
#define PIN_I2S_BCLK 5
#define PIN_I2S_WS   7
#define PIN_I2S_DOUT 8   // ESP32 -> codec (speaker)
#define PIN_I2S_DIN  6   // codec -> ESP32 (microphone)
#define PIN_AMP_EN   1   // speaker amplifier (FM8002E)

#define PIN_BOOT_BTN 0
#define PIN_BAT_ADC  9   // battery voltage / 2

esp_err_t board_init(void);
i2c_master_bus_handle_t board_i2c(void);

/** LCD + touch + LVGL. Returns the LVGL display (landscape when CONFIG_ELEARN_LANDSCAPE). */
lv_display_t *board_display_init(void);

/** Backlight 0..100 %. */
void board_backlight(int percent);

/** Battery voltage in mV (0 if unknown) and an estimated charge 0..100 (-1 = no battery / USB only). */
int board_battery_mv(void);
int board_battery_percent(void);

/** BOOT key (IO0), active low. */
bool board_boot_pressed(void);

/** Screen sleep: backlight + panel off (ILI9341 sleep mode), LVGL paused, CPU at 80 MHz. */
void board_display_sleep(bool sleep);
/** A finger on the screen (read directly, for waking up while LVGL is paused). */
bool board_touch_pressed(void);
