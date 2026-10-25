package handlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
)

// maxPracticeWords: words per "Practise these now" (one AI request at most).
const maxPracticeWords = 20

type wordQuestionRow struct {
	ID          int64
	CardID      int64
	Kind        string
	Prompt      string
	Choices     models.JSON
	Answer      int
	Explanation string
}

type deviceWordQuestion struct {
	CardID      int64    `json:"card_id"`
	Word        string   `json:"word"`
	Kind        string   `json:"kind"`
	Prompt      string   `json:"prompt"`
	Choices     []string `json:"choices"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"`
}

// DevicePractice ("Practise these now" on the device): one multiple-choice
// question per word (usage, closest meaning or collocation). Questions are made
// by AI only for words that have none yet (one request for all of them), then
// kept and reused. Answers are sent as quiz reviews: a wrong one makes the word due.
func (h *Handler) DevicePractice(c *gin.Context) {
	uid, ok := h.pairedUser(c)
	if !ok {
		return
	}
	h.servePracticeQuestions(c, uid)
}

// WordQuestions: POST /api/practice/word-questions {"card_ids":[...]} (web):
// the same questions as the device (shared, made once).
func (h *Handler) WordQuestions(c *gin.Context) { h.servePracticeQuestions(c, userID(c)) }

// WordQuestionAnswer: POST /api/practice/word-questions/answer {"card_id":1,"correct":false}.
// A wrong answer makes the word due now (its memory state is not changed); a
// right one changes nothing.
func (h *Handler) WordQuestionAnswer(c *gin.Context) {
	var req struct {
		CardID  int64 `json:"card_id" binding:"required"`
		Correct bool  `json:"correct"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if !req.Correct {
		if _, err := h.Study.ApplyAnswer(study.Answer{UserID: userID(c), CardID: req.CardID, Rating: srs.Again, At: time.Now(),
			Source: "practice", OnlyIfDue: true, DueNowIfWrong: true}); err != nil {
			internal(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) servePracticeQuestions(c *gin.Context, uid int64) {
	var req struct {
		CardIDs []int64 `json:"card_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	out, warning, err := h.practiceQuestions(uid, req.CardIDs)
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": out, "warning": warning})
}

// practiceQuestions: one question per word (in the order asked), making the
// missing ones with AI in one request.
func (h *Handler) practiceQuestions(uid int64, cardIDs []int64) ([]deviceWordQuestion, string, error) {
	ids := uniqueIDs(cardIDs, maxPracticeWords)
	var cards []models.Card
	if err := h.DB.Where("id IN ? AND kind IN ? AND deck_id IN ("+study.VisibleDecksSQL+")",
		ids, []string{models.KindVocab, models.KindPhrasal}, uid).Find(&cards).Error; err != nil {
		return nil, "", err
	}
	have, err := h.wordQuestions(ids)
	if err != nil {
		return nil, "", err
	}

	warning := ""
	var need []models.Card
	for _, card := range cards {
		if len(have[card.ID]) == 0 {
			need = append(need, card)
		}
	}
	if len(need) > 0 {
		if !h.AI.Enabled() {
			warning = ai.ErrNotConfigured.Error()
		} else {
			inputs := make([]ai.WordQuestionInput, len(need))
			for i, card := range need {
				cc := content.ToCompact(card)
				inputs[i] = ai.WordQuestionInput{Word: card.Front, POS: cc.POS, MeaningVi: cc.MeaningVi, MeaningEn: cc.MeaningEn}
			}
			// not tied to the request: if the device gives up waiting, the questions
			// are still made and kept for next time
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			res, err := h.AI.GenerateWordQuestions(ctx, inputs)
			cancel()
			if err != nil {
				log.Printf("[ai] word questions: %v", err)
				warning = "AI could not prepare the questions right now."
			} else {
				for i, qs := range res {
					for _, q := range qs {
						ch, _ := json.Marshal(q.Choices)
						h.DB.Exec(`INSERT INTO word_questions (card_id, kind, prompt, choices, answer, explanation) VALUES (?, ?, ?, ?, ?, ?)`,
							need[i].ID, q.Kind, q.Prompt, models.JSON(ch), q.Answer, q.Explanation)
					}
				}
				if have, err = h.wordQuestions(ids); err != nil {
					return nil, "", err
				}
			}
		}
	}

	byID := map[int64]models.Card{}
	for _, card := range cards {
		byID[card.ID] = card
	}
	out := []deviceWordQuestion{}
	for _, id := range ids {
		card, ok := byID[id]
		qs := have[id]
		if !ok || len(qs) == 0 {
			continue
		}
		q := qs[rand.IntN(len(qs))] // a different kind from time to time
		var choices []string
		if json.Unmarshal(q.Choices, &choices) != nil || len(choices) != 4 {
			continue
		}
		// shuffle the choices so the right one is not always at the same place
		order := rand.Perm(4)
		shuffled := make([]string, 4)
		answer := 0
		for to, from := range order {
			shuffled[to] = choices[from]
			if from == q.Answer {
				answer = to
			}
		}
		out = append(out, deviceWordQuestion{CardID: card.ID, Word: card.Front, Kind: q.Kind, Prompt: q.Prompt,
			Choices: shuffled, Answer: answer, Explanation: q.Explanation})
	}
	return out, warning, nil
}

func (h *Handler) wordQuestions(cardIDs []int64) (map[int64][]wordQuestionRow, error) {
	var rows []wordQuestionRow
	err := h.DB.Raw(`SELECT id, card_id, kind, prompt, choices, answer, explanation FROM word_questions WHERE card_id IN ?`, cardIDs).
		Scan(&rows).Error
	out := map[int64][]wordQuestionRow{}
	for _, r := range rows {
		out[r.CardID] = append(out[r.CardID], r)
	}
	return out, err
}
