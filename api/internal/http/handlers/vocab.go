package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"
)

func ListVocabs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var vocabs []models.Vocab
		topic := c.Query("topic")
		q := db
		if topic != "" {
			q = q.Where("topic = ?", topic)
		}
		if err := q.Limit(10).Find(&vocabs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, vocabs)
	}
}

func SelectVocab(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.SelectVocabRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
		for _, vid := range req.VocabIDs {
			uv := models.UserVocab{
				UserID:       req.UserID,
				VocabID:      vid,
				Difficulty:   "medium",
				LastResult:   false,
				Stability:    2.5,
				IntervalDays: 1,
				NextReviewAt: time.Now().Add(24 * time.Hour),
			}
			db.Create(&uv)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func GetTopics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var topics []string
		if err := db.Model(&models.Vocab{}).Distinct("topic").Pluck("topic", &topics).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, topics)
	}
}
