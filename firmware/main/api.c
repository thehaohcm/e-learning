#include "api.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "cJSON.h"
#include "esp_app_desc.h"
#include "esp_crt_bundle.h"
#include "esp_heap_caps.h"
#include "esp_http_client.h"
#include "esp_log.h"
#include "net.h"
#include "store.h"

static const char *TAG = "api";

#define MAX_BODY (256 * 1024)

// all calls run in the app task, one at a time; AI calls may take longer
static int s_timeout_ms = 15000;

typedef struct {
    char *buf;
    int len, cap;
} body_t;

static esp_err_t on_http(esp_http_client_event_t *e)
{
    body_t *b = e->user_data;
    if (e->event_id == HTTP_EVENT_ON_DATA && b) {
        if (b->len + e->data_len + 1 > b->cap) {
            int cap = b->cap ? b->cap * 2 : 8192;
            while (cap < b->len + e->data_len + 1) cap *= 2;
            if (cap > MAX_BODY) return ESP_FAIL;
            char *nb = heap_caps_realloc_prefer(b->buf, cap, 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
            if (!nb) return ESP_FAIL;
            b->buf = nb;
            b->cap = cap;
        }
        memcpy(b->buf + b->len, e->data, e->data_len);
        b->len += e->data_len;
        b->buf[b->len] = 0;
    }
    return ESP_OK;
}

/** One request. Returns the HTTP status (or -1); the body is in out (caller frees out->buf). */
static int request(esp_http_client_method_t method, const char *path, const char *json, bool auth, body_t *out)
{
    if (!net_connected()) return -1;
    char url[256];
    snprintf(url, sizeof(url), "%s%s", cfg()->server, path);
    esp_http_client_config_t hc = {
        .url = url,
        .method = method,
        .timeout_ms = s_timeout_ms,
        .crt_bundle_attach = esp_crt_bundle_attach,
        .event_handler = on_http,
        .user_data = out,
        .buffer_size = 2048,
        .buffer_size_tx = 1024,
    };
    esp_http_client_handle_t c = esp_http_client_init(&hc);
    if (!c) return -1;
    esp_http_client_set_header(c, "Accept", "application/json");
    if (auth && cfg()->token[0]) {
        char h[100];
        snprintf(h, sizeof(h), "Bearer %s", cfg()->token);
        esp_http_client_set_header(c, "Authorization", h);
    }
    if (json) {
        esp_http_client_set_header(c, "Content-Type", "application/json");
        esp_http_client_set_post_field(c, json, strlen(json));
    }
    esp_err_t err = esp_http_client_perform(c);
    int status = err == ESP_OK ? esp_http_client_get_status_code(c) : -1;
    if (err != ESP_OK) ESP_LOGW(TAG, "%s: %s", path, esp_err_to_name(err));
    esp_http_client_cleanup(c);
    return status;
}

esp_err_t api_register(void)
{
    if (cfg()->token[0]) return ESP_OK;
    const esp_app_desc_t *app = esp_app_get_description();
    char json[128];
    snprintf(json, sizeof(json), "{\"name\":\"ES3C28P study device\",\"firmware\":\"%s\"}", app->version);
    body_t b = {0};
    int st = request(HTTP_METHOD_POST, "/api/device/register", json, false, &b);
    esp_err_t res = ESP_FAIL;
    if (st == 201 || st == 200) {
        cJSON *r = cJSON_Parse(b.buf);
        const cJSON *tok = cJSON_GetObjectItem(r, "token");
        if (cJSON_IsString(tok)) {
            strlcpy(cfg()->token, tok->valuestring, sizeof(cfg()->token));
            cfg_save();
            res = ESP_OK;
        }
        cJSON_Delete(r);
    } else {
        ESP_LOGW(TAG, "register: HTTP %d", st);
    }
    free(b.buf);
    return res;
}

esp_err_t api_status(api_status_t *out)
{
    memset(out, 0, sizeof(*out));
    body_t b = {0};
    int st = request(HTTP_METHOD_GET, "/api/device/status", NULL, true, &b);
    esp_err_t res = ESP_FAIL;
    if (st == 401) {
        res = ESP_ERR_INVALID_STATE; // token unknown: register again
    } else if (st == 200) {
        cJSON *r = cJSON_Parse(b.buf);
        out->paired = cJSON_IsTrue(cJSON_GetObjectItem(r, "paired"));
        const cJSON *v = cJSON_GetObjectItem(r, "pair_code");
        if (cJSON_IsString(v)) strlcpy(out->pair_code, v->valuestring, sizeof(out->pair_code));
        v = cJSON_GetObjectItem(r, "user_name");
        if (cJSON_IsString(v)) strlcpy(out->user_name, v->valuestring, sizeof(out->user_name));
        cJSON_Delete(r);
        res = ESP_OK;
    }
    free(b.buf);
    return res;
}

esp_err_t api_sync(void)
{
    body_t b = {0};
    char path[80];
    snprintf(path, sizeof(path), "/api/device/sync?limit=40&fw=%s", esp_app_get_description()->version);
    int st = request(HTTP_METHOD_GET, path, NULL, true, &b);
    esp_err_t res = ESP_FAIL;
    if (st == 401) {
        res = ESP_ERR_INVALID_STATE;
    } else if (st == 409) {
        res = ESP_ERR_NOT_FOUND; // not paired (any more)
    } else if (st == 200 && b.buf) {
        deck_t *d = deck_parse(b.buf, b.len);
        if (d) {
            clock_set_from_server(d->server_time);
            if (!d->synced_at) d->synced_at = d->server_time;
            deck_set(d);
            deck_cache_save(b.buf, b.len);
            res = ESP_OK;
            ESP_LOGI(TAG, "synced: %d cards, %d due", d->n_cards, d->due_now);
        }
    } else {
        ESP_LOGW(TAG, "sync: HTTP %d", st);
    }
    free(b.buf);
    return res;
}

int api_extra_new(int n)
{
    char json[32];
    snprintf(json, sizeof(json), "{\"count\":%d}", n);
    body_t b = {0};
    int st = request(HTTP_METHOD_POST, "/api/device/extra-new", json, true, &b);
    if (st != 200) ESP_LOGW(TAG, "extra-new: HTTP %d %s", st, b.buf ? b.buf : "");
    free(b.buf);
    return st;
}

static char *dupstr(const cJSON *o, const char *key)
{
    const cJSON *v = cJSON_GetObjectItem(o, key);
    return cJSON_IsString(v) ? strdup(v->valuestring) : NULL;
}

void word_q_free(word_q_t *q, int n)
{
    for (int i = 0; q && i < n; i++) {
        free(q[i].prompt);
        free(q[i].explanation);
        for (int k = 0; k < 4; k++) free(q[i].choices[k]);
    }
    free(q);
}

int api_practice(const int64_t *card_ids, int n, word_q_t **out)
{
    *out = NULL;
    cJSON *root = cJSON_CreateObject();
    cJSON *arr = cJSON_AddArrayToObject(root, "card_ids");
    for (int i = 0; i < n; i++) cJSON_AddItemToArray(arr, cJSON_CreateNumber((double)card_ids[i]));
    char *json = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);
    body_t b = {0};
    s_timeout_ms = 50000; // the server may ask AI to write the questions (~10-40 s)
    int st = request(HTTP_METHOD_POST, "/api/device/practice", json, true, &b);
    s_timeout_ms = 15000;
    free(json);
    int count = -1;
    if (st == 200 && b.buf) {
        cJSON *r = cJSON_Parse(b.buf);
        const cJSON *qs = cJSON_GetObjectItem(r, "questions");
        int m = cJSON_GetArraySize(qs);
        word_q_t *q = m > 0 ? calloc(m, sizeof(word_q_t)) : NULL;
        count = 0;
        const cJSON *it;
        cJSON_ArrayForEach(it, qs)
        {
            if (!q) break;
            word_q_t *w = &q[count];
            const cJSON *ch = cJSON_GetObjectItem(it, "choices");
            const cJSON *id = cJSON_GetObjectItem(it, "card_id");
            const cJSON *ans = cJSON_GetObjectItem(it, "answer");
            if (cJSON_GetArraySize(ch) != 4 || !cJSON_IsNumber(id) || !cJSON_IsNumber(ans)) continue;
            w->card_id = (int64_t)id->valuedouble;
            w->answer = ans->valueint;
            const cJSON *k = cJSON_GetObjectItem(it, "kind");
            strlcpy(w->kind, cJSON_IsString(k) ? k->valuestring : "usage", sizeof(w->kind));
            w->prompt = dupstr(it, "prompt");
            w->explanation = dupstr(it, "explanation");
            for (int i = 0; i < 4; i++) {
                const cJSON *c = cJSON_GetArrayItem(ch, i);
                w->choices[i] = strdup(cJSON_IsString(c) ? c->valuestring : "");
            }
            if (!w->prompt || w->answer < 0 || w->answer > 3) {
                free(w->prompt);
                free(w->explanation);
                for (int i = 0; i < 4; i++) free(w->choices[i]);
                memset(w, 0, sizeof(*w));
                continue;
            }
            count++;
        }
        cJSON_Delete(r);
        *out = q;
        if (count == 0) {
            free(q);
            *out = NULL;
        }
    } else {
        ESP_LOGW(TAG, "practice: HTTP %d", st);
    }
    free(b.buf);
    return count;
}

void api_prepare_forms(void)
{
    body_t b = {0};
    int st = request(HTTP_METHOD_POST, "/api/device/prepare-forms", "{}", true, &b);
    if (st != 200 && st != 202) ESP_LOGW(TAG, "prepare-forms: HTTP %d", st);
    free(b.buf);
}

static void iso_time(int64_t t, char *dst, size_t len)
{
    time_t tt = (time_t)t;
    struct tm tm;
    gmtime_r(&tt, &tm);
    strftime(dst, len, "%Y-%m-%dT%H:%M:%SZ", &tm);
}

int api_push_reviews(void)
{
    if (!reviews_count() || !net_connected()) return 0;
    review_t *rs = heap_caps_malloc_prefer(sizeof(review_t) * 100, 2, MALLOC_CAP_SPIRAM | MALLOC_CAP_8BIT, MALLOC_CAP_DEFAULT);
    if (!rs) return 0;
    int n = reviews_peek(rs, 100);
    cJSON *root = cJSON_CreateObject();
    cJSON *arr = cJSON_AddArrayToObject(root, "reviews");
    for (int i = 0; i < n; i++) {
        cJSON *o = cJSON_CreateObject();
        cJSON_AddStringToObject(o, "client_review_id", rs[i].id);
        cJSON_AddNumberToObject(o, "card_id", (double)rs[i].card_id);
        cJSON_AddNumberToObject(o, "rating", rs[i].rating & 0x0F);
        if (rs[i].rating & REVIEW_QUIZ) cJSON_AddTrueToObject(o, "quiz");
        if (rs[i].at > 0) { // unknown time: the server uses "now"
            char ts[32];
            iso_time(rs[i].at, ts, sizeof(ts));
            cJSON_AddStringToObject(o, "reviewed_at", ts);
        }
        if (rs[i].duration_ms > 0) cJSON_AddNumberToObject(o, "duration_ms", rs[i].duration_ms);
        cJSON_AddItemToArray(arr, o);
    }
    char *json = cJSON_PrintUnformatted(root);
    cJSON_Delete(root);
    free(rs);
    body_t b = {0};
    int st = request(HTTP_METHOD_POST, "/api/device/reviews", json, true, &b);
    free(json);
    free(b.buf);
    if (st == 200) {
        reviews_drop(n); // duplicates are fine: the server ignores ids it already has
        ESP_LOGI(TAG, "sent %d answers", n);
        return n;
    }
    ESP_LOGW(TAG, "reviews: HTTP %d", st);
    return 0;
}
