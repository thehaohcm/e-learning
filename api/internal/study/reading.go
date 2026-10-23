package study

import (
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
)

// ReadingWord is a vocabulary item chosen for a reading passage.
type ReadingWord struct {
	CardID    int64  `json:"card_id"`
	Word      string `json:"word"`
	MeaningVi string `json:"meaning_vi"`
	MeaningEn string `json:"meaning_en,omitempty"`
	IsNew     bool   `json:"is_new"`
}

// ReadingWords picks words for a passage: up to `learned` English words the
// user has already studied (the ones they forget most and that are due soonest
// first) plus up to `upcoming` words they will meet next. If the user has
// learned few words yet, more upcoming words are used instead.
func (s *Service) ReadingWords(userID int64, learned, upcoming int) ([]ReadingWord, error) {
	englishKinds := []string{models.KindVocab, models.KindPhrasal}

	var known []models.Card
	if err := s.DB.Table("cards c").Select("c.*").
		Joins("JOIN user_cards uc ON uc.card_id = c.id AND uc.user_id = ?", userID).
		Where("c.kind IN ? AND uc.state <> ? AND NOT uc.suspended", englishKinds, srs.StateNew).
		Order("uc.lapses DESC, uc.due ASC").Limit(learned).Scan(&known).Error; err != nil {
		return nil, err
	}

	want := upcoming + (learned - len(known)) // fill the gap with upcoming words
	var next []models.Card
	if want > 0 {
		// cards already in the list but not studied yet, then never-seen cards from the core deck
		if err := s.DB.Table("cards c").Select("c.*").
			Joins("JOIN user_cards uc ON uc.card_id = c.id AND uc.user_id = ?", userID).
			Where("c.kind IN ? AND uc.state = ? AND NOT uc.suspended", englishKinds, srs.StateNew).
			Order("uc.created_at, c.id").Limit(want).Scan(&next).Error; err != nil {
			return nil, err
		}
		if missing := want - len(next); missing > 0 {
			var more []models.Card
			if err := s.DB.Table("cards c").Select("c.*").
				Joins("JOIN decks d ON d.id = c.deck_id JOIN subjects sub ON sub.id = d.subject_id").
				Where("sub.slug = 'english' AND c.kind IN ? AND (d.owner_id IS NULL OR d.owner_id = ?)", englishKinds, userID).
				Where("NOT EXISTS (SELECT 1 FROM user_cards uc WHERE uc.user_id = ? AND uc.card_id = c.id)", userID).
				Order("c.id").Limit(missing).Scan(&more).Error; err != nil {
				return nil, err
			}
			next = append(next, more...)
		}
	}

	out := make([]ReadingWord, 0, len(known)+len(next))
	add := func(c models.Card, isNew bool) {
		cc := content.ToCompact(c)
		out = append(out, ReadingWord{CardID: c.ID, Word: c.Front, MeaningVi: cc.MeaningVi, MeaningEn: cc.MeaningEn, IsNew: isNew})
	}
	for _, c := range known {
		add(c, false)
	}
	for _, c := range next {
		add(c, true)
	}
	return out, nil
}
