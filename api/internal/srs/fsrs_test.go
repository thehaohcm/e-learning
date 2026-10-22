package srs

import (
	"testing"
	"time"
)

func TestNewCardGoodGoesToLearningThenReview(t *testing.T) {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	s := State{State: StateNew, Due: now}

	r1, err := Review(s, Good, now)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Next.State != StateLearning {
		t.Fatalf("expected learning, got %d", r1.Next.State)
	}
	if got := r1.Next.Due.Sub(now); got != 10*time.Minute {
		t.Fatalf("expected 10m learning step, got %v", got)
	}

	later := r1.Next.Due
	r2, err := Review(r1.Next, Good, later)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Next.State != StateReview {
		t.Fatalf("expected review, got %d", r2.Next.State)
	}
	if r2.ScheduledDays < 1 {
		t.Fatalf("expected at least 1 day interval, got %d", r2.ScheduledDays)
	}
}

func TestIntervalsGrowAndAgainResets(t *testing.T) {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	s := State{State: StateNew, Due: now}
	r, _ := Review(s, Easy, now)
	prev := r.ScheduledDays
	cur := r.Next
	for i := 0; i < 4; i++ {
		now = cur.Due
		r, _ = Review(cur, Good, now)
		if r.ScheduledDays < prev {
			t.Fatalf("interval shrank on Good: %d -> %d", prev, r.ScheduledDays)
		}
		prev = r.ScheduledDays
		cur = r.Next
	}
	if prev < 20 {
		t.Fatalf("expected interval to grow beyond 20 days after 5 good answers, got %d", prev)
	}

	now = cur.Due
	r, _ = Review(cur, Again, now)
	if r.Next.State != StateRelearning {
		t.Fatalf("expected relearning after Again, got %d", r.Next.State)
	}
	if r.Next.Lapses != cur.Lapses+1 {
		t.Fatalf("lapses not incremented")
	}
	if r.Next.Due.Sub(now) > time.Hour {
		t.Fatalf("forgotten card should come back within the hour, got %v", r.Next.Due.Sub(now))
	}
}

func TestHardIsNotALapse(t *testing.T) {
	now := time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
	lr := now.Add(-10 * 24 * time.Hour)
	s := State{State: StateReview, Stability: 10, Difficulty: 5, Due: now, LastReview: &lr, ScheduledDays: 10, Reps: 3}
	r, err := Review(s, Hard, now)
	if err != nil {
		t.Fatal(err)
	}
	if r.Next.Lapses != 0 || r.Next.State != StateReview {
		t.Fatalf("Hard must keep the card in review without a lapse: %+v", r.Next)
	}
}

func TestBrokenStateIsTreatedAsNew(t *testing.T) {
	now := time.Now()
	s := State{State: StateReview, Stability: 0, Difficulty: 0.3, Due: now}
	if _, err := Review(s, Good, now); err != nil {
		t.Fatal(err)
	}
}

func TestPreviewHasFourRatings(t *testing.T) {
	p := Preview(State{State: StateNew, Due: time.Now()}, time.Now())
	if len(p) != 4 {
		t.Fatalf("expected 4 previews, got %d", len(p))
	}
}
