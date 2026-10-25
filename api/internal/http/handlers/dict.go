package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/dict"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// DictLookup: GET /api/dict?word=reluctant
// The Vietnamese meanings of an English word from dict.minhqnd.com (cached),
// shown while a word is typed in the notes panel. A dictionary that does not
// answer is not an error for the panel: {"found":false,"error":"..."}.
func (h *Handler) DictLookup(c *gin.Context) {
	word := strings.TrimSpace(c.Query("word"))
	if dict.Normalize(word) == "" {
		abort(c, http.StatusBadRequest, "word is required (up to 60 characters)")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
	defer cancel()
	e, err := h.Dict.Lookup(ctx, word)
	if e.Meanings == nil {
		e.Meanings = []dict.Meaning{}
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"word": e.Word, "found": false, "meanings": e.Meanings, "error": "dictionary unavailable"})
		return
	}
	c.JSON(http.StatusOK, e)
}

// dictInfo fills a new note's meaning from the dictionary (a few seconds at
// most), so the panel shows it at once. The note stays "new": Explain (AI)
// still writes the full entry with examples.
func (h *Handler) dictInfo(ctx context.Context, n *models.WordNote, text string) {
	if !h.Dict.Enabled() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	e, err := h.Dict.Lookup(ctx, text)
	if err != nil || !e.Found {
		return
	}
	kind := models.KindVocab
	if strings.Contains(strings.TrimSpace(text), " ") {
		kind = models.KindPhrasal
	}
	info, _ := json.Marshal(ai.NoteInfo{
		Base: text, Kind: kind, IPA: e.IPA, POS: e.FirstPOS(), MeaningVi: e.Short(3), Dict: true,
	})
	n.Info = models.JSON(info)
}
