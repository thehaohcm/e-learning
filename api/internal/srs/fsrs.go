// Package srs wraps the official FSRS implementation
// (github.com/open-spaced-repetition/go-fsrs) so the rest of the code never
// touches scheduling maths directly.
package srs

import (
	"fmt"
	"math"
	"time"

	fsrs "github.com/open-spaced-repetition/go-fsrs/v3"
)

// Ratings sent by clients.
const (
	Again = 1 // forgot
	Hard  = 2
	Good  = 3
	Easy  = 4
)

// Card states (same numbers as the user_cards.state column).
const (
	StateNew        = 0
	StateLearning   = 1
	StateReview     = 2
	StateRelearning = 3
)

// MasteredStabilityDays: a card counts as "mastered" once FSRS expects it to be
// remembered (90%) for at least three weeks.
const MasteredStabilityDays = 21.0

// State is the scheduling state of one card for one user.
type State struct {
	State         int
	Stability     float64
	Difficulty    float64
	Due           time.Time
	LastReview    *time.Time
	ScheduledDays int
	Reps          int
	Lapses        int
}

// Result of answering a card.
type Result struct {
	Next          State
	StateBefore   int
	ElapsedDays   int
	ScheduledDays int
}

func params() fsrs.Parameters {
	p := fsrs.DefaultParam()
	p.RequestRetention = 0.9
	p.MaximumInterval = 365 * 3
	p.EnableFuzz = true // spread reviews so cards learned together don't all come back the same day
	p.EnableShortTerm = true
	return p
}

func ValidRating(r int) bool { return r >= Again && r <= Easy }

func toFSRS(s State) fsrs.Card {
	c := fsrs.NewCard()
	c.State = fsrs.State(s.State)
	c.Stability = s.Stability
	c.Difficulty = s.Difficulty
	c.Due = s.Due
	c.ScheduledDays = uint64(max(0, s.ScheduledDays))
	c.Reps = uint64(max(0, s.Reps))
	c.Lapses = uint64(max(0, s.Lapses))
	if s.LastReview != nil {
		c.LastReview = *s.LastReview
	}
	// A card in review state must have valid memory values; guard against
	// broken data (e.g. imported legacy rows) instead of producing NaN.
	if c.State != fsrs.New && (c.Stability <= 0 || c.Difficulty < 1 || c.LastReview.IsZero()) {
		c.State = fsrs.New
		c.Stability, c.Difficulty = 0, 0
		c.LastReview = time.Time{}
	}
	return c
}

func fromFSRS(c fsrs.Card) State {
	lr := c.LastReview
	return State{
		State:         int(c.State),
		Stability:     c.Stability,
		Difficulty:    c.Difficulty,
		Due:           c.Due,
		LastReview:    &lr,
		ScheduledDays: int(c.ScheduledDays),
		Reps:          int(c.Reps),
		Lapses:        int(c.Lapses),
	}
}

// Review applies one answer at time `now` and returns the new state.
// A new FSRS instance is created per call because the library keeps
// per-call fuzz state inside its Parameters (not safe to share).
func Review(s State, rating int, now time.Time) (Result, error) {
	if !ValidRating(rating) {
		return Result{}, fmt.Errorf("rating must be 1-4, got %d", rating)
	}
	f := fsrs.NewFSRS(params())
	info := f.Next(toFSRS(s), now.UTC(), fsrs.Rating(rating))
	next := fromFSRS(info.Card)
	if math.IsNaN(next.Stability) || math.IsNaN(next.Difficulty) {
		return Result{}, fmt.Errorf("fsrs produced NaN for state %+v", s)
	}
	return Result{
		Next:          next,
		StateBefore:   int(info.ReviewLog.State),
		ElapsedDays:   int(info.ReviewLog.ElapsedDays),
		ScheduledDays: int(info.Card.ScheduledDays),
	}, nil
}

// Preview returns when the card would be due again for each rating, so the
// UI can show "Again: 1m · Hard: 5m · Good: 10m · Easy: 3d" on the buttons.
func Preview(s State, now time.Time) map[int]time.Time {
	f := fsrs.NewFSRS(params())
	log := f.Repeat(toFSRS(s), now.UTC())
	out := make(map[int]time.Time, 4)
	for r, info := range log {
		out[int(r)] = info.Card.Due
	}
	return out
}

// Retrievability is the estimated probability (0-1) that the user still
// remembers the card at `now`.
func Retrievability(s State, now time.Time) float64 {
	f := fsrs.NewFSRS(params())
	return f.GetRetrievability(toFSRS(s), now.UTC())
}

// HumanInterval formats a duration the way flashcard apps do: 1m, 10m, 3h, 4d, 2mo.
func HumanInterval(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", max(1, int(math.Round(d.Minutes()))))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(math.Round(d.Hours())))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd", int(math.Round(d.Hours()/24)))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%.1fmo", d.Hours()/24/30)
	default:
		return fmt.Sprintf("%.1fy", d.Hours()/24/365)
	}
}
