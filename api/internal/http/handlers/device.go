package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/auth"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

const pairCodeTTL = 15 * time.Minute

func jsonMarshal(v any) ([]byte, error) { return json.Marshal(v) }

// newPairCode sets a fresh, unique pairing code on the device row.
func (h *Handler) newPairCode(d *models.Device) error {
	var lastErr error
	for i := 0; i < 5; i++ {
		code, err := auth.NewPairCode()
		if err != nil {
			return err
		}
		exp := time.Now().UTC().Add(pairCodeTTL)
		// free codes that expired so the unique index does not block reuse
		h.DB.Model(&models.Device{}).Where("pair_code = ? AND pair_code_expires_at < NOW()", code).
			Updates(map[string]any{"pair_code": nil, "pair_code_expires_at": nil})
		lastErr = h.DB.Model(d).Updates(map[string]any{"pair_code": code, "pair_code_expires_at": exp}).Error
		if lastErr == nil {
			d.PairCode, d.PairCodeExpiresAt = &code, &exp
			return nil
		}
	}
	return lastErr
}

// ---------------------------------------------------------------- device-side endpoints

// DeviceRegister is called once by a new device (no auth). It returns a secret
// token to keep in flash and a 6-digit code to show on screen. The user then
// types the code on the website (Devices page) to link the device.
func (h *Handler) DeviceRegister(c *gin.Context) {
	var req dto.DeviceRegisterRequest
	_ = c.ShouldBindJSON(&req)
	token, hash, err := auth.NewDeviceToken()
	if err != nil {
		internal(c, err)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "Study device"
	}
	d := models.Device{Name: name, TokenHash: hash, Firmware: req.Firmware, Settings: models.JSON("{}")}
	if err := h.DB.Create(&d).Error; err != nil {
		internal(c, err)
		return
	}
	if err := h.newPairCode(&d); err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.DeviceRegisterResponse{
		DeviceID: d.ID, Token: token, PairCode: *d.PairCode, PairExpiresAt: *d.PairCodeExpiresAt,
	})
}

// DeviceStatus tells the device whether it is linked to a user yet. While not
// linked it returns the current pairing code (renewed when expired).
func (h *Handler) DeviceStatus(c *gin.Context) {
	d := device(c)
	if d.UserID != nil {
		var u models.User
		h.DB.Select("id, display_name, email").First(&u, *d.UserID)
		name := u.DisplayName
		if name == "" {
			name = strings.Split(u.Email, "@")[0]
		}
		c.JSON(http.StatusOK, gin.H{"paired": true, "user_name": name, "device_name": d.Name, "server_time": time.Now().UTC()})
		return
	}
	if d.PairCode == nil || d.PairCodeExpiresAt == nil || time.Now().After(*d.PairCodeExpiresAt) {
		if err := h.newPairCode(&d); err != nil {
			internal(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"paired": false, "pair_code": *d.PairCode, "pair_expires_at": d.PairCodeExpiresAt, "server_time": time.Now().UTC()})
}

func (h *Handler) pairedUser(c *gin.Context) (int64, bool) {
	d := device(c)
	if d.UserID == nil {
		abort(c, http.StatusConflict, "device is not paired yet")
		return 0, false
	}
	return *d.UserID, true
}

type deviceCard struct {
	content.Compact
	IsNew   bool           `json:"is_new"`
	State   int            `json:"state"`
	Due     time.Time      `json:"due"`
	Options []study.Choice `json:"options,omitempty"` // 4 meanings, correct one has id == card id
	// How to show the card: "flashcard", "quiz" (options, or choices for mcq)
	// or "listening" (play audio_url, then the options). From settings.device_mode.
	Mode string `json:"mode"`
	// Cloze (mode "cloze"): the card's example sentence with the word replaced
	// by "____" and 4 words to pick from (correct one has id == card id).
	// The device never types, it only picks.
	Cloze *deviceCloze `json:"cloze,omitempty"`
}

type deviceCloze struct {
	// Kind: "word" (pick the word: 4 different words) or "form" (TOEIC Part 5:
	// 4 forms of the same word family, e.g. fluctuate / fluctuated / fluctuation).
	Kind          string         `json:"kind"`
	Sentence      string         `json:"sentence"`
	TranslationVi string         `json:"translation_vi,omitempty"` // shown after answering
	Options       []study.Choice `json:"options"`
}

// DeviceSync returns everything the device needs to work offline for a while:
// settings, the current study queue in a compact format, and counters.
// ?limit= (default 30, max 100) caps the batch for the device's memory.
func (h *Handler) DeviceSync(c *gin.Context) {
	uid, ok := h.pairedUser(c)
	if !ok {
		return
	}
	limit := queryInt(c, "limit", 30)
	if limit < 1 || limit > 100 {
		limit = 30
	}
	now := time.Now().UTC()
	items, err := h.Study.Queue(uid, study.Filter{}, limit, now)
	if err != nil {
		internal(c, err)
		return
	}
	var u models.User
	h.DB.Select("id, settings, timezone").First(&u, uid)
	settings := study.ParseSettings(u.Settings)

	clozeWords := h.clozeWords(items, settings.DeviceMode)
	clozeForms := h.clozeFormsFor(items, settings.DeviceMode)
	cards := make([]deviceCard, 0, len(items))
	nth := 0 // cards that can be quizzed, for "mix"
	for _, it := range items {
		cc := content.ToCompact(it.Card)
		cloze := buildCloze(it, cc, clozeWords, clozeForms)
		mode := study.DeviceCardMode(settings.DeviceMode, study.DeviceCard{
			Kind: it.Card.Kind, IsNew: it.IsNew, HasOptions: len(it.Choices) >= 2, HasAudio: cc.AudioURL != "", HasCloze: cloze != nil,
		}, nth)
		if !it.IsNew {
			nth++
		}
		dc := deviceCard{Compact: cc, IsNew: it.IsNew, State: it.Progress.State, Due: it.Progress.Due, Options: it.Choices, Mode: mode}
		if mode == study.DeviceModeCloze {
			dc.Cloze = cloze
		}
		cards = append(cards, dc)
	}
	stats, err := h.Study.Stats(uid, now)
	if err != nil {
		internal(c, err)
		return
	}
	loc := study.UserLocation(u.Timezone)
	_, offset := now.In(loc).Zone()

	c.JSON(http.StatusOK, gin.H{
		"server_time":    now,
		"utc_offset_sec": offset,
		"settings":       settings,
		"due_now":        stats.DueNow,
		"new_left_today": stats.NewLeftToday,
		"reviews_today":  stats.ReviewsToday,
		"streak_days":    stats.StreakDays,
		"cards":          cards,
	})
}

// DeviceReviews receives answers recorded on the device (possibly offline).
// Each review carries a client_review_id so re-sending the same batch after a
// network error is safe. Reviews are applied in the order they happened.
func (h *Handler) DeviceReviews(c *gin.Context) {
	uid, ok := h.pairedUser(c)
	if !ok {
		return
	}
	var req dto.DeviceReviewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	reviews := req.Reviews
	sortByTime(reviews)

	devID := device(c).ID
	accepted, duplicates := 0, 0
	type failure struct {
		ClientReviewID string `json:"client_review_id"`
		Error          string `json:"error"`
	}
	failed := []failure{}
	for _, r := range reviews {
		cid := r.ClientReviewID
		res, err := h.Study.ApplyAnswer(study.Answer{
			UserID: uid, CardID: r.CardID, Rating: r.Rating, At: r.ReviewedAt,
			DurationMs: r.DurationMs, Source: "device", DeviceID: &devID, ClientReviewID: &cid,
		})
		switch {
		case err != nil:
			failed = append(failed, failure{ClientReviewID: cid, Error: err.Error()})
		case res.Duplicate:
			duplicates++
		default:
			accepted++
		}
	}
	c.JSON(http.StatusOK, gin.H{"accepted": accepted, "duplicates": duplicates, "failed": failed})
}

func sortByTime(rs []dto.DeviceReview) {
	for i := 1; i < len(rs); i++ {
		for j := i; j > 0 && rs[j].ReviewedAt.Before(rs[j-1].ReviewedAt); j-- {
			rs[j], rs[j-1] = rs[j-1], rs[j]
		}
	}
}

// ---------------------------------------------------------------- user-side device management

func (h *Handler) ListDevices(c *gin.Context) {
	ds := []models.Device{}
	if err := h.DB.Where("user_id = ?", userID(c)).Order("id").Find(&ds).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, ds)
}

func (h *Handler) PairDevice(c *gin.Context) {
	var req dto.PairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	var d models.Device
	err := h.DB.Where("pair_code = ? AND pair_code_expires_at > NOW() AND user_id IS NULL", req.Code).First(&d).Error
	if err != nil {
		abort(c, http.StatusNotFound, "code not found or expired: check the code on the device screen")
		return
	}
	uid := userID(c)
	updates := map[string]any{"user_id": uid, "pair_code": nil, "pair_code_expires_at": nil}
	if name := strings.TrimSpace(req.Name); name != "" {
		updates["name"] = name
	}
	if err := h.DB.Model(&d).Updates(updates).Error; err != nil {
		internal(c, err)
		return
	}
	h.DB.First(&d, d.ID)
	c.JSON(http.StatusOK, d)
}

func (h *Handler) RenameDevice(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.RenameDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res := h.DB.Model(&models.Device{}).Where("id = ? AND user_id = ?", id, userID(c)).Update("name", strings.TrimSpace(req.Name))
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		abort(c, http.StatusNotFound, "device not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// UnpairDevice deletes the device; it must register again to be used.
func (h *Handler) UnpairDevice(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	res := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).Delete(&models.Device{})
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		abort(c, http.StatusNotFound, "device not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// clozeWords loads the English words of the distractor cards (the cards behind
// each item's meaning options), used as the wrong words of cloze questions.
func (h *Handler) clozeWords(items []study.QueueItem, mode string) map[int64]string {
	out := map[int64]string{}
	if mode != study.DeviceModeCloze && mode != study.DeviceModeMix {
		return out
	}
	var ids []int64
	for _, it := range items {
		for _, ch := range it.Choices {
			if ch.ID != it.Card.ID {
				ids = append(ids, ch.ID)
			}
		}
	}
	if len(ids) == 0 {
		return out
	}
	var rows []struct {
		ID    int64
		Front string
	}
	h.DB.Raw(`SELECT id, front FROM cards WHERE id IN ?`, ids).Scan(&rows)
	for _, r := range rows {
		out[r.ID] = r.Front
	}
	return out
}

// buildCloze: the example sentence with the word blanked and 4 shuffled words
// (the card's own + the words of its 3 meaning distractors). nil if not possible.
func buildCloze(it study.QueueItem, cc content.Compact, words map[int64]string, forms map[string][]string) *deviceCloze {
	if it.IsNew || (it.Card.Kind != models.KindVocab && it.Card.Kind != models.KindPhrasal) || cc.ExampleEn == "" {
		return nil
	}
	sentence, ok := study.BlankWord(cc.ExampleEn, it.Card.Front)
	if !ok {
		return nil
	}
	// half of the time, when prepared: which FORM of the word fits (TOEIC Part 5)
	if f := forms[formKey(it.Card.ID, cc.ExampleEn)]; len(f) == 3 && rand.IntN(2) == 0 {
		if answer, ok := study.FindWord(cc.ExampleEn, it.Card.Front); ok {
			opts := []study.Choice{{ID: it.Card.ID, Text: answer}, {ID: -1, Text: f[0]}, {ID: -2, Text: f[1]}, {ID: -3, Text: f[2]}}
			rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
			return &deviceCloze{Kind: "form", Sentence: sentence, TranslationVi: cc.ExampleVi, Options: opts}
		}
	}
	opts := []study.Choice{{ID: it.Card.ID, Text: it.Card.Front}}
	seen := map[string]bool{strings.ToLower(it.Card.Front): true}
	for _, ch := range it.Choices {
		w := words[ch.ID]
		if ch.ID == it.Card.ID || w == "" || seen[strings.ToLower(w)] {
			continue
		}
		seen[strings.ToLower(w)] = true
		opts = append(opts, study.Choice{ID: ch.ID, Text: w})
	}
	if len(opts) < 4 {
		return nil
	}
	opts = opts[:4]
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	return &deviceCloze{Kind: "word", Sentence: sentence, TranslationVi: cc.ExampleVi, Options: opts}
}

func formKey(cardID int64, sentence string) string {
	return fmt.Sprintf("%d|%s", cardID, strings.ToLower(strings.TrimSpace(sentence)))
}

// clozeFormsFor: "Which form?" options already prepared (on the web) for the
// sentences of these cards, keyed by formKey. No AI call here.
func (h *Handler) clozeFormsFor(items []study.QueueItem, mode string) map[string][]string {
	out := map[string][]string{}
	if mode != study.DeviceModeCloze && mode != study.DeviceModeMix {
		return out
	}
	var ids []int64
	for _, it := range items {
		ids = append(ids, it.Card.ID)
	}
	if len(ids) == 0 {
		return out
	}
	var rows []struct {
		CardID      int64
		SentenceEn  string
		FormOptions []byte
	}
	h.DB.Raw(`SELECT card_id, sentence_en, form_options FROM card_examples
	          WHERE card_id IN ? AND form_options IS NOT NULL AND jsonb_array_length(form_options) = 3`, ids).Scan(&rows)
	for _, r := range rows {
		var f []string
		if json.Unmarshal(r.FormOptions, &f) == nil && len(f) == 3 {
			out[formKey(r.CardID, r.SentenceEn)] = f
		}
	}
	return out
}
