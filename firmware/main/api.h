// HTTPS client for the E-Learning server's device API:
//   POST /api/device/register   → token + 6-digit pairing code
//   GET  /api/device/status     → linked to a user yet?
//   GET  /api/device/sync       → settings, counters, study batch
//   POST /api/device/reviews    → answers (idempotent, offline-safe)
#pragma once

#include <stdbool.h>
#include "esp_err.h"

typedef struct {
    bool paired;
    char pair_code[8];
    char user_name[48];
} api_status_t;

/** Registers the device when it has no token yet (stores the token). */
esp_err_t api_register(void);
/** ESP_ERR_INVALID_STATE: the server no longer knows this token (device removed). */
esp_err_t api_status(api_status_t *out);
/** Downloads the study batch and makes it the current deck (also cached for offline use). */
esp_err_t api_sync(void);
/** Sends the pending answers. Returns how many were sent (0 if none / offline). */
int api_push_reviews(void);
