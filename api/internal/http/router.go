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
	r := gin.Default()

	// Build CORS origins from env
	corsOrigins := []string{
		"http://localhost:8085",
		"http://localhost:5173",
		"http://localhost:8080",
	}
	if frontendURL := cfg.FrontendURL; frontendURL != "" {
		corsOrigins = append(corsOrigins, frontendURL)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	h := &handlers.Handler{DB: db, Cfg: cfg}

	r.GET("/health", h.Health)

	api := r.Group("/api")
	{
		// Auth
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)

		// Vocabulary
		api.GET("/vocabs", h.ListVocabs)
		api.GET("/vocabs/:id", h.GetVocab)
		api.POST("/vocabs/generate-random", h.GenerateRandomVocabs)
		api.GET("/topics", h.GetTopics)

		// SRS
		api.POST("/srs/start", h.StartLearning)
		api.GET("/srs/due-reviews", h.GetDueReviews)
		api.POST("/srs/review", h.SubmitReview)

		// Quiz
		api.GET("/quiz", h.GetQuiz)
		api.POST("/quiz/answer", h.SubmitAnswer)

		// Writing
		api.POST("/writing/submit", h.SubmitWriting)
		api.GET("/writing/:id", h.GetWritingFeedback)
		api.GET("/writing/history", h.GetWritingHistory)

		// Dashboard
		api.GET("/dashboard", h.GetDashboard)

		// Search
		api.GET("/search", h.Search)

		// Audio
		api.POST("/audio/upload", h.UploadAudio)

		// Preferences
		api.GET("/preferences", h.GetPreferences)
		api.PUT("/preferences", h.UpdatePreferences)
	}

	return r
}
