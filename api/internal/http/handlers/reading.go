package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// readingDoc is what is stored in readings.content.
type readingDoc struct {
	ai.ReadingContent
	Vocab []study.ReadingWord `json:"vocab"`
}

type readingSummary struct {
	ID        int64     `json:"id"`
	Kind      string    `json:"kind"`
	Level     string    `json:"level"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Score     *int      `json:"score"`
	Total     *int      `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

type questionResult struct {
	Chosen      int    `json:"chosen"`
	Answer      int    `json:"answer"`
	Correct     bool   `json:"correct"`
	Explanation string `json:"explanation"`
	Evidence    string `json:"evidence"`
}

// CreateReading starts generating a passage or conversation in the background.
// Body: {"kind": "reading" | "conversation", "level": "easy" | "medium" | "ielts"}
func (h *Handler) CreateReading(c *gin.Context) {
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	var req struct {
		Kind  string `json:"kind" binding:"required,oneof=reading conversation"`
		Level string `json:"level" binding:"required,oneof=easy medium ielts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	uid := userID(c)

	var busy int64
	h.DB.Model(&models.Reading{}).Where("user_id = ? AND status = 'processing' AND created_at > NOW() - INTERVAL '5 minutes'", uid).Count(&busy)
	if busy > 0 {
		abort(c, http.StatusTooManyRequests, "a reading is already being created, please wait for it")
		return
	}

	learned, upcoming := 6, 3
	if req.Kind == "conversation" {
		learned, upcoming = 4, 2
	}
	words, err := h.Study.ReadingWords(uid, learned, upcoming)
	if err != nil {
		internal(c, err)
		return
	}
	if len(words) < 3 {
		abort(c, http.StatusBadRequest, "not enough English words yet: add some words to your learning list first")
		return
	}

	doc, _ := json.Marshal(readingDoc{Vocab: words})
	r := models.Reading{UserID: uid, Kind: req.Kind, Level: req.Level, Status: "processing", Content: models.JSON(doc), CreatedAt: time.Now().UTC()}
	if err := h.DB.Create(&r).Error; err != nil {
		internal(c, err)
		return
	}
	go h.generateReading(r.ID, req.Kind, req.Level, words)
	c.JSON(http.StatusAccepted, readingSummary{ID: r.ID, Kind: r.Kind, Level: r.Level, Status: r.Status, CreatedAt: r.CreatedAt})
}

func (h *Handler) generateReading(id int64, kind, level string, words []study.ReadingWord) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	targets := make([]ai.TargetWord, len(words))
	for i, w := range words {
		meaning := w.MeaningVi
		if meaning == "" {
			meaning = w.MeaningEn
		}
		targets[i] = ai.TargetWord{Word: w.Word, MeaningVi: meaning, IsNew: w.IsNew}
	}
	now := time.Now().UTC()
	rc, err := h.AI.GenerateReading(ctx, kind, level, targets)
	if err != nil {
		log.Printf("[ai] reading %d: %v", id, err)
		h.DB.Model(&models.Reading{}).Where("id = ?", id).Updates(map[string]any{
			"status": "failed", "error_message": "AI could not create this reading, please try again", "completed_at": now,
		})
		return
	}
	doc, _ := json.Marshal(readingDoc{ReadingContent: *rc, Vocab: words})
	h.DB.Model(&models.Reading{}).Where("id = ?", id).Updates(map[string]any{
		"status": "completed", "title": rc.Title, "content": models.JSON(doc), "completed_at": now,
	})
}

func (h *Handler) ListReadings(c *gin.Context) {
	rows := []readingSummary{} // never null in JSON
	if err := h.DB.Model(&models.Reading{}).
		Select("id, kind, level, title, status, score, total, created_at").
		Where("user_id = ?", userID(c)).Order("created_at DESC").Limit(50).Scan(&rows).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) loadReading(c *gin.Context) (*models.Reading, *readingDoc, bool) {
	id, ok := paramID(c, "id")
	if !ok {
		return nil, nil, false
	}
	var r models.Reading
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&r).Error; err != nil {
		notFoundOr500(c, err, "reading")
		return nil, nil, false
	}
	var doc readingDoc
	_ = json.Unmarshal(r.Content, &doc)
	return &r, &doc, true
}

// GetReading returns the reading. Correct answers are hidden until the quiz is submitted.
func (h *Handler) GetReading(c *gin.Context) {
	r, doc, ok := h.loadReading(c)
	if !ok {
		return
	}
	var results []questionResult
	if r.AnsweredAt != nil {
		var chosen []int
		_ = json.Unmarshal(r.Answers, &chosen)
		results = gradeReading(doc.Questions, chosen)
	} else {
		for i := range doc.Questions {
			doc.Questions[i].Answer = -1
			doc.Questions[i].Explanation = ""
			doc.Questions[i].Evidence = ""
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"id": r.ID, "kind": r.Kind, "level": r.Level, "title": r.Title, "status": r.Status,
		"error_message": r.ErrorMessage, "created_at": r.CreatedAt,
		"score": r.Score, "total": r.Total, "content": doc, "results": results,
	})
}

// AnswerReading grades the quiz. Questions about a vocabulary word also count
// as a review of that word when it is due (correct = Good, wrong = Again).
func (h *Handler) AnswerReading(c *gin.Context) {
	r, doc, ok := h.loadReading(c)
	if !ok {
		return
	}
	if r.Status != "completed" {
		abort(c, http.StatusConflict, "reading is not ready")
		return
	}
	var req struct {
		Answers []int `json:"answers" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if len(req.Answers) != len(doc.Questions) {
		abort(c, http.StatusBadRequest, "answers must have one entry per question (-1 = skipped)")
		return
	}
	if r.AnsweredAt != nil { // already submitted: just show the stored result
		var chosen []int
		_ = json.Unmarshal(r.Answers, &chosen)
		res := gradeReading(doc.Questions, chosen)
		c.JSON(http.StatusOK, gin.H{"score": r.Score, "total": r.Total, "results": res})
		return
	}

	results := gradeReading(doc.Questions, req.Answers)
	score := 0
	for i, res := range results {
		if res.Correct {
			score++
		}
		word := strings.TrimSpace(doc.Questions[i].Word)
		if word == "" {
			continue
		}
		for _, v := range doc.Vocab {
			if strings.EqualFold(v.Word, word) {
				rating := srs.Again
				if res.Correct {
					rating = srs.Good
				}
				if _, err := h.Study.ApplyAnswer(study.Answer{UserID: r.UserID, CardID: v.CardID, Rating: rating,
					At: time.Now(), Source: "reading", OnlyIfDue: true}); err != nil {
					log.Printf("[reading] srs update card %d: %v", v.CardID, err)
				}
				break
			}
		}
	}
	total := len(results)
	answers, _ := json.Marshal(req.Answers)
	now := time.Now().UTC()
	if err := h.DB.Model(&models.Reading{}).Where("id = ?", r.ID).Updates(map[string]any{
		"answers": models.JSON(answers), "score": score, "total": total, "answered_at": now,
	}).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"score": score, "total": total, "results": results})
}

func gradeReading(qs []ai.ReadingQuestion, chosen []int) []questionResult {
	out := make([]questionResult, len(qs))
	for i, q := range qs {
		ch := -1
		if i < len(chosen) {
			ch = chosen[i]
		}
		out[i] = questionResult{Chosen: ch, Answer: q.Answer, Correct: ch == q.Answer, Explanation: q.Explanation, Evidence: q.Evidence}
	}
	return out
}

func (h *Handler) DeleteReading(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	res := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).Delete(&models.Reading{})
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		abort(c, http.StatusNotFound, "reading not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
