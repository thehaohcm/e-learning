package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

const maxPronBatch = 20

type pronView struct {
	Syllables []string `json:"syllables"`
	Stress    int      `json:"stress"`
	TipsVi    []string `json:"tips_vi"`
}

// Pronunciations: POST /api/cards/pronunciations {"card_ids":[1,2,3]}
// Returns {"items":{"<card id>":{syllables, stress, tips_vi}}} for the English
// cards that already have one. Never calls AI: the web app asks for a word
// explicitly with GeneratePronunciation ("How to say it" button).
func (h *Handler) Pronunciations(c *gin.Context) {
	var req struct {
		CardIDs []int64 `json:"card_ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	ids := uniqueIDs(req.CardIDs, maxPronBatch)
	var cards []models.Card
	if err := h.DB.Where("id IN ? AND kind IN ? AND deck_id IN ("+study.VisibleDecksSQL+")", ids, englishKinds, userID(c)).
		Find(&cards).Error; err != nil {
		internal(c, err)
		return
	}
	items := map[int64]pronView{}
	if len(cards) == 0 {
		c.JSON(http.StatusOK, gin.H{"items": items})
		return
	}
	stored, err := h.loadPronunciations(cards)
	if err != nil {
		internal(c, err)
		return
	}
	for id, p := range stored {
		items[id] = p
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GeneratePronunciation: POST /api/cards/:id/pronunciation
// One AI request for this word (only if it has none yet), kept for everyone.
func (h *Handler) GeneratePronunciation(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var card models.Card
	if err := h.DB.Where("kind IN ? AND deck_id IN ("+study.VisibleDecksSQL+")", englishKinds, userID(c)).
		First(&card, id).Error; err != nil {
		notFoundOr500(c, err, "word")
		return
	}
	stored, err := h.loadPronunciations([]models.Card{card})
	if err != nil {
		internal(c, err)
		return
	}
	if p, ok := stored[card.ID]; ok {
		c.JSON(http.StatusOK, p)
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()
	res, err := h.AI.GeneratePronunciations(ctx, []ai.PronWord{{Word: card.Front, IPA: content.ToCompact(card).IPA}})
	if err != nil || len(res) == 0 || res[0] == nil {
		log.Printf("[ai] pronunciation of %q: %v", card.Front, err)
		abort(c, http.StatusBadGateway, "AI could not explain this word right now, please try again")
		return
	}
	p := res[0]
	h.savePron(card.ID, p)
	c.JSON(http.StatusOK, pronView{Syllables: p.Syllables, Stress: p.Stress, TipsVi: p.TipsVi})
}

func (h *Handler) loadPronunciations(cards []models.Card) (map[int64]pronView, error) {
	ids := make([]int64, len(cards))
	for i, cd := range cards {
		ids[i] = cd.ID
	}
	var rows []models.CardPronunciation
	if err := h.DB.Where("card_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[int64]pronView{}
	for _, r := range rows {
		v := pronView{Stress: r.Stress, Syllables: []string{}, TipsVi: []string{}}
		_ = json.Unmarshal(r.Syllables, &v.Syllables)
		_ = json.Unmarshal(r.TipsVi, &v.TipsVi)
		out[r.CardID] = v
	}
	return out, nil
}

// QuizStress: POST /api/quiz/stress {card_id, syllable} — "Which syllable is
// stressed?" Grades it and updates the word's pronunciation schedule.
func (h *Handler) QuizStress(c *gin.Context) {
	var req struct {
		CardID   int64 `json:"card_id" binding:"required"`
		Syllable *int  `json:"syllable" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res, err := h.Study.CheckStress(userID(c), req.CardID, *req.Syllable, time.Now())
	if err != nil {
		notFoundOr500(c, err, "word")
		return
	}
	c.JSON(http.StatusOK, res)
}

// PronForgot: POST /api/pron/forgot {card_id} — "I forgot how to say it".
// Only the pronunciation schedule changes (the word comes back soon as a
// stress question); the meaning schedule is not touched.
func (h *Handler) PronForgot(c *gin.Context) {
	var req struct {
		CardID int64 `json:"card_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res, err := h.Study.ForgotPron(userID(c), req.CardID, time.Now())
	if err != nil {
		notFoundOr500(c, err, "word")
		return
	}
	c.JSON(http.StatusOK, res)
}

// PronStatus: GET /api/quiz/pron-status?deck_id= — words of the deck without
// a stored pronunciation (stress questions need one).
func (h *Handler) PronStatus(c *gin.Context) {
	missing, total, err := h.Study.PronMissing(userID(c), filterFromQuery(c))
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"missing": missing, "total": total, "can_generate": h.AI.Enabled()})
}

const maxPronGenerate = 25

// PronGenerate: POST /api/quiz/pron-generate {deck_id} — ONE AI request
// writes syllables + stress for up to 25 words of the deck that have none
// (kept for everyone). Pressed by the learner; call again for the next 25.
func (h *Handler) PronGenerate(c *gin.Context) {
	var req struct {
		DeckID int64 `json:"deck_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	var cards []models.Card
	if err := h.DB.Table("cards c").Select("c.*").
		Joins("LEFT JOIN card_pronunciations cp ON cp.card_id = c.id").
		Where("(cp.card_id IS NULL OR cp.sound IS NULL) AND c.deck_id = ? AND c.kind = ? AND c.front NOT LIKE '% %' AND c.deck_id IN ("+study.VisibleDecksSQL+")",
			req.DeckID, models.KindVocab, userID(c)).
		Order("c.id").Limit(maxPronGenerate).Scan(&cards).Error; err != nil {
		internal(c, err)
		return
	}
	created := 0
	if len(cards) > 0 {
		words := make([]ai.PronWord, len(cards))
		for i, cd := range cards {
			words[i] = ai.PronWord{Word: cd.Front, IPA: content.ToCompact(cd).IPA}
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 150*time.Second)
		defer cancel()
		res, err := h.AI.GeneratePronunciations(ctx, words)
		if err != nil {
			log.Printf("[ai] pronunciations for deck %d: %v", req.DeckID, err)
			abort(c, http.StatusBadGateway, "AI could not prepare the words right now, please try again")
			return
		}
		for i, p := range res {
			if i >= len(cards) || p == nil {
				continue
			}
			if h.savePron(cards[i].ID, p) {
				created++
			}
		}
	}
	missing, total, _ := h.Study.PronMissing(userID(c), study.Filter{DeckID: req.DeckID})
	c.JSON(http.StatusOK, gin.H{"created": created, "missing": missing, "total": total})
}

// savePron stores a word's pronunciation (kept if one exists) and its "Which
// sound?" data when still missing. Returns true if anything was written.
func (h *Handler) savePron(cardID int64, p *ai.Pronunciation) bool {
	syl, _ := json.Marshal(p.Syllables)
	tips, _ := json.Marshal(p.TipsVi)
	n := h.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.CardPronunciation{
		CardID: cardID, Syllables: models.JSON(syl), Stress: p.Stress, TipsVi: models.JSON(tips),
	}).RowsAffected
	sound := []byte("{}") // {} = nothing tricky in this word
	if p.Sound != nil {
		sound, _ = json.Marshal(p.Sound)
	}
	m := h.DB.Exec(`UPDATE card_pronunciations SET sound = ?::jsonb WHERE card_id = ? AND sound IS NULL`, string(sound), cardID).RowsAffected
	return n > 0 || m > 0
}

// QuizSound: POST /api/quiz/sound {card_id, option} — "How do you say the “u”
// in FLUC?" Grades it and updates the word's pronunciation schedule.
func (h *Handler) QuizSound(c *gin.Context) {
	var req struct {
		CardID int64 `json:"card_id" binding:"required"`
		Option *int  `json:"option" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	res, err := h.Study.CheckSound(userID(c), req.CardID, *req.Option, time.Now())
	if err != nil {
		notFoundOr500(c, err, "word")
		return
	}
	c.JSON(http.StatusOK, res)
}
