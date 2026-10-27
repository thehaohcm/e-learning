package study

import (
	"encoding/json"
	"fmt"
	"time"

	"mvp-lang-app/internal/models"
)

// Auto-pace adjusts how many new cards a topic introduces per day, when the
// user turns it on. The idea: new words are cheap to add but each one brings
// many future reviews, so the daily intake should rise only when memory is
// comfortably keeping up, and fall back when it isn't. It runs once a day,
// uses a sustained signal (several "strong" days in a row, not one snapshot),
// is bounded, and is always visible to the user (AutoPaceNotes). New cards are
// still introduced in order by the queue — auto-pace never picks at random.
const (
	AutoPaceStep       = 2  // change the daily limit by this many cards
	AutoPaceCap        = 20 // never raise a topic above this
	AutoPaceFloor      = 3  // never lower a topic below this
	AutoPaceStrongDays = 7  // consecutive "strong" days before a raise
	autoPaceMinStudied = 10 // need at least this many studied cards to judge
)

// RetentionLevel grades a topic's memory health from its progress, with the
// same thresholds the Home card shows: "strong", "okay", "focus" (needs focus)
// or "early" (too few cards studied to tell).
func RetentionLevel(p Progress) string {
	done := p.Buckets.Total - p.Buckets.New
	if done < autoPaceMinStudied {
		return "early"
	}
	knownWell := float64(p.Buckets.Mastered+p.Buckets.Familiar) / float64(done)
	struggle := float64(p.Buckets.Struggling) / float64(done)
	hasAcc := p.Answers30 > 0
	acc := 0.0
	if hasAcc {
		acc = float64(p.Correct30) / float64(p.Answers30)
	}
	if struggle >= 0.2 || knownWell < 0.3 || (hasAcc && acc < 0.65) {
		return "focus"
	}
	if knownWell >= 0.6 && struggle < 0.1 && (!hasAcc || acc >= 0.8) {
		return "strong"
	}
	return "okay"
}

// AutoPace runs the once-a-day adjustment for a user. It is a no-op unless the
// user turned auto-pace on and a new day has started since the last run.
func (s *Service) AutoPace(userID int64, now time.Time) error {
	u, settings, err := s.loadUser(userID)
	if err != nil || !settings.AutoPace {
		return err
	}
	day := dayKey(now, UserLocation(u.Timezone))
	if settings.AutoPaceDate == day {
		return nil // already done today
	}
	settings.AutoPaceDate = day
	if settings.AutoPaceStrong == nil {
		settings.AutoPaceStrong = map[string]int{}
	}

	subs, err := s.subjectsIn(Filter{})
	if err != nil {
		return err
	}
	for _, sub := range subs {
		p, err := s.Progress(userID, now, sub.ID)
		if err != nil {
			return err
		}
		cur := settings.DailyNewFor(sub.Slug)
		switch RetentionLevel(p) {
		case "strong":
			streak := settings.AutoPaceStrong[sub.Slug] + 1
			if streak >= AutoPaceStrongDays && cur < AutoPaceCap {
				n := min(AutoPaceCap, cur+AutoPaceStep)
				settings.setDailyNew(sub.Slug, n)
				settings.addPaceNote(day, sub.Slug, cur, n, "memory strong")
				streak = 0
			}
			settings.AutoPaceStrong[sub.Slug] = streak
		case "focus":
			if cur > AutoPaceFloor {
				n := max(AutoPaceFloor, cur-AutoPaceStep)
				settings.setDailyNew(sub.Slug, n)
				settings.addPaceNote(day, sub.Slug, cur, n, "focus on review")
			}
			settings.AutoPaceStrong[sub.Slug] = 0
		default: // "okay" / "early": the streak must be unbroken, so reset it
			settings.AutoPaceStrong[sub.Slug] = 0
		}
	}

	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return s.DB.Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]any{"settings": models.JSON(b), "updated_at": now}).Error
}

// addPaceNote records a change, newest first, keeping the last few for the UI.
func (s *Settings) addPaceNote(day, slug string, from, to int, why string) {
	note := fmt.Sprintf("%s · %s %d → %d/day (%s)", day, subjectShort(slug), from, to, why)
	s.AutoPaceNotes = append([]string{note}, s.AutoPaceNotes...)
	if len(s.AutoPaceNotes) > 6 {
		s.AutoPaceNotes = s.AutoPaceNotes[:6]
	}
}

func subjectShort(slug string) string {
	switch slug {
	case "english":
		return "English"
	case "pmp":
		return "PMP"
	case "pentest":
		return "Pentest"
	}
	return slug
}
