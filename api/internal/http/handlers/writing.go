package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// SubmitWriting stores the text and evaluates it with AI in the background.
// The client polls GET /api/writing/:id until status is "completed" or "failed".
func (h *Handler) SubmitWriting(c *gin.Context) {
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	var req dto.WritingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	sub := models.WritingSubmission{
		UserID: userID(c), Prompt: strings.TrimSpace(req.Prompt), OriginalText: strings.TrimSpace(req.Text),
		WordCount: len(strings.Fields(req.Text)), Status: "processing", AIFeedbackJSON: models.JSON("{}"),
		CreatedAt: time.Now().UTC(),
	}
	if err := h.DB.Create(&sub).Error; err != nil {
		internal(c, err)
		return
	}
	go h.evaluateWriting(sub.ID, sub.Prompt, sub.OriginalText)
	c.JSON(http.StatusAccepted, sub)
}

func (h *Handler) evaluateWriting(id int64, prompt, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	now := time.Now().UTC()
	fb, err := h.AI.EvaluateWriting(ctx, prompt, text, h.phraseLines())
	if err != nil {
		log.Printf("[ai] writing %d: %v", id, err)
		msg := "AI evaluation failed, please submit again"
		h.DB.Model(&models.WritingSubmission{}).Where("id = ?", id).
			Updates(map[string]any{"status": "failed", "error_message": msg, "completed_at": now})
		return
	}
	b, _ := json.Marshal(fb)
	model := h.AI.Model()
	h.DB.Model(&models.WritingSubmission{}).Where("id = ?", id).Updates(map[string]any{
		"status": "completed", "overall_score": fb.Overall, "grammar_score": fb.Grammar,
		"vocab_score": fb.Vocab, "coherence_score": fb.Coherence,
		"ai_feedback_json": models.JSON(b), "ai_model_used": model, "completed_at": now,
	})
}

func (h *Handler) GetWriting(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var sub models.WritingSubmission
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&sub).Error; err != nil {
		notFoundOr500(c, err, "submission")
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *Handler) ListWriting(c *gin.Context) {
	subs := []models.WritingSubmission{}
	if err := h.DB.Where("user_id = ?", userID(c)).Order("created_at DESC").Limit(30).Find(&subs).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, subs)
}
