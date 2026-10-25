package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Word notes: the floating notes panel. A note is matched to an existing card
// when possible ("ready"); otherwise AI explains it (one request for all new
// notes). "Add to Study" enrolls the card, creating it first in the user's
// personal "My words" deck when the word is not in any deck.

const (
	noteNew    = "new"
	noteReady  = "ready"
	noteAdded  = "added"
	maxNoteLen = 60
	maxFill    = 20
)

var englishKinds = []string{models.KindVocab, models.KindPhrasal}

// ListNotes: GET /api/notes
func (h *Handler) ListNotes(c *gin.Context) {
	notes := []models.WordNote{}
	if err := h.DB.Where("user_id = ?", userID(c)).Order("created_at DESC").Limit(300).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, notes)
}

// CreateNote: POST /api/notes {"text":"...","note":"","context":"sentence","source":"Reading: ..."}
// Noting the same word again updates it (and moves it to the top).
func (h *Handler) CreateNote(c *gin.Context) {
	var req noteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	h.createNote(c, userID(c), req)
}

type noteRequest struct {
	Text    string `json:"text"`
	Note    string `json:"note"`
	Context string `json:"context"`
	Source  string `json:"source"`
}

// createNote is shared by the web (CreateNote) and the browser extension (ExtAddNote).
func (h *Handler) createNote(c *gin.Context, uid int64, req noteRequest) {
	text := cleanNoteText(req.Text)
	if text == "" {
		abort(c, http.StatusBadRequest, "only English words or phrases can be noted (up to 60 characters)")
		return
	}
	now := time.Now().UTC()
	var n models.WordNote
	err := h.DB.Where("user_id = ? AND lower(text) = lower(?)", uid, text).First(&n).Error
	switch {
	case err == nil: // already noted: keep it, refresh what was given
		updates := map[string]any{"created_at": now, "updated_at": now}
		if s := clip(req.Note, 300); s != "" {
			updates["note"] = s
		}
		if s := clip(req.Context, 400); s != "" {
			updates["context"] = s
		}
		if s := clip(req.Source, 80); s != "" {
			updates["source"] = s
		}
		h.DB.Model(&n).Updates(updates)
		h.DB.First(&n, n.ID)
		c.JSON(http.StatusOK, n)
		return
	case !errors.Is(err, gorm.ErrRecordNotFound):
		internal(c, err)
		return
	}
	n = models.WordNote{
		UserID: uid, Text: text, Note: clip(req.Note, 300), Context: clip(req.Context, 400), Source: clip(req.Source, 80),
		Status: noteNew, Info: models.JSON("{}"), CreatedAt: now, UpdatedAt: now,
	}
	if !h.matchNote(uid, &n, text) {
		h.dictInfo(c.Request.Context(), &n, text)
	}
	if err := h.DB.Create(&n).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, n)
}

// UpdateNote: PATCH /api/notes/:id {"note":"..."} or {"text":"..."} (a new text is matched again)
func (h *Handler) UpdateNote(c *gin.Context) {
	n, ok := h.loadNote(c)
	if !ok {
		return
	}
	var req struct {
		Text *string `json:"text"`
		Note *string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if req.Note != nil {
		n.Note = clip(*req.Note, 300)
	}
	if req.Text != nil {
		text := cleanNoteText(*req.Text)
		if text == "" {
			abort(c, http.StatusBadRequest, "only English words or phrases can be noted (up to 60 characters)")
			return
		}
		if !strings.EqualFold(text, n.Text) {
			n.Text, n.CardID, n.Status, n.Info = text, nil, noteNew, models.JSON("{}")
			if !h.matchNote(n.UserID, &n, text) {
				h.dictInfo(c.Request.Context(), &n, text)
			}
		}
	}
	n.UpdatedAt = time.Now().UTC()
	if err := h.DB.Save(&n).Error; err != nil {
		abort(c, http.StatusConflict, "this word is already in your notes")
		return
	}
	c.JSON(http.StatusOK, n)
}

// DeleteNote: DELETE /api/notes/:id (the card stays in Study if it was added)
func (h *Handler) DeleteNote(c *gin.Context) {
	n, ok := h.loadNote(c)
	if !ok {
		return
	}
	h.DB.Delete(&n)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// FillNotes: POST /api/notes/fill
// One AI request explains every note that has no meaning yet (up to 20).
// Words the AI says are not English are deleted and listed in "removed".
func (h *Handler) FillNotes(c *gin.Context) {
	uid := userID(c)
	removed, err := h.fillNotes(c.Request.Context(), uid, nil)
	if err != nil {
		h.notesError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": h.userNotes(uid), "removed": removed})
}

func (h *Handler) userNotes(uid int64) []models.WordNote {
	notes := []models.WordNote{}
	h.DB.Where("user_id = ?", uid).Order("created_at DESC").Limit(300).Find(&notes)
	return notes
}

// AddNotes: POST /api/notes/add {"ids":[1,2]} (empty = every note not in Study yet)
// Notes without a meaning are explained first (one AI request), then each word
// is enrolled: its existing card, or a new card in the personal "My words" deck.
func (h *Handler) AddNotes(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	_ = c.ShouldBindJSON(&req)
	uid := userID(c)
	removed, err := h.fillNotes(c.Request.Context(), uid, req.IDs)
	if err != nil && !errors.Is(err, ai.ErrNotConfigured) {
		h.notesError(c, err)
		return
	}

	q := h.DB.Where("user_id = ? AND status = ?", uid, noteReady)
	if len(req.IDs) > 0 {
		q = q.Where("id IN ?", req.IDs)
	}
	var notes []models.WordNote
	if err := q.Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	now := time.Now().UTC()
	added, created := 0, 0
	var deck *models.Deck
	for i := range notes {
		n := &notes[i]
		cardID := int64(0)
		if n.CardID != nil {
			cardID = *n.CardID
		} else {
			var info ai.NoteInfo
			_ = json.Unmarshal(n.Info, &info)
			if info.MeaningVi == "" {
				continue
			}
			if deck == nil {
				d, err := h.myWordsDeck(uid)
				if err != nil {
					internal(c, err)
					return
				}
				deck = d
			}
			card, isNew, err := h.cardFromNote(deck, info)
			if err != nil {
				log.Printf("[notes] card for %q: %v", n.Text, err)
				continue
			}
			if isNew {
				created++
			}
			cardID = card.ID
		}
		if _, err := h.Study.Enroll(uid, []int64{cardID}, now); err != nil {
			internal(c, err)
			return
		}
		h.DB.Model(n).Updates(map[string]any{"card_id": cardID, "status": noteAdded, "updated_at": now})
		added++
	}
	c.JSON(http.StatusOK, gin.H{"added": added, "created": created, "removed": removed, "notes": h.userNotes(uid)})
}

// ---------------------------------------------------------------- helpers

func (h *Handler) loadNote(c *gin.Context) (models.WordNote, bool) {
	var n models.WordNote
	id, ok := paramID(c, "id")
	if !ok {
		return n, false
	}
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&n).Error; err != nil {
		notFoundOr500(c, err, "note")
		return n, false
	}
	return n, true
}

func (h *Handler) notesError(c *gin.Context, err error) {
	if errors.Is(err, ai.ErrNotConfigured) {
		abort(c, http.StatusServiceUnavailable, "AI is not configured: only words that are already in a deck can be explained")
		return
	}
	log.Printf("[ai] notes: %v", err)
	abort(c, http.StatusBadGateway, "AI could not explain the words right now, please try again")
}

// matchNote links the note to an existing English card with the same front
// (a shared deck or the user's own), so no AI is needed for it.
func (h *Handler) matchNote(uid int64, n *models.WordNote, text string) bool {
	var card models.Card
	err := h.DB.Where("lower(front) = lower(?) AND kind IN ? AND deck_id IN ("+study.VisibleDecksSQL+")", text, englishKinds, uid).
		Order("(source = 'note'), id").First(&card).Error
	if err != nil {
		return false
	}
	cc := content.ToCompact(card)
	info, _ := json.Marshal(ai.NoteInfo{
		Base: card.Front, Kind: card.Kind, IPA: cc.IPA, POS: cc.POS, MeaningVi: cc.MeaningVi, MeaningEn: cc.MeaningEn,
		ExampleEn: cc.ExampleEn, ExampleVi: cc.ExampleVi,
	})
	id := card.ID
	n.CardID, n.Info, n.Status = &id, models.JSON(info), noteReady
	var inList int64
	h.DB.Model(&models.UserCard{}).Where("user_id = ? AND card_id = ?", uid, id).Count(&inList)
	if inList > 0 {
		n.Status = noteAdded
	}
	return true
}

// fillNotes explains the user's notes that have no meaning yet (only `ids` if
// given). Notes the AI says are not English words are deleted; their texts are returned.
func (h *Handler) fillNotes(ctx context.Context, uid int64, ids []int64) ([]string, error) {
	removed := []string{}
	q := h.DB.Where("user_id = ? AND status = ?", uid, noteNew)
	if len(ids) > 0 {
		q = q.Where("id IN ?", ids)
	}
	var notes []models.WordNote
	if err := q.Order("created_at DESC").Limit(maxFill).Find(&notes).Error; err != nil {
		return removed, err
	}
	if len(notes) == 0 {
		return removed, nil
	}
	if !h.AI.Enabled() {
		return removed, ai.ErrNotConfigured
	}
	inputs := make([]ai.NoteInput, len(notes))
	for i, n := range notes {
		inputs[i] = ai.NoteInput{Text: n.Text, Context: n.Context}
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	infos, err := h.AI.FillNotes(ctx, inputs)
	if err != nil {
		return removed, err
	}
	now := time.Now().UTC()
	for i := range notes {
		n := &notes[i]
		info := infos[i]
		if info.MeaningVi == "" {
			if info.Answered { // the AI says it is not an English word: no point keeping it
				h.DB.Delete(n)
				removed = append(removed, n.Text)
			}
			continue // missing from the reply: stays "new", explained next time
		}
		// the dictionary form may already be a card ("negotiated" -> "negotiate")
		if !strings.EqualFold(info.Base, n.Text) && h.matchNote(uid, n, info.Base) {
			h.DB.Model(n).Updates(map[string]any{"card_id": n.CardID, "info": n.Info, "status": n.Status, "updated_at": now})
			continue
		}
		b, _ := json.Marshal(info)
		h.DB.Model(n).Updates(map[string]any{"info": models.JSON(b), "status": noteReady, "updated_at": now})
	}
	return removed, nil
}

// myWordsDeck returns the user's personal deck, creating it on first use.
func (h *Handler) myWordsDeck(uid int64) (*models.Deck, error) {
	var d models.Deck
	err := h.DB.Where("owner_id = ? AND slug = ?", uid, myWordsSlug(uid)).First(&d).Error
	if err == nil {
		return &d, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var subject models.Subject
	if err := h.DB.Where("slug = ?", "english").First(&subject).Error; err != nil {
		return nil, fmt.Errorf("english subject: %w", err)
	}
	owner := uid
	d = models.Deck{
		SubjectID: subject.ID, Slug: myWordsSlug(uid), Name: "My words", OwnerID: &owner,
		Description: "Words you noted while studying", DefaultKind: models.KindVocab, SortOrder: 1000,
	}
	if err := h.DB.Create(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func myWordsSlug(uid int64) string { return fmt.Sprintf("my-words-%d", uid) }

// cardFromNote creates (or finds) the card for a noted word in the deck.
func (h *Handler) cardFromNote(deck *models.Deck, info ai.NoteInfo) (models.Card, bool, error) {
	var existing models.Card
	if err := h.DB.Where("deck_id = ? AND lower(front) = lower(?)", deck.ID, info.Base).First(&existing).Error; err == nil {
		return existing, false, nil
	}
	sense := content.Sense{Level: 1, MeaningVi: info.MeaningVi, MeaningEn: info.MeaningEn, ExampleEn: info.ExampleEn, ExampleVi: info.ExampleVi}
	var raw []byte
	kind := models.KindVocab
	if info.Kind == "phrasal" {
		kind = models.KindPhrasal
		parts := strings.SplitN(info.Base, " ", 2)
		raw, _ = json.Marshal(content.PhrasalPayload{Verb: parts[0], Particle: parts[len(parts)-1], Senses: []content.Sense{sense}})
	} else {
		raw, _ = json.Marshal(content.VocabPayload{IPA: info.IPA, POS: info.POS, Senses: []content.Sense{sense}})
	}
	payload, err := content.Normalize(kind, info.Base, raw)
	if err != nil {
		return models.Card{}, false, err
	}
	card := models.Card{DeckID: deck.ID, Kind: kind, Front: info.Base, Payload: models.JSON(payload), Source: "note"}
	if err := h.DB.Create(&card).Error; err != nil {
		return models.Card{}, false, err
	}
	return card, true, nil
}

var spaces = regexp.MustCompile(`\s+`)

// cleanNoteText trims spaces and surrounding punctuation ("“negotiate,”" -> "negotiate").
func cleanNoteText(s string) string {
	s = spaces.ReplaceAllString(strings.TrimSpace(s), " ")
	s = strings.TrimFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	if s == "" || len([]rune(s)) > maxNoteLen || !englishText.MatchString(s) {
		return ""
	}
	return s
}

// English words/phrases only: latin letters without accents, digits, ' - . and spaces.
var englishText = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9'’\-. ]*$`)

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
