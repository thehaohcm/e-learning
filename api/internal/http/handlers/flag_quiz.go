package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"

	"github.com/gin-gonic/gin"
)

// FlagQuiz: POST /api/study/flag-quiz {command, context?, model?} — writes an
// ephemeral multi-select quiz question about one command's flags/options
// (e.g. for "hydra", which of -l -L -p -P ... are real). Not persisted as a
// card: it is generated and answered on the spot, right where the learner
// just reviewed the command (the "Không nhớ" reveal screen in Study/Labs).
func (h *Handler) FlagQuiz(c *gin.Context) {
	var req struct {
		Command string `json:"command"`
		Context string `json:"context"`
		Model   string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Command) == "" {
		abort(c, http.StatusBadRequest, "command is required")
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	if req.Model != "" && !h.chatModelAllowed(ctx, req.Model) {
		abort(c, http.StatusBadRequest, "model not allowed")
		return
	}
	quiz, err := h.AI.GenerateFlagQuiz(ctx, req.Model, strings.TrimSpace(req.Command), strings.TrimSpace(req.Context))
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("[ai] flag-quiz user=%d: %v", userID(c), err)
		abort(c, http.StatusBadGateway, "AI could not generate a quiz: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, quiz)
}
