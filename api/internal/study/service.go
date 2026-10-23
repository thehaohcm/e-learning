// Package study contains the learning logic shared by the web app and the
// device: building the study queue, applying answers with FSRS, and stats.
package study

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrCardNotFound = errors.New("card not found")

type Service struct{ DB *gorm.DB }

// Filter narrows the queue to one subject or deck (both optional).
type Filter struct {
	SubjectID int64
	DeckID    int64
}

// QueueItem is a card the user should study now.
type QueueItem struct {
	Card     models.Card       `json:"card"`
	Progress models.UserCard   `json:"progress"`
	IsNew    bool              `json:"is_new"`
	Preview  map[string]string `json:"preview"` // rating -> "10m", "3d"...
	// Choices: 4 meanings (the card's own one + 3 from the same deck) so the
	// client can check recall with a multiple-choice question instead of
	// self-rating. The correct choice has ID == card.id. Empty for mcq cards
	// (their own choices are in the payload) or when the deck is too small.
	Choices []Choice `json:"choices,omitempty"`
}

func (s *Service) loadUser(userID int64) (models.User, Settings, error) {
	var u models.User
	if err := s.DB.First(&u, userID).Error; err != nil {
		return u, Settings{}, err
	}
	return u, ParseSettings(u.Settings), nil
}

// UserLocation returns the user's time zone (falls back to Asia/Ho_Chi_Minh).
// The name is also passed to PostgreSQL (AT TIME ZONE), so it must be an IANA
// name, never a fixed offset like "UTC+7" (POSIX reads that as UTC-7).
func UserLocation(tz string) *time.Location {
	if loc, err := time.LoadLocation(tz); err == nil && tz != "" && tz != "Local" {
		return loc
	}
	if loc, err := time.LoadLocation("Asia/Ho_Chi_Minh"); err == nil {
		return loc
	}
	return time.UTC
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	l := t.In(loc)
	return time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, loc)
}

// VisibleDecksSQL selects the decks a user may see: shared decks and the user's
// own personal decks. Use as "deck_id IN (" + VisibleDecksSQL + ")" with the user id.
const VisibleDecksSQL = "SELECT id FROM decks WHERE owner_id IS NULL OR owner_id = ?"

// scope adds the subject/deck filter to a query that has `cards c` joined.
func scope(q *gorm.DB, f Filter) *gorm.DB {
	if f.DeckID > 0 {
		q = q.Where("c.deck_id = ?", f.DeckID)
	}
	if f.SubjectID > 0 {
		q = q.Where("c.deck_id IN (SELECT id FROM decks WHERE subject_id = ?)", f.SubjectID)
	}
	return q
}

// dayKey is the user's calendar day ("2026-10-22").
func dayKey(now time.Time, loc *time.Location) string {
	return startOfDay(now, loc).Format("2006-01-02")
}

// AddExtraNew allows n more new cards today (on top of the daily setting).
func (s *Service) AddExtraNew(userID int64, n int, now time.Time) (Settings, error) {
	u, settings, err := s.loadUser(userID)
	if err != nil {
		return settings, err
	}
	day := dayKey(now, UserLocation(u.Timezone))
	if settings.ExtraNewDate != day {
		settings.ExtraNewDate, settings.ExtraNew = day, 0
	}
	settings.ExtraNew = min(MaxExtraNew, settings.ExtraNew+n)
	b, err := json.Marshal(settings)
	if err != nil {
		return settings, err
	}
	err = s.DB.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{"settings": models.JSON(b), "updated_at": now}).Error
	return settings, err
}

// NewIntroducedToday counts cards answered for the first time today.
func (s *Service) NewIntroducedToday(userID int64, loc *time.Location, now time.Time) (int, error) {
	var n int64
	err := s.DB.Model(&models.ReviewLog{}).
		Where("user_id = ? AND state_before = ? AND reviewed_at >= ?", userID, srs.StateNew, startOfDay(now, loc)).
		Distinct("card_id").Count(&n).Error
	return int(n), err
}

// Queue returns due cards first (oldest due first), then new cards up to the
// remaining daily quota. New cards come from cards the user added manually
// and, if needed, are auto-introduced from the user's active decks.
func (s *Service) Queue(userID int64, f Filter, limit int, now time.Time) ([]QueueItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	u, settings, err := s.loadUser(userID)
	if err != nil {
		return nil, err
	}
	loc := UserLocation(u.Timezone)

	type row struct {
		models.UserCard
		CDeckID  int64       `gorm:"column:c_deck_id"`
		CKind    string      `gorm:"column:c_kind"`
		CFront   string      `gorm:"column:c_front"`
		CPayload models.JSON `gorm:"column:c_payload"`
		CSource  string      `gorm:"column:c_source"`
	}
	sel := "uc.*, c.deck_id AS c_deck_id, c.kind AS c_kind, c.front AS c_front, c.payload AS c_payload, c.source AS c_source"

	var due []row
	q := s.DB.Table("user_cards uc").Select(sel).
		Joins("JOIN cards c ON c.id = uc.card_id").
		Where("uc.user_id = ? AND NOT uc.suspended AND uc.state <> ? AND uc.due <= ?", userID, srs.StateNew, now)
	if err := scope(q, f).Order("uc.due ASC").Limit(limit).Scan(&due).Error; err != nil {
		return nil, err
	}

	var fresh []row
	if room := limit - len(due); room > 0 {
		introduced, err := s.NewIntroducedToday(userID, loc, now)
		if err != nil {
			return nil, err
		}
		quota := min(room, settings.NewLimit(dayKey(now, loc))-introduced)
		if quota > 0 {
			q := s.DB.Table("user_cards uc").Select(sel).
				Joins("JOIN cards c ON c.id = uc.card_id").
				Where("uc.user_id = ? AND NOT uc.suspended AND uc.state = ?", userID, srs.StateNew)
			if err := scope(q, f).Order("uc.created_at ASC, c.id ASC").Limit(quota).Scan(&fresh).Error; err != nil {
				return nil, err
			}
			if missing := quota - len(fresh); missing > 0 {
				added, err := s.autoIntroduce(userID, settings, f, missing, now)
				if err != nil {
					return nil, err
				}
				if added > 0 {
					fresh = fresh[:0]
					q := s.DB.Table("user_cards uc").Select(sel).
						Joins("JOIN cards c ON c.id = uc.card_id").
						Where("uc.user_id = ? AND NOT uc.suspended AND uc.state = ?", userID, srs.StateNew)
					if err := scope(q, f).Order("uc.created_at ASC, c.id ASC").Limit(quota).Scan(&fresh).Error; err != nil {
						return nil, err
					}
				}
			}
		}
	}

	items := make([]QueueItem, 0, len(due)+len(fresh))
	add := func(r row, isNew bool) {
		card := models.Card{ID: r.CardID, DeckID: r.CDeckID, Kind: r.CKind, Front: r.CFront, Payload: r.CPayload, Source: r.CSource}
		items = append(items, QueueItem{Card: card, Progress: r.UserCard, IsNew: isNew, Preview: PreviewLabels(r.UserCard, now)})
	}
	for _, r := range due {
		add(r, false)
	}
	for _, r := range fresh {
		add(r, true)
	}
	for i := range items {
		if items[i].IsNew || items[i].Card.Kind == models.KindMCQ {
			continue // new cards are shown with the answer; mcq has its own choices
		}
		ch, err := s.MeaningChoices(items[i].Card)
		if err != nil {
			return nil, err
		}
		items[i].Choices = ch
	}
	s.RotateExamples(items) // a new sentence each time the word comes back
	return items, nil
}

// autoIntroduce enrolls up to n cards the user has never seen from the active
// decks (or from the filtered deck/subject).
func (s *Service) autoIntroduce(userID int64, settings Settings, f Filter, n int, now time.Time) (int, error) {
	q := s.DB.Table("cards c").Select("c.id").
		Where("NOT EXISTS (SELECT 1 FROM user_cards uc WHERE uc.user_id = ? AND uc.card_id = c.id)", userID)
	switch {
	case f.DeckID > 0 || f.SubjectID > 0:
		q = scope(q, f)
	case len(settings.ActiveDeckIDs) > 0:
		q = q.Where("c.deck_id IN ?", settings.ActiveDeckIDs)
	default:
		q = q.Where("c.deck_id IN (SELECT d.id FROM decks d JOIN subjects s ON s.id = d.subject_id WHERE s.slug = 'english' AND d.slug = 'core-vocab')")
	}
	q = q.Where("c.deck_id IN ("+VisibleDecksSQL+")", userID) // never another user's personal words
	var ids []int64
	if err := q.Order("c.id ASC").Limit(n).Pluck("c.id", &ids).Error; err != nil {
		return 0, err
	}
	return s.Enroll(userID, ids, now)
}

// Enroll adds cards to the user's learning list (no-op for cards already there).
func (s *Service) Enroll(userID int64, cardIDs []int64, now time.Time) (int, error) {
	if len(cardIDs) == 0 {
		return 0, nil
	}
	rows := make([]models.UserCard, 0, len(cardIDs))
	for _, id := range cardIDs {
		rows = append(rows, models.UserCard{UserID: userID, CardID: id, State: srs.StateNew, Due: now, CreatedAt: now, UpdatedAt: now})
	}
	res := s.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows)
	return int(res.RowsAffected), res.Error
}

// Answer is one review of a card.
type Answer struct {
	UserID         int64
	CardID         int64
	Rating         int
	At             time.Time
	DurationMs     *int
	Source         string // web | device | quiz
	DeviceID       *int64
	ClientReviewID *string
	// OnlyIfDue: used by quizzes, so a quiz never reschedules a card that is
	// not due yet (and never enrolls a card the user did not add).
	OnlyIfDue bool
}

type AnswerResult struct {
	Applied   bool            `json:"applied"`
	Duplicate bool            `json:"duplicate,omitempty"`
	Progress  models.UserCard `json:"progress"`
	NextIn    string          `json:"next_in,omitempty"`
}

// ApplyAnswer schedules the card with FSRS and appends a review log, in one
// transaction. Idempotent when ClientReviewID is set (device retries).
func (s *Service) ApplyAnswer(a Answer) (AnswerResult, error) {
	if !srs.ValidRating(a.Rating) {
		return AnswerResult{}, fmt.Errorf("rating must be 1-4")
	}
	now := time.Now().UTC()
	at := a.At.UTC()
	if at.IsZero() || at.After(now.Add(5*time.Minute)) {
		at = now // device clock not set or wrong
	}

	var out AnswerResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if a.ClientReviewID != nil {
			var n int64
			tx.Model(&models.ReviewLog{}).Where("user_id = ? AND client_review_id = ?", a.UserID, *a.ClientReviewID).Count(&n)
			if n > 0 {
				out.Duplicate = true
				return tx.Where("user_id = ? AND card_id = ?", a.UserID, a.CardID).First(&out.Progress).Error
			}
		}

		var uc models.UserCard
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND card_id = ?", a.UserID, a.CardID).First(&uc).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if a.OnlyIfDue {
				return nil
			}
			var exists int64
			tx.Model(&models.Card{}).Where("id = ?", a.CardID).Count(&exists)
			if exists == 0 {
				return ErrCardNotFound
			}
			uc = models.UserCard{UserID: a.UserID, CardID: a.CardID, State: srs.StateNew, Due: at, CreatedAt: now}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		}
		if a.OnlyIfDue && uc.Due.After(at) {
			out.Progress = uc
			return nil
		}
		if uc.LastReview != nil && at.Before(*uc.LastReview) {
			at = *uc.LastReview // late-arriving offline answer: never go back in time
		}

		res, err := srs.Review(toState(uc), a.Rating, at)
		if err != nil {
			return err
		}
		n := res.Next
		uc.State, uc.Stability, uc.Difficulty = n.State, n.Stability, n.Difficulty
		uc.Due, uc.LastReview, uc.ScheduledDays = n.Due, n.LastReview, n.ScheduledDays
		uc.Reps, uc.Lapses, uc.UpdatedAt = n.Reps, n.Lapses, now
		if err := tx.Save(&uc).Error; err != nil {
			return err
		}

		source := a.Source
		if source == "" {
			source = "web"
		}
		log := models.ReviewLog{
			UserID: a.UserID, CardID: a.CardID, Rating: a.Rating,
			StateBefore: res.StateBefore, ElapsedDays: res.ElapsedDays, ScheduledDays: res.ScheduledDays,
			DurationMs: a.DurationMs, Source: source, DeviceID: a.DeviceID,
			ClientReviewID: a.ClientReviewID, ReviewedAt: at,
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		out.Applied = true
		out.Progress = uc
		out.NextIn = srs.HumanInterval(uc.Due.Sub(at))
		return nil
	})
	return out, err
}

func toState(uc models.UserCard) srs.State {
	return srs.State{
		State: uc.State, Stability: uc.Stability, Difficulty: uc.Difficulty,
		Due: uc.Due, LastReview: uc.LastReview, ScheduledDays: uc.ScheduledDays,
		Reps: uc.Reps, Lapses: uc.Lapses,
	}
}

// PreviewLabels returns the next interval for each rating ("1": "1m", "3": "2d", ...).
func PreviewLabels(uc models.UserCard, now time.Time) map[string]string {
	out := map[string]string{}
	for r, due := range srs.Preview(toState(uc), now) {
		out[fmt.Sprint(r)] = srs.HumanInterval(due.Sub(now))
	}
	return out
}
