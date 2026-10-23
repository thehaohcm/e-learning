#include "audio.h"

#include <math.h>
#include <stdlib.h>
#include <string.h>

#include "board.h"
#include "driver/i2s_std.h"
#include "esp_check.h"
#include "esp_codec_dev.h"
#include "esp_codec_dev_defaults.h"
#include "esp_crt_bundle.h"
#include "esp_heap_caps.h"
#include "esp_http_client.h"
#include "esp_log.h"
#include "freertos/FreeRTOS.h"
#include "freertos/queue.h"
#include "freertos/task.h"

static const char *TAG = "audio";

#define AUDIO_RATE 16000

static i2s_chan_handle_t s_tx, s_rx;
static esp_codec_dev_handle_t s_spk;
static QueueHandle_t s_q;
static int s_volume = 70;
static int s_open_rate;
static volatile bool s_stop;

typedef struct {
    int kind; // 0 = effect, 1 = url
    audio_fx_t fx;
    char *url;
} audio_msg_t;

static esp_err_t i2s_init(void)
{
    i2s_chan_config_t chan_cfg = I2S_CHANNEL_DEFAULT_CONFIG(I2S_NUM_0, I2S_ROLE_MASTER);
    chan_cfg.auto_clear = true;
    ESP_RETURN_ON_ERROR(i2s_new_channel(&chan_cfg, &s_tx, &s_rx), TAG, "i2s channel");
    i2s_std_config_t std = {
        .clk_cfg = I2S_STD_CLK_DEFAULT_CONFIG(AUDIO_RATE),
        .slot_cfg = I2S_STD_PHILIPS_SLOT_DEFAULT_CONFIG(I2S_DATA_BIT_WIDTH_16BIT, I2S_SLOT_MODE_MONO),
        .gpio_cfg = {
            .mclk = PIN_I2S_MCLK,
            .bclk = PIN_I2S_BCLK,
            .ws = PIN_I2S_WS,
            .dout = PIN_I2S_DOUT,
            .din = PIN_I2S_DIN,
        },
    };
    std.clk_cfg.mclk_multiple = I2S_MCLK_MULTIPLE_256;
    ESP_RETURN_ON_ERROR(i2s_channel_init_std_mode(s_tx, &std), TAG, "i2s tx");
    ESP_RETURN_ON_ERROR(i2s_channel_init_std_mode(s_rx, &std), TAG, "i2s rx");
    i2s_channel_enable(s_tx);
    i2s_channel_enable(s_rx);
    return ESP_OK;
}

static esp_err_t codec_init(void)
{
    audio_codec_i2s_cfg_t i2s_cfg = {.port = I2S_NUM_0, .rx_handle = s_rx, .tx_handle = s_tx};
    const audio_codec_data_if_t *data_if = audio_codec_new_i2s_data(&i2s_cfg);
    audio_codec_i2c_cfg_t i2c_cfg = {.port = I2C_NUM_0, .addr = ES8311_CODEC_DEFAULT_ADDR, .bus_handle = board_i2c()};
    const audio_codec_ctrl_if_t *ctrl_if = audio_codec_new_i2c_ctrl(&i2c_cfg);
    const audio_codec_gpio_if_t *gpio_if = audio_codec_new_gpio();
    if (!data_if || !ctrl_if || !gpio_if) return ESP_FAIL;

    es8311_codec_cfg_t es = {
        .ctrl_if = ctrl_if,
        .gpio_if = gpio_if,
        .codec_mode = ESP_CODEC_DEV_WORK_MODE_BOTH,
        .pa_pin = PIN_AMP_EN,
#if CONFIG_ELEARN_AMP_ACTIVE_LOW
        .pa_reverted = true,
#else
        .pa_reverted = false,
#endif
        .master_mode = false,
        .use_mclk = true,
        .hw_gain = {.pa_voltage = 5.0, .codec_dac_voltage = 3.3},
    };
    const audio_codec_if_t *codec_if = es8311_codec_new(&es);
    if (!codec_if) return ESP_FAIL;
    esp_codec_dev_cfg_t dev_cfg = {
        .dev_type = ESP_CODEC_DEV_TYPE_OUT,
        .codec_if = codec_if,
        .data_if = data_if,
    };
    s_spk = esp_codec_dev_new(&dev_cfg);
    return s_spk ? ESP_OK : ESP_FAIL;
}

static void open_rate(int rate, int channels)
{
    if (s_open_rate == rate * 10 + channels) return;
    if (s_open_rate) esp_codec_dev_close(s_spk);
    esp_codec_dev_sample_info_t fs = {.sample_rate = rate, .channel = channels, .bits_per_sample = 16};
    if (esp_codec_dev_open(s_spk, &fs) == ESP_CODEC_DEV_OK) {
        esp_codec_dev_set_out_vol(s_spk, s_volume);
        s_open_rate = rate * 10 + channels;
    }
}

// ---------------------------------------------------------------- effects

/** Writes a tone with a short fade in/out (no clicks). */
static void tone(int freq, int ms, float amp)
{
    const int n = AUDIO_RATE * ms / 1000;
    const int chunk = 256;
    int16_t buf[256];
    const int fade = AUDIO_RATE / 200; // 5 ms
    for (int i = 0; i < n && !s_stop; i += chunk) {
        int m = (n - i < chunk) ? n - i : chunk;
        for (int k = 0; k < m; k++) {
            int idx = i + k;
            float env = 1.0f;
            if (idx < fade) env = (float)idx / fade;
            else if (n - idx < fade) env = (float)(n - idx) / fade;
            buf[k] = (int16_t)(sinf(2.0f * (float)M_PI * freq * idx / AUDIO_RATE) * 12000.0f * amp * env);
        }
        esp_codec_dev_write(s_spk, buf, m * sizeof(int16_t));
    }
}

static void silence(int ms)
{
    int16_t z[160] = {0};
    for (int i = 0; i < ms / 10 && !s_stop; i++) esp_codec_dev_write(s_spk, z, sizeof(z));
}

static void play_fx(audio_fx_t fx)
{
    open_rate(AUDIO_RATE, 1);
    switch (fx) {
    case AUDIO_FX_TAP:
        tone(1800, 18, 0.35f);
        break;
    case AUDIO_FX_CORRECT:
        tone(880, 90, 0.6f);
        tone(1320, 140, 0.6f);
        break;
    case AUDIO_FX_WRONG:
        tone(220, 260, 0.7f);
        break;
    case AUDIO_FX_CHIME:
        tone(988, 120, 0.6f);
        silence(30);
        tone(1319, 120, 0.6f);
        silence(30);
        tone(1568, 220, 0.6f);
        break;
    }
    silence(40);
}

// ---------------------------------------------------------------- WAV over HTTP

static bool read_full(esp_http_client_handle_t c, uint8_t *buf, int len)
{
    int got = 0;
    while (got < len) {
        int r = esp_http_client_read(c, (char *)buf + got, len - got);
        if (r <= 0) return false;
        got += r;
    }
    return true;
}

static void play_wav_url(const char *url)
{
    esp_http_client_config_t cfg = {
        .url = url,
        .timeout_ms = 10000,
        .crt_bundle_attach = esp_crt_bundle_attach,
        .buffer_size = 2048,
    };
    esp_http_client_handle_t c = esp_http_client_init(&cfg);
    if (!c) return;
    if (esp_http_client_open(c, 0) != ESP_OK || esp_http_client_fetch_headers(c) < 0 ||
        esp_http_client_get_status_code(c) != 200) {
        ESP_LOGW(TAG, "cannot open %s", url);
        goto done;
    }
    // RIFF header, then chunks until "data"
    uint8_t h[12];
    if (!read_full(c, h, 12) || memcmp(h, "RIFF", 4) || memcmp(h + 8, "WAVE", 4)) {
        ESP_LOGW(TAG, "not a WAV file");
        goto done;
    }
    int rate = AUDIO_RATE, channels = 1, bits = 16;
    for (;;) {
        uint8_t ch[8];
        if (!read_full(c, ch, 8)) goto done;
        uint32_t size = ch[4] | ch[5] << 8 | ch[6] << 16 | (uint32_t)ch[7] << 24;
        if (!memcmp(ch, "fmt ", 4)) {
            uint8_t f[40];
            uint32_t n = size < sizeof(f) ? size : sizeof(f);
            if (!read_full(c, f, n)) goto done;
            for (uint32_t skip = size - n; skip > 0; skip--) {
                uint8_t b;
                if (!read_full(c, &b, 1)) goto done;
            }
            channels = f[2] | f[3] << 8;
            rate = f[4] | f[5] << 8 | f[6] << 16 | f[7] << 24;
            bits = f[14] | f[15] << 8;
        } else if (!memcmp(ch, "data", 4)) {
            break;
        } else {
            uint8_t tmp[64];
            while (size > 0) {
                uint32_t n = size < sizeof(tmp) ? size : sizeof(tmp);
                if (!read_full(c, tmp, n)) goto done;
                size -= n;
            }
        }
    }
    if (bits != 16 || channels < 1 || channels > 2 || rate < 8000 || rate > 48000) {
        ESP_LOGW(TAG, "unsupported WAV: %d Hz, %d bit, %d ch", rate, bits, channels);
        goto done;
    }
    open_rate(rate, channels);
    uint8_t *buf = heap_caps_malloc(4096, MALLOC_CAP_DEFAULT);
    if (!buf) goto done;
    int r;
    while (!s_stop && (r = esp_http_client_read(c, (char *)buf, 4096)) > 0) {
        esp_codec_dev_write(s_spk, buf, r & ~1);
    }
    free(buf);
done:
    esp_http_client_close(c);
    esp_http_client_cleanup(c);
}

static void audio_task(void *arg)
{
    audio_msg_t m;
    for (;;) {
        if (xQueueReceive(s_q, &m, portMAX_DELAY) != pdTRUE) continue;
        s_stop = false;
        if (m.kind == 0) {
            play_fx(m.fx);
        } else if (m.url) {
            play_wav_url(m.url);
            free(m.url);
        }
    }
}

esp_err_t audio_init(void)
{
    ESP_RETURN_ON_ERROR(i2s_init(), TAG, "i2s");
    if (codec_init() != ESP_OK) {
        ESP_LOGE(TAG, "ES8311 codec not found: no sound");
        s_spk = NULL;
        return ESP_FAIL;
    }
    s_q = xQueueCreate(6, sizeof(audio_msg_t));
    xTaskCreatePinnedToCore(audio_task, "audio", 6144, NULL, 5, NULL, 0);
    return ESP_OK;
}

void audio_set_volume(int volume)
{
    if (volume < 0) volume = 0;
    if (volume > 100) volume = 100;
    s_volume = volume;
    if (s_spk && s_open_rate) esp_codec_dev_set_out_vol(s_spk, volume);
}

int audio_get_volume(void) { return s_volume; }

void audio_fx(audio_fx_t fx)
{
    if (!s_q || s_volume == 0) return;
    audio_msg_t m = {.kind = 0, .fx = fx};
    xQueueSend(s_q, &m, 0);
}

void audio_play_url(const char *url)
{
    if (!s_q || !url || !*url) return;
    audio_msg_t m = {.kind = 1, .url = strdup(url)};
    if (m.url && xQueueSend(s_q, &m, 0) != pdTRUE) free(m.url);
}

void audio_stop(void) { s_stop = true; }
