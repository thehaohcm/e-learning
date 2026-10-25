// Firmware update over Wi-Fi: the server announces its newest firmware in the
// sync answer; the user starts the update from Settings.
#pragma once

#include <stdbool.h>
#include "esp_err.h"

/** true if version "a.b.c" is newer than the running firmware. */
bool ota_is_newer(const char *version);

/** Downloads and installs the image at url (device token sent), calling
 *  progress(0..100) on the way. On success the device must restart. */
esp_err_t ota_update(const char *url, void (*progress)(int percent));
