package http

import (
    "mvp-lang-app/internal/http/handlers"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
    r := gin.Default()

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

