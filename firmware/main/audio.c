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
#include "esp_timer.h"
#include "freertos/FreeRTOS.h"
#include "freertos/queue.h"
#include "freertos/semphr.h"
#include "freertos/task.h"

static const char *TAG = "audio";

#define AUDIO_RATE 16000

static i2s_chan_handle_t s_tx, s_rx;
static esp_codec_dev_handle_t s_spk;
static QueueHandle_t s_q;
static int s_volume = 70;
static int s_open_rate;
// Every new word to play (or audio_stop) bumps s_gen: whatever is playing or
// waiting with an older number is dropped, so fast answers never pile up sounds.
static volatile uint32_t s_gen;
static uint32_t s_cur; // number of what the task is playing now

static inline bool stale(void) { return s_cur != s_gen; }

typedef enum { MSG_FX, MSG_PLAY, MSG_PREFETCH } msg_kind_t;

typedef struct {
    msg_kind_t kind;
    audio_fx_t fx;
    char *url;
    uint32_t gen;
    uint32_t batch; // MSG_PREFETCH from audio_preload: counted in s_pre_done
} audio_msg_t;

// audio_preload: words of the round downloaded before it starts
static volatile uint32_t s_pre_batch;
static volatile int s_pre_done;

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
    for (int i = 0; i < n; i += chunk) {
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
    for (int i = 0; i < ms / 10; i++) esp_codec_dev_write(s_spk, z, sizeof(z));
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

// ---------------------------------------------------------------- words (WAV over HTTPS)
//
// A word is downloaded completely before it plays (no gaps when the network is
// slow) and kept in a small cache in PSRAM, so "play again", the quiz and the
// next card (prefetched while the current one is on screen) start at once.
// One HTTPS connection is kept open and reused (no TLS handshake per word).

#define CACHE_SLOTS 48
#define CACHE_MAX_BYTES (1536 * 1024)
#define WAV_MAX_BYTES (256 * 1024)

typedef struct {
    char *url;
    uint8_t *data;
    int len;
    uint32_t used; // for LRU
} cache_t;

static cache_t s_cache[CACHE_SLOTS];
static SemaphoreHandle_t s_cache_lock; // the UI asks audio_ready() from another task
static int s_cache_bytes;
static uint32_t s_tick;
static esp_http_client_handle_t s_http;
static int64_t s_http_used_us; // last use of the kept-open connection

// The server (nginx/Cloudflare) closes an idle connection after about a minute;
// reusing a closed one only fails after a long timeout (the word came 10 s late
// or never). So it is reused only for words close together, as during a round.
#define KEEP_OPEN_US (20 * 1000000LL)

typedef struct {
    uint8_t *buf;
    int len, cap;
    bool too_big;
} dl_t;

static esp_err_t on_data(esp_http_client_event_t *e)
{
    dl_t *d = e->user_data;
    if (e->event_id != HTTP_EVENT_ON_DATA || !d || d->too_big) return ESP_OK;
    if (d->len + e->data_len > d->cap) {
        int cap = d->cap ? d->cap * 2 : 48 * 1024;
        while (cap < d->len + e->data_len) cap *= 2;
        if (cap > WAV_MAX_BYTES) {
            d->too_big = true;
            return ESP_OK;
        }
        uint8_t *nb = heap_caps_realloc_prefer(d->buf, cap, 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
        if (!nb) {
            d->too_big = true;
            return ESP_OK;
        }
        d->buf = nb;
        d->cap = cap;
    }
    memcpy(d->buf + d->len, e->data, e->data_len);
    d->len += e->data_len;
    return ESP_OK;
}

static cache_t *cache_find(const char *url)
{
    if (!url) return NULL;
    for (int i = 0; i < CACHE_SLOTS; i++) {
        if (s_cache[i].url && !strcmp(s_cache[i].url, url)) {
            s_cache[i].used = ++s_tick;
            return &s_cache[i];
        }
    }
    return NULL;
}

static void cache_drop(cache_t *c)
{
    s_cache_bytes -= c->len;
    free(c->url);
    free(c->data);
    memset(c, 0, sizeof(*c));
}

static cache_t *cache_put(const char *url, uint8_t *data, int len)
{
    for (;;) { // make room: drop the least recently used
        int free_slot = -1, oldest = -1;
        for (int i = 0; i < CACHE_SLOTS; i++) {
            if (!s_cache[i].url) {
                if (free_slot < 0) free_slot = i;
            } else if (oldest < 0 || s_cache[i].used < s_cache[oldest].used) {
                oldest = i;
            }
        }
        if (free_slot >= 0 && s_cache_bytes + len <= CACHE_MAX_BYTES) {
            cache_t *c = &s_cache[free_slot];
            c->url = strdup(url);
            if (!c->url) break;
            c->data = data;
            c->len = len;
            c->used = ++s_tick;
            s_cache_bytes += len;
            return c;
        }
        if (oldest < 0) break;
        cache_drop(&s_cache[oldest]);
    }
    free(data);
    return NULL;
}

/** The WAV at url: from the cache, or downloaded (and cached). NULL on error. */
static cache_t *fetch(const char *url)
{
    xSemaphoreTake(s_cache_lock, portMAX_DELAY);
    cache_t *hit = cache_find(url);
    xSemaphoreGive(s_cache_lock);
    if (hit) return hit; // entries are only dropped by this task, so it stays valid
    dl_t d = {0};
    if (s_http && esp_timer_get_time() - s_http_used_us > KEEP_OPEN_US) {
        esp_http_client_cleanup(s_http);
        s_http = NULL;
    }
    for (int attempt = 0; attempt < 2; attempt++) {
        if (!s_http) {
            esp_http_client_config_t cfg = {
                .url = url,
                .timeout_ms = 12000, // a word not made yet is made on request (a few seconds)
                .crt_bundle_attach = esp_crt_bundle_attach,
                .buffer_size = 4096,
                .keep_alive_enable = true,
                .event_handler = on_data,
            };
            s_http = esp_http_client_init(&cfg);
            if (!s_http) return NULL;
        }
        esp_http_client_set_url(s_http, url);
        esp_http_client_set_user_data(s_http, &d);
        d.len = 0;
        d.too_big = false;
        int64_t t0 = esp_timer_get_time();
        esp_err_t err = esp_http_client_perform(s_http);
        int status = err == ESP_OK ? esp_http_client_get_status_code(s_http) : -1;
        s_http_used_us = esp_timer_get_time();
        if (err == ESP_OK && status == 200 && !d.too_big && d.len > 44) {
            ESP_LOGI(TAG, "got %s: %d bytes in %d ms", url, d.len, (int)((s_http_used_us - t0) / 1000));
            xSemaphoreTake(s_cache_lock, portMAX_DELAY);
            cache_t *c = cache_put(url, d.buf, d.len);
            xSemaphoreGive(s_cache_lock);
            return c;
        }
        ESP_LOGW(TAG, "%s: %s, HTTP %d", url, esp_err_to_name(err), status);
        // a broken kept-alive connection: start a new one and try once more
        esp_http_client_cleanup(s_http);
        s_http = NULL;
        if (status > 0) break; // the server answered (e.g. 404): no retry
    }
    free(d.buf);
    return NULL;
}

static inline uint32_t le32(const uint8_t *p) { return p[0] | p[1] << 8 | p[2] << 16 | (uint32_t)p[3] << 24; }

static void play_wav(const cache_t *w)
{
    const uint8_t *p = w->data, *end = w->data + w->len;
    if (w->len < 12 || memcmp(p, "RIFF", 4) || memcmp(p + 8, "WAVE", 4)) {
        ESP_LOGW(TAG, "not a WAV file");
        return;
    }
    int rate = AUDIO_RATE, channels = 1, bits = 16;
    const uint8_t *pcm = NULL;
    uint32_t pcm_len = 0;
    for (p += 12; p + 8 <= end;) {
        uint32_t size = le32(p + 4);
        const uint8_t *body = p + 8;
        if (!memcmp(p, "fmt ", 4) && size >= 16 && body + 16 <= end) {
            channels = body[2] | body[3] << 8;
            rate = (int)le32(body + 4);
            bits = body[14] | body[15] << 8;
        } else if (!memcmp(p, "data", 4)) {
            pcm = body;
            pcm_len = size > (uint32_t)(end - body) ? (uint32_t)(end - body) : size;
            break;
        }
        p = body + size + (size & 1);
    }
    if (!pcm || bits != 16 || channels < 1 || channels > 2 || rate < 8000 || rate > 48000) {
        ESP_LOGW(TAG, "unsupported WAV: %d Hz, %d bit, %d ch", rate, bits, channels);
        return;
    }
    open_rate(rate, channels);
    for (uint32_t off = 0; off < pcm_len && !stale();) {
        uint32_t n = pcm_len - off > 2048 ? 2048 : pcm_len - off;
        esp_codec_dev_write(s_spk, (void *)(pcm + off), n & ~1u);
        off += n;
    }
}

static void audio_task(void *arg)
{
    audio_msg_t m;
    for (;;) {
        if (xQueueReceive(s_q, &m, portMAX_DELAY) != pdTRUE) continue;
        s_cur = m.gen;
        switch (m.kind) {
        case MSG_FX:
            play_fx(m.fx); // short: always played, so right/wrong is never lost
            break;
        case MSG_PLAY: {
            if (stale()) break; // a newer word was asked for meanwhile
            cache_t *w = fetch(m.url);
            if (w && !stale()) play_wav(w);
            break;
        }
        case MSG_PREFETCH:
            fetch(m.url);
            if (m.batch && m.batch == s_pre_batch) s_pre_done++;
            break;
        }
        free(m.url);
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
    s_q = xQueueCreate(32, sizeof(audio_msg_t));
    s_cache_lock = xSemaphoreCreateMutex();
    xTaskCreatePinnedToCore(audio_task, "audio", 12288, NULL, 5, NULL, 0); // HTTPS (TLS) needs a big stack
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

static void enqueue(msg_kind_t kind, audio_fx_t fx, const char *url, bool front)
{
    audio_msg_t m = {.kind = kind, .fx = fx, .gen = s_gen};
    if (url && !(m.url = strdup(url))) return;
    BaseType_t ok = front ? xQueueSendToFront(s_q, &m, 0) : xQueueSend(s_q, &m, 0);
    if (ok != pdTRUE) free(m.url);
}

void audio_fx(audio_fx_t fx)
{
    if (!s_q || s_volume == 0) return;
    enqueue(MSG_FX, fx, NULL, false);
}

void audio_play_url(const char *url)
{
    if (!s_q || !url || !*url || s_volume == 0) return;
    s_gen++; // stops the word playing now and drops the ones still waiting
    enqueue(MSG_PLAY, 0, url, false);
}

void audio_prefetch(const char *url)
{
    // queued after what was asked for already, so it never delays it
    if (!s_q || !url || !*url || s_volume == 0 || uxQueueSpacesAvailable(s_q) < 3) return;
    enqueue(MSG_PREFETCH, 0, url, false);
}

void audio_stop(void) { s_gen++; }

bool audio_ready(const char *url)
{
    if (!url || !s_cache_lock) return false;
    xSemaphoreTake(s_cache_lock, portMAX_DELAY);
    bool ok = false;
    for (int i = 0; i < CACHE_SLOTS && !ok; i++) ok = s_cache[i].url && !strcmp(s_cache[i].url, url);
    xSemaphoreGive(s_cache_lock);
    return ok;
}

int audio_preload(const char *const *urls, int n)
{
    if (!s_q || s_volume == 0) return 0;
    s_pre_batch++;
    s_pre_done = 0;
    int queued = 0;
    for (int i = 0; i < n; i++) {
        if (!urls[i] || !*urls[i]) continue;
        audio_msg_t m = {.kind = MSG_PREFETCH, .gen = s_gen, .batch = s_pre_batch, .url = strdup(urls[i])};
        if (!m.url) continue;
        if (xQueueSend(s_q, &m, 0) != pdTRUE) {
            free(m.url);
            break;
        }
        queued++;
    }
    return queued;
}

int audio_preload_done(void) { return s_pre_done; }
