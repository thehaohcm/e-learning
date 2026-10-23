package handlers

import (
	"errors"
	"net/http"
	"time"

	"mvp-lang-app/internal/push"

	"github.com/gin-gonic/gin"
)

type pushSubReq struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys"`
}

func (h *Handler) pushReady(c *gin.Context) bool {
	if h.Push == nil {
		abort(c, http.StatusServiceUnavailable, "notifications are not available on this server")
		return false
	}
	return true
}

// PushKey GET /api/push/key?endpoint=... → the VAPID public key, and whether
// this browser's subscription (endpoint) is stored for the user.
func (h *Handler) PushKey(c *gin.Context) {
	if !h.pushReady(c) {
		return
	}
	ep := c.Query("endpoint")
	c.JSON(http.StatusOK, gin.H{"public_key": h.Push.PublicKey(), "subscribed": ep != "" && h.Push.Has(userID(c), ep)})
}

// PushSubscribe POST /api/push/subscribe {endpoint, keys:{p256dh, auth}}: the
// JSON of PushSubscription from the browser.
func (h *Handler) PushSubscribe(c *gin.Context) {
	if !h.pushReady(c) {
		return
	}
	var req pushSubReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	sub := push.Subscription{Endpoint: req.Endpoint, P256dh: req.Keys.P256dh, Auth: req.Keys.Auth}
	if err := h.Push.Validate(sub); err != nil {
		abort(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.Push.Save(userID(c), sub, c.GetHeader("User-Agent")); err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PushUnsubscribe POST /api/push/unsubscribe {endpoint}
func (h *Handler) PushUnsubscribe(c *gin.Context) {
	if !h.pushReady(c) {
		return
	}
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if err := h.Push.Delete(userID(c), req.Endpoint); err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PushTest POST /api/push/test {endpoint}: sends a reminder now to this
// browser, with the real counts (or an "all done" note when nothing is due).
func (h *Handler) PushTest(c *gin.Context) {
	if !h.pushReady(c) {
		return
	}
	var req struct {
		Endpoint string `json:"endpoint" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	st, err := h.Study.Stats(userID(c), time.Now())
	if err != nil {
		internal(c, err)
		return
	}
	msg, _ := push.StudyReminder(st)
	n, err := h.Push.SendToUser(c.Request.Context(), userID(c), msg, req.Endpoint)
	if err != nil {
		abort(c, http.StatusBadGateway, "the push service refused the notification: "+err.Error())
		return
	}
	if n == 0 {
		abort(c, http.StatusNotFound, errors.New("this browser is not subscribed (or the subscription expired): turn notifications on again").Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": n, "message": msg})
}
