// Device storage:
//  - settings in NVS (Wi-Fi, server, device token, volume, brightness);
//  - the last synced study batch ("deck") parsed into memory and cached on
//    SPIFFS (/data/deck.json), so the device can be used offline;
//  - answers not yet sent to the server (/data/pending.bin).
#pragma once

#include <stdbool.h>
#include <stdint.h>
#include <time.h>
#include "esp_err.h"

// ---------------------------------------------------------------- settings (NVS)
typedef struct {
    char ssid[33];
    char pass[65];
    char server[128];   // https://elearn.example.com (no trailing slash)
    char token[80];     // "dev_..." from /api/device/register
    int volume;         // 0..100
    int brightness;     // 10..100
    bool speak;         // read English words aloud when a card is shown
    bool recall_first;  // multiple choice: options hidden until "Show options"
} device_cfg_t;

esp_err_t store_init(void);
device_cfg_t *cfg(void);
esp_err_t cfg_save(void);
void cfg_forget_wifi(void);
void cfg_forget_token(void);

// ---------------------------------------------------------------- study batch
typedef struct {
    int64_t id;
    char *text;
} option_t;

typedef struct {
    int64_t id;
    bool done;       // answered on the device since the last sync
    char kind[12];   // vocab | phrasal | concept | mcq
    char mode[12];   // flashcard | quiz | listening | cloze
    bool is_new;
    char *front, *ipa, *pos, *meaning_vi, *meaning_en, *example_en, *example_vi, *back, *audio_url;
    int n_choices;   // mcq: its own choices, answer = index
    char *choices[6];
    int answer;
    int n_options;   // quiz/listening: 4 meanings, the right one has id == card id
    option_t options[4];
    char *cloze_sentence, *cloze_vi, cloze_kind[8];
    int n_cloze;     // cloze: 4 words / forms, the right one has id == card id
    option_t cloze[4];
} card_t;

typedef struct {
    int refs;
    time_t synced_at;       // device clock when received
    time_t server_time;     // server clock at sync (fallback clock)
    int utc_offset_sec;     // user's time zone
    // user settings
    int daily_new, reminder_every_min, popup_cards;
    char active_from[6], active_to[6], device_mode[12];
    // counters at sync time
    int due_now, new_left_today, reviews_today, streak_days;
    int n_cards;
    card_t *cards;
    int n_quiz;      // words already learned, for the Quiz button
    card_t *quiz;
    int extra_new_left;  // "+5 new words" still allowed today
    char fw_version[16]; // newest firmware on the server ("" = none)
    char *fw_url;        // its download URL
} deck_t;

/** Parses /api/device/sync JSON into a new deck (refs = 1). NULL on error. */
deck_t *deck_parse(const char *json, size_t len);
/** Current deck (+1 ref) or NULL. Release with deck_unref. */
deck_t *deck_get(void);
void deck_unref(deck_t *d);
/** Replaces the current deck (takes the caller's reference). */
void deck_set(deck_t *d);
/** Saves / loads the raw sync JSON (offline start). */
void deck_cache_save(const char *json, size_t len);
deck_t *deck_cache_load(void);

// ---------------------------------------------------------------- answers waiting to be sent
/** Added to a rating: the answer comes from the Quiz button (the server only
 *  reschedules cards that are due). */
#define REVIEW_QUIZ 0x40

typedef struct {
    char id[28];       // client_review_id (idempotent on the server)
    int64_t card_id;
    int8_t rating;     // 1 again .. 4 easy (| REVIEW_QUIZ)
    int64_t at;        // unix seconds (UTC)
    int32_t duration_ms;
} review_t;

void reviews_add(int64_t card_id, int rating, int duration_ms);
/** Copies up to max pending reviews; returns the count. */
int reviews_peek(review_t *out, int max);
/** Removes the first n (after the server accepted them). */
void reviews_drop(int n);
int reviews_count(void);

/** Unix time, or 0 if the clock is not set yet (SNTP / server time). */
time_t now_utc(void);
void clock_set_from_server(time_t t);
