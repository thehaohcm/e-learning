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
        .timeout_ms = 15000,
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
    int st = request(HTTP_METHOD_GET, "/api/device/sync?limit=40", NULL, true, &b);
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
        cJSON_AddNumberToObject(o, "rating", rs[i].rating);
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
