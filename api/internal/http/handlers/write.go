package handlers

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// "Write it" (Quiz mode): the user sees a Vietnamese sentence and writes it in
// English with a given word. Prompts come from card_examples (Vietnamese
// translation = prompt, English sentence = reference); missing ones are written
// by AI in one request. All answers of a round are graded in one AI request.

type writePrompt struct {
	CardID      int64  `json:"card_id"`
	Word        string `json:"word"`
	IPA         string `json:"ipa,omitempty"`
	MeaningVi   string `json:"meaning_vi"`
	ExampleID   int64  `json:"example_id"`
	PromptVi    string `json:"prompt_vi"`
	ReferenceEn string `json:"reference_en"`
}

// WritePrompts: POST /api/practice/write/prompts {"deck_id":2,"count":10}
// Words the user is learning come first (soonest due), then other words of the deck.
func (h *Handler) WritePrompts(c *gin.Context) {
	var req struct {
		DeckID int64 `json:"deck_id" binding:"required"`
		Count  int   `json:"count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if req.Count <= 0 || req.Count > practiceMaxWords {
		req.Count = practiceMaxWords
	}
	uid := userID(c)
	var cards []models.Card
	err := h.DB.Raw(`
		SELECT c.* FROM cards c
		LEFT JOIN user_cards uc ON uc.card_id = c.id AND uc.user_id = ?
		WHERE c.deck_id = ? AND c.kind IN ?
		ORDER BY (uc.card_id IS NOT NULL AND uc.state <> 0 AND NOT uc.suspended) DESC, uc.due ASC NULLS LAST, random()
		LIMIT ?`, uid, req.DeckID, []string{models.KindVocab, models.KindPhrasal}, req.Count).Scan(&cards).Error
	if err != nil {
		internal(c, err)
		return
	}
	if len(cards) == 0 {
		abort(c, http.StatusBadRequest, "Write it works with English decks (words and phrasal verbs)")
		return
	}

	pools, err := h.examplePools(cards)
	if err != nil {
		internal(c, err)
		return
	}
	warning := ""
	var need []models.Card
	for _, cd := range cards {
		if countAI(pools[cd.ID]) == 0 {
			need = append(need, cd)
		}
	}
	if len(need) > 0 {
		if !h.AI.Enabled() {
			warning = "AI is not configured: only words with an example sentence are used."
		} else if err := h.generateExamples(c.Request.Context(), need); err != nil {
			log.Printf("[ai] write prompts: %v", err)
			warning = "AI could not write new sentences right now: some words were left out."
		}
		if pools, err = h.examplePools(cards); err != nil {
			internal(c, err)
			return
		}
	}

	out := []writePrompt{}
	for _, cd := range cards {
		var usable []models.CardExample
		for _, ex := range pools[cd.ID] {
			if ex.TranslationVi != "" {
				usable = append(usable, ex)
			}
		}
		if len(usable) == 0 {
			// the card's own example, if it has a Vietnamese translation
			cc := content.ToCompact(cd)
			if cc.ExampleEn == "" || cc.ExampleVi == "" {
				continue
			}
			ex := models.CardExample{CardID: cd.ID, SentenceEn: cc.ExampleEn, TranslationVi: cc.ExampleVi, Source: "card"}
			h.DB.Where("card_id = ? AND lower(sentence_en) = lower(?)", cd.ID, cc.ExampleEn).FirstOrCreate(&ex)
			usable = []models.CardExample{ex}
		}
		ex := usable[rand.Intn(len(usable))]
		cc := content.ToCompact(cd)
		out = append(out, writePrompt{
			CardID: cd.ID, Word: cd.Front, IPA: cc.IPA, MeaningVi: cc.MeaningVi,
			ExampleID: ex.ID, PromptVi: ex.TranslationVi, ReferenceEn: ex.SentenceEn,
		})
	}
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	c.JSON(http.StatusOK, gin.H{"prompts": out, "warning": warning})
}

// WriteGrade: POST /api/practice/write/grade {"items":[{"example_id":1,"answer":"..."}]}
// One AI request for all answers. A correct sentence with the word counts as a
// "Good" review for cards that are due (like the quiz), unless the learner had
// to reveal the word ("hinted"). Mistakes are not punished: this exercise is
// about using the word, the schedule is set by Study.
func (h *Handler) WriteGrade(c *gin.Context) {
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	var req struct {
		Items []struct {
			ExampleID int64  `json:"example_id"`
			Answer    string `json:"answer"`
			Hinted    bool   `json:"hinted"` // the word was revealed before writing
		} `json:"items" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if len(req.Items) > practiceMaxGrading {
		req.Items = req.Items[:practiceMaxGrading]
	}
	ids := make([]int64, 0, len(req.Items))
	for _, it := range req.Items {
		ids = append(ids, it.ExampleID)
	}
	var exs []models.CardExample
	if err := h.DB.Where("id IN ?", ids).Find(&exs).Error; err != nil {
		internal(c, err)
		return
	}
	exByID := map[int64]models.CardExample{}
	cardIDs := []int64{}
	for _, ex := range exs {
		exByID[ex.ID] = ex
		cardIDs = append(cardIDs, ex.CardID)
	}
	var cards []models.Card
	h.DB.Where("id IN ?", cardIDs).Find(&cards)
	cardByID := map[int64]models.Card{}
	for _, cd := range cards {
		cardByID[cd.ID] = cd
	}

	type result struct {
		ExampleID int64 `json:"example_id"`
		CardID    int64 `json:"card_id"`
		UsedWord  bool  `json:"used_word"`
		ai.WrittenResult
		Graded    bool `json:"graded"`
		Scheduled bool `json:"scheduled"` // counted as a review (card was due)
	}
	results := make([]result, len(req.Items))
	var items []ai.WrittenItem
	var slots []int
	hinted := map[int]bool{}
	for i, it := range req.Items {
		hinted[i] = it.Hinted
		results[i].ExampleID = it.ExampleID
		ex, ok := exByID[it.ExampleID]
		answer := strings.TrimSpace(it.Answer)
		if !ok || answer == "" {
			continue
		}
		if len([]rune(answer)) > 400 {
			answer = string([]rune(answer)[:400])
		}
		cd := cardByID[ex.CardID]
		results[i].CardID = cd.ID
		results[i].UsedWord = ai.ContainsWord(answer, cd.Front)
		cc := content.ToCompact(cd)
		items = append(items, ai.WrittenItem{
			Word: cd.Front, MeaningVi: cc.MeaningVi, PromptVi: ex.TranslationVi, ReferenceEn: ex.SentenceEn, LearnerEn: answer,
		})
		slots = append(slots, i)
	}
	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{"results": results})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
	defer cancel()
	graded, err := h.AI.GradeWrittenSentences(ctx, items)
	if err != nil {
		log.Printf("[ai] write grade: %v", err)
		abort(c, http.StatusBadGateway, "AI could not check the sentences right now, please try again")
		return
	}
	uid := userID(c)
	for k, g := range graded {
		if g.Score < 0 {
			continue
		}
		r := &results[slots[k]]
		if !r.UsedWord {
			// the point of the exercise is to use the word: say so, whatever the AI thought
			g.Correct = false
			g.Score = min(g.Score, 40)
			g.FeedbackVi = "Câu chưa dùng từ “" + cardByID[r.CardID].Front + "”. Hãy viết lại với từ này."
		}
		r.WrittenResult = g
		r.Graded = true
		if g.Correct && !hinted[slots[k]] {
			ar, err := h.Study.ApplyAnswer(study.Answer{UserID: uid, CardID: r.CardID, Rating: srs.Good, At: time.Now(), Source: "write", OnlyIfDue: true})
			if err == nil {
				r.Scheduled = ar.Applied
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
