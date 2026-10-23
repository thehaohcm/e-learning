package handlers

import (
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/auth"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// Browser extension: "select a word on any page → Add to E-Learn notes".
// It authenticates with a key made in Settings ("ext_..."), which can only add
// and read the user's notes.

type extKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"-"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"-"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func (extKey) TableName() string { return "ext_keys" }

const maxExtKeys = 10

// ListExtKeys: GET /api/extension/keys
func (h *Handler) ListExtKeys(c *gin.Context) {
	keys := []extKey{}
	if err := h.DB.Where("user_id = ?", userID(c)).Order("created_at DESC").Find(&keys).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, keys)
}

// CreateExtKey: POST /api/extension/keys {"name":"Chrome at work"} → the key is
// shown ONCE ({"key": "ext_...", ...}); only its hash is stored.
func (h *Handler) CreateExtKey(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&req)
	uid := userID(c)
	var n int64
	h.DB.Model(&extKey{}).Where("user_id = ?", uid).Count(&n)
	if n >= maxExtKeys {
		abort(c, http.StatusBadRequest, "too many keys: remove one first")
		return
	}
	token, hash, err := auth.NewExtKey()
	if err != nil {
		internal(c, err)
		return
	}
	name := clip(strings.TrimSpace(req.Name), 60)
	if name == "" {
		name = "Browser extension"
	}
	k := extKey{UserID: uid, Name: name, TokenHash: hash, Prefix: token[:10], CreatedAt: time.Now().UTC()}
	if err := h.DB.Create(&k).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"key": token, "id": k.ID, "name": k.Name, "prefix": k.Prefix, "created_at": k.CreatedAt})
}

// DeleteExtKey: DELETE /api/extension/keys/:id — the extension stops working at once.
func (h *Handler) DeleteExtKey(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).Delete(&extKey{}).Error; err != nil {
		internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RequireExtKey accepts an extension key ("Authorization: Bearer ext_...").
func (h *Handler) RequireExtKey(c *gin.Context) {
	tok := bearer(c)
	if !strings.HasPrefix(tok, "ext_") {
		abort(c, http.StatusUnauthorized, "extension key required (Settings → Browser extension)")
		return
	}
	var k extKey
	if err := h.DB.Where("token_hash = ?", auth.HashDeviceToken(tok)).First(&k).Error; err != nil {
		abort(c, http.StatusUnauthorized, "this key was removed or is wrong: make a new one in Settings → Browser extension")
		return
	}
	now := time.Now().UTC()
	if k.LastUsedAt == nil || now.Sub(*k.LastUsedAt) > time.Minute {
		h.DB.Model(&extKey{}).Where("id = ?", k.ID).Update("last_used_at", now)
	}
	c.Set(ctxUserID, k.UserID)
	c.Next()
}

// ExtMe: GET /api/ext/me — "connected as ..." in the extension options.
func (h *Handler) ExtMe(c *gin.Context) {
	var u models.User
	if err := h.DB.Select("id, email, display_name").First(&u, userID(c)).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"email": u.Email, "display_name": u.DisplayName})
}

// ExtAddNote: POST /api/ext/notes {"text":"negotiate","context":"sentence","source":"page title","url":"..."}
func (h *Handler) ExtAddNote(c *gin.Context) {
	var req struct {
		noteRequest
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	src := strings.TrimSpace(req.Source)
	if src == "" {
		src = hostOf(req.URL)
	}
	if src == "" {
		src = "Web"
	}
	req.noteRequest.Source = src
	h.createNote(c, userID(c), req.noteRequest)
}

// ExtRecentNotes: GET /api/ext/notes — the last 8 notes, for the extension popup.
func (h *Handler) ExtRecentNotes(c *gin.Context) {
	notes := []models.WordNote{}
	if err := h.DB.Where("user_id = ?", userID(c)).Order("created_at DESC").Limit(8).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, notes)
}

func hostOf(u string) string {
	u = strings.TrimSpace(u)
	for _, p := range []string{"https://", "http://"} {
		u = strings.TrimPrefix(u, p)
	}
	if i := strings.IndexAny(u, "/?#"); i >= 0 {
		u = u[:i]
	}
	return clip(strings.TrimPrefix(u, "www."), 80)
}
