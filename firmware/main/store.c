#include "store.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

#include "cJSON.h"
#include "esp_heap_caps.h"
#include "esp_log.h"
#include "esp_mac.h"
#include "esp_spiffs.h"
#include "freertos/FreeRTOS.h"
#include "freertos/semphr.h"
#include "nvs.h"
#include "nvs_flash.h"

static const char *TAG = "store";
static device_cfg_t s_cfg;
static SemaphoreHandle_t s_lock;
static deck_t *s_deck;

#define DECK_FILE "/data/deck.json"
#define PENDING_FILE "/data/pending.bin"
#define MAX_PENDING 400

static review_t *s_pending; // in PSRAM
static int s_n_pending;
static uint32_t s_review_seq;

// ---------------------------------------------------------------- settings

static void nvs_str(nvs_handle_t h, const char *key, char *dst, size_t len, const char *def)
{
    size_t n = len;
    if (nvs_get_str(h, key, dst, &n) != ESP_OK) {
        strlcpy(dst, def, len);
    }
}

static void load_cfg(void)
{
    nvs_handle_t h;
    memset(&s_cfg, 0, sizeof(s_cfg));
    s_cfg.volume = 70;
    s_cfg.brightness = 80;
    s_cfg.speak = true;
    s_cfg.recall_first = true;
    strlcpy(s_cfg.server, CONFIG_ELEARN_SERVER_URL, sizeof(s_cfg.server));
    if (nvs_open("cfg", NVS_READONLY, &h) != ESP_OK) return;
    nvs_str(h, "ssid", s_cfg.ssid, sizeof(s_cfg.ssid), "");
    nvs_str(h, "pass", s_cfg.pass, sizeof(s_cfg.pass), "");
    nvs_str(h, "server", s_cfg.server, sizeof(s_cfg.server), CONFIG_ELEARN_SERVER_URL);
    nvs_str(h, "token", s_cfg.token, sizeof(s_cfg.token), "");
    int32_t v;
    if (nvs_get_i32(h, "volume", &v) == ESP_OK) s_cfg.volume = v;
    if (nvs_get_i32(h, "bright", &v) == ESP_OK) s_cfg.brightness = v;
    uint8_t sp;
    if (nvs_get_u8(h, "speak", &sp) == ESP_OK) s_cfg.speak = sp != 0;
    if (nvs_get_u8(h, "recall", &sp) == ESP_OK) s_cfg.recall_first = sp != 0;
    uint32_t seq;
    if (nvs_get_u32(h, "seq", &seq) == ESP_OK) s_review_seq = seq;
    nvs_close(h);
}

device_cfg_t *cfg(void) { return &s_cfg; }

esp_err_t cfg_save(void)
{
    nvs_handle_t h;
    esp_err_t err = nvs_open("cfg", NVS_READWRITE, &h);
    if (err != ESP_OK) return err;
    nvs_set_str(h, "ssid", s_cfg.ssid);
    nvs_set_str(h, "pass", s_cfg.pass);
    nvs_set_str(h, "server", s_cfg.server);
    nvs_set_str(h, "token", s_cfg.token);
    nvs_set_i32(h, "volume", s_cfg.volume);
    nvs_set_i32(h, "bright", s_cfg.brightness);
    nvs_set_u8(h, "speak", s_cfg.speak ? 1 : 0);
    nvs_set_u8(h, "recall", s_cfg.recall_first ? 1 : 0);
    nvs_set_u32(h, "seq", s_review_seq);
    err = nvs_commit(h);
    nvs_close(h);
    return err;
}

void cfg_forget_wifi(void)
{
    s_cfg.ssid[0] = s_cfg.pass[0] = 0;
    cfg_save();
}

void cfg_forget_token(void)
{
    s_cfg.token[0] = 0;
    cfg_save();
    remove(DECK_FILE);
    remove(PENDING_FILE);
    s_n_pending = 0;
}

// ---------------------------------------------------------------- pending reviews

static void pending_save(void)
{
    FILE *f = fopen(PENDING_FILE, "wb");
    if (!f) return;
    fwrite(&s_n_pending, sizeof(s_n_pending), 1, f);
    fwrite(s_pending, sizeof(review_t), s_n_pending, f);
    fclose(f);
}

static void pending_load(void)
{
    FILE *f = fopen(PENDING_FILE, "rb");
    if (!f) return;
    int n = 0;
    if (fread(&n, sizeof(n), 1, f) == 1 && n > 0 && n <= MAX_PENDING) {
        s_n_pending = fread(s_pending, sizeof(review_t), n, f);
    }
    fclose(f);
}

void reviews_add(int64_t card_id, int rating, int duration_ms)
{
    xSemaphoreTake(s_lock, portMAX_DELAY);
    if (s_n_pending >= MAX_PENDING) { // keep the newest answers
        memmove(s_pending, s_pending + 1, sizeof(review_t) * (MAX_PENDING - 1));
        s_n_pending--;
    }
    uint8_t mac[6] = {0};
    esp_efuse_mac_get_default(mac);
    review_t *r = &s_pending[s_n_pending++];
    memset(r, 0, sizeof(*r));
    snprintf(r->id, sizeof(r->id), "%02x%02x%02x-%lu", mac[3], mac[4], mac[5], (unsigned long)++s_review_seq);
    r->card_id = card_id;
    r->rating = rating;
    r->at = now_utc();
    r->duration_ms = duration_ms;
    pending_save();
    xSemaphoreGive(s_lock);
    cfg_save(); // keeps the sequence number unique across restarts
}

int reviews_peek(review_t *out, int max)
{
    xSemaphoreTake(s_lock, portMAX_DELAY);
    int n = s_n_pending < max ? s_n_pending : max;
    memcpy(out, s_pending, sizeof(review_t) * n);
    xSemaphoreGive(s_lock);
    return n;
}

void reviews_drop(int n)
{
    xSemaphoreTake(s_lock, portMAX_DELAY);
    if (n > s_n_pending) n = s_n_pending;
    memmove(s_pending, s_pending + n, sizeof(review_t) * (s_n_pending - n));
    s_n_pending -= n;
    pending_save();
    xSemaphoreGive(s_lock);
}

int reviews_count(void) { return s_n_pending; }

// ---------------------------------------------------------------- clock

time_t now_utc(void)
{
    time_t t = time(NULL);
    return t > 1700000000 ? t : 0; // not set yet
}

void clock_set_from_server(time_t t)
{
    if (t <= 0) return;
    time_t now = time(NULL);
    if (now < 1700000000 || llabs((long long)(now - t)) > 120) {
        struct timeval tv = {.tv_sec = t};
        settimeofday(&tv, NULL);
    }
}

// ---------------------------------------------------------------- deck

static char *jstr(const cJSON *o, const char *key)
{
    const cJSON *v = cJSON_GetObjectItemCaseSensitive(o, key);
    if (!cJSON_IsString(v) || !v->valuestring[0]) return NULL;
    return strdup(v->valuestring);
}

static int jint(const cJSON *o, const char *key, int def)
{
    const cJSON *v = cJSON_GetObjectItemCaseSensitive(o, key);
    return cJSON_IsNumber(v) ? (int)v->valuedouble : def;
}

static int64_t jid(const cJSON *o, const char *key)
{
    const cJSON *v = cJSON_GetObjectItemCaseSensitive(o, key);
    return cJSON_IsNumber(v) ? (int64_t)v->valuedouble : 0;
}

static void jcopy(const cJSON *o, const char *key, char *dst, size_t len, const char *def)
{
    const cJSON *v = cJSON_GetObjectItemCaseSensitive(o, key);
    strlcpy(dst, cJSON_IsString(v) ? v->valuestring : def, len);
}

/** "2026-10-23T08:00:00.123Z" -> unix time (UTC). */
static time_t parse_time(const cJSON *v)
{
    if (!cJSON_IsString(v)) return 0;
    struct tm tm = {0};
    if (sscanf(v->valuestring, "%d-%d-%dT%d:%d:%d", &tm.tm_year, &tm.tm_mon, &tm.tm_mday, &tm.tm_hour, &tm.tm_min,
               &tm.tm_sec) != 6)
        return 0;
    tm.tm_year -= 1900;
    tm.tm_mon -= 1;
    // timegm without TZ: days from civil
    int y = tm.tm_year + 1900, m = tm.tm_mon + 1, d = tm.tm_mday;
    y -= m <= 2;
    int era = (y >= 0 ? y : y - 399) / 400;
    int yoe = y - era * 400;
    int doy = (153 * (m + (m > 2 ? -3 : 9)) + 2) / 5 + d - 1;
    int doe = yoe * 365 + yoe / 4 - yoe / 100 + doy;
    long days = (long)era * 146097 + doe - 719468;
    return (time_t)days * 86400 + tm.tm_hour * 3600 + tm.tm_min * 60 + tm.tm_sec;
}

static int parse_options(const cJSON *arr, option_t *out, int max)
{
    int n = 0;
    const cJSON *it;
    cJSON_ArrayForEach(it, arr)
    {
        if (n >= max) break;
        char *t = jstr(it, "text");
        if (!t) continue;
        out[n].id = jid(it, "id");
        out[n].text = t;
        n++;
    }
    return n;
}

static void card_free(card_t *c)
{
    free(c->front); free(c->ipa); free(c->pos); free(c->meaning_vi); free(c->meaning_en);
    free(c->example_en); free(c->example_vi); free(c->back); free(c->audio_url);
    for (int i = 0; i < c->n_choices; i++) free(c->choices[i]);
    for (int i = 0; i < c->n_options; i++) free(c->options[i].text);
    free(c->cloze_sentence); free(c->cloze_vi);
    for (int i = 0; i < c->n_cloze; i++) free(c->cloze[i].text);
}

static void deck_free(deck_t *d)
{
    if (!d) return;
    for (int i = 0; i < d->n_cards; i++) card_free(&d->cards[i]);
    free(d->cards);
    for (int i = 0; i < d->n_quiz; i++) card_free(&d->quiz[i]);
    free(d->quiz);
    free(d->fw_url);
    free(d);
}

/** "/api/..." -> "https://server/api/..." (takes ownership of url). */
static char *absolute_url(char *url)
{
    if (!url || url[0] != '/') return url;
    char *abs = malloc(strlen(s_cfg.server) + strlen(url) + 1);
    if (abs) {
        strcpy(abs, s_cfg.server);
        strcat(abs, url);
    }
    free(url);
    return abs;
}

/** One card of the sync JSON. false if it has no front text. */
static bool parse_card(const cJSON *c, card_t *k)
{
    k->id = jid(c, "id");
    jcopy(c, "kind", k->kind, sizeof(k->kind), "vocab");
    jcopy(c, "mode", k->mode, sizeof(k->mode), "flashcard");
    k->is_new = cJSON_IsTrue(cJSON_GetObjectItem(c, "is_new"));
    k->front = jstr(c, "front");
    if (!k->front) return false;
    k->ipa = jstr(c, "ipa");
    k->pos = jstr(c, "pos");
    k->meaning_vi = jstr(c, "meaning_vi");
    k->meaning_en = jstr(c, "meaning_en");
    k->example_en = jstr(c, "example_en");
    k->example_vi = jstr(c, "example_vi");
    k->back = jstr(c, "back");
    k->audio_url = absolute_url(jstr(c, "audio_url"));
    k->answer = jint(c, "answer", -1);
    const cJSON *ch;
    cJSON_ArrayForEach(ch, cJSON_GetObjectItem(c, "choices"))
    {
        if (k->n_choices < 6 && cJSON_IsString(ch)) k->choices[k->n_choices++] = strdup(ch->valuestring);
    }
    k->n_options = parse_options(cJSON_GetObjectItem(c, "options"), k->options, 4);
    const cJSON *cz = cJSON_GetObjectItem(c, "cloze");
    if (cz) {
        k->cloze_sentence = jstr(cz, "sentence");
        k->cloze_vi = jstr(cz, "translation_vi");
        jcopy(cz, "kind", k->cloze_kind, sizeof(k->cloze_kind), "word");
        k->n_cloze = parse_options(cJSON_GetObjectItem(cz, "options"), k->cloze, 4);
    }
    // a mode the device cannot show falls back to a flashcard
    if (!strcmp(k->mode, "cloze") && (!k->cloze_sentence || k->n_cloze < 2)) strcpy(k->mode, "flashcard");
    if ((!strcmp(k->mode, "quiz") || !strcmp(k->mode, "listening")) && k->n_options < 2 && k->n_choices < 2)
        strcpy(k->mode, "flashcard");
    return true;
}

deck_t *deck_parse(const char *json, size_t len)
{
    cJSON *root = cJSON_ParseWithLength(json, len);
    if (!root) return NULL;
    deck_t *d = heap_caps_calloc_prefer(1, sizeof(deck_t), 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
    if (!d) {
        cJSON_Delete(root);
        return NULL;
    }
    d->refs = 1;
    d->synced_at = now_utc();
    d->server_time = parse_time(cJSON_GetObjectItem(root, "server_time"));
    d->utc_offset_sec = jint(root, "utc_offset_sec", 7 * 3600);
    d->due_now = jint(root, "due_now", 0);
    d->new_left_today = jint(root, "new_left_today", 0);
    d->reviews_today = jint(root, "reviews_today", 0);
    d->streak_days = jint(root, "streak_days", 0);
    d->extra_new_left = jint(root, "extra_new_left", 0);
    const cJSON *st = cJSON_GetObjectItem(root, "settings");
    d->daily_new = jint(st, "daily_new", 10);
    d->reminder_every_min = jint(st, "reminder_every_min", 45);
    d->popup_cards = jint(st, "popup_cards", 10);
    jcopy(st, "active_from", d->active_from, sizeof(d->active_from), "07:30");
    jcopy(st, "active_to", d->active_to, sizeof(d->active_to), "22:00");
    jcopy(st, "device_mode", d->device_mode, sizeof(d->device_mode), "flashcard");

    const cJSON *cards = cJSON_GetObjectItem(root, "cards");
    int n = cJSON_GetArraySize(cards);
    d->cards = heap_caps_calloc_prefer(n > 0 ? n : 1, sizeof(card_t), 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
    if (!d->cards) {
        free(d);
        cJSON_Delete(root);
        return NULL;
    }
    const cJSON *c;
    cJSON_ArrayForEach(c, cards)
    {
        if (parse_card(c, &d->cards[d->n_cards])) d->n_cards++;
        else card_free(&d->cards[d->n_cards]), memset(&d->cards[d->n_cards], 0, sizeof(card_t));
    }
    const cJSON *fw = cJSON_GetObjectItem(root, "firmware");
    if (fw) {
        jcopy(fw, "version", d->fw_version, sizeof(d->fw_version), "");
        d->fw_url = absolute_url(jstr(fw, "url"));
    }
    const cJSON *quiz = cJSON_GetObjectItem(root, "quiz");
    int nq = cJSON_GetArraySize(quiz);
    if (nq > 0) {
        d->quiz = heap_caps_calloc_prefer(nq, sizeof(card_t), 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
        cJSON_ArrayForEach(c, quiz)
        {
            if (!d->quiz) break;
            if (parse_card(c, &d->quiz[d->n_quiz])) d->n_quiz++;
            else card_free(&d->quiz[d->n_quiz]), memset(&d->quiz[d->n_quiz], 0, sizeof(card_t));
        }
    }
    cJSON_Delete(root);
    return d;
}

deck_t *deck_get(void)
{
    xSemaphoreTake(s_lock, portMAX_DELAY);
    deck_t *d = s_deck;
    if (d) d->refs++;
    xSemaphoreGive(s_lock);
    return d;
}

void deck_unref(deck_t *d)
{
    if (!d) return;
    xSemaphoreTake(s_lock, portMAX_DELAY);
    bool last = --d->refs == 0;
    xSemaphoreGive(s_lock);
    if (last) deck_free(d);
}

void deck_set(deck_t *d)
{
    xSemaphoreTake(s_lock, portMAX_DELAY);
    deck_t *old = s_deck;
    s_deck = d;
    xSemaphoreGive(s_lock);
    deck_unref(old);
}

void deck_cache_save(const char *json, size_t len)
{
    FILE *f = fopen(DECK_FILE, "wb");
    if (!f) return;
    fwrite(json, 1, len, f);
    fclose(f);
}

deck_t *deck_cache_load(void)
{
    FILE *f = fopen(DECK_FILE, "rb");
    if (!f) return NULL;
    fseek(f, 0, SEEK_END);
    long n = ftell(f);
    fseek(f, 0, SEEK_SET);
    deck_t *d = NULL;
    if (n > 0 && n < 512 * 1024) {
        char *buf = heap_caps_malloc_prefer(n + 1, 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
        if (buf && fread(buf, 1, n, f) == (size_t)n) {
            buf[n] = 0;
            d = deck_parse(buf, n);
        }
        free(buf);
    }
    fclose(f);
    return d;
}

// ---------------------------------------------------------------- init

esp_err_t store_init(void)
{
    esp_err_t err = nvs_flash_init();
    if (err == ESP_ERR_NVS_NO_FREE_PAGES || err == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        err = nvs_flash_init();
    }
    if (err != ESP_OK) return err;
    s_lock = xSemaphoreCreateMutex();
    esp_vfs_spiffs_conf_t fs = {.base_path = "/data", .partition_label = "storage", .max_files = 4,
                                .format_if_mount_failed = true};
    if (esp_vfs_spiffs_register(&fs) != ESP_OK) {
        ESP_LOGE(TAG, "SPIFFS mount failed: offline cache disabled");
    }
    s_pending = heap_caps_calloc_prefer(MAX_PENDING, sizeof(review_t), 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
    load_cfg();
    pending_load();
    ESP_LOGI(TAG, "%d answers waiting to be sent", s_n_pending);
    return ESP_OK;
}
