// A study round on the device: flashcard / quiz / fill-in-the-blank / listening
// / PMP multiple choice, as decided by the server for each card ("mode").
#pragma once

#include <stdbool.h>
#include "lvgl.h"

/** Starts a round with the current deck. max_cards 0 = all; quick = reminder (one card, then home). */
void study_start(int max_cards, bool quick);
/** Quiz: n (max 10) random words already learned (any due date). */
void quiz_start(int n);

#include <stdint.h>
#include "api.h"
/** "Practise these now": the ticked words (card ids) of the finished round. */
int study_practice_ids(int64_t *out, int max);
/** Starts the practice with these AI questions (takes them; may be NULL / 0). */
void study_practice_begin(word_q_t *qs, int n);
bool study_active(void);
/** Ends the round (the user went home in the middle). Answers given so far are kept. */
void study_stop(void);
/** Home-screen button callback ("Study N cards"). */
void study_clicked_cb(lv_event_t *e);
