package handlers

import (
	"log"
	"net/http"

	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// "More examples" on a word card: up to practicePoolSize AI sentences per card
// in different situations, shared by all users and also used by the practice
// drills (Forgotten words, Write it).

type cardExample struct {
	ID      int64  `json:"id"`
	En      string `json:"en"`
	Vi      string `json:"vi"`
	Context string `json:"context,omitempty"`
	Source  string `json:"source"` // ai | card
}

// CardExamples: GET /api/cards/:id/examples (stored sentences only, no AI)
func (h *Handler) CardExamples(c *gin.Context) {
	card, ok := h.exampleCard(c)
	if !ok {
		return
	}
	h.respondExamples(c, card, "")
}

// GenerateCardExamples: POST /api/cards/:id/examples
// Fills the card up to practicePoolSize AI sentences (one AI request), then returns them all.
func (h *Handler) GenerateCardExamples(c *gin.Context) {
	card, ok := h.exampleCard(c)
	if !ok {
		return
	}
	if !h.AI.Enabled() {
		h.respondExamples(c, card, "AI is not configured on the server.")
		return
	}
	pools, err := h.examplePools([]models.Card{card})
	if err != nil {
		internal(c, err)
		return
	}
	warning := ""
	if missing := practicePoolSize - countAI(pools[card.ID]); missing > 0 {
		if err := h.generateExamplesN(c.Request.Context(), []models.Card{card}, missing); err != nil {
			log.Printf("[ai] examples for card %d: %v", card.ID, err)
			warning = "AI could not write examples right now, please try again later."
		}
	}
	h.respondExamples(c, card, warning)
}

func (h *Handler) exampleCard(c *gin.Context) (models.Card, bool) {
	var card models.Card
	id, ok := paramID(c, "id")
	if !ok {
		return card, false
	}
	if err := h.DB.Where("deck_id IN ("+study.VisibleDecksSQL+")", userID(c)).First(&card, id).Error; err != nil {
		notFoundOr500(c, err, "card")
		return card, false
	}
	if card.Kind != models.KindVocab && card.Kind != models.KindPhrasal {
		abort(c, http.StatusBadRequest, "examples are only for English words")
		return card, false
	}
	return card, true
}

func (h *Handler) respondExamples(c *gin.Context, card models.Card, warning string) {
	pools, err := h.examplePools([]models.Card{card})
	if err != nil {
		internal(c, err)
		return
	}
	out := []cardExample{}
	ai := 0
	for _, ex := range pools[card.ID] {
		if ex.Source != "ai" {
			continue // the card's own example is already shown on the card
		}
		if ai++; ai > practicePoolSize {
			break
		}
		out = append(out, cardExample{ID: ex.ID, En: ex.SentenceEn, Vi: ex.TranslationVi, Context: ex.Context, Source: ex.Source})
	}
	c.JSON(http.StatusOK, gin.H{
		"card_id":      card.ID,
		"word":         card.Front,
		"examples":     out,
		"can_generate": h.AI.Enabled() && len(out) < practicePoolSize,
		"warning":      warning,
	})
}
