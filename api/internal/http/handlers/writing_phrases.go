package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// The writing phrase bank: the "writing-phrases" deck (migration 0018), concept
// cards whose front is what the learner wants to say (Vietnamese) and whose back
// is the English frame. Studied like any deck; also shown next to the Writing
// editor, and given to the AI so its feedback can point to a frame.

const phraseDeckSlug = "writing-phrases"

type phraseItem struct {
	ID        int64  `json:"id"`
	Vi        string `json:"vi"`      // what to say
	Pattern   string `json:"pattern"` // how to write it
	NoteVi    string `json:"note_vi,omitempty"`
	Example   string `json:"example,omitempty"`
	ExampleVi string `json:"example_vi,omitempty"`
	InStudy   bool   `json:"in_study"`
	order     int
}

type phraseGroup struct {
	Name  string       `json:"name"`
	Items []phraseItem `json:"items"`
}

// phraseCards returns the cards of the bank in their order, with their payload.
func (h *Handler) phraseCards() ([]models.Card, []content.ConceptPayload, error) {
	var cards []models.Card
	err := h.DB.Where(`kind = ? AND deck_id IN (
			SELECT d.id FROM decks d JOIN subjects s ON s.id = d.subject_id
			WHERE d.slug = ? AND s.slug = 'english' AND d.owner_id IS NULL)`, models.KindConcept, phraseDeckSlug).
		Order("id").Find(&cards).Error
	if err != nil {
		return nil, nil, err
	}
	ps := make([]content.ConceptPayload, len(cards))
	for i, c := range cards {
		_ = json.Unmarshal(c.Payload, &ps[i])
	}
	return cards, ps, nil
}

// phraseLines: "group: frame" for the AI (empty if the bank is missing).
func (h *Handler) phraseLines() []string {
	_, ps, err := h.phraseCards()
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		if p.Back != "" {
			out = append(out, p.Group+": "+p.Back)
		}
	}
	return out
}

// WritingPhrases: GET /api/writing/phrases
// {"deck_id": 7, "in_study": 12, "total": 81, "groups": [{"name": "Lý do", "items": [...]}]}
func (h *Handler) WritingPhrases(c *gin.Context) {
	cards, ps, err := h.phraseCards()
	if err != nil {
		internal(c, err)
		return
	}
	inStudy := map[int64]bool{}
	if len(cards) > 0 {
		ids := make([]int64, len(cards))
		for i, cd := range cards {
			ids[i] = cd.ID
		}
		var have []int64
		h.DB.Model(&models.UserCard{}).Where("user_id = ? AND card_id IN ?", userID(c), ids).Pluck("card_id", &have)
		for _, id := range have {
			inStudy[id] = true
		}
	}
	groups := []phraseGroup{}
	index := map[string]int{}
	var deckID int64
	for i, cd := range cards {
		p := ps[i]
		deckID = cd.DeckID
		name := p.Group
		if name == "" {
			name = "Other"
		}
		gi, ok := index[name]
		if !ok {
			gi = len(groups)
			index[name] = gi
			groups = append(groups, phraseGroup{Name: name})
		}
		groups[gi].Items = append(groups[gi].Items, phraseItem{
			ID: cd.ID, Vi: cd.Front, Pattern: p.Back, NoteVi: p.NoteVi, Example: p.Example, ExampleVi: p.ExampleVi,
			InStudy: inStudy[cd.ID], order: p.Order,
		})
	}
	// groups in the order of their first frame, frames in bank order
	for _, g := range groups {
		sort.SliceStable(g.Items, func(a, b int) bool { return g.Items[a].order < g.Items[b].order })
	}
	sort.SliceStable(groups, func(a, b int) bool { return groups[a].Items[0].order < groups[b].Items[0].order })
	c.JSON(http.StatusOK, gin.H{"deck_id": deckID, "groups": groups, "total": len(cards), "in_study": len(inStudy)})
}

// StudyWritingPhrases: POST /api/writing/phrases/study
// Adds every frame to the user's Study list (new cards: they come in with the
// daily new-card limit, like words).
func (h *Handler) StudyWritingPhrases(c *gin.Context) {
	cards, _, err := h.phraseCards()
	if err != nil {
		internal(c, err)
		return
	}
	ids := make([]int64, len(cards))
	for i, cd := range cards {
		ids[i] = cd.ID
	}
	added, err := h.Study.Enroll(userID(c), ids, time.Now().UTC())
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": added, "total": len(ids)})
}
