#include "ui_study.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "app.h"
#include "audio.h"
#include "esp_lvgl_port.h"
#include "esp_random.h"
#include "esp_timer.h"
#include "net.h"
#include "store.h"
#include "ui.h"
#include "ui_theme.h"

// Ratings (FSRS)
#define AGAIN 1
#define HARD 2
#define GOOD 3
#define EASY 4

#define QUIZ_LEN 10 // questions per quiz

enum { R_NONE, R_OK, R_WRONG, R_GUESSED }; // how a card of the round went

typedef struct {
    deck_t *deck;
    int *order;          // card indexes of this round
    int n, pos;
    bool quick;          // pop-up: a few cards, then "Keep studying" / "Done"
    bool quiz;           // Quiz button: learned words, any due date
    card_t *cards;       // deck->cards, or deck->quiz for a quiz
    bool practice;       // "Practise these now": flashcards, nothing recorded
    int n_alloc;         // size of order / res / pick
    uint8_t *res;        // per position: R_OK / R_WRONG / R_GUESSED (R_NONE = not answered)
    bool *pick;          // end of round: ticked for "Practise these now"
    lv_obj_t *practise_btn;
    int pending;         // a choice was tapped: its rating, saved when the card is left
    bool recorded;       // this card's answer is saved
    lv_obj_t *recall_row; // "Show options" / "I don't know" (recall first)
    word_q_t *wq;        // practice: AI questions (order code: card | (question + 1) << 16)
    int n_wq;
    int answered, correct;
    int64_t shown_us;    // when the current card appeared (answer time)
    bool answered_this;  // quiz/cloze: first tap counts
    lv_obj_t *opts[7]; // up to 6 options + "I don't know"
    int n_opts;
    lv_obj_t *after;     // area filled after answering
} session_t;

static session_t S;
static bool s_active;

static void show_card(void);
static lv_timer_t *s_pre_timer; // loading the words of a round (see begin_round)
static void start_locked(int max_cards, bool quick);

bool study_active(void) { return s_active; }

static void record(int rating, int result);

/** Frees what the current round holds (deck reference, lists). */
static void session_free(void)
{
    if (S.deck) deck_unref(S.deck);
    free(S.order);
    free(S.res);
    free(S.pick);
    word_q_free(S.wq, S.n_wq);
    memset(&S, 0, sizeof(S));
}

/** Room for a round of n cards (and repeats in a practice). */
static bool session_alloc(int n)
{
    S.n_alloc = n > 0 ? n : 1;
    S.order = calloc(S.n_alloc, sizeof(int));
    S.res = calloc(S.n_alloc, 1);
    S.pick = calloc(S.n_alloc, sizeof(bool));
    return S.order && S.res && S.pick;
}

void study_stop(void)
{
    // a choice tapped but the card left without "Next" (home, screen off): keep the answer
    if (s_active && !S.practice && S.deck && S.pos < S.n && S.answered_this && !S.recorded)
        record(S.pending, S.pending >= GOOD ? R_OK : R_WRONG);
    s_active = false;
    audio_stop();
    if (s_pre_timer) { // left while loading the words
        lv_timer_delete(s_pre_timer);
        s_pre_timer = NULL;
        ui_busy_end_locked();
    }
    session_free();
}

// In a practice, an order entry also says which AI question goes with the card.
static inline int card_of(int code) { return code & 0xFFFF; }
static card_t *cur(void) { return &S.cards[card_of(S.order[S.pos])]; }
static word_q_t *cur_q(void)
{
    int q = (S.order[S.pos] >> 16) - 1;
    return S.practice && q >= 0 && q < S.n_wq ? &S.wq[q] : NULL;
}

static void shuffle(int *a, int n)
{
    for (int i = n - 1; i > 0; i--) {
        int j = (int)(esp_random() % (uint32_t)(i + 1));
        int t = a[i];
        a[i] = a[j];
        a[j] = t;
    }
}

static int elapsed_ms(void) { return (int)((esp_timer_get_time() - S.shown_us) / 1000); }

static void finish_round(void);

static void next_card(void)
{
    S.pos++;
    if (S.pos >= S.n) { // quick (pop-up): S.n = cards per pop-up
        finish_round();
        return;
    }
    show_card();
}

/** Saves the answer of the current card (once). result: R_OK / R_WRONG / R_GUESSED. */
static void record(int rating, int result)
{
    card_t *c = cur();
    if (S.recorded || S.practice) return;
    S.recorded = true;
    if (S.res) S.res[S.pos] = result;
    S.answered++;
    if (rating >= GOOD) S.correct++;
    if (S.quiz) {
        reviews_add(c->id, rating | REVIEW_QUIZ, elapsed_ms());
        return;
    }
    c->done = true; // not offered again until the next sync
    reviews_add(c->id, rating, elapsed_ms());
}

// ---------------------------------------------------------------- layout helpers

static lv_obj_t *card_screen(lv_obj_t **content_out)
{
    char title[40];
    card_t *c = cur();
    const char *mode = !strcmp(c->kind, "mcq") ? "Question"
                     : !strcmp(c->mode, "quiz") ? "Meaning"
                     : !strcmp(c->mode, "cloze") ? "Fill blank"
                     : !strcmp(c->mode, "listening") ? "Listen"
                     : c->is_new ? "New word" : "Review";
    if (S.practice) {
        word_q_t *q = cur_q();
        mode = !q ? "Practice" : !strcmp(q->kind, "synonym") ? "Closest meaning"
             : !strcmp(q->kind, "collocation") ? "Goes with" : "Usage";
    }
    snprintf(title, sizeof(title), "%s%d/%d · %s", S.quiz && !S.practice ? "Quiz " : "", S.pos + 1, S.n, mode);
    lv_obj_t *scr = ui_screen_with_header(title, true);
    lv_obj_t *content = lv_obj_create(scr);
    lv_obj_set_size(content, lv_pct(100), lv_display_get_vertical_resolution(NULL) - 28);
    lv_obj_align(content, LV_ALIGN_BOTTOM_MID, 0, 0);
    lv_obj_set_style_bg_opa(content, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(content, 0, 0);
    lv_obj_set_style_pad_all(content, 8, 0);
    lv_obj_set_style_pad_row(content, 6, 0);
    lv_obj_set_flex_flow(content, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(content, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_set_scroll_dir(content, LV_DIR_VER);
    *content_out = content;
    return scr;
}

static lv_obj_t *row_of(lv_obj_t *parent)
{
    lv_obj_t *r = lv_obj_create(parent);
    lv_obj_set_size(r, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(r, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(r, 0, 0);
    lv_obj_set_style_pad_all(r, 0, 0);
    lv_obj_set_style_pad_column(r, 6, 0);
    lv_obj_set_style_pad_row(r, 6, 0);
    lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW_WRAP);
    lv_obj_set_flex_align(r, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_remove_flag(r, LV_OBJ_FLAG_SCROLLABLE);
    return r;
}

/** Plays the card's English word (if it has audio). A newer word replaces it. */
static void say(card_t *c)
{
    if (!c || !c->audio_url) return;
    audio_play_url(c->audio_url);
}

/** Downloads the next card's word now, so it plays at once when it comes. */
static void prefetch_next(void)
{
    if (S.pos + 1 < S.n) {
        card_t *n = &S.cards[card_of(S.order[S.pos + 1])];
        if (n->audio_url) audio_prefetch(n->audio_url);
    }
}

static void say_clicked(lv_event_t *e) { say(cur()); }

/** The biggest font in which every single word of text fits in max_w pixels,
 *  so a long word ("comprehensive") is never broken in the middle. */
static const lv_font_t *fit_font(const char *text, int max_w)
{
    const lv_font_t *fonts[] = {font_word, font_title, font_bold, font_body};
    for (int f = 0; f < 4; f++) {
        bool fits = true;
        const char *p = text;
        while (*p && fits) {
            const char *e = strchr(p, ' ');
            int len = e ? (int)(e - p) : (int)strlen(p);
            char w[64];
            if (len >= (int)sizeof(w)) len = sizeof(w) - 1;
            memcpy(w, p, len);
            w[len] = 0;
            lv_point_t sz;
            lv_text_get_size(&sz, w, fonts[f], 0, 0, LV_COORD_MAX, LV_TEXT_FLAG_NONE);
            fits = sz.x <= max_w;
            p = e ? e + 1 : p + len;
        }
        if (fits) return fonts[f];
    }
    return font_small;
}

static void word_header(lv_obj_t *p, card_t *c)
{
    lv_obj_t *w;
    int full = lv_display_get_horizontal_resolution(NULL) - 16; // card content width
    if (c->audio_url) {
        // the word with a speaker button on its right
        lv_obj_t *r = lv_obj_create(p);
        lv_obj_set_size(r, lv_pct(100), LV_SIZE_CONTENT);
        lv_obj_set_style_bg_opa(r, LV_OPA_TRANSP, 0);
        lv_obj_set_style_border_width(r, 0, 0);
        lv_obj_set_style_pad_all(r, 0, 0);
        lv_obj_set_style_pad_column(r, 6, 0);
        lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
        lv_obj_set_flex_align(r, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
        lv_obj_remove_flag(r, LV_OBJ_FLAG_SCROLLABLE);
        int max_w = full - 40 - 6 - 4; // minus the speaker button
        w = ui_label(r, c->front, fit_font(c->front, max_w), COL_TEXT, LV_SIZE_CONTENT);
        lv_obj_set_style_max_width(w, max_w, 0);
        lv_obj_t *b = ui_button(r, LV_SYMBOL_VOLUME_MAX, COL_INFO, font_body);
        lv_obj_set_size(b, 40, 36);
        lv_obj_set_style_pad_all(b, 0, 0);
        lv_obj_add_event_cb(b, say_clicked, LV_EVENT_CLICKED, NULL);
    } else {
        w = ui_label(p, c->front, fit_font(c->front, full - 4), COL_TEXT, -1);
    }
    lv_obj_set_style_text_align(w, LV_TEXT_ALIGN_CENTER, 0);
    if (c->ipa || c->pos) {
        char line[96];
        snprintf(line, sizeof(line), "%s%s%s", c->ipa ? c->ipa : "", c->ipa && c->pos ? "  ·  " : "", c->pos ? c->pos : "");
        lv_obj_t *l = ui_label(p, line, font_small, COL_MUTED, -1);
        lv_obj_set_style_text_align(l, LV_TEXT_ALIGN_CENTER, 0);
    }
}

/** Meaning, example and translation, in a white box. */
static void answer_box(lv_obj_t *p, card_t *c)
{
    lv_obj_t *box = lv_obj_create(p);
    lv_obj_set_size(box, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_color(box, COL_CARD, 0);
    lv_obj_set_style_border_width(box, 0, 0);
    lv_obj_set_style_radius(box, 10, 0);
    lv_obj_set_style_pad_all(box, 8, 0);
    lv_obj_set_style_pad_row(box, 4, 0);
    lv_obj_set_flex_flow(box, LV_FLEX_FLOW_COLUMN);
    lv_obj_remove_flag(box, LV_OBJ_FLAG_SCROLLABLE);
    if (!strcmp(c->kind, "concept") && c->back) {
        // term / writing frame: the answer first, its note under it
        ui_label(box, c->back, font_bold, COL_PRIMARY, -1);
        if (c->meaning_vi) ui_label(box, c->meaning_vi, font_small, COL_MUTED, -1);
    } else {
        if (c->meaning_vi) ui_label(box, c->meaning_vi, font_bold, COL_PRIMARY, -1);
        if (c->meaning_en) ui_label(box, c->meaning_en, font_small, COL_MUTED, -1);
        if (c->back) ui_label(box, c->back, font_body, COL_TEXT, -1);
    }
    if (c->example_en) ui_label(box, c->example_en, font_body, COL_TEXT, -1);
    if (c->example_vi) ui_label(box, c->example_vi, font_small, COL_MUTED, -1);
}

// ---------------------------------------------------------------- flashcard

static void rate_clicked(lv_event_t *e)
{
    int rating = (int)(intptr_t)lv_event_get_user_data(e);
    audio_fx(rating == AGAIN ? AUDIO_FX_WRONG : AUDIO_FX_TAP);
    record(rating, rating == AGAIN ? R_WRONG : R_OK);
    next_card();
}

static void rating_buttons(lv_obj_t *p)
{
    lv_obj_t *r = row_of(p);
    lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
    static const struct { const char *t; uint32_t col; int rating; } B[] = {
        {"Again", 0xDC2626, AGAIN}, {"Hard", 0xEA580C, HARD}, {"Good", 0x16A34A, GOOD}, {"Easy", 0x2563EB, EASY}};
    for (int i = 0; i < 4; i++) {
        lv_obj_t *b = ui_button(r, B[i].t, lv_color_hex(B[i].col), font_bold);
        lv_obj_set_flex_grow(b, 1);
        lv_obj_set_height(b, 44);
        lv_obj_add_event_cb(b, rate_clicked, LV_EVENT_CLICKED, (void *)(intptr_t)B[i].rating);
    }
}

static void show_clicked(lv_event_t *e)
{
    lv_obj_t *btn = lv_event_get_target(e);
    lv_obj_t *p = lv_obj_get_parent(btn);
    audio_fx(AUDIO_FX_TAP);
    lv_obj_delete(btn);
    answer_box(p, cur());
    rating_buttons(p);
    lv_obj_scroll_to_y(p, 0, LV_ANIM_OFF);
}

static void build_flashcard(lv_obj_t *p, card_t *c)
{
    word_header(p, c);
    if (c->is_new) {
        // a new word: learn it first (meaning shown), then rate how well it is known
        answer_box(p, c);
        rating_buttons(p);
        return;
    }
    lv_obj_t *b = ui_button(p, LV_SYMBOL_EYE_OPEN "  Show the answer", COL_PRIMARY, font_bold);
    lv_obj_set_width(b, lv_pct(100));
    lv_obj_set_height(b, 50);
    lv_obj_add_event_cb(b, show_clicked, LV_EVENT_CLICKED, NULL);
}

// ---------------------------------------------------------------- choices (quiz, listening, cloze, mcq)

static void next_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    if (S.answered_this) record(S.pending, S.pending >= GOOD ? R_OK : R_WRONG);
    next_card();
}

/** Right answer, but only found by looking at the options: counted as forgotten. */
static void guessed_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    record(AGAIN, R_GUESSED);
    next_card();
}

static void reveal(bool ok, int right_index)
{
    for (int i = 0; i < S.n_opts; i++) {
        lv_obj_t *b = S.opts[i];
        lv_obj_remove_flag(b, LV_OBJ_FLAG_CLICKABLE);
        if (i == right_index) lv_obj_set_style_bg_color(b, COL_OK, 0);
    }
    card_t *c = cur();
    lv_obj_t *p = S.after;
    word_q_t *q = cur_q();
    bool cloze = !q && !strcmp(c->mode, "cloze");
    if (cloze && cfg()->speak) say(c); // the word was hidden until now
    if (q) {
        if (q->explanation && *q->explanation) ui_label(p, q->explanation, font_small, COL_MUTED, -1);
        answer_box(p, c);
    } else if (cloze && c->cloze_vi) {
        ui_label(p, c->cloze_vi, font_small, COL_MUTED, -1);
    } else if (strcmp(c->kind, "mcq")) {
        answer_box(p, c);
    }
    lv_obj_t *r = row_of(p);
    lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
    if (ok && !S.practice) {
        // found it only thanks to the options? then it should come back soon
        lv_obj_t *g = ui_button(r, "Guessed", COL_WARN, font_bold);
        lv_obj_set_flex_grow(g, 1);
        lv_obj_set_height(g, 44);
        lv_obj_add_event_cb(g, guessed_clicked, LV_EVENT_CLICKED, NULL);
    }
    lv_obj_t *n = ui_button(r, S.pos + 1 < S.n ? "Next " LV_SYMBOL_RIGHT : "Done " LV_SYMBOL_OK,
                            ok ? COL_OK : COL_PRIMARY, font_bold);
    lv_obj_set_flex_grow(n, 1);
    lv_obj_set_height(n, 44);
    lv_obj_add_event_cb(n, next_clicked, LV_EVENT_CLICKED, NULL);
    // just enough to show "Next" at the bottom: the content above stays in view
    lv_obj_update_layout(lv_obj_get_screen(r));
    lv_obj_scroll_to_view_recursive(r, LV_ANIM_ON);
}

typedef struct {
    int index;
    bool right;
} opt_data_t;

static opt_data_t s_opt_data[6];
static int s_right_index;

static void option_clicked(lv_event_t *e)
{
    opt_data_t *o = lv_event_get_user_data(e);
    lv_obj_t *b = lv_event_get_target(e);
    if (S.answered_this) return;
    S.answered_this = true;
    if (S.practice) {
        // a wrong answer here: not really known, so the word becomes due (quiz review, Again)
        if (!o->right) reviews_add(cur()->id, AGAIN | REVIEW_QUIZ, elapsed_ms());
        audio_fx(o->right ? AUDIO_FX_CORRECT : AUDIO_FX_WRONG);
        if (!o->right) lv_obj_set_style_bg_color(b, COL_BAD, 0);
        reveal(o->right, s_right_index);
        return;
    }
    // saved when the card is left ("Next" or "I guessed")
    if (o->right) {
        audio_fx(AUDIO_FX_CORRECT);
        S.pending = GOOD;
    } else {
        audio_fx(AUDIO_FX_WRONG);
        lv_obj_set_style_bg_color(b, COL_BAD, 0);
        S.pending = AGAIN;
    }
    reveal(o->right, s_right_index);
}

static void unhide_options(void)
{
    for (int i = 0; i < S.n_opts; i++) lv_obj_remove_flag(S.opts[i], LV_OBJ_FLAG_HIDDEN);
    if (S.recall_row) {
        // hidden at once (no more room taken, so scrolling to "Next" is right),
        // deleted later: its button is still handling this tap
        lv_obj_add_flag(S.recall_row, LV_OBJ_FLAG_HIDDEN);
        lv_obj_delete_async(S.recall_row);
        S.recall_row = NULL;
    }
}

static void show_options_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    unhide_options();
}

/** Could not remember: the answer is shown, counted as forgotten. */
static void dont_know_clicked(lv_event_t *e)
{
    if (S.answered_this) return;
    audio_fx(AUDIO_FX_WRONG);
    unhide_options();
    S.answered_this = true;
    S.pending = AGAIN;
    // practice: not knowing the AI question = not really known (as a wrong answer)
    if (S.practice && cur_q()) reviews_add(cur()->id, AGAIN | REVIEW_QUIZ, elapsed_ms());
    reveal(false, s_right_index);
}

/** Option buttons; right = index of the right one. */
static void option_list(lv_obj_t *p, const char *const *texts, int n, int right)
{
    S.n_opts = 0;
    s_right_index = right;
    for (int i = 0; i < n && i < 6; i++) {
        lv_obj_t *b = lv_button_create(p);
        lv_obj_set_width(b, lv_pct(100));
        lv_obj_set_style_bg_color(b, COL_CARD, 0);
        lv_obj_set_style_border_color(b, COL_SOFT, 0);
        lv_obj_set_style_border_width(b, 2, 0);
        lv_obj_set_style_radius(b, 10, 0);
        lv_obj_set_style_shadow_width(b, 0, 0);
        lv_obj_set_style_pad_all(b, 8, 0);
        lv_obj_t *l = ui_label(b, NULL, font_body, COL_TEXT, -1);
        char t[300];
        snprintf(t, sizeof(t), "%c.  %s", 'A' + i, texts[i]);
        lv_label_set_text(l, t);
        s_opt_data[i] = (opt_data_t){.index = i, .right = i == right};
        lv_obj_add_event_cb(b, option_clicked, LV_EVENT_CLICKED, &s_opt_data[i]);
        S.opts[S.n_opts++] = b;
    }
    // "I don't know": honest instead of a lucky guess (counted as forgotten, the answer is shown)
    {
        lv_obj_t *b = lv_button_create(p);
        lv_obj_set_width(b, lv_pct(100));
        lv_obj_set_style_bg_color(b, COL_BG, 0);
        lv_obj_set_style_border_color(b, COL_SOFT, 0);
        lv_obj_set_style_border_width(b, 2, 0);
        lv_obj_set_style_border_opa(b, LV_OPA_COVER, 0);
        lv_obj_set_style_radius(b, 10, 0);
        lv_obj_set_style_shadow_width(b, 0, 0);
        lv_obj_set_style_pad_all(b, 6, 0);
        lv_obj_t *l = ui_label(b, "?  I don't know", font_body, COL_MUTED, LV_SIZE_CONTENT);
        lv_obj_center(l);
        lv_obj_add_event_cb(b, dont_know_clicked, LV_EVENT_CLICKED, NULL);
        S.opts[S.n_opts++] = b;
    }
    S.recall_row = NULL;
    // only for "what does it mean?" (quiz, listening): a sentence with a blank or
    // a PMP question cannot be answered without seeing the options
    card_t *c = cur();
    bool meaning = strcmp(c->kind, "mcq") && (!strcmp(c->mode, "quiz") || !strcmp(c->mode, "listening"));
    if (cfg()->recall_first && meaning && !S.practice) {
        // recall first: the options stay hidden until the learner has tried to remember
        for (int i = 0; i < S.n_opts; i++) lv_obj_add_flag(S.opts[i], LV_OBJ_FLAG_HIDDEN);
        S.recall_row = row_of(p);
        lv_obj_set_flex_flow(S.recall_row, LV_FLEX_FLOW_ROW);
        lv_obj_t *dk = ui_button(S.recall_row, "Don't know", COL_MUTED, font_bold);
        lv_obj_set_flex_grow(dk, 1);
        lv_obj_set_height(dk, 48);
        lv_obj_add_event_cb(dk, dont_know_clicked, LV_EVENT_CLICKED, NULL);
        lv_obj_t *so = ui_button(S.recall_row, LV_SYMBOL_EYE_OPEN " Options", COL_PRIMARY, font_bold);
        lv_obj_set_flex_grow(so, 1);
        lv_obj_set_height(so, 48);
        lv_obj_add_event_cb(so, show_options_clicked, LV_EVENT_CLICKED, NULL);
    }
    S.after = lv_obj_create(p);
    lv_obj_set_size(S.after, lv_pct(100), LV_SIZE_CONTENT);
    lv_obj_set_style_bg_opa(S.after, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(S.after, 0, 0);
    lv_obj_set_style_pad_all(S.after, 0, 0);
    lv_obj_set_style_pad_row(S.after, 6, 0);
    lv_obj_set_flex_flow(S.after, LV_FLEX_FLOW_COLUMN);
    lv_obj_remove_flag(S.after, LV_OBJ_FLAG_SCROLLABLE);
}

static void replay_clicked(lv_event_t *e)
{
    audio_play_url(cur()->audio_url);
}

static void build_choices(lv_obj_t *p, card_t *c)
{
    const char *texts[6];
    int n = 0, right = -1;
    if (!strcmp(c->kind, "mcq")) {
        lv_obj_t *q = ui_label(p, c->front, font_bold, COL_TEXT, -1);
        (void)q;
        for (int i = 0; i < c->n_choices; i++) texts[n++] = c->choices[i];
        right = c->answer;
    } else if (!strcmp(c->mode, "cloze")) {
        lv_obj_t *s = ui_label(p, c->cloze_sentence, font_body, COL_TEXT, -1);
        lv_obj_set_style_text_align(s, LV_TEXT_ALIGN_CENTER, 0);
        ui_label(p, !strcmp(c->cloze_kind, "form") ? "Which form fits?" : "Which word fits?", font_small, COL_MUTED,
                 LV_SIZE_CONTENT);
        for (int i = 0; i < c->n_cloze; i++) {
            texts[n] = c->cloze[i].text;
            if (c->cloze[i].id == c->id) right = n;
            n++;
        }
    } else {
        if (!strcmp(c->mode, "listening") && c->audio_url) {
            lv_obj_t *pl = ui_button(p, LV_SYMBOL_VOLUME_MAX "  Play again", COL_INFO, font_bold);
            lv_obj_add_event_cb(pl, replay_clicked, LV_EVENT_CLICKED, NULL);
            ui_label(p, "Listen and choose the meaning", font_small, COL_MUTED, LV_SIZE_CONTENT);
            say(c);
        } else {
            word_header(p, c);
            ui_label(p, "Choose the meaning", font_small, COL_MUTED, LV_SIZE_CONTENT);
        }
        for (int i = 0; i < c->n_options; i++) {
            texts[n] = c->options[i].text;
            if (c->options[i].id == c->id) right = n;
            n++;
        }
    }
    option_list(p, texts, n, right);
}

// ---------------------------------------------------------------- practice ("Practise these now")
// The words ticked at the end of a round, as flashcards: the word, try to
// remember, see the answer. Nothing is sent: these answers were saved already.

static void practice_next(lv_event_t *e)
{
    bool again = (bool)(intptr_t)lv_event_get_user_data(e);
    audio_fx(AUDIO_FX_TAP);
    if (again && S.n < S.n_alloc) S.order[S.n++] = S.order[S.pos]; // once more at the end
    next_card();
}

static void practice_show(lv_event_t *e)
{
    lv_obj_t *btn = lv_event_get_target(e);
    lv_obj_t *p = lv_obj_get_parent(btn);
    audio_fx(AUDIO_FX_TAP);
    lv_obj_delete(btn);
    answer_box(p, cur());
    lv_obj_t *r = row_of(p);
    lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
    lv_obj_t *a = ui_button(r, LV_SYMBOL_REFRESH "  Again", COL_WARN, font_bold);
    lv_obj_set_flex_grow(a, 1);
    lv_obj_set_height(a, 44);
    lv_obj_add_event_cb(a, practice_next, LV_EVENT_CLICKED, (void *)(intptr_t) true);
    lv_obj_t *g = ui_button(r, LV_SYMBOL_OK "  Got it", COL_OK, font_bold);
    lv_obj_set_flex_grow(g, 1);
    lv_obj_set_height(g, 44);
    lv_obj_add_event_cb(g, practice_next, LV_EVENT_CLICKED, (void *)(intptr_t) false);
    lv_obj_scroll_to_view_recursive(r, LV_ANIM_ON);
}

/** An AI question about the word (usage, closest meaning, collocation). */
static void build_word_question(lv_obj_t *p, word_q_t *q)
{
    ui_label(p, q->prompt, font_body, COL_TEXT, -1);
    const char *texts[4] = {q->choices[0], q->choices[1], q->choices[2], q->choices[3]};
    option_list(p, texts, 4, q->answer);
}

static void build_practice(lv_obj_t *p, card_t *c)
{
    word_header(p, c);
    lv_obj_t *b = ui_button(p, LV_SYMBOL_EYE_OPEN "  Show the answer", COL_PRIMARY, font_bold);
    lv_obj_set_width(b, lv_pct(100));
    lv_obj_set_height(b, 50);
    lv_obj_add_event_cb(b, practice_show, LV_EVENT_CLICKED, NULL);
}

// ---------------------------------------------------------------- round

static void show_card(void)
{
    card_t *c = cur();
    // listening needs its sound: without it the card is asked with the word shown
    if (!strcmp(c->mode, "listening") && !(c->audio_url && audio_ready(c->audio_url))) strcpy(c->mode, "quiz");
    lv_obj_t *content;
    lv_obj_t *scr = card_screen(&content);
    S.shown_us = esp_timer_get_time();
    S.answered_this = false;
    S.recorded = false;
    S.pending = 0;
    S.n_opts = 0;
    bool choices = !strcmp(c->kind, "mcq") || !strcmp(c->mode, "quiz") || !strcmp(c->mode, "cloze") ||
                   !strcmp(c->mode, "listening");
    if (S.practice && cur_q()) build_word_question(content, cur_q());
    else if (S.practice) build_practice(content, c);
    else if (choices && !c->is_new) build_choices(content, c);
    else build_flashcard(content, c);
    lv_screen_load_anim(scr, LV_SCR_LOAD_ANIM_MOVE_LEFT, 180, 0, true);
    // read the word aloud when it is on screen (cloze: after answering; listening plays it anyway)
    bool shows_word = strcmp(c->kind, "mcq") && !cur_q() &&
                      (S.practice || c->is_new || (strcmp(c->mode, "cloze") && strcmp(c->mode, "listening")));
    if (cfg()->speak && shows_word) say(c);
    prefetch_next();
}

static void keep_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    start_locked(0, false); // all the cards left
}

/** Cards of the current batch not answered yet. */
static int cards_left(void)
{
    int n = 0;
    for (int i = 0; S.deck && i < S.deck->n_cards; i++) n += !S.deck->cards[i].done;
    return n;
}

static void quiz_again_clicked(lv_event_t *e);

static void summary_home_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    ui_go_home_locked();
}

// ---------------------------------------------------------------- loading the words of a round
// Before the first card, the words of the round are downloaded (grey veil,
// "Loading sounds 3/10"), at most PRELOAD_MS: then every card speaks at once.
// What is not ready by then keeps loading in the background.

#define PRELOAD_MAX 20
#define PRELOAD_MS 8000

static int s_pre_total;
static int64_t s_pre_start;

static void preload_finish(void)
{
    if (s_pre_timer) {
        lv_timer_delete(s_pre_timer);
        s_pre_timer = NULL;
    }
    ui_busy_end_locked();
    if (s_active) show_card();
}

static void preload_tick(lv_timer_t *t)
{
    int done = audio_preload_done();
    char m[48];
    snprintf(m, sizeof(m), "Loading sounds %d/%d…", done < s_pre_total ? done : s_pre_total, s_pre_total);
    ui_busy_text_locked(m);
    if (done >= s_pre_total || esp_timer_get_time() - s_pre_start > PRELOAD_MS * 1000LL) preload_finish();
}

/** Starts the round prepared in S (lock held): loads its words, then the first card. */
static void begin_round(void)
{
    const char *urls[PRELOAD_MAX];
    int n = 0;
    for (int i = 0; i < S.n && n < PRELOAD_MAX; i++) {
        card_t *c = &S.cards[S.order[i]];
        bool needed = !strcmp(c->mode, "listening") || cfg()->speak;
        if (needed && c->audio_url && !audio_ready(c->audio_url)) urls[n++] = c->audio_url;
    }
    s_pre_total = n && net_connected() ? audio_preload(urls, n) : 0;
    if (s_pre_total == 0) {
        show_card();
        return;
    }
    s_pre_start = esp_timer_get_time();
    char m[48];
    snprintf(m, sizeof(m), "Loading sounds 0/%d…", s_pre_total);
    ui_busy_locked(m);
    s_pre_timer = lv_timer_create(preload_tick, 200, NULL);
}
static void extra_new_clicked(lv_event_t *e);

static int picked_count(void)
{
    int n = 0;
    for (int i = 0; i < S.n; i++) n += S.pick[i];
    return n;
}

static void update_practise_btn(void)
{
    if (!S.practise_btn) return;
    int n = picked_count();
    char t[48];
    snprintf(t, sizeof(t), LV_SYMBOL_LOOP "  Practise these now (%d)", n);
    lv_label_set_text(lv_obj_get_child(S.practise_btn, 0), t);
    if (n) lv_obj_remove_state(S.practise_btn, LV_STATE_DISABLED);
    else lv_obj_add_state(S.practise_btn, LV_STATE_DISABLED);
}

/** A tap anywhere on a word's row ticks / unticks it. */
static void pick_row_clicked(lv_event_t *e)
{
    int i = (int)(intptr_t)lv_event_get_user_data(e);
    lv_obj_t *row = lv_event_get_current_target(e);
    lv_obj_t *cb = lv_obj_get_child(row, 0);
    S.pick[i] = !S.pick[i];
    if (S.pick[i]) lv_obj_add_state(cb, LV_STATE_CHECKED);
    else lv_obj_remove_state(cb, LV_STATE_CHECKED);
    update_practise_btn();
}

/** Starts the practice of the ticked words: for each, a flashcard then (if the
 *  server made one) an AI question. Lock held. */
static void practice_start(word_q_t *qs, int nq)
{
    int words[64], nw = 0;
    for (int i = 0; i < S.n && nw < 64; i++) {
        if (S.pick[i]) words[nw++] = S.order[i];
    }
    if (!nw) {
        word_q_free(qs, nq);
        return;
    }
    shuffle(words, nw);
    int cap = nw * 2 * 3; // flashcard + question, room for "Again" repeats
    int *order = calloc(cap, sizeof(int));
    if (!order) {
        word_q_free(qs, nq);
        return;
    }
    int n = 0;
    for (int w = 0; w < nw; w++) {
        order[n++] = words[w];
        for (int k = 0; k < nq; k++) {
            if (qs[k].card_id == S.cards[words[w]].id) {
                order[n++] = words[w] | ((k + 1) << 16);
                break;
            }
        }
    }
    free(S.order);
    S.order = order;
    S.n_alloc = cap;
    S.n = n;
    S.pos = 0;
    S.practice = true;
    S.practise_btn = NULL;
    word_q_free(S.wq, S.n_wq);
    S.wq = qs;
    S.n_wq = nq;
    s_active = true;
    show_card();
}

static void practise_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    if (!picked_count()) return;
    if (!net_connected()) {
        practice_start(NULL, 0); // offline: flashcards only
        return;
    }
    s_active = true; // no pop-up / batch change while waiting
    ui_busy_locked("Preparing questions…");
    app_request(APP_REQ_PRACTICE);
}

int study_practice_ids(int64_t *out, int max)
{
    int n = 0;
    if (lvgl_port_lock(0)) {
        for (int i = 0; S.pick && i < S.n && n < max; i++) {
            if (S.pick[i]) out[n++] = S.cards[S.order[i]].id;
        }
        lvgl_port_unlock();
    }
    return n;
}

void study_practice_begin(word_q_t *qs, int n)
{
    if (lvgl_port_lock(0)) {
        ui_busy_end_locked();
        if (S.deck && S.pick && !S.practice) practice_start(qs, n);
        else word_q_free(qs, n); // left meanwhile
        lvgl_port_unlock();
    }
}

/** End of a round (study, pop-up or quiz): score, every word with how it went;
 *  the ones missed or guessed are ticked for "Practise these now". */
static void round_summary(int left)
{
    lv_obj_t *scr = ui_screen_with_header(S.quiz ? "Quiz finished" : "Round finished", true);
    lv_obj_t *b = lv_obj_create(scr);
    lv_obj_set_size(b, lv_pct(100), lv_display_get_vertical_resolution(NULL) - 28);
    lv_obj_align(b, LV_ALIGN_BOTTOM_MID, 0, 0);
    lv_obj_set_style_bg_opa(b, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(b, 0, 0);
    lv_obj_set_style_pad_all(b, 8, 0);
    lv_obj_set_style_pad_row(b, 6, 0);
    lv_obj_set_flex_flow(b, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_set_scroll_dir(b, LV_DIR_VER);

    char m[300];
    snprintf(m, sizeof(m), "%d / %d remembered", S.correct, S.answered);
    ui_label(b, m, font_title, S.correct * 10 >= S.answered * 8 ? COL_OK : COL_PRIMARY, LV_SIZE_CONTENT);
    int missed = 0;
    for (int i = 0; i < S.n; i++) {
        S.pick[i] = S.res[i] == R_WRONG || S.res[i] == R_GUESSED;
        missed += S.pick[i];
    }
    if (S.answered) {
        ui_label(b, missed ? "Ticked words can be practised now (tick any other you are not sure of):"
                           : "Tick the words you are not sure of to practise them now:",
                 font_small, COL_MUTED, -1);
    }
    for (int i = 0; i < S.n; i++) {
        if (S.res[i] == R_NONE) continue;
        card_t *c = &S.cards[S.order[i]];
        const char *mark = S.res[i] == R_OK ? LV_SYMBOL_OK : S.res[i] == R_WRONG ? LV_SYMBOL_CLOSE : "?";
        const char *mean = c->meaning_vi ? c->meaning_vi : (c->meaning_en ? c->meaning_en : "");
        snprintf(m, sizeof(m), "%s %s  —  %s", mark, c->front, mean);
        // [box] word — meaning (long meanings wrap); the whole row is the tap target
        lv_obj_t *row = row_of(b);
        lv_obj_set_flex_flow(row, LV_FLEX_FLOW_ROW);
        lv_obj_set_flex_align(row, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_START, LV_FLEX_ALIGN_START);
        lv_obj_add_flag(row, LV_OBJ_FLAG_CLICKABLE);
        lv_obj_t *cb = lv_checkbox_create(row);
        lv_checkbox_set_text(cb, "");
        lv_obj_remove_flag(cb, LV_OBJ_FLAG_CLICKABLE); // the row handles the tap
        if (S.pick[i]) lv_obj_add_state(cb, LV_STATE_CHECKED);
        lv_obj_t *l = ui_label(row, m, font_small, S.res[i] == R_OK ? COL_TEXT : (S.res[i] == R_WRONG ? COL_BAD : COL_WARN),
                               LV_SIZE_CONTENT);
        lv_obj_set_flex_grow(l, 1);
        lv_obj_set_width(l, 1); // let flex give it the rest of the row, so it wraps
        lv_obj_add_event_cb(row, pick_row_clicked, LV_EVENT_CLICKED, (void *)(intptr_t)i);
    }
    if (S.answered) {
        S.practise_btn = ui_button(b, "", COL_WARN, font_bold);
        lv_obj_set_width(S.practise_btn, lv_pct(100));
        lv_obj_add_event_cb(S.practise_btn, practise_clicked, LV_EVENT_CLICKED, NULL);
        update_practise_btn();
    }

    if (S.quiz) { // [Quiz again][Home]: a big Home too (the one at the top left is small)
        lv_obj_t *r = row_of(b);
        lv_obj_set_flex_flow(r, LV_FLEX_FLOW_ROW);
        lv_obj_t *again = ui_button(r, LV_SYMBOL_REFRESH "  Quiz again", COL_PRIMARY, font_bold);
        lv_obj_set_flex_grow(again, 2);
        lv_obj_add_event_cb(again, quiz_again_clicked, LV_EVENT_CLICKED, NULL);
        lv_obj_t *home = ui_button(r, LV_SYMBOL_HOME "  Home", COL_MUTED, font_bold);
        lv_obj_set_flex_grow(home, 1);
        lv_obj_add_event_cb(home, summary_home_clicked, LV_EVENT_CLICKED, NULL);
    } else if (S.quick && left > 0) {
        snprintf(m, sizeof(m), LV_SYMBOL_PLAY "  Keep studying (%d)", left);
        lv_obj_t *k = ui_button(b, m, COL_PRIMARY, font_bold);
        lv_obj_set_width(k, lv_pct(100));
        lv_obj_add_event_cb(k, keep_clicked, LV_EVENT_CLICKED, NULL);
    }

    int study_left = 0;
    for (int i = 0; i < S.deck->n_cards; i++) study_left += !S.deck->cards[i].done;
    if (study_left == 0 && S.deck->extra_new_left > 0) {
        lv_obj_t *x = ui_button(b, LV_SYMBOL_PLUS "  +5 new words today", COL_INFO, font_bold);
        lv_obj_set_width(x, lv_pct(100));
        lv_obj_add_event_cb(x, extra_new_clicked, LV_EVENT_CLICKED, NULL);
    }
    if (!S.quiz) { // study / pop-up summary: a big Home at the bottom (not only the small one on top)
        lv_obj_t *home = ui_button(b, LV_SYMBOL_HOME "  Home", COL_MUTED, font_bold);
        lv_obj_set_width(home, lv_pct(100));
        lv_obj_add_event_cb(home, summary_home_clicked, LV_EVENT_CLICKED, NULL);
    }
    lv_screen_load_anim(scr, LV_SCR_LOAD_ANIM_FADE_IN, 150, 0, true);
    audio_fx(AUDIO_FX_CORRECT);
}

static void finish_round(void)
{
    s_active = false;
    if (S.practice) { // practice done
        audio_fx(AUDIO_FX_CORRECT);
        ui_go_home_locked();
        return;
    }
    // the round keeps its deck reference until Home / the next round (the list uses it)
    app_request(S.quiz ? APP_REQ_PUSH : APP_REQ_SYNC); // send the answers (+ next batch)
    round_summary(S.quiz ? 0 : cards_left());
}

/** Called with the LVGL lock held (from a button) or takes it (from the app task). */
static void start_locked(int max_cards, bool quick)
{
    session_free();
    S.deck = deck_get();
    if (!S.deck || !S.deck->n_cards || !session_alloc(S.deck->n_cards)) {
        session_free();
        ui_go_home_locked();
        return;
    }
    S.quick = quick;
    S.cards = S.deck->cards;
    // due reviews first, then new words; each group in a random order (a pop-up
    // mixes them all), so the same words do not always come first
    int n_due = 0;
    for (int pass = 0; pass < 2; pass++) {
        for (int i = 0; i < S.deck->n_cards; i++) {
            if (!S.deck->cards[i].done && S.deck->cards[i].is_new == (pass == 1)) S.order[S.n++] = i;
        }
        if (pass == 0) n_due = S.n;
    }
    if (quick) {
        shuffle(S.order, S.n);
    } else {
        shuffle(S.order, n_due);
        shuffle(S.order + n_due, S.n - n_due);
    }
    if (max_cards > 0 && S.n > max_cards) S.n = max_cards;
    if (S.n == 0) {
        session_free();
        ui_go_home_locked();
        return;
    }
    s_active = true;
    begin_round();
}

/** A quiz: QUIZ_LEN random words already learned (from the last sync). */
static void quiz_start_locked(int len)
{
    session_free();
    S.deck = deck_get();
    if (!S.deck || S.deck->n_quiz == 0 || !session_alloc(S.deck->n_quiz)) {
        session_free();
        ui_go_home_locked();
        return;
    }
    S.quiz = true;
    S.cards = S.deck->quiz;
    int n = S.deck->n_quiz;
    for (int i = 0; i < n; i++) S.order[i] = i;
    shuffle(S.order, n);
    if (len <= 0 || len > QUIZ_LEN) len = QUIZ_LEN;
    S.n = n < len ? n : len;
    s_active = true;
    begin_round();
}

static void quiz_again_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    quiz_start_locked(QUIZ_LEN);
}

static void extra_new_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    ui_go_home_locked();
    ui_busy("Getting 5 new words…"); // the lock is recursive
    app_request(APP_REQ_EXTRA_NEW);
}

void quiz_start(int n)
{
    if (lvgl_port_lock(0)) {
        quiz_start_locked(n);
        lvgl_port_unlock();
    }
}

void study_start(int max_cards, bool quick)
{
    // from a button the lock is already held by the LVGL task; from other tasks take it
    if (lvgl_port_lock(0)) {
        start_locked(max_cards, quick);
        lvgl_port_unlock();
    }
}
