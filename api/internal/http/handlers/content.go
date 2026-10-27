package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/http/dto"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/study"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type deckView struct {
	models.Deck
	CardCount     int `json:"card_count"`
	LearningCount int `json:"learning_count"`
}

type subjectView struct {
	models.Subject
	Decks []deckView `json:"decks"`
}

// ListSubjects returns subjects with their decks and card counts.
func (h *Handler) ListSubjects(c *gin.Context) {
	subjects := []models.Subject{}
	if err := h.DB.Order("sort_order, id").Find(&subjects).Error; err != nil {
		internal(c, err)
		return
	}
	decks := []deckView{}
	if err := h.DB.Raw(`
		SELECT d.*,
		       (SELECT COUNT(*) FROM cards c WHERE c.deck_id = d.id) AS card_count,
		       (SELECT COUNT(*) FROM user_cards uc JOIN cards c ON c.id = uc.card_id
		         WHERE c.deck_id = d.id AND uc.user_id = ? AND NOT uc.suspended) AS learning_count
		FROM decks d WHERE d.owner_id IS NULL OR d.owner_id = ?
		ORDER BY d.sort_order, d.id`, userID(c), userID(c)).Scan(&decks).Error; err != nil {
		internal(c, err)
		return
	}
	out := make([]subjectView, 0, len(subjects))
	for _, s := range subjects {
		sv := subjectView{Subject: s, Decks: []deckView{}}
		for _, d := range decks {
			if d.SubjectID == s.ID {
				sv.Decks = append(sv.Decks, d)
			}
		}
		out = append(out, sv)
	}
	c.JSON(http.StatusOK, out)
}

type cardWithProgress struct {
	models.Card
	Progress *models.UserCard `json:"progress"`
}

// ListCards supports ?deck_id, ?subject_id, ?q (search), ?page, ?limit.
func (h *Handler) ListCards(c *gin.Context) {
	f := filterFromQuery(c)
	page := max(1, queryInt(c, "page", 1))
	limit := queryInt(c, "limit", 30)
	if limit < 1 || limit > 100 {
		limit = 30
	}
	search := strings.TrimSpace(c.Query("q"))
	build := func() *gorm.DB {
		q := h.DB.Table("cards c").Where("c.deck_id IN ("+study.VisibleDecksSQL+")", userID(c))
		if f.DeckID > 0 {
			q = q.Where("c.deck_id = ?", f.DeckID)
		}
		if f.SubjectID > 0 {
			q = q.Where("c.deck_id IN (SELECT id FROM decks WHERE subject_id = ?)", f.SubjectID)
		}
		if search != "" {
			like := "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(search) + "%"
			q = q.Where("(c.front ILIKE ? OR c.payload::text ILIKE ?)", like, like)
		}
		return q
	}

	var total int64
	if err := build().Count(&total).Error; err != nil {
		internal(c, err)
		return
	}
	cards := []models.Card{}
	q := build().Select("c.*")
	if s := search; s != "" {
		// exact matches first, then prefix matches, then the rest
		q = q.Order(clause.OrderBy{Expression: clause.Expr{
			SQL:                "(lower(c.front) = lower(?)) DESC, (lower(c.front) LIKE lower(?)) DESC, c.id ASC",
			Vars:               []any{s, s + "%"},
			WithoutParentheses: true,
		}})
	} else {
		q = q.Order("c.id ASC")
	}
	if err := q.Limit(limit).Offset((page - 1) * limit).Scan(&cards).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": h.attachProgress(userID(c), cards), "total": total, "page": page, "limit": limit})
}

func (h *Handler) attachProgress(uid int64, cards []models.Card) []cardWithProgress {
	ids := make([]int64, len(cards))
	for i, cd := range cards {
		ids[i] = cd.ID
	}
	var progress []models.UserCard
	if len(ids) > 0 {
		h.DB.Where("user_id = ? AND card_id IN ?", uid, ids).Find(&progress)
	}
	byCard := map[int64]*models.UserCard{}
	for i := range progress {
		byCard[progress[i].CardID] = &progress[i]
	}
	out := make([]cardWithProgress, len(cards))
	for i, cd := range cards {
		out[i] = cardWithProgress{Card: cd, Progress: byCard[cd.ID]}
	}
	return out
}

func (h *Handler) GetCard(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var card models.Card
	if err := h.DB.Where("deck_id IN ("+study.VisibleDecksSQL+")", userID(c)).First(&card, id).Error; err != nil {
		notFoundOr500(c, err, "card")
		return
	}
	c.JSON(http.StatusOK, h.attachProgress(userID(c), []models.Card{card})[0])
}

// ---------------------------------------------------------------- admin: content management

func (h *Handler) CreateSubject(c *gin.Context) {
	var req dto.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	s := models.Subject{Slug: strings.ToLower(req.Slug), Name: req.Name, Description: req.Description, SortOrder: req.SortOrder}
	if err := h.DB.Create(&s).Error; err != nil {
		abort(c, http.StatusConflict, "subject slug already exists")
		return
	}
	c.JSON(http.StatusCreated, s)
}

// UpdateSubject: PATCH-like PUT /subjects/:id (name, description, order). Admin only.
func (h *Handler) UpdateSubject(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		abort(c, http.StatusBadRequest, "nothing to update")
		return
	}
	var s models.Subject
	if err := h.DB.First(&s, id).Error; err != nil {
		notFoundOr500(c, err, "subject")
		return
	}
	if err := h.DB.Model(&s).Updates(updates).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

// UpdateDeck: PUT /decks/:id (name, description, default_kind, order). Admin only.
func (h *Handler) UpdateDeck(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DefaultKind != nil {
		updates["default_kind"] = *req.DefaultKind
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if len(updates) == 0 {
		abort(c, http.StatusBadRequest, "nothing to update")
		return
	}
	var d models.Deck
	if err := h.DB.First(&d, id).Error; err != nil {
		notFoundOr500(c, err, "deck")
		return
	}
	if d.OwnerID != nil {
		abort(c, http.StatusForbidden, "personal decks cannot be edited here")
		return
	}
	if err := h.DB.Model(&d).Updates(updates).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) CreateDeck(c *gin.Context) {
	var req dto.CreateDeckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	d := models.Deck{SubjectID: req.SubjectID, Slug: strings.ToLower(req.Slug), Name: req.Name, Description: req.Description, DefaultKind: req.DefaultKind, SortOrder: req.SortOrder}
	if err := h.DB.Create(&d).Error; err != nil {
		abort(c, http.StatusConflict, "deck already exists or subject not found")
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) CreateCard(c *gin.Context) {
	var req dto.CardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	var deck models.Deck
	if err := h.DB.First(&deck, req.DeckID).Error; err != nil {
		notFoundOr500(c, err, "deck")
		return
	}
	kind := req.Kind
	if kind == "" {
		kind = deck.DefaultKind
	}
	if !content.ValidKind(kind) {
		abort(c, http.StatusBadRequest, "invalid kind")
		return
	}
	payload, err := content.Normalize(kind, req.Front, req.Payload)
	if err != nil {
		bad(c, err)
		return
	}
	card := models.Card{DeckID: deck.ID, Kind: kind, Front: strings.TrimSpace(req.Front), Payload: models.JSON(payload), Source: "manual"}
	if err := h.DB.Create(&card).Error; err != nil {
		abort(c, http.StatusConflict, "a card with this front already exists in the deck")
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (h *Handler) UpdateCard(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.CardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	var card models.Card
	if err := h.DB.First(&card, id).Error; err != nil {
		notFoundOr500(c, err, "card")
		return
	}
	payload, err := content.Normalize(card.Kind, req.Front, req.Payload)
	if err != nil {
		bad(c, err)
		return
	}
	card.Front = strings.TrimSpace(req.Front)
	card.Payload = models.JSON(payload)
	card.UpdatedAt = time.Now().UTC()
	if err := h.DB.Save(&card).Error; err != nil {
		abort(c, http.StatusConflict, "a card with this front already exists in the deck")
		return
	}
	c.JSON(http.StatusOK, card)
}

// DeleteCard removes one card (and, by cascade, its progress and logs).
func (h *Handler) DeleteCard(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	res := h.DB.Delete(&models.Card{}, id)
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		abort(c, http.StatusNotFound, "card not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GenerateCards asks the AI for new cards and APPENDS them to the deck.
// Existing cards and everyone's progress are never touched.
func (h *Handler) GenerateCards(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	var req dto.GenerateRequest
	_ = c.ShouldBindJSON(&req)
	if req.Count <= 0 {
		req.Count = 10
	}
	req.Count = min(req.Count, 20)

	var deck models.Deck
	if err := h.DB.First(&deck, id).Error; err != nil {
		notFoundOr500(c, err, "deck")
		return
	}
	var subject models.Subject
	h.DB.First(&subject, deck.SubjectID)

	var existing []string
	h.DB.Model(&models.Card{}).Where("deck_id = ?", deck.ID).Order("id DESC").Limit(300).Pluck("front", &existing)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	generated, err := h.AI.GenerateCards(ctx, ai.DeckContext{
		SubjectName: subject.Name, DeckName: deck.Name, Description: deck.Description, Kind: deck.DefaultKind,
	}, req.Count, existing)
	if err != nil {
		log.Printf("[ai] generate deck=%d: %v", deck.ID, err)
		abort(c, http.StatusBadGateway, "AI generation failed, please try again")
		return
	}

	created := []models.Card{}
	skipped := 0
	for _, g := range generated {
		front := strings.TrimSpace(g.Front)
		payload, err := content.Normalize(deck.DefaultKind, front, g.Payload)
		if err != nil {
			skipped++
			continue
		}
		card := models.Card{DeckID: deck.ID, Kind: deck.DefaultKind, Front: front, Payload: models.JSON(payload), Source: "ai"}
		res := h.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&card)
		if res.Error != nil || res.RowsAffected == 0 {
			skipped++
			continue
		}
		created = append(created, card)
	}
	c.JSON(http.StatusOK, gin.H{"created": len(created), "skipped": skipped, "cards": created})
}
