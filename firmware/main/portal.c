#include "portal.h"

#include <ctype.h>
#include <stdlib.h>
#include <string.h>

#include "esp_http_server.h"
#include "esp_log.h"
#include "esp_system.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "lwip/sockets.h"
#include "net.h"
#include "store.h"

static const char *TAG = "portal";

static const char PAGE[] =
    "<!doctype html><html><head><meta charset=utf-8><meta name=viewport content='width=device-width,initial-scale=1'>"
    "<title>E-Learn device setup</title><style>"
    "body{font:16px system-ui,sans-serif;margin:0;background:#f1f5f9;color:#1e293b}"
    ".c{max-width:420px;margin:24px auto;background:#fff;border-radius:14px;padding:20px;box-shadow:0 4px 18px #0001}"
    "h1{font-size:20px;margin:0 0 6px}p{color:#64748b;font-size:14px;margin:0 0 14px}"
    "label{display:block;font-weight:600;margin:12px 0 4px}"
    "input,select{width:100%;box-sizing:border-box;padding:10px;border:1px solid #cbd5e1;border-radius:8px;font:inherit}"
    "button{margin-top:18px;width:100%;padding:12px;border:0;border-radius:8px;background:#4f46e5;color:#fff;font:600 16px system-ui}"
    ".s{font-size:13px;color:#64748b}a{color:#4f46e5}"
    "</style></head><body><div class=c><h1>&#127891; E-Learn device</h1>"
    "<p>Choose the Wi-Fi network the device should use.</p>"
    "<form method=post action=/save>"
    "<label>Wi-Fi network</label><select id=s name=ssid><option>Scanning…</option></select>"
    "<div class=s><a href=# onclick=\"m();return false\">Type the name instead</a></div>"
    "<label>Password</label><input name=pass type=password autocomplete=off placeholder='(empty for an open network)'>"
    "<label>Server</label><input name=server value='%SERVER%'>"
    "<button>Save &amp; connect</button></form></div><script>"
    "function m(){var s=document.getElementById('s'),i=document.createElement('input');i.name='ssid';i.id='s';"
    "i.placeholder='Network name';s.replaceWith(i);i.focus()}"
    "fetch('/scan').then(r=>r.json()).then(l=>{var s=document.getElementById('s');if(s.tagName!='SELECT')return;"
    "s.innerHTML='';l.sort((a,b)=>b.rssi-a.rssi).forEach(n=>{var o=document.createElement('option');"
    "o.value=n.ssid;o.textContent=n.ssid+(n.lock?' \\uD83D\\uDD12':'');s.appendChild(o)});"
    "if(!l.length)m()}).catch(m)"
    "</script></body></html>";

static esp_err_t page_get(httpd_req_t *req)
{
    // fill in the current server address
    const char *mark = strstr(PAGE, "%SERVER%");
    httpd_resp_set_type(req, "text/html; charset=utf-8");
    httpd_resp_send_chunk(req, PAGE, mark - PAGE);
    httpd_resp_send_chunk(req, cfg()->server, HTTPD_RESP_USE_STRLEN);
    httpd_resp_send_chunk(req, mark + 8, HTTPD_RESP_USE_STRLEN);
    return httpd_resp_send_chunk(req, NULL, 0);
}

static esp_err_t scan_get(httpd_req_t *req)
{
    char *json = net_scan_json();
    httpd_resp_set_type(req, "application/json");
    httpd_resp_sendstr(req, json ? json : "[]");
    free(json);
    return ESP_OK;
}

/** application/x-www-form-urlencoded value of key (decoded) into dst. */
static void form_value(const char *body, const char *key, char *dst, size_t len)
{
    dst[0] = 0;
    size_t kl = strlen(key);
    const char *p = body;
    while (p && *p) {
        if (!strncmp(p, key, kl) && p[kl] == '=') {
            p += kl + 1;
            size_t o = 0;
            while (*p && *p != '&' && o + 1 < len) {
                if (*p == '+') {
                    dst[o++] = ' ';
                    p++;
                } else if (*p == '%' && isxdigit((unsigned char)p[1]) && isxdigit((unsigned char)p[2])) {
                    char hex[3] = {p[1], p[2], 0};
                    dst[o++] = (char)strtol(hex, NULL, 16);
                    p += 3;
                } else {
                    dst[o++] = *p++;
                }
            }
            dst[o] = 0;
            return;
        }
        p = strchr(p, '&');
        if (p) p++;
    }
}

static void restart_task(void *arg)
{
    vTaskDelay(pdMS_TO_TICKS(1500));
    esp_restart();
}

static esp_err_t save_post(httpd_req_t *req)
{
    char body[512];
    int n = req->content_len < (int)sizeof(body) - 1 ? req->content_len : (int)sizeof(body) - 1;
    int got = 0;
    while (got < n) {
        int r = httpd_req_recv(req, body + got, n - got);
        if (r <= 0) return ESP_FAIL;
        got += r;
    }
    body[got] = 0;
    device_cfg_t *c = cfg();
    char ssid[33], pass[65], server[128];
    form_value(body, "ssid", ssid, sizeof(ssid));
    form_value(body, "pass", pass, sizeof(pass));
    form_value(body, "server", server, sizeof(server));
    if (!ssid[0]) {
        httpd_resp_sendstr(req, "<p>Please choose a Wi-Fi network. <a href=/>Back</a></p>");
        return ESP_OK;
    }
    size_t sl = strlen(server);
    while (sl && server[sl - 1] == '/') server[--sl] = 0;
    if (sl && strncmp(server, "http://", 7) && strncmp(server, "https://", 8)) {
        httpd_resp_sendstr(req, "<p>The server must start with https:// <a href=/>Back</a></p>");
        return ESP_OK;
    }
    bool server_changed = sl && strcmp(server, c->server);
    strlcpy(c->ssid, ssid, sizeof(c->ssid));
    strlcpy(c->pass, pass, sizeof(c->pass));
    if (sl) strlcpy(c->server, server, sizeof(c->server));
    if (server_changed) c->token[0] = 0; // a new server needs a new registration
    cfg_save();
    httpd_resp_set_type(req, "text/html; charset=utf-8");
    httpd_resp_sendstr(req, "<!doctype html><meta name=viewport content='width=device-width,initial-scale=1'>"
                            "<body style='font:16px system-ui;text-align:center;padding:40px'>"
                            "<h2>&#10003; Saved</h2><p>The device restarts and connects to the Wi-Fi.<br>"
                            "You can close this page.</p></body>");
    xTaskCreate(restart_task, "restart", 2048, NULL, 5, NULL);
    return ESP_OK;
}

/** Any other address (captive portal checks of phones) → the setup page. */
static esp_err_t redirect_404(httpd_req_t *req, httpd_err_code_t err)
{
    httpd_resp_set_status(req, "302 Found");
    httpd_resp_set_hdr(req, "Location", "http://192.168.4.1/");
    httpd_resp_send(req, NULL, 0);
    return ESP_OK;
}

// ---------------------------------------------------------------- DNS: every name → 192.168.4.1

static void dns_task(void *arg)
{
    int sock = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    struct sockaddr_in addr = {.sin_family = AF_INET, .sin_port = htons(53), .sin_addr.s_addr = htonl(INADDR_ANY)};
    if (sock < 0 || bind(sock, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        ESP_LOGE(TAG, "DNS socket");
        vTaskDelete(NULL);
        return;
    }
    uint8_t buf[512];
    for (;;) {
        struct sockaddr_in from;
        socklen_t fl = sizeof(from);
        int n = recvfrom(sock, buf, sizeof(buf) - 16, 0, (struct sockaddr *)&from, &fl);
        if (n < 12) continue;
        // find the end of the question (QNAME + QTYPE + QCLASS)
        int p = 12;
        while (p < n && buf[p]) p += buf[p] + 1;
        p += 5;
        if (p > n) continue;
        buf[2] = 0x81; // response, recursion desired
        buf[3] = 0x80; // recursion available, no error
        buf[6] = 0; buf[7] = 1; // 1 answer
        buf[8] = buf[9] = buf[10] = buf[11] = 0;
        uint8_t ans[] = {0xC0, 0x0C, 0, 1, 0, 1, 0, 0, 0, 60, 0, 4, 192, 168, 4, 1};
        memcpy(buf + p, ans, sizeof(ans));
        sendto(sock, buf, p + sizeof(ans), 0, (struct sockaddr *)&from, fl);
    }
}

void portal_start(void)
{
    static httpd_handle_t srv;
    if (srv) return;
    httpd_config_t hc = HTTPD_DEFAULT_CONFIG();
    hc.max_uri_handlers = 8;
    hc.lru_purge_enable = true;
    if (httpd_start(&srv, &hc) != ESP_OK) {
        ESP_LOGE(TAG, "http server");
        return;
    }
    httpd_uri_t root = {.uri = "/", .method = HTTP_GET, .handler = page_get};
    httpd_uri_t scan = {.uri = "/scan", .method = HTTP_GET, .handler = scan_get};
    httpd_uri_t save = {.uri = "/save", .method = HTTP_POST, .handler = save_post};
    httpd_register_uri_handler(srv, &root);
    httpd_register_uri_handler(srv, &scan);
    httpd_register_uri_handler(srv, &save);
    httpd_register_err_handler(srv, HTTPD_404_NOT_FOUND, redirect_404);
    xTaskCreate(dns_task, "dns", 3072, NULL, 4, NULL);
}
