// HTTPS client for the E-Learning server's device API:
//   POST /api/device/register   → token + 6-digit pairing code
//   GET  /api/device/status     → linked to a user yet?
//   GET  /api/device/sync       → settings, counters, study batch
//   POST /api/device/reviews    → answers (idempotent, offline-safe)
#pragma once

#include <stdint.h>

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
/** An AI question about a word ("Practise these now"): 4 choices, answer = index. */
typedef struct {
    int64_t card_id;
    char kind[16]; // usage | synonym | collocation
    char *prompt;
    char *choices[4];
    int answer;
    char *explanation;
} word_q_t;

/** One question per word (made by AI on the server if needed). Returns the
 *  count (*out to free with word_q_free), or -1 on error. */
int api_practice(const int64_t *card_ids, int n, word_q_t **out);
void word_q_free(word_q_t *q, int n);

/** Asks the server to prepare "which form?" questions (AI, in the background). */
void api_prepare_forms(void);

/** "+5 new words": allows n more new cards today (sync afterwards).
 *  Returns the HTTP status (200 = done), or -1 without network. */
int api_extra_new(int n);
