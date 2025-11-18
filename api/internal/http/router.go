package http

import (
	"mvp-lang-app/internal/http/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8085", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Accept-Language", "Connection", "Referer", "Sec-Fetch-Dest", "Sec-Fetch-Mode", "Sec-Fetch-Site", "User-Agent", "sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", handlers.Health())

	api := r.Group("/api")
	{
		api.GET("/vocabs", handlers.ListVocabs(db))
		api.POST("/vocabs/select", handlers.SelectVocab(db))
		api.GET("/quiz", handlers.GetQuizForUser(db))
		api.POST("/quiz/attempt", handlers.SubmitQuizAttempt(db))
		api.POST("/vocab/feedback", handlers.SubmitDifficulty(db))
		api.GET("/dashboard", handlers.GetDashboard(db))
	}

	return r
}
