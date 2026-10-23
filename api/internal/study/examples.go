package study

import (
	"encoding/json"

	"mvp-lang-app/internal/models"
)

// RotateExamples shows a different example each time a word comes back: the
// main example of vocab/phrasal cards (the lowest-level sense, shown first) is
// replaced by a random AI example already stored for the card ("More
// examples"). Cards without stored AI examples keep their own example. No AI
// call is made here. The sense gets "example_id" so the web can leave that
// sentence out of the "More examples" list.
func (s *Service) RotateExamples(items []QueueItem) {
	var ids []int64
	for _, it := range items {
		if it.Card.Kind == models.KindVocab || it.Card.Kind == models.KindPhrasal {
			ids = append(ids, it.Card.ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	var rows []struct {
		ID            int64
		CardID        int64
		SentenceEn    string
		TranslationVi string
	}
	if err := s.DB.Raw(`
		SELECT DISTINCT ON (card_id) id, card_id, sentence_en, translation_vi
		FROM card_examples
		WHERE card_id IN ? AND source = 'ai' AND sentence_en <> ''
		ORDER BY card_id, random()`, ids).Scan(&rows).Error; err != nil {
		return // keep the cards' own examples
	}
	type pick struct {
		id     int64
		en, vi string
	}
	picks := make(map[int64]pick, len(rows))
	for _, r := range rows {
		picks[r.CardID] = pick{r.ID, r.SentenceEn, r.TranslationVi}
	}
	for i := range items {
		p, ok := picks[items[i].Card.ID]
		if !ok {
			continue
		}
		if payload, ok := withExample(items[i].Card.Payload, p.id, p.en, p.vi); ok {
			items[i].Card.Payload = payload
		}
	}
}

// withExample sets the example of the lowest-level sense in a card payload.
func withExample(raw models.JSON, id int64, en, vi string) (models.JSON, bool) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return raw, false
	}
	senses, _ := payload["senses"].([]any)
	best := -1
	bestLevel := 0.0
	for i, s := range senses {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		level, _ := m["level"].(float64)
		if best < 0 || level < bestLevel {
			best, bestLevel = i, level
		}
	}
	if best < 0 {
		return raw, false
	}
	m := senses[best].(map[string]any)
	m["example_en"] = en
	m["example_vi"] = vi
	m["example_id"] = id
	out, err := json.Marshal(payload)
	if err != nil {
		return raw, false
	}
	return models.JSON(out), true
}
