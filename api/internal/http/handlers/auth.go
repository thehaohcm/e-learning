package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/auth"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type userView struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Timezone    string `json:"timezone"`
}

func viewUser(u models.User) userView {
	return userView{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, Role: u.Role, Timezone: u.Timezone}
}

func (h *Handler) issue(c *gin.Context, u models.User, status int) {
	tok, exp, err := h.Tokens.Issue(u.ID, u.Role)
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(status, dto.AuthResponse{Token: tok, ExpiresAt: exp, User: viewUser(u)})
}

// roleFor decides the role at sign-up: emails in ADMIN_EMAILS are admins, and
// the very first account becomes admin when no admin exists yet.
func (h *Handler) roleFor(email string) string {
	if h.Cfg.IsAdminEmail(email) {
		return models.RoleAdmin
	}
	var admins int64
	h.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&admins)
	if admins == 0 {
		return models.RoleAdmin
	}
	return models.RoleUser
}

// Register creates an account. A legacy account (created by the old app,
// without a password) can be claimed once by registering with its email.
func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		internal(c, err)
		return
	}

	var u models.User
	err = h.DB.Where("lower(email) = ?", email).First(&u).Error
	switch {
	case err == nil && u.PasswordHash != "":
		abort(c, http.StatusConflict, "email already registered")
		return
	case err == nil: // legacy account without password: claim it
		u.PasswordHash = hash
		if req.DisplayName != "" {
			u.DisplayName = req.DisplayName
		}
		if h.Cfg.IsAdminEmail(email) {
			u.Role = models.RoleAdmin
		}
		u.UpdatedAt = time.Now().UTC()
		if err := h.DB.Save(&u).Error; err != nil {
			internal(c, err)
			return
		}
		h.issue(c, u, http.StatusOK)
		return
	case !errors.Is(err, gorm.ErrRecordNotFound):
		internal(c, err)
		return
	}

	if !h.Cfg.OpenSignup && !h.Cfg.IsAdminEmail(email) {
		abort(c, http.StatusForbidden, "sign-up is closed")
		return
	}
	u = models.User{
		Email: email, DisplayName: strings.TrimSpace(req.DisplayName), PasswordHash: hash,
		Role: h.roleFor(email), Timezone: h.Cfg.DefaultTZ, Settings: models.JSON("{}"),
	}
	if err := h.DB.Create(&u).Error; err != nil {
		internal(c, err)
		return
	}
	h.issue(c, u, http.StatusCreated)
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	var u models.User
	err := h.DB.Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&u).Error
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		if err == nil && u.PasswordHash == "" {
			abort(c, http.StatusUnauthorized, "this account has no password yet: use Register with the same email to set one")
			return
		}
		abort(c, http.StatusUnauthorized, "wrong email or password")
		return
	}
	if h.Cfg.IsAdminEmail(u.Email) && u.Role != models.RoleAdmin {
		u.Role = models.RoleAdmin
		h.DB.Model(&u).Update("role", u.Role)
	}
	h.issue(c, u, http.StatusOK)
}

func (h *Handler) Me(c *gin.Context) {
	var u models.User
	if err := h.DB.First(&u, userID(c)).Error; err != nil {
		abort(c, http.StatusUnauthorized, "account not found")
		return
	}
	c.JSON(http.StatusOK, viewUser(u))
}
