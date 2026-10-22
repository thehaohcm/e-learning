// Package http wires routes and middleware.
package http

import (
	"time"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/http/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies(nil)
	r.MaxMultipartMemory = 1 << 20

	r.Use(cors.New(cors.Config{
		AllowOrigins:  cfg.CORSOrigins,
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	h := handlers.New(db, cfg)

	r.GET("/health", h.Health)
	api := r.Group("/api")

	// ---- public
	authLimit := handlers.RateLimit(20)
	api.POST("/auth/register", authLimit, h.Register)
	api.POST("/auth/login", authLimit, h.Login)

	// ---- study device (ESP32). Token: "Authorization: Bearer dev_..."
	api.POST("/device/register", handlers.RateLimit(10), h.DeviceRegister)
	dev := api.Group("/device", h.RequireDevice)
	{
		dev.GET("/status", h.DeviceStatus)
		dev.GET("/sync", h.DeviceSync)
		dev.POST("/reviews", h.DeviceReviews)
	}

	// ---- logged-in user (web)
	u := api.Group("", h.RequireUser)
	{
		u.GET("/auth/me", h.Me)

		u.GET("/subjects", h.ListSubjects)
		u.GET("/cards", h.ListCards)
		u.GET("/cards/:id", h.GetCard)

		u.POST("/study/enroll", h.Enroll)
		u.GET("/study/queue", h.Queue)
		u.POST("/study/review", h.Review)
		u.POST("/study/suspend", h.Suspend)

		u.GET("/quiz", h.Quiz)
		u.POST("/quiz/answer", h.QuizAnswer)

		u.GET("/stats", h.Stats)
		u.GET("/settings", h.GetSettings)
		u.PUT("/settings", h.UpdateSettings)

		u.GET("/devices", h.ListDevices)
		u.POST("/devices/pair", h.PairDevice)
		u.PATCH("/devices/:id", h.RenameDevice)
		u.DELETE("/devices/:id", h.UnpairDevice)

		u.POST("/writing", h.SubmitWriting)
		u.GET("/writing", h.ListWriting)
		u.GET("/writing/:id", h.GetWriting)
	}

	// ---- admin: manage content
	a := api.Group("", h.RequireUser, h.RequireAdmin)
	{
		a.POST("/subjects", h.CreateSubject)
		a.POST("/decks", h.CreateDeck)
		a.POST("/decks/:id/generate", h.GenerateCards)
		a.POST("/cards", h.CreateCard)
		a.PUT("/cards/:id", h.UpdateCard)
		a.DELETE("/cards/:id", h.DeleteCard)
	}

	return r
}
