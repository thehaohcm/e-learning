package handlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// clozeItem is one fill-in-the-blank sentence of the round check.
type clozeItem struct {
	writePrompt
	POS string `json:"pos,omitempty"`
	// Answer: the form of the word used in the sentence ("fluctuated").
	Answer string `json:"answer"`
	// WordOptions: 3 other words, same part of speech when possible ("Which word?").
	WordOptions []string `json:"word_options"`
	// FormOptions: 3 wrong forms of the same word family ("Which form?", TOEIC
	// Part 5). null = not prepared yet (POST /practice/cloze/forms), [] = none.
	FormOptions *[]string `json:"form_options"`
}

// ClozePrompts: POST /api/practice/cloze {"card_ids":[1,2,3]}
// End of a Study/Quiz round: one sentence per word (different from the card's
// own example when possible), shuffled, for a fill-in-the-blank exercise with a
// word bank. Missing sentences are written by AI in one request.
func (h *Handler) ClozePrompts(c *gin.Context) {
	var req struct {
		CardIDs []int64 `json:"card_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	ids := uniqueIDs(req.CardIDs, practiceMaxWords)
	var cards []models.Card
	if err := h.DB.Where("id IN ? AND kind IN ? AND deck_id IN ("+study.VisibleDecksSQL+")", ids,
		[]string{models.KindVocab, models.KindPhrasal}, userID(c)).Find(&cards).Error; err != nil {
		internal(c, err)
		return
	}
	if len(cards) == 0 {
		abort(c, http.StatusBadRequest, "no English words in this round")
		return
	}
	prompts, warning, err := h.examplePrompts(c.Request.Context(), cards)
	if err != nil {
		internal(c, err)
		return
	}

	pos := map[int64]string{}
	for _, cd := range cards {
		pos[cd.ID] = posKey(content.ToCompact(cd).POS, cd.Kind)
	}
	wordOpts := h.sameKindWords(cards, pos)
	exIDs := make([]int64, len(prompts))
	for i, p := range prompts {
		exIDs[i] = p.ExampleID
	}
	forms := h.storedForms(exIDs)

	out := make([]clozeItem, 0, len(prompts))
	for _, p := range prompts {
		answer, _ := study.FindWord(p.ReferenceEn, p.Word)
		it := clozeItem{writePrompt: p, POS: pos[p.CardID], Answer: answer, WordOptions: wordOpts[p.CardID]}
		if f, ok := forms[p.ExampleID]; ok {
			it.FormOptions = &f
		}
		out = append(out, it)
	}
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	c.JSON(http.StatusOK, gin.H{"items": out, "warning": warning})
}

// posKey: a comparable part of speech ("noun", "verb", "adjective", "adverb").
func posKey(pos, kind string) string {
	if kind == models.KindPhrasal {
		return "phrasal verb"
	}
	p := strings.ToLower(strings.TrimSpace(pos))
	switch {
	case strings.HasPrefix(p, "phrasal"):
		return "phrasal verb"
	case strings.HasPrefix(p, "adj"):
		return "adjective"
	case strings.HasPrefix(p, "adv"):
		return "adverb"
	case strings.HasPrefix(p, "noun"), p == "n", p == "n.":
		return "noun"
	case strings.HasPrefix(p, "verb"), p == "v", p == "v.":
		return "verb"
	}
	return p
}

// sameKindWords: for each card, 3 other English words for "Which word?":
// same part of speech first (from the round, then from the card's deck), so
// grammar alone does not give the answer away.
func (h *Handler) sameKindWords(cards []models.Card, pos map[int64]string) map[int64][]string {
	inRound := map[int64]bool{}
	decks := map[int64]bool{}
	for _, cd := range cards {
		inRound[cd.ID] = true
		decks[cd.DeckID] = true
	}
	type cand struct {
		front, pos string
		deck       int64
	}
	var pool []cand
	for _, cd := range cards {
		pool = append(pool, cand{cd.Front, pos[cd.ID], cd.DeckID})
	}
	for deck := range decks {
		var more []models.Card
		h.DB.Where("deck_id = ? AND kind IN ?", deck, []string{models.KindVocab, models.KindPhrasal}).
			Order("random()").Limit(60).Find(&more)
		for _, m := range more {
			if !inRound[m.ID] {
				pool = append(pool, cand{m.Front, posKey(content.ToCompact(m).POS, m.Kind), m.DeckID})
			}
		}
	}
	out := map[int64][]string{}
	for _, cd := range cards {
		seen := map[string]bool{strings.ToLower(cd.Front): true}
		var picked []string
		add := func(match func(cand) bool) {
			for _, i := range rand.Perm(len(pool)) {
				if len(picked) == 3 {
					return
				}
				p := pool[i]
				if !seen[strings.ToLower(p.front)] && match(p) {
					seen[strings.ToLower(p.front)] = true
					picked = append(picked, p.front)
				}
			}
		}
		add(func(p cand) bool { return p.pos != "" && p.pos == pos[cd.ID] }) // same part of speech
		add(func(p cand) bool { return true })                               // otherwise anything
		out[cd.ID] = picked
	}
	return out
}

// storedForms: form options already prepared for these examples ([] = none).
func (h *Handler) storedForms(exampleIDs []int64) map[int64][]string {
	out := map[int64][]string{}
	if len(exampleIDs) == 0 {
		return out
	}
	var rows []struct {
		ID          int64
		FormOptions []byte
	}
	h.DB.Raw(`SELECT id, form_options FROM card_examples WHERE id IN ? AND form_options IS NOT NULL`, exampleIDs).Scan(&rows)
	for _, r := range rows {
		var f []string
		if json.Unmarshal(r.FormOptions, &f) == nil {
			if f == nil {
				f = []string{}
			}
			out[r.ID] = f
		}
	}
	return out
}

const maxFormBatch = 12

// ClozeForms: POST /api/practice/cloze/forms {"example_ids":[...]}
// "Which form?" options (TOEIC Part 5) for the round's sentences. Sentences
// without them are prepared by AI in ONE request and kept for everyone.
// Returns {"forms": {"<example id>": {"answer": "...", "options": [3] or []}}}.
func (h *Handler) ClozeForms(c *gin.Context) {
	var req struct {
		ExampleIDs []int64 `json:"example_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	ids := uniqueIDs(req.ExampleIDs, maxFormBatch)
	var rows []struct {
		ID          int64
		SentenceEn  string
		Front       string
		FormOptions []byte
	}
	if err := h.DB.Raw(`
		SELECT e.id, e.sentence_en, c.front, e.form_options
		FROM card_examples e JOIN cards c ON c.id = e.card_id
		WHERE e.id IN ? AND c.kind IN ? AND c.deck_id IN (`+study.VisibleDecksSQL+`)`,
		ids, []string{models.KindVocab, models.KindPhrasal}, userID(c)).Scan(&rows).Error; err != nil {
		internal(c, err)
		return
	}

	type form struct {
		Answer  string   `json:"answer"`
		Options []string `json:"options"`
	}
	result := map[int64]form{}
	var need []ai.FormItem
	var needIDs []int64
	for _, r := range rows {
		answer, found := study.FindWord(r.SentenceEn, r.Front)
		if r.FormOptions != nil {
			var f []string
			_ = json.Unmarshal(r.FormOptions, &f)
			result[r.ID] = form{answer, nonNil(f)}
			continue
		}
		if !found || strings.Contains(r.Front, " ") {
			h.saveForms(r.ID, nil) // phrases, or the word is not in the sentence: no form question
			result[r.ID] = form{answer, []string{}}
			continue
		}
		need = append(need, ai.FormItem{Sentence: r.SentenceEn, Answer: answer, Word: r.Front})
		needIDs = append(needIDs, r.ID)
	}

	warning := ""
	if len(need) > 0 {
		if !h.AI.Enabled() {
			warning = ai.ErrNotConfigured.Error()
		} else {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
			defer cancel()
			res, err := h.AI.GenerateWordForms(ctx, need)
			if err != nil {
				log.Printf("[ai] word forms: %v", err)
				warning = "AI could not prepare the word-form options right now."
			} else {
				for i, id := range needIDs {
					opts := res[i]
					h.saveForms(id, opts)
					result[id] = form{need[i].Answer, nonNil(opts)}
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"forms": result, "warning": warning})
}

func (h *Handler) saveForms(exampleID int64, opts []string) {
	b, _ := json.Marshal(nonNil(opts))
	h.DB.Exec(`UPDATE card_examples SET form_options = ?::jsonb WHERE id = ?`, string(b), exampleID)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
