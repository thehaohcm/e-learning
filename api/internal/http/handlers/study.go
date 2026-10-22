package handlers

import (
	"errors"
	"net/http"
	"time"

	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Enroll(c *gin.Context) {
	var req dto.EnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	// only enroll ids that exist
	var ids []int64
	h.DB.Model(&models.Card{}).Where("id IN ?", req.CardIDs).Pluck("id", &ids)
	n, err := h.Study.Enroll(userID(c), ids, time.Now().UTC())
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": n})
}

// Queue returns the cards to study now (?deck_id, ?subject_id, ?limit).
func (h *Handler) Queue(c *gin.Context) {
	items, err := h.Study.Queue(userID(c), filterFromQuery(c), queryInt(c, "limit", 20), time.Now().UTC())
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) Review(c *gin.Context) {
	var req dto.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res, err := h.Study.ApplyAnswer(study.Answer{
		UserID: userID(c), CardID: req.CardID, Rating: req.Rating,
		At: time.Now().UTC(), DurationMs: req.DurationMs, Source: "web",
	})
	if err != nil {
		notFoundOr500(c, err, "card")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Suspend(c *gin.Context) {
	var req dto.SuspendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res := h.DB.Model(&models.UserCard{}).
		Where("user_id = ? AND card_id = ?", userID(c), req.CardID).
		Updates(map[string]any{"suspended": req.Suspended, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		abort(c, http.StatusNotFound, "card is not in your learning list")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Quiz(c *gin.Context) {
	qs, err := h.Study.BuildQuiz(userID(c), filterFromQuery(c), queryInt(c, "count", 10))
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, qs)
}

func (h *Handler) QuizAnswer(c *gin.Context) {
	var req dto.QuizAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res, err := h.Study.CheckQuizAnswer(userID(c), req.CardID, req.ChoiceID)
	if err != nil {
		if errors.Is(err, study.ErrCardNotFound) {
			abort(c, http.StatusNotFound, "card not found")
			return
		}
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Stats(c *gin.Context) {
	st, err := h.Study.Stats(userID(c), time.Now().UTC())
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

// ---------------------------------------------------------------- settings

func (h *Handler) GetSettings(c *gin.Context) {
	var u models.User
	if err := h.DB.First(&u, userID(c)).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": study.ParseSettings(u.Settings), "timezone": u.Timezone, "display_name": u.DisplayName})
}

// UpdateSettings accepts a partial update: {"settings": {...}, "timezone": "...", "display_name": "..."}.
func (h *Handler) UpdateSettings(c *gin.Context) {
	var req struct {
		Settings    *map[string]any `json:"settings"`
		Timezone    *string         `json:"timezone"`
		DisplayName *string         `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	var u models.User
	if err := h.DB.First(&u, userID(c)).Error; err != nil {
		internal(c, err)
		return
	}
	updates := map[string]any{"updated_at": time.Now().UTC()}
	settings := study.ParseSettings(u.Settings)
	if req.Settings != nil {
		patch, _ := jsonMarshal(*req.Settings)
		merged, err := settings.Merge(patch)
		if err != nil {
			bad(c, err)
			return
		}
		settings = merged
		b, _ := jsonMarshal(settings)
		updates["settings"] = models.JSON(b)
	}
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil || *req.Timezone == "" {
			abort(c, http.StatusBadRequest, "unknown time zone")
			return
		}
		updates["timezone"] = *req.Timezone
		u.Timezone = *req.Timezone
	}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
		u.DisplayName = *req.DisplayName
	}
	if err := h.DB.Model(&models.User{}).Where("id = ?", u.ID).Updates(updates).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": settings, "timezone": u.Timezone, "display_name": u.DisplayName})
}
