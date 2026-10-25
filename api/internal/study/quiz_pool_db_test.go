package study

import (
	"os"
	"testing"
	"time"

	"mvp-lang-app/internal/db"
	"mvp-lang-app/internal/srs"
	"mvp-lang-app/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Needs an empty scratch PostgreSQL database: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestQuizPoolAndQuizAnswers(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(gdb, migrations.FS); err != nil {
		t.Fatal(err)
	}
	var uid int64
	gdb.Raw(`INSERT INTO users (email) VALUES ('quiz@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)

	var english, pmp int64
	gdb.Raw(`SELECT id FROM subjects ORDER BY sort_order, id LIMIT 1`).Scan(&english)
	gdb.Raw(`SELECT id FROM subjects WHERE id <> ? ORDER BY sort_order, id LIMIT 1`, english).Scan(&pmp)
	future := time.Now().Add(10 * 24 * time.Hour)
	// learned (not due) cards in both topics, plus some new ones that must not be quizzed
	for _, sub := range []int64{english, pmp} {
		gdb.Exec(`INSERT INTO user_cards (user_id, card_id, state, due, stability, difficulty, created_at)
			SELECT ?, c.id, ?, ?, 10, 5, NOW() FROM cards c JOIN decks d ON d.id = c.deck_id
			WHERE d.subject_id = ? AND d.owner_id IS NULL AND c.kind <> 'mcq' ORDER BY c.id LIMIT 12`, uid, srs.StateReview, future, sub)
	}
	gdb.Exec(`INSERT INTO user_cards (user_id, card_id, state, due, created_at)
		SELECT ?, c.id, 0, NOW(), NOW() FROM cards c JOIN decks d ON d.id = c.deck_id
		WHERE d.subject_id = ? AND c.id NOT IN (SELECT card_id FROM user_cards WHERE user_id = ?) LIMIT 5`, uid, english, uid)

	s := &Service{DB: gdb}
	pool, err := s.QuizPool(uid, Filter{SubjectIDs: []int64{english}}, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(pool) < 4 {
		t.Fatalf("pool too small: %d", len(pool))
	}
	for _, it := range pool {
		var sub int64
		gdb.Raw(`SELECT d.subject_id FROM decks d WHERE d.id = ?`, it.Card.DeckID).Scan(&sub)
		if sub != english {
			t.Fatalf("card %d is not English", it.Card.ID)
		}
		if it.Progress.State == srs.StateNew {
			t.Fatalf("new card %d in the quiz", it.Card.ID)
		}
		if len(it.Choices) != 4 {
			t.Fatalf("card %d has %d choices", it.Card.ID, len(it.Choices))
		}
		right := 0
		for _, ch := range it.Choices {
			if ch.ID == it.Card.ID {
				right++
			}
		}
		if right != 1 {
			t.Fatalf("card %d: %d right answers", it.Card.ID, right)
		}
	}

	// right answer on a card not due: nothing changes
	a, b := pool[0].Card.ID, pool[1].Card.ID
	res, err := s.ApplyAnswer(Answer{UserID: uid, CardID: a, Rating: srs.Good, At: time.Now(), Source: "device-quiz", OnlyIfDue: true, DueNowIfWrong: true})
	if err != nil || res.Applied || res.Progress.Due.Before(future.Add(-time.Minute)) {
		t.Fatalf("right answer changed a card not due: %+v %v", res, err)
	}
	// wrong answer on a card not due: due now, memory unchanged
	res, err = s.ApplyAnswer(Answer{UserID: uid, CardID: b, Rating: srs.Again, At: time.Now(), Source: "device-quiz", OnlyIfDue: true, DueNowIfWrong: true})
	if err != nil || res.Applied {
		t.Fatalf("wrong answer: %+v %v", res, err)
	}
	var due time.Time
	var stab float64
	gdb.Raw(`SELECT due, stability FROM user_cards WHERE user_id = ? AND card_id = ?`, uid, b).Row().Scan(&due, &stab)
	if due.After(time.Now()) || stab != 10 {
		t.Fatalf("wrong answer: due %v stability %v", due, stab)
	}
	// it is now in Study's queue
	q, _ := s.Queue(uid, Filter{SubjectIDs: []int64{english}}, 50, time.Now())
	found := false
	for _, it := range q {
		found = found || it.Card.ID == b
	}
	if !found {
		t.Fatal("the missed card should be due in Study")
	}
	// now due: a quiz answer reschedules it with FSRS
	res, err = s.ApplyAnswer(Answer{UserID: uid, CardID: b, Rating: srs.Good, At: time.Now(), Source: "device-quiz", OnlyIfDue: true, DueNowIfWrong: true})
	if err != nil || !res.Applied {
		t.Fatalf("due card should be reviewed: %+v %v", res, err)
	}
}
