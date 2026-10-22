package study

import (
	"testing"
	"time"
)

func TestSettingsDefaultsAndSanitize(t *testing.T) {
	s := ParseSettings([]byte(`{"daily_new": 999, "active_from": "25:00", "reminder_every_min": 30}`))
	if s.DailyNew != 10 || s.ActiveFrom != "07:30" || s.ReminderEveryMin != 30 {
		t.Fatalf("unexpected %+v", s)
	}
	m, err := s.Merge([]byte(`{"daily_new": 5}`))
	if err != nil || m.DailyNew != 5 || m.ReminderEveryMin != 30 {
		t.Fatalf("merge failed: %+v %v", m, err)
	}
}

func TestStreak(t *testing.T) {
	loc := time.UTC
	today := time.Date(2026, 10, 22, 0, 0, 0, 0, loc)
	d := func(n int) time.Time { return today.AddDate(0, 0, -n) }
	if got := streak([]time.Time{d(0), d(1), d(2), d(4)}, today); got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
	// not studied yet today: streak from yesterday still counts
	if got := streak([]time.Time{d(1), d(2)}, today); got != 2 {
		t.Fatalf("want 2, got %d", got)
	}
	if got := streak([]time.Time{d(3)}, today); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}
