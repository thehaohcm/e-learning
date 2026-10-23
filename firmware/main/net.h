// Wi-Fi: station mode with saved credentials, or a setup access point with a
// small web page (portal.c) to choose the network and the server.
#pragma once

#include <stdbool.h>
#include "esp_err.h"

void net_init(void);
/** Connects with the saved SSID/password. Waits up to timeout_ms. */
esp_err_t net_connect(int timeout_ms);
bool net_connected(void);
int net_rssi(void);
const char *net_ip(void);

/** Starts the setup access point ("ELearn-XXXX", open) + web page at 192.168.4.1. */
void net_start_setup_ap(void);
const char *net_ap_ssid(void);
/** Scan results as JSON [{"ssid":"..","rssi":-50,"lock":true}], caller frees. */
char *net_scan_json(void);

/** Starts SNTP (time for review timestamps). */
void net_start_time(void);
