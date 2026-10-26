package worker

import (
	"testing"
	"time"

	"mvp-lang-app/internal/study"
)

func TestRecentlyStudied(t *testing.T) {
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	at := func(d time.Duration) study.Stats { x := now.Add(-d); return study.Stats{LastReviewAt: &x} }
	if RecentlyStudied(study.Stats{}, now) {
		t.Fatal("never studied")
	}
	if !RecentlyStudied(at(20*time.Minute), now) {
		t.Fatal("20 min ago")
	}
	if RecentlyStudied(at(2*time.Hour), now) {
		t.Fatal("2 h ago")
	}
}
