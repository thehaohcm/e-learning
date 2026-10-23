// Package handlers implements the HTTP endpoints. Business logic lives in
// internal/study, internal/content and internal/ai; handlers only parse input,
// check access and format output.
package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/auth"
	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/push"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Cfg    config.Config
	AI     *ai.Client
	Tokens *auth.Issuer
	Study  *study.Service
	Push   *push.Store // nil if the VAPID key could not be loaded
}

func New(db *gorm.DB, cfg config.Config) *Handler {
	h := &Handler{
		DB:     db,
		Cfg:    cfg,
		AI:     ai.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL, cfg.DeepSeekModel),
		Tokens: auth.NewIssuer(cfg.JWTSecret),
		Study:  &study.Service{DB: db},
	}
	if store, err := push.NewStore(db, cfg.VAPIDSubject, cfg.PushAnyHost); err != nil {
		log.Printf("push notifications disabled: %v", err)
	} else {
		h.Push = store
	}
	return h
}

func (h *Handler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err == nil {
		err = sqlDB.PingContext(c.Request.Context())
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC(), "ai": h.AI.Enabled()})
}

// ---------------------------------------------------------------- context keys

const (
	ctxUserID   = "uid"
	ctxRole     = "role"
	ctxDeviceID = "device_id"
	ctxDevice   = "device"
)

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// RequireUser accepts a web JWT.
func (h *Handler) RequireUser(c *gin.Context) {
	tok := bearer(c)
	if tok == "" || strings.HasPrefix(tok, "dev_") || strings.HasPrefix(tok, "ext_") {
		abort(c, http.StatusUnauthorized, "login required")
		return
	}
	claims, err := h.Tokens.Parse(tok)
	if err != nil {
		abort(c, http.StatusUnauthorized, "invalid or expired session")
		return
	}
	c.Set(ctxUserID, claims.UserID)
	c.Set(ctxRole, claims.Role)
	c.Next()
}

// RequireAdmin must run after RequireUser. The role is re-read from the DB so
// demoting an admin takes effect immediately.
func (h *Handler) RequireAdmin(c *gin.Context) {
	var role string
	h.DB.Model(&models.User{}).Where("id = ?", userID(c)).Pluck("role", &role)
	if role != models.RoleAdmin {
		abort(c, http.StatusForbidden, "admin only")
		return
	}
	c.Next()
}

// RequireDevice accepts a device token ("dev_..."). Unpaired devices are let
// through; handlers that need a user call pairedUser.
func (h *Handler) RequireDevice(c *gin.Context) {
	tok := bearer(c)
	if !strings.HasPrefix(tok, "dev_") {
		abort(c, http.StatusUnauthorized, "device token required")
		return
	}
	var d models.Device
	if err := h.DB.Where("token_hash = ?", auth.HashDeviceToken(tok)).First(&d).Error; err != nil {
		abort(c, http.StatusUnauthorized, "unknown device")
		return
	}
	now := time.Now().UTC()
	if d.LastSeenAt == nil || now.Sub(*d.LastSeenAt) > time.Minute {
		h.DB.Model(&models.Device{}).Where("id = ?", d.ID).Update("last_seen_at", now)
	}
	c.Set(ctxDeviceID, d.ID)
	c.Set(ctxDevice, d)
	c.Next()
}

func userID(c *gin.Context) int64 { return c.GetInt64(ctxUserID) }

func device(c *gin.Context) models.Device {
	d, _ := c.Get(ctxDevice)
	dev, _ := d.(models.Device)
	return dev
}

// ---------------------------------------------------------------- helpers

type apiError struct {
	Error string `json:"error"`
}

func abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, apiError{Error: msg})
}

func bad(c *gin.Context, err error) {
	abort(c, http.StatusBadRequest, err.Error())
}

// internal logs the real error and returns a generic message.
func internal(c *gin.Context, err error) {
	log.Printf("[error] %s %s: %v", c.Request.Method, c.FullPath(), err)
	abort(c, http.StatusInternalServerError, "internal error")
}

func notFoundOr500(c *gin.Context, err error, what string) {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, study.ErrCardNotFound) {
		abort(c, http.StatusNotFound, what+" not found")
		return
	}
	internal(c, err)
}

func paramID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		abort(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, name string, def int) int {
	if v, err := strconv.Atoi(c.Query(name)); err == nil {
		return v
	}
	return def
}

func queryInt64(c *gin.Context, name string) int64 {
	v, _ := strconv.ParseInt(c.Query(name), 10, 64)
	return v
}

func filterFromQuery(c *gin.Context) study.Filter {
	return study.Filter{SubjectID: queryInt64(c, "subject_id"), DeckID: queryInt64(c, "deck_id")}
}

// ---------------------------------------------------------------- rate limiting

// RateLimit is a small fixed-window limiter per client IP, used on auth and
// device registration endpoints to slow down brute force.
func RateLimit(perMinute int) gin.HandlerFunc {
	type window struct {
		start time.Time
		n     int
	}
	var mu sync.Mutex
	hits := map[string]*window{}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		w := hits[ip]
		if w == nil || now.Sub(w.start) > time.Minute {
			w = &window{start: now}
			hits[ip] = w
		}
		w.n++
		over := w.n > perMinute
		if len(hits) > 10000 { // keep memory bounded
			for k, v := range hits {
				if now.Sub(v.start) > time.Minute {
					delete(hits, k)
				}
			}
		}
		mu.Unlock()
		if over {
			abort(c, http.StatusTooManyRequests, "too many requests, try again in a minute")
			return
		}
		c.Next()
	}
}
