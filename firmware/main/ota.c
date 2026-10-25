#include "ota.h"

#include <stdio.h>
#include <stdlib.h>

#include "esp_app_desc.h"
#include "esp_crt_bundle.h"
#include "esp_http_client.h"
#include "esp_https_ota.h"
#include "esp_log.h"
#include "store.h"

static const char *TAG = "ota";

static void parse(const char *v, int out[3])
{
    out[0] = out[1] = out[2] = 0;
    if (v) sscanf(v, "%d.%d.%d", &out[0], &out[1], &out[2]);
}

bool ota_is_newer(const char *version)
{
    if (!version || !*version) return false;
    int a[3], b[3];
    parse(version, a);
    parse(esp_app_get_description()->version, b);
    for (int i = 0; i < 3; i++) {
        if (a[i] != b[i]) return a[i] > b[i];
    }
    return false;
}

static esp_err_t add_token(esp_http_client_handle_t c)
{
    char h[100];
    snprintf(h, sizeof(h), "Bearer %s", cfg()->token);
    return esp_http_client_set_header(c, "Authorization", h);
}

esp_err_t ota_update(const char *url, void (*progress)(int percent))
{
    esp_http_client_config_t hc = {
        .url = url,
        .timeout_ms = 30000,
        .crt_bundle_attach = esp_crt_bundle_attach,
        .keep_alive_enable = true,
        .buffer_size = 4096,
        .buffer_size_tx = 1024,
    };
    esp_https_ota_config_t oc = {.http_config = &hc, .http_client_init_cb = add_token};
    esp_https_ota_handle_t h = NULL;
    esp_err_t err = esp_https_ota_begin(&oc, &h);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "begin %s: %s", url, esp_err_to_name(err));
        return err;
    }
    int total = esp_https_ota_get_image_size(h), last = -1;
    while ((err = esp_https_ota_perform(h)) == ESP_ERR_HTTPS_OTA_IN_PROGRESS) {
        int done = esp_https_ota_get_image_len_read(h);
        int pct = total > 0 ? (int)((int64_t)done * 100 / total) : 0;
        if (progress && pct / 5 != last / 5) {
            last = pct;
            progress(pct);
        }
    }
    if (err == ESP_OK && !esp_https_ota_is_complete_data_received(h)) err = ESP_ERR_INVALID_SIZE;
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "download failed: %s", esp_err_to_name(err));
        esp_https_ota_abort(h);
        return err;
    }
    err = esp_https_ota_finish(h); // checks the image and makes it the next one to start
    if (err != ESP_OK) ESP_LOGE(TAG, "finish: %s", esp_err_to_name(err));
    else if (progress) progress(100);
    return err;
}
