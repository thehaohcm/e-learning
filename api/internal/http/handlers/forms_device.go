package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// formsRunning: users whose word forms are being prepared (one run at a time).
var formsRunning sync.Map

// DevicePrepareForms (the device calls it when the Quiz button is pressed):
// prepares "which form?" options (TOEIC Part 5) for the example sentences of up
// to maxFormBatch learned words that have none yet, in ONE AI request, in the
// background. They are kept (card_examples.form_options, shared with the web)
// and reach the device at its next sync.
func (h *Handler) DevicePrepareForms(c *gin.Context) {
	uid, ok := h.pairedUser(c)
	if !ok {
		return
	}
	if !h.AI.Enabled() {
		c.JSON(http.StatusOK, gin.H{"queued": 0})
		return
	}
	if _, busy := formsRunning.LoadOrStore(uid, true); busy {
		c.JSON(http.StatusOK, gin.H{"queued": 0, "running": true})
		return
	}
	devSettings := study.ParseDeviceSettings(device(c).Settings)
	pool, err := h.Study.QuizPool(uid, study.Filter{SubjectIDs: devSettings.SubjectIDs}, 60)
	if err != nil {
		formsRunning.Delete(uid)
		internal(c, err)
		return
	}
	var need []ai.FormItem
	var needIDs []int64
	for _, it := range pool {
		if len(need) >= maxFormBatch {
			break
		}
		if it.Card.Kind != models.KindVocab || strings.Contains(strings.TrimSpace(it.Card.Front), " ") {
			continue // phrases have no word family
		}
		cc := content.ToCompact(it.Card)
		answer, found := study.FindWord(cc.ExampleEn, it.Card.Front)
		if !found {
			continue
		}
		id, prepared := h.exampleRow(it.Card.ID, cc.ExampleEn, cc.ExampleVi)
		if id == 0 || prepared {
			continue
		}
		need = append(need, ai.FormItem{Sentence: cc.ExampleEn, Answer: answer, Word: it.Card.Front})
		needIDs = append(needIDs, id)
	}
	if len(need) == 0 {
		formsRunning.Delete(uid)
		c.JSON(http.StatusOK, gin.H{"queued": 0})
		return
	}
	go func() {
		defer formsRunning.Delete(uid)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		res, err := h.AI.GenerateWordForms(ctx, need)
		if err != nil {
			log.Printf("[ai] device word forms: %v", err)
			return
		}
		for i, id := range needIDs {
			if i < len(res) {
				h.saveForms(id, res[i])
			}
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"queued": len(need)})
}

// exampleRow returns the card_examples row of this sentence (adding the card's
// own example if needed) and whether its form options are prepared already.
func (h *Handler) exampleRow(cardID int64, en, vi string) (int64, bool) {
	h.DB.Exec(`INSERT INTO card_examples (card_id, sentence_en, translation_vi, source) VALUES (?, ?, ?, 'card')
		ON CONFLICT DO NOTHING`, cardID, en, vi)
	var row struct {
		ID       int64
		Prepared bool
	}
	h.DB.Raw(`SELECT id, form_options IS NOT NULL AS prepared FROM card_examples
		WHERE card_id = ? AND lower(sentence_en) = lower(?)`, cardID, en).Scan(&row)
	return row.ID, row.Prepared
}
