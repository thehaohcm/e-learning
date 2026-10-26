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

// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestProgress(t *testing.T) {
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
	gdb.Raw(`INSERT INTO users (email) VALUES ('progress@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)
	gdb.Exec(`DELETE FROM review_logs WHERE user_id = ?`, uid)
	var ids []int64
	gdb.Raw(`SELECT c.id FROM cards c JOIN decks d ON d.id = c.deck_id WHERE d.slug = 'core-vocab' ORDER BY c.id LIMIT 7`).Scan(&ids)
	if len(ids) != 7 {
		t.Fatal("seed cards missing")
	}
	now := time.Now().UTC()
	// state, stability, lapses, suspended → expected group
	rows := []struct {
		state     int
		stability float64
		lapses    int
		suspended bool
	}{
		{0, 0, 0, false},   // new
		{1, 0.5, 0, false}, // learning
		{2, 10, 0, false},  // familiar
		{2, 30, 3, false},  // mastered (solid now, even if forgotten before)
		{3, 2, 1, false},   // struggling (relearning)
		{2, 4, 2, false},   // struggling (2 lapses)
		{2, 40, 0, true},   // suspended: not counted
	}
	for i, r := range rows {
		gdb.Exec(`INSERT INTO user_cards (user_id, card_id, state, stability, lapses, reps, suspended, due, created_at, updated_at)
		          VALUES (?, ?, ?, ?, ?, 3, ?, ?, ?, ?)`, uid, ids[i], r.state, r.stability, r.lapses, r.suspended, now, now, now)
	}
	gdb.Exec(`INSERT INTO review_logs (user_id, card_id, rating, state_before) VALUES (?, ?, 3, 2), (?, ?, 1, 2)`, uid, ids[2], uid, ids[5])

	p, err := (&Service{DB: gdb}).Progress(uid, now)
	if err != nil {
		t.Fatal(err)
	}
	want := Buckets{Total: 6, New: 1, Struggling: 2, Learning: 1, Familiar: 1, Mastered: 1}
	if p.Buckets != want {
		t.Fatalf("buckets %+v, want %+v", p.Buckets, want)
	}
	if len(p.Subjects) != 1 || p.Subjects[0].Buckets != want {
		t.Fatalf("subjects %+v", p.Subjects)
	}
	if p.Answers30 != 2 || p.Correct30 != 1 {
		t.Fatalf("answers %d / %d", p.Correct30, p.Answers30)
	}
	if len(p.Weakest) != 2 || p.Weakest[0].CardID != ids[5] || p.Weakest[0].Meaning == "" || p.Weakest[0].Front == "" {
		t.Fatalf("weakest %+v", p.Weakest)
	}
}
