package study

import (
	"os"
	"testing"
	"time"

	"mvp-lang-app/internal/db"
	"mvp-lang-app/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Each topic has its own new-card quota; stats and progress can be per topic.
// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestQueuePerSubject(t *testing.T) {
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
	var uid, eng, pmp, pmpDeck, engDeck int64
	gdb.Raw(`INSERT INTO users (email, settings) VALUES ('subjects@test', '{}') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)
	gdb.Exec(`DELETE FROM review_logs WHERE user_id = ?`, uid)
	gdb.Raw(`SELECT id FROM subjects WHERE slug = 'english'`).Scan(&eng)
	gdb.Raw(`SELECT id FROM subjects WHERE slug = 'pmp'`).Scan(&pmp)
	gdb.Raw(`SELECT id FROM decks WHERE slug = 'terms' AND subject_id = ?`, pmp).Scan(&pmpDeck)
	gdb.Raw(`SELECT id FROM decks WHERE slug = 'core-vocab' AND subject_id = ?`, eng).Scan(&engDeck)
	// English 4 a day, PMP 2 a day (other subjects 0); both decks active
	gdb.Exec(`UPDATE users SET settings = ?::jsonb WHERE id = ?`,
		`{"daily_new": 0, "daily_new_by_subject": {"english": 4, "pmp": 2}, "active_deck_ids": [`+itoa(engDeck)+`,`+itoa(pmpDeck)+`]}`, uid)

	s := &Service{DB: gdb}
	now := time.Now().UTC()
	count := func(items []QueueItem) (e, p int) {
		for _, it := range items {
			var sub int64
			gdb.Raw(`SELECT subject_id FROM decks WHERE id = ?`, it.Card.DeckID).Scan(&sub)
			if sub == eng {
				e++
			} else if sub == pmp {
				p++
			}
		}
		return
	}
	all, err := s.Queue(uid, Filter{}, 20, now)
	if err != nil {
		t.Fatal(err)
	}
	if e, p := count(all); e != 4 || p != 2 {
		t.Fatalf("all topics: english %d, pmp %d", e, p)
	}
	// the topics take turns: PMP is not pushed behind all the English cards
	if e, p := count(all[:2]); e != 1 || p != 1 {
		t.Fatalf("not mixed: %d %d", e, p)
	}
	only, _ := s.Queue(uid, Filter{SubjectID: pmp}, 20, now)
	if e, p := count(only); e != 0 || p != 2 {
		t.Fatalf("pmp only: english %d, pmp %d", e, p)
	}

	// +5 for PMP only
	if _, err := s.AddExtraNew(uid, "pmp", 5, now); err != nil {
		t.Fatal(err)
	}
	st, err := s.Stats(uid, now, pmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Subjects) != 1 || st.Subjects[0].SubjectID != pmp || st.NewLimitToday != 7 {
		t.Fatalf("pmp stats: %+v", st)
	}
	all2, _ := s.Stats(uid, now)
	if len(all2.Subjects) < 2 || all2.NewLimitToday != 4+7 {
		t.Fatalf("all stats: limit %d", all2.NewLimitToday)
	}
	// an answer in English is not counted in PMP
	gdb.Exec(`INSERT INTO review_logs (user_id, card_id, rating, state_before) VALUES (?, ?, 3, 0)`, uid, all[0].Card.ID)
	if st, _ := s.Stats(uid, now, pmp); st.ReviewsToday != 0 || st.LastReviewAt != nil {
		t.Fatalf("pmp reviews %d", st.ReviewsToday)
	}
	if st, _ := s.Stats(uid, now); st.ReviewsToday != 1 {
		t.Fatalf("all reviews %d", st.ReviewsToday)
	}
	p, err := s.Progress(uid, now, pmp)
	if err != nil || p.Buckets.Total != 2 || len(p.Subjects) != 1 {
		t.Fatalf("progress pmp %+v %v", p.Buckets, err)
	}
	if p, _ := s.Progress(uid, now); p.Buckets.Total != 6 {
		t.Fatalf("progress all %+v", p.Buckets)
	}
}

func itoa(n int64) string {
	b := []byte{}
	if n == 0 {
		return "0"
	}
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}
