package handlers

import (
	"net/http"

	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetQuizForUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var questions []models.QuizQuestion
		if err := db.Limit(5).Order("random()").Find(&questions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, questions)
	}
}

func SubmitQuizAttempt(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.QuizAttemptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}

		var q models.QuizQuestion
		if err := db.First(&q, req.QuestionID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		correct := (req.AnswerIdx == q.AnswerIdx)

		attempt := models.QuizAttempt{
			UserID:     req.UserID,
			QuestionID: req.QuestionID,
			Correct:    correct,
		}
		db.Create(&attempt)

		c.JSON(http.StatusOK, gin.H{"correct": correct, "correct_answer_idx": q.AnswerIdx})
	}
}
