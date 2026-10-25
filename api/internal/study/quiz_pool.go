package study

import (
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
)

// QuizPool returns up to n random cards the user has already learned (any due
// date), for the device's Quiz button. Each has 4 meaning choices (mcq cards
// use their own choices). Answers go through ApplyAnswer with OnlyIfDue.
func (s *Service) QuizPool(userID int64, f Filter, n int) ([]QueueItem, error) {
	if n <= 0 || n > 100 {
		n = 40
	}
	type row struct {
		models.UserCard
		CDeckID  int64       `gorm:"column:c_deck_id"`
		CKind    string      `gorm:"column:c_kind"`
		CFront   string      `gorm:"column:c_front"`
		CPayload models.JSON `gorm:"column:c_payload"`
		CSource  string      `gorm:"column:c_source"`
	}
	var rows []row
	q := s.DB.Table("user_cards uc").
		Select("uc.*, c.deck_id AS c_deck_id, c.kind AS c_kind, c.front AS c_front, c.payload AS c_payload, c.source AS c_source").
		Joins("JOIN cards c ON c.id = uc.card_id").
		Where("uc.user_id = ? AND NOT uc.suspended AND uc.state <> ?", userID, srs.StateNew).
		Where("c.deck_id IN ("+VisibleDecksSQL+")", userID)
	if err := scope(q, f).Order("random()").Limit(n).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]QueueItem, 0, len(rows))
	for _, r := range rows {
		card := models.Card{ID: r.CardID, DeckID: r.CDeckID, Kind: r.CKind, Front: r.CFront, Payload: r.CPayload, Source: r.CSource}
		it := QueueItem{Card: card, Progress: r.UserCard}
		if card.Kind != models.KindMCQ {
			choices, err := s.MeaningChoices(card)
			if err != nil {
				return nil, err
			}
			if choices == nil {
				continue // deck too small for 4 different meanings
			}
			it.Choices = choices
		}
		out = append(out, it)
	}
	s.RotateExamples(out) // a different (AI) example sentence from time to time, like Study
	return out, nil
}
