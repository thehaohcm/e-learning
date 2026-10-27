package study

import (
	"os"
	"strconv"
	"testing"
	"time"

	"mvp-lang-app/internal/db"
	"mvp-lang-app/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestAutoPace(t *testing.T) {
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
	var uid, eng int64
	gdb.Raw(`INSERT INTO users (email) VALUES ('autopace@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Raw(`SELECT id FROM subjects WHERE slug = 'english'`).Scan(&eng)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)
	gdb.Exec(`DELETE FROM review_logs WHERE user_id = ?`, uid)
	var ids []int64
	gdb.Raw(`SELECT c.id FROM cards c JOIN decks d ON d.id = c.deck_id WHERE d.slug = 'core-vocab' ORDER BY c.id LIMIT 12`).Scan(&ids)
	if len(ids) != 12 {
		t.Fatal("seed cards missing")
	}
	now := time.Now().UTC()
	// 12 well-known English cards (state review, high stability) → "strong"
	for _, id := range ids {
		gdb.Exec(`INSERT INTO user_cards (user_id, card_id, state, stability, lapses, reps, suspended, due, created_at, updated_at)
		          VALUES (?, ?, 2, 40, 0, 5, false, ?, ?, ?)`, uid, id, now.Add(48*time.Hour), now, now)
	}
	s := &Service{DB: gdb}
	if lvl := func() string { p, _ := s.Progress(uid, now, eng); return RetentionLevel(p) }(); lvl != "strong" {
		t.Fatalf("expected strong, got %s", lvl)
	}

	// auto-pace on, english at 10/day, 6 strong days already banked
	gdb.Exec(`UPDATE users SET settings = ?::jsonb WHERE id = ?`,
		`{"auto_pace":true,"daily_new":10,"auto_pace_strong":{"english":6}}`, uid)
	if err := s.AutoPace(uid, now); err != nil {
		t.Fatal(err)
	}
	_, st, _ := s.loadUser(uid)
	if st.DailyNewFor("english") != 12 {
		t.Fatalf("english should be raised to 12, got %d", st.DailyNewFor("english"))
	}
	if st.AutoPaceStrong["english"] != 0 || len(st.AutoPaceNotes) != 1 {
		t.Fatalf("streak/notes: %d %v", st.AutoPaceStrong["english"], st.AutoPaceNotes)
	}

	// same day → no further change
	if err := s.AutoPace(uid, now); err != nil {
		t.Fatal(err)
	}
	_, st, _ = s.loadUser(uid)
	if st.DailyNewFor("english") != 12 || len(st.AutoPaceNotes) != 1 {
		t.Fatalf("should be idempotent within a day: %d %v", st.DailyNewFor("english"), st.AutoPaceNotes)
	}

	// next day, still strong but streak reset → only +1 banked, no raise yet
	tomorrow := now.Add(24 * time.Hour)
	if err := s.AutoPace(uid, tomorrow); err != nil {
		t.Fatal(err)
	}
	_, st, _ = s.loadUser(uid)
	if st.DailyNewFor("english") != 12 || st.AutoPaceStrong["english"] != 1 {
		t.Fatalf("next day: limit %d streak %d", st.DailyNewFor("english"), st.AutoPaceStrong["english"])
	}

	// turn many cards into "struggling" → "focus" lowers the limit
	for _, id := range ids {
		gdb.Exec(`UPDATE user_cards SET state = 3, stability = 1, lapses = 3 WHERE user_id = ? AND card_id = ?`, uid, id)
	}
	twoDays := now.Add(48 * time.Hour)
	if err := s.AutoPace(uid, twoDays); err != nil {
		t.Fatal(err)
	}
	_, st, _ = s.loadUser(uid)
	if st.DailyNewFor("english") != 10 || st.AutoPaceStrong["english"] != 0 {
		t.Fatalf("focus should lower 12→10: limit %d", st.DailyNewFor("english"))
	}
	_ = strconv.Itoa
}
