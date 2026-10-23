#include "ui_study.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "app.h"
#include "audio.h"
#include "esp_lvgl_port.h"
#include "esp_timer.h"
#include "store.h"
#include "ui.h"
#include "ui_theme.h"

// Ratings (FSRS)
#define AGAIN 1
#define HARD 2
#define GOOD 3
#define EASY 4

typedef struct {
    deck_t *deck;
    int *order;          // card indexes of this round
    int n, pos;
    bool quick;          // reminder: one card, then home
    int answered, correct;
    int64_t shown_us;    // when the current card appeared (answer time)
    bool answered_this;  // quiz/cloze: first tap counts
    lv_obj_t *opts[6];
    int n_opts;
    lv_obj_t *after;     // area filled after answering
} session_t;

static session_t S;
static bool s_active;

static void show_card(void);

bool study_active(void) { return s_active; }

static card_t *cur(void) { return &S.deck->cards[S.order[S.pos]]; }

static int elapsed_ms(void) { return (int)((esp_timer_get_time() - S.shown_us) / 1000); }

static void finish_round(void);

static void next_card(void)
{
    S.pos++;
    if (S.pos >= S.n || (S.quick && S.answered >= 1)) {
        finish_round();
        return;
    }
    show_card();
}

static void record(int rating)
{
    card_t *c = cur();
    c->done = true; // not offered again until the next sync
    reviews_add(c->id, rating, elapsed_ms());
    S.answered++;
    if (rating >= GOOD) S.correct++;
}

// ---------------------------------------------------------------- layout helpers

static lv_obj_t *card_screen(lv_obj_t **content_out)
{
    char title[40];
    card_t *c = cur();
    const char *mode = !strcmp(c->kind, "mcq") ? "Question"
                     : !strcmp(c->mode, "quiz") ? "Quiz"
                     : !strcmp(c->mode, "cloze") ? "Fill the blank"
                     : !strcmp(c->mode, "listening") ? "Listening"
                     : c->is_new ? "New word" : "Review";
    if (S.quick) snprintf(title, sizeof(title), "Quick review · %s", mode);
    else snprintf(title, sizeof(title), "%d/%d · %s", S.pos + 1, S.n, mode);
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

static void word_header(lv_obj_t *p, card_t *c)
{
    lv_obj_t *w = ui_label(p, c->front, strlen(c->front) > 22 ? font_title : font_word, COL_TEXT, -1);
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
    if (c->meaning_vi) ui_label(box, c->meaning_vi, font_bold, COL_PRIMARY, -1);
    if (c->meaning_en) ui_label(box, c->meaning_en, font_small, COL_MUTED, -1);
    if (c->back) ui_label(box, c->back, font_body, COL_TEXT, -1);
    if (c->example_en) ui_label(box, c->example_en, font_body, COL_TEXT, -1);
    if (c->example_vi) ui_label(box, c->example_vi, font_small, COL_MUTED, -1);
}

// ---------------------------------------------------------------- flashcard

static void rate_clicked(lv_event_t *e)
{
    int rating = (int)(intptr_t)lv_event_get_user_data(e);
    audio_fx(rating == AGAIN ? AUDIO_FX_WRONG : AUDIO_FX_TAP);
    record(rating);
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
    if (!strcmp(c->mode, "cloze") && c->cloze_vi) {
        ui_label(p, c->cloze_vi, font_small, COL_MUTED, -1);
    } else if (strcmp(c->kind, "mcq")) {
        answer_box(p, c);
    }
    lv_obj_t *n = ui_button(p, S.pos + 1 < S.n && !S.quick ? "Next " LV_SYMBOL_RIGHT : "Done " LV_SYMBOL_OK,
                            ok ? COL_OK : COL_PRIMARY, font_bold);
    lv_obj_set_width(n, lv_pct(100));
    lv_obj_set_height(n, 44);
    lv_obj_add_event_cb(n, next_clicked, LV_EVENT_CLICKED, NULL);
    lv_obj_scroll_to_view_recursive(n, LV_ANIM_ON);
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
    if (o->right) {
        audio_fx(AUDIO_FX_CORRECT);
        record(GOOD);
    } else {
        audio_fx(AUDIO_FX_WRONG);
        lv_obj_set_style_bg_color(b, COL_BAD, 0);
        record(AGAIN);
    }
    reveal(o->right, s_right_index);
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
            audio_play_url(c->audio_url);
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

// ---------------------------------------------------------------- round

static void show_card(void)
{
    card_t *c = cur();
    lv_obj_t *content;
    lv_obj_t *scr = card_screen(&content);
    S.shown_us = esp_timer_get_time();
    S.answered_this = false;
    S.n_opts = 0;
    bool choices = !strcmp(c->kind, "mcq") || !strcmp(c->mode, "quiz") || !strcmp(c->mode, "cloze") ||
                   !strcmp(c->mode, "listening");
    if (choices && !c->is_new) build_choices(content, c);
    else build_flashcard(content, c);
    lv_screen_load_anim(scr, LV_SCR_LOAD_ANIM_MOVE_LEFT, 180, 0, true);
}

static void home_clicked(lv_event_t *e)
{
    audio_fx(AUDIO_FX_TAP);
    ui_go_home_locked();
}

static void finish_round(void)
{
    s_active = false;
    app_request(APP_REQ_SYNC); // send the answers and get the next batch
    if (S.quick) {
        deck_unref(S.deck);
        free(S.order);
        S.deck = NULL;
        ui_go_home_locked();
        return;
    }
    lv_obj_t *scr = ui_screen_with_header("Round finished", true);
    lv_obj_t *b = lv_obj_create(scr);
    lv_obj_set_size(b, lv_pct(100), lv_display_get_vertical_resolution(NULL) - 28);
    lv_obj_align(b, LV_ALIGN_BOTTOM_MID, 0, 0);
    lv_obj_set_style_bg_opa(b, LV_OPA_TRANSP, 0);
    lv_obj_set_style_border_width(b, 0, 0);
    lv_obj_set_flex_flow(b, LV_FLEX_FLOW_COLUMN);
    lv_obj_set_flex_align(b, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER, LV_FLEX_ALIGN_CENTER);
    lv_obj_t *t = lv_label_create(b);
    lv_label_set_text(t, LV_SYMBOL_OK);
    lv_obj_set_style_text_font(t, &lv_font_montserrat_20, 0);
    lv_obj_set_style_text_color(t, COL_OK, 0);
    char m[80];
    snprintf(m, sizeof(m), "%d card%s done", S.answered, S.answered == 1 ? "" : "s");
    ui_label(b, m, font_title, COL_TEXT, LV_SIZE_CONTENT);
    snprintf(m, sizeof(m), "%d remembered · %d to practise again", S.correct, S.answered - S.correct);
    ui_label(b, m, font_body, COL_MUTED, LV_SIZE_CONTENT);
    lv_obj_t *h = ui_button(b, LV_SYMBOL_HOME "  Home", COL_PRIMARY, font_bold);
    lv_obj_add_event_cb(h, home_clicked, LV_EVENT_CLICKED, NULL);
    lv_screen_load_anim(scr, LV_SCR_LOAD_ANIM_FADE_IN, 150, 0, true);
    audio_fx(AUDIO_FX_CORRECT);
    deck_unref(S.deck);
    free(S.order);
    S.deck = NULL;
}

/** Called with the LVGL lock held (from a button) or takes it (from the app task). */
static void start_locked(int max_cards, bool quick)
{
    if (S.deck) {
        deck_unref(S.deck);
        free(S.order);
    }
    memset(&S, 0, sizeof(S));
    S.deck = deck_get();
    if (!S.deck || !S.deck->n_cards) {
        deck_unref(S.deck);
        S.deck = NULL;
        ui_go_home_locked();
        return;
    }
    S.quick = quick;
    S.order = malloc(sizeof(int) * S.deck->n_cards);
    // due reviews first (already ordered by the server), then new words
    for (int pass = 0; pass < 2; pass++) {
        for (int i = 0; i < S.deck->n_cards; i++) {
            if (!S.deck->cards[i].done && S.deck->cards[i].is_new == (pass == 1)) S.order[S.n++] = i;
        }
    }
    if (max_cards > 0 && S.n > max_cards) S.n = max_cards;
    if (S.n == 0) {
        deck_unref(S.deck);
        free(S.order);
        S.deck = NULL;
        ui_go_home_locked();
        return;
    }
    s_active = true;
    show_card();
}

void study_start(int max_cards, bool quick)
{
    // from a button the lock is already held by the LVGL task; from other tasks take it
    if (lvgl_port_lock(0)) {
        start_locked(max_cards, quick);
        lvgl_port_unlock();
    }
}
