#include "net.h"

#include <stdio.h>
#include <string.h>

#include "cJSON.h"
#include "esp_event.h"
#include "esp_timer.h"
#include "esp_log.h"
#include "esp_mac.h"
#include "esp_netif.h"
#include "esp_netif_sntp.h"
#include "esp_wifi.h"
#include "freertos/FreeRTOS.h"
#include "freertos/event_groups.h"
#include "portal.h"
#include "store.h"

static const char *TAG = "net";

#define BIT_CONNECTED BIT0
#define BIT_FAILED BIT1

static EventGroupHandle_t s_ev;
static esp_netif_t *s_sta, *s_ap;
static char s_ip[16] = "";
static char s_ap_ssid[24] = "";
static int s_retries;
static bool s_want_sta;
static esp_timer_handle_t s_retry_timer;

static void retry_cb(void *arg)
{
    if (s_want_sta) esp_wifi_connect();
}

static void on_event(void *arg, esp_event_base_t base, int32_t id, void *data)
{
    if (base == WIFI_EVENT && id == WIFI_EVENT_STA_START) {
        if (s_want_sta) esp_wifi_connect();
    } else if (base == WIFI_EVENT && id == WIFI_EVENT_STA_DISCONNECTED) {
        xEventGroupClearBits(s_ev, BIT_CONNECTED);
        s_ip[0] = 0;
        if (!s_want_sta) return;
        // keep trying in the background (never block the event loop); report failure after a few attempts
        if (++s_retries == 5) xEventGroupSetBits(s_ev, BIT_FAILED);
        esp_timer_stop(s_retry_timer);
        esp_timer_start_once(s_retry_timer, (s_retries < 5 ? 500 : 5000) * 1000ULL);
    } else if (base == IP_EVENT && id == IP_EVENT_STA_GOT_IP) {
        ip_event_got_ip_t *e = data;
        snprintf(s_ip, sizeof(s_ip), IPSTR, IP2STR(&e->ip_info.ip));
        s_retries = 0;
        xEventGroupClearBits(s_ev, BIT_FAILED);
        xEventGroupSetBits(s_ev, BIT_CONNECTED);
        ESP_LOGI(TAG, "connected, IP %s", s_ip);
    }
}

void net_init(void)
{
    s_ev = xEventGroupCreate();
    esp_timer_create_args_t ta = {.callback = retry_cb, .name = "wifi_retry"};
    esp_timer_create(&ta, &s_retry_timer);
    ESP_ERROR_CHECK(esp_netif_init());
    ESP_ERROR_CHECK(esp_event_loop_create_default());
    s_sta = esp_netif_create_default_wifi_sta();
    wifi_init_config_t wcfg = WIFI_INIT_CONFIG_DEFAULT();
    ESP_ERROR_CHECK(esp_wifi_init(&wcfg));
    esp_wifi_set_storage(WIFI_STORAGE_RAM); // credentials live in our own NVS namespace
    esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID, on_event, NULL);
    esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP, on_event, NULL);

    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_WIFI_SOFTAP);
    snprintf(s_ap_ssid, sizeof(s_ap_ssid), "ELearn-%02X%02X", mac[4], mac[5]);
}

esp_err_t net_connect(int timeout_ms)
{
    device_cfg_t *c = cfg();
    if (!c->ssid[0]) return ESP_ERR_INVALID_STATE;
    wifi_config_t wc = {0};
    strlcpy((char *)wc.sta.ssid, c->ssid, sizeof(wc.sta.ssid));
    strlcpy((char *)wc.sta.password, c->pass, sizeof(wc.sta.password));
    wc.sta.threshold.authmode = c->pass[0] ? WIFI_AUTH_WPA_PSK : WIFI_AUTH_OPEN;
    wc.sta.pmf_cfg.capable = true;
    s_want_sta = true;
    s_retries = 0;
    esp_wifi_set_mode(WIFI_MODE_STA);
    esp_wifi_set_config(WIFI_IF_STA, &wc);
    esp_wifi_set_ps(WIFI_PS_MIN_MODEM);
    esp_wifi_start();
    EventBits_t b = xEventGroupWaitBits(s_ev, BIT_CONNECTED | BIT_FAILED, pdFALSE, pdFALSE, pdMS_TO_TICKS(timeout_ms));
    return (b & BIT_CONNECTED) ? ESP_OK : ESP_FAIL;
}

bool net_connected(void) { return s_ev && (xEventGroupGetBits(s_ev) & BIT_CONNECTED); }

int net_rssi(void)
{
    wifi_ap_record_t ap;
    return (net_connected() && esp_wifi_sta_get_ap_info(&ap) == ESP_OK) ? ap.rssi : 0;
}

const char *net_ip(void) { return s_ip; }
const char *net_ap_ssid(void) { return s_ap_ssid; }

void net_start_setup_ap(void)
{
    s_want_sta = false;
    esp_wifi_stop();
    if (!s_ap) s_ap = esp_netif_create_default_wifi_ap();
    wifi_config_t ap = {0};
    strlcpy((char *)ap.ap.ssid, s_ap_ssid, sizeof(ap.ap.ssid));
    ap.ap.ssid_len = strlen(s_ap_ssid);
    ap.ap.channel = 6;
    ap.ap.max_connection = 4;
    ap.ap.authmode = WIFI_AUTH_OPEN;
    esp_wifi_set_mode(WIFI_MODE_APSTA); // STA part is used to scan for networks
    esp_wifi_set_config(WIFI_IF_AP, &ap);
    esp_wifi_start();
    portal_start();
    ESP_LOGI(TAG, "setup access point %s at 192.168.4.1", s_ap_ssid);
}

char *net_scan_json(void)
{
    wifi_scan_config_t sc = {.show_hidden = false};
    uint16_t n = 0;
    cJSON *arr = cJSON_CreateArray();
    if (esp_wifi_scan_start(&sc, true) == ESP_OK && esp_wifi_scan_get_ap_num(&n) == ESP_OK && n > 0) {
        if (n > 20) n = 20;
        wifi_ap_record_t *recs = calloc(n, sizeof(wifi_ap_record_t));
        if (recs && esp_wifi_scan_get_ap_records(&n, recs) == ESP_OK) {
            for (int i = 0; i < n; i++) {
                if (!recs[i].ssid[0]) continue;
                bool dup = false;
                cJSON *it;
                cJSON_ArrayForEach(it, arr)
                {
                    if (!strcmp(cJSON_GetObjectItem(it, "ssid")->valuestring, (char *)recs[i].ssid)) dup = true;
                }
                if (dup) continue;
                cJSON *o = cJSON_CreateObject();
                cJSON_AddStringToObject(o, "ssid", (char *)recs[i].ssid);
                cJSON_AddNumberToObject(o, "rssi", recs[i].rssi);
                cJSON_AddBoolToObject(o, "lock", recs[i].authmode != WIFI_AUTH_OPEN);
                cJSON_AddItemToArray(arr, o);
            }
        }
        free(recs);
    }
    char *s = cJSON_PrintUnformatted(arr);
    cJSON_Delete(arr);
    return s;
}

void net_start_time(void)
{
    static bool started;
    if (started) return;
    started = true;
    esp_sntp_config_t sc = ESP_NETIF_SNTP_DEFAULT_CONFIG("pool.ntp.org");
    esp_netif_sntp_init(&sc);
}
