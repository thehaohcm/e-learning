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
	q, ok := StudyReminder(study.Stats{Learning: 30})
	if !ok || q.URL != "/quiz?count=5" {
		t.Fatalf("all done with learned words: quiz suggestion, got %+v %v", q, ok)
	}
	m, _ = StudyReminder(study.Stats{DueNow: 1, ReviewsToday: 3, StreakDays: 4})
	if m.Body != "1 card to review" {
		t.Fatalf("%q", m.Body)
	}
}

func TestStudyReminderPerSubject(t *testing.T) {
	m, ok := StudyReminder(study.Stats{DueNow: 11, Subjects: []study.SubjectStats{
		{Slug: "english", Name: "English (TOEIC / IELTS)", Due: 8}, {Slug: "pmp", Name: "PMI PMP", Due: 3}}})
	if !ok || m.Body != "8 English · 3 PMP to review" {
		t.Fatalf("got %q", m.Body)
	}
	m, _ = StudyReminder(study.Stats{DueNow: 3, Subjects: []study.SubjectStats{{Slug: "english", Due: 0}, {Slug: "pmp", Due: 3}}})
	if m.Body != "3 cards to review" {
		t.Fatalf("one topic: %q", m.Body)
	}
}
