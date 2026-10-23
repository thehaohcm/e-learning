// A study round on the device: flashcard / quiz / fill-in-the-blank / listening
// / PMP multiple choice, as decided by the server for each card ("mode").
#pragma once

#include <stdbool.h>
#include "lvgl.h"

/** Starts a round with the current deck. max_cards 0 = all; quick = reminder (one card, then home). */
void study_start(int max_cards, bool quick);
bool study_active(void);
/** Home-screen button callback ("Study N cards"). */
void study_clicked_cb(lv_event_t *e);
