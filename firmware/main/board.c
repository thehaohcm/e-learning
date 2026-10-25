#include "board.h"

#include "driver/gpio.h"
#include "driver/ledc.h"
#include "driver/spi_master.h"
#include "esp_adc/adc_cali.h"
#include "esp_adc/adc_cali_scheme.h"
#include "esp_adc/adc_oneshot.h"
#include "esp_check.h"
#include "esp_lcd_ili9341.h"
#include "esp_lcd_panel_io.h"
#include "esp_lcd_panel_ops.h"
#include "esp_lcd_touch_ft5x06.h"
#include "esp_log.h"
#include "esp_lvgl_port.h"
#include "esp_pm.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

static const char *TAG = "board";

static i2c_master_bus_handle_t s_i2c;
static esp_lcd_panel_handle_t s_panel;
static esp_lcd_touch_handle_t s_touch;
static volatile bool s_asleep;
static adc_oneshot_unit_handle_t s_adc;
static adc_cali_handle_t s_adc_cali;

#define LCD_SPI_HOST SPI2_HOST
#define LCD_PIXEL_CLOCK_HZ (40 * 1000 * 1000)
#define BL_LEDC_TIMER LEDC_TIMER_0
#define BL_LEDC_CH LEDC_CHANNEL_0

esp_err_t board_init(void)
{
    // shared I2C bus (touch + codec), new I2C master driver
    i2c_master_bus_config_t bus = {
        .i2c_port = I2C_NUM_0,
        .sda_io_num = PIN_I2C_SDA,
        .scl_io_num = PIN_I2C_SCL,
        .clk_source = I2C_CLK_SRC_DEFAULT,
        .glitch_ignore_cnt = 7,
        .flags.enable_internal_pullup = true,
    };
    ESP_RETURN_ON_ERROR(i2c_new_master_bus(&bus, &s_i2c), TAG, "i2c bus");

    // backlight PWM (off until the first screen is drawn)
    ledc_timer_config_t t = {
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .duty_resolution = LEDC_TIMER_10_BIT,
        .timer_num = BL_LEDC_TIMER,
        .freq_hz = 5000,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    ESP_RETURN_ON_ERROR(ledc_timer_config(&t), TAG, "ledc timer");
    ledc_channel_config_t ch = {
        .gpio_num = PIN_LCD_BL,
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .channel = BL_LEDC_CH,
        .timer_sel = BL_LEDC_TIMER,
        .duty = 0,
    };
    ESP_RETURN_ON_ERROR(ledc_channel_config(&ch), TAG, "ledc channel");

    // BOOT key
    gpio_config_t btn = {
        .pin_bit_mask = 1ULL << PIN_BOOT_BTN,
        .mode = GPIO_MODE_INPUT,
        .pull_up_en = GPIO_PULLUP_ENABLE,
    };
    gpio_config(&btn);

    // battery ADC (IO9 = ADC1 channel 8, voltage divided by 2)
    adc_oneshot_unit_init_cfg_t ucfg = {.unit_id = ADC_UNIT_1};
    if (adc_oneshot_new_unit(&ucfg, &s_adc) == ESP_OK) {
        adc_oneshot_chan_cfg_t ccfg = {.atten = ADC_ATTEN_DB_12, .bitwidth = ADC_BITWIDTH_DEFAULT};
        adc_oneshot_config_channel(s_adc, ADC_CHANNEL_8, &ccfg);
        adc_cali_curve_fitting_config_t cal = {.unit_id = ADC_UNIT_1, .chan = ADC_CHANNEL_8, .atten = ADC_ATTEN_DB_12,
                                               .bitwidth = ADC_BITWIDTH_DEFAULT};
        if (adc_cali_create_scheme_curve_fitting(&cal, &s_adc_cali) != ESP_OK) {
            s_adc_cali = NULL;
        }
    }
    return ESP_OK;
}

i2c_master_bus_handle_t board_i2c(void) { return s_i2c; }

void board_backlight(int percent)
{
    if (percent < 0) percent = 0;
    if (percent > 100) percent = 100;
    uint32_t duty = (1023 * percent) / 100;
    ledc_set_duty(LEDC_LOW_SPEED_MODE, BL_LEDC_CH, duty);
    ledc_update_duty(LEDC_LOW_SPEED_MODE, BL_LEDC_CH);
}

lv_display_t *board_display_init(void)
{
    // ---- LCD panel
    spi_bus_config_t buscfg = {
        .sclk_io_num = PIN_LCD_SCLK,
        .mosi_io_num = PIN_LCD_MOSI,
        .miso_io_num = PIN_LCD_MISO,
        .quadwp_io_num = -1,
        .quadhd_io_num = -1,
        .max_transfer_sz = LCD_H_RES * 40 * sizeof(uint16_t),
    };
    ESP_ERROR_CHECK(spi_bus_initialize(LCD_SPI_HOST, &buscfg, SPI_DMA_CH_AUTO));

    esp_lcd_panel_io_handle_t io = NULL;
    esp_lcd_panel_io_spi_config_t io_cfg = {
        .dc_gpio_num = PIN_LCD_DC,
        .cs_gpio_num = PIN_LCD_CS,
        .pclk_hz = LCD_PIXEL_CLOCK_HZ,
        .lcd_cmd_bits = 8,
        .lcd_param_bits = 8,
        .spi_mode = 0,
        .trans_queue_depth = 10,
    };
    ESP_ERROR_CHECK(esp_lcd_new_panel_io_spi((esp_lcd_spi_bus_handle_t)LCD_SPI_HOST, &io_cfg, &io));

    esp_lcd_panel_handle_t panel = NULL;
    esp_lcd_panel_dev_config_t panel_cfg = {
        .reset_gpio_num = -1, // shares the ESP32-S3 reset (EN)
#if CONFIG_ELEARN_LCD_BGR
        .rgb_ele_order = LCD_RGB_ELEMENT_ORDER_BGR,
#else
        .rgb_ele_order = LCD_RGB_ELEMENT_ORDER_RGB,
#endif
        .bits_per_pixel = 16,
    };
    ESP_ERROR_CHECK(esp_lcd_new_panel_ili9341(io, &panel_cfg, &panel));
    esp_lcd_panel_reset(panel);
    esp_lcd_panel_init(panel);
#if CONFIG_ELEARN_LCD_INVERT
    esp_lcd_panel_invert_color(panel, true);
#endif
    esp_lcd_panel_disp_on_off(panel, true);
    s_panel = panel;

    // ---- touch (FT6336G is FT5x06 compatible)
    esp_lcd_panel_io_handle_t tp_io = NULL;
    esp_lcd_panel_io_i2c_config_t tp_io_cfg = ESP_LCD_TOUCH_IO_I2C_FT5x06_CONFIG();
    tp_io_cfg.scl_speed_hz = 400000;
    ESP_ERROR_CHECK(esp_lcd_new_panel_io_i2c(s_i2c, &tp_io_cfg, &tp_io));
    esp_lcd_touch_config_t tp_cfg = {
        .x_max = LCD_H_RES,
        .y_max = LCD_V_RES,
        .rst_gpio_num = PIN_TOUCH_RST,
        .int_gpio_num = PIN_TOUCH_INT,
        .levels = {.reset = 0, .interrupt = 0},
        .flags = {
            .swap_xy = 0,
#if CONFIG_ELEARN_TOUCH_MIRROR_X
            .mirror_x = 1,
#else
            .mirror_x = 0,
#endif
#if CONFIG_ELEARN_TOUCH_MIRROR_Y
            .mirror_y = 1,
#else
            .mirror_y = 0,
#endif
        },
    };
    esp_lcd_touch_handle_t touch = NULL;
    if (esp_lcd_touch_new_i2c_ft5x06(tp_io, &tp_cfg, &touch) != ESP_OK) {
        ESP_LOGE(TAG, "touch controller not found (FT6336G at 0x38)");
        touch = NULL;
    }

    // ---- LVGL
    lvgl_port_cfg_t port_cfg = ESP_LVGL_PORT_INIT_CONFIG();
    port_cfg.task_priority = 4;
    port_cfg.task_stack = 8192;
    port_cfg.task_affinity = 1; // LVGL on core 1, Wi-Fi on core 0
    ESP_ERROR_CHECK(lvgl_port_init(&port_cfg));

    lvgl_port_display_cfg_t disp_cfg = {
        .io_handle = io,
        .panel_handle = panel,
        .buffer_size = LCD_H_RES * 40,
        .double_buffer = true,
        .hres = LCD_H_RES,
        .vres = LCD_V_RES,
        .monochrome = false,
        .color_format = LV_COLOR_FORMAT_RGB565,
        .rotation = {
            .swap_xy = false,
#if CONFIG_ELEARN_LCD_MIRROR_X
            .mirror_x = true,
#else
            .mirror_x = false,
#endif
            .mirror_y = false,
        },
        .flags = {
            .buff_dma = true,
            .swap_bytes = true,
        },
    };
    lv_display_t *disp = lvgl_port_add_disp(&disp_cfg);
    s_touch = touch;
    if (touch) {
        lvgl_port_touch_cfg_t tcfg = {.disp = disp, .handle = touch};
        lvgl_port_add_touch(&tcfg);
    }
#if CONFIG_ELEARN_LANDSCAPE
    if (lvgl_port_lock(0)) {
#if CONFIG_ELEARN_LANDSCAPE_FLIP
        lv_display_set_rotation(disp, LV_DISPLAY_ROTATION_270);
#else
        lv_display_set_rotation(disp, LV_DISPLAY_ROTATION_90);
#endif
        lvgl_port_unlock();
    }
#endif
    return disp;
}

int board_battery_mv(void)
{
    if (!s_adc) return 0;
    int raw = 0, mv = 0, sum = 0;
    for (int i = 0; i < 8; i++) {
        if (adc_oneshot_read(s_adc, ADC_CHANNEL_8, &raw) != ESP_OK) return 0;
        if (s_adc_cali && adc_cali_raw_to_voltage(s_adc_cali, raw, &mv) == ESP_OK) {
            sum += mv;
        } else {
            sum += raw * 3100 / 4095;
        }
    }
    return (sum / 8) * 2; // divider 1:2
}

int board_battery_percent(void)
{
    int mv = board_battery_mv();
    if (mv < 2500 || mv > 4500) return -1; // no battery (USB only) or not measurable
    // simple Li-ion curve: 3.3 V = 0 %, 4.15 V = 100 %
    int p = (mv - 3300) * 100 / (4150 - 3300);
    if (p < 0) p = 0;
    if (p > 100) p = 100;
    return p;
}

bool board_boot_pressed(void) { return gpio_get_level(PIN_BOOT_BTN) == 0; }

// ---------------------------------------------------------------- screen sleep

/** While the screen sleeps LVGL does not read the touch panel: poll it here. */
bool board_touch_pressed(void)
{
    if (!s_touch) return false;
    uint16_t x, y, strength;
    uint8_t n = 0;
    esp_lcd_touch_read_data(s_touch);
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wdeprecated-declarations" // available in every esp_lcd_touch 1.x
    bool down = esp_lcd_touch_get_coordinates(s_touch, &x, &y, &strength, &n, 1) && n > 0;
#pragma GCC diagnostic pop
    return down;
}

void board_display_sleep(bool sleep)
{
    // Only the backlight goes off (it uses most of the power). LVGL keeps
    // running and reading the touch panel, so a touch always wakes the screen.
    if (sleep == s_asleep) return;
    s_asleep = sleep;
    if (sleep) board_backlight(0);
}
