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

func TestNewLimitExtraOnlyToday(t *testing.T) {
	s := ParseSettings([]byte(`{"daily_new": 10, "extra_new_date": "2026-10-22", "extra_new": 5}`))
	if s.NewLimit("english", "2026-10-22") != 15 || s.NewLimit("english", "2026-10-23") != 10 || s.NewLimit("pmp", "2026-10-22") != 10 {
		t.Fatalf("got %d / %d", s.NewLimit("english", "2026-10-22"), s.NewLimit("english", "2026-10-23"))
	}
	if s := ParseSettings([]byte(`{"extra_new": 9999}`)); s.ExtraNew != 0 || len(s.ExtraNewBySubject) != 0 {
		t.Fatalf("not capped: %d %v", s.ExtraNew, s.ExtraNewBySubject)
	}
}

func TestDailyNewBySubject(t *testing.T) {
	s := ParseSettings([]byte(`{"daily_new": 10, "daily_new_by_subject": {"pmp": 5, "x": 999}}`))
	if s.DailyNewFor("pmp") != 5 || s.DailyNewFor("english") != 10 || s.DailyNewFor("x") != 10 {
		t.Fatalf("got %+v", s.DailyNewBySubject)
	}
	m, err := s.Merge([]byte(`{"daily_new_by_subject": {"english": 15}}`))
	if err != nil || m.DailyNewFor("english") != 15 || m.DailyNewFor("pmp") != 5 {
		t.Fatalf("merge %+v %v", m.DailyNewBySubject, err)
	}
}

func TestPushTimesCleaned(t *testing.T) {
	s := ParseSettings([]byte(`{"push_times": ["20:30", "08:00", "25:00", "20:30", "7:00", "06:00", "09:00", "10:00", "11:00", "12:00"]}`))
	want := []string{"06:00", "08:00", "09:00", "10:00", "11:00", "12:00"}
	if len(s.PushTimes) != len(want) {
		t.Fatalf("got %v", s.PushTimes)
	}
	for i := range want {
		if s.PushTimes[i] != want[i] {
			t.Fatalf("got %v", s.PushTimes)
		}
	}
	if d := ParseSettings(nil); d.PushTimes == nil || len(d.PushTimes) != 0 {
		t.Fatalf("default push_times should be []: %v", d.PushTimes)
	}
}
