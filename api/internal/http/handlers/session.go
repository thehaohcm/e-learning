package handlers

import (
    "net/http"
    "time"

    "mvp-lang-app/internal/http/dto"
    "mvp-lang-app/internal/models"
    "mvp-lang-app/internal/srs"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func SubmitDifficulty(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req dto.FeedbackRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
            return
        }

        var uv models.UserVocab
        if err := db.Where("user_id = ? AND vocab_id = ?", req.UserID, req.VocabID).First(&uv).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "user_vocab not found"})
            return
        }

        rating := map[string]int{"easy":5,"medium":4,"hard":2}[req.Difficulty]
        nextInterval, nextStability, nextReview := srs.Update(uv.IntervalDays, uv.Stability, rating)

        uv.Difficulty = req.Difficulty
        uv.LastResult = req.Correct
        uv.IntervalDays = nextInterval
        uv.Stability = nextStability
        uv.NextReviewAt = nextReview
        uv.UpdatedAt = time.Now()
        if err := db.Save(&uv).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "next_interval_days": nextInterval,
            "next_review_at":     nextReview,
            "stability":          nextStability,
        })
    }
}

func GetDashboard(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // MVP: fixed userID=1 for demo
        userID := int64(1)
        type Summary struct {
            Learned int64
            Weak    int64
            NextDue int64
        }
        var learned int64
        db.Model(&models.UserVocab{}).Where("user_id = ?", userID).Count(&learned)
        var weak int64
        db.Model(&models.UserVocab{}).Where("user_id = ? AND difficulty = 'hard'", userID).Count(&weak)
        var due int64
        db.Model(&models.UserVocab{}).Where("user_id = ? AND next_review_at <= NOW()", userID).Count(&due)

        c.JSON(http.StatusOK, gin.H{
            "learned": learned,
            "weak": weak,
            "due": due,
        })
    }
}

