package push

import (
	"testing"

	"mvp-lang-app/internal/study"
)

func TestStudyReminder(t *testing.T) {
	if _, ok := StudyReminder(study.Stats{}); ok {
		t.Fatal("nothing to study should not send")
	}
	m, ok := StudyReminder(study.Stats{DueNow: 12, NewLeftToday: 1, StreakDays: 4})
	if m1, _ := StudyReminder(study.Stats{DueNow: 2, StreakDays: 1}); m1.Body != "2 cards to review" {
		t.Fatalf("%q", m1.Body)
	}
	if !ok || m.Body != "12 cards to review · 1 new word left today — keep your 4-day streak!" || m.URL != "/review" {
		t.Fatalf("%+v", m)
	}
	m, _ = StudyReminder(study.Stats{DueNow: 1, ReviewsToday: 3, StreakDays: 4})
	if m.Body != "1 card to review" {
		t.Fatalf("%q", m.Body)
	}
}
