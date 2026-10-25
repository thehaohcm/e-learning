package handlers

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"
	"mvp-lang-app/internal/tts"

	"github.com/gin-gonic/gin"
)

// readAloud: card kinds whose front is an English word / phrase to read.
func readAloud(kind string) bool { return kind == models.KindVocab || kind == models.KindPhrasal }

// addAudio sets audio_url on English word cards (made by the tts container if
// needed). It waits at most a few seconds: words not ready yet get their audio
// at a later sync.
func (h *Handler) addAudio(ctx context.Context, items []study.QueueItem, ccs []content.Compact) {
	if h.TTS == nil {
		return
	}
	var texts []string
	for i, it := range items {
		if readAloud(it.Card.Kind) && ccs[i].AudioURL == "" && tts.Normalize(it.Card.Front) != "" {
			texts = append(texts, it.Card.Front)
		}
	}
	if len(texts) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	ready := h.TTS.Ensure(ctx, texts)
	var used []string
	for i, it := range items {
		if !readAloud(it.Card.Kind) || ccs[i].AudioURL != "" || tts.Normalize(it.Card.Front) == "" {
			continue
		}
		k := h.TTS.Key(it.Card.Front)
		if ready[k] {
			ccs[i].AudioURL = tts.URL(k)
			used = append(used, k)
		} else if h.TTS.Enabled() {
			// not made yet: the device asks with the text and it is made then
			ccs[i].AudioURL = tts.URL(k) + "?t=" + url.QueryEscape(tts.Normalize(it.Card.Front))
		}
	}
	h.TTS.Touch(used)
}

var audioFile = regexp.MustCompile(`^[0-9a-f]{32}\.wav$`)

// Audio serves a stored word (public: the key can only be known from the text).
func (h *Handler) Audio(c *gin.Context) {
	f := c.Param("file")
	if !audioFile.MatchString(f) {
		abort(c, http.StatusNotFound, "not found")
		return
	}
	key := strings.TrimSuffix(f, ".wav")
	wav, ok := h.TTS.Get(key)
	if !ok && c.Query("t") != "" {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		if h.TTS.MakeFor(ctx, key, c.Query("t")) {
			wav, ok = h.TTS.Get(key)
		}
		cancel()
	}
	if !ok {
		abort(c, http.StatusNotFound, "not found")
		return
	}
	c.Header("Cache-Control", "public, max-age=2592000")
	c.Data(http.StatusOK, "audio/wav", wav)
}

// AudioStats (admin): how much audio is stored.
func (h *Handler) AudioStats(c *gin.Context) {
	st, err := h.TTS.Stats()
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

// AudioTest (admin): ?text=hello makes the audio now and returns its URL, or the error.
func (h *Handler) AudioTest(c *gin.Context) {
	text := c.DefaultQuery("text", "hello")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	k, err := h.TTS.Test(ctx, text)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "url": tts.URL(k)})
}

// AudioPurge (admin): ?older_days=N deletes audio not played for N days, none = all.
// Deleted words are made again when they are studied.
func (h *Handler) AudioPurge(c *gin.Context) {
	days := queryInt(c, "older_days", 0)
	n, err := h.TTS.Purge(time.Duration(max(days, 0)) * 24 * time.Hour)
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": n})
}
