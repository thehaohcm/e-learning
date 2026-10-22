package handlers

import (
	"encoding/json"
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

	cards := make([]deviceCard, 0, len(items))
	nth := 0 // cards that can be quizzed, for "mix"
	for _, it := range items {
		cc := content.ToCompact(it.Card)
		mode := study.DeviceCardMode(settings.DeviceMode, study.DeviceCard{
			Kind: it.Card.Kind, IsNew: it.IsNew, HasOptions: len(it.Choices) >= 2, HasAudio: cc.AudioURL != "",
		}, nth)
		if !it.IsNew {
			nth++
		}
		cards = append(cards, deviceCard{Compact: cc, IsNew: it.IsNew, State: it.Progress.State, Due: it.Progress.Due, Options: it.Choices, Mode: mode})
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
