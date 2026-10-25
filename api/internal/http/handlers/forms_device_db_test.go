package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mvp-lang-app/internal/auth"
	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/db"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/migrations"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDevicePrepareForms(t *testing.T) {
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
	gdb.Raw(`INSERT INTO users (email) VALUES ('forms@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)
	// learned single words that have an example sentence
	gdb.Exec(`INSERT INTO user_cards (user_id, card_id, state, due, stability, difficulty, created_at)
		SELECT ?, c.id, 2, NOW() + INTERVAL '5 days', 10, 5, NOW() FROM cards c JOIN decks d ON d.id = c.deck_id
		WHERE d.owner_id IS NULL AND c.kind = 'vocab' AND position(' ' in c.front) = 0
		  AND c.payload::text ILIKE '%example%' ORDER BY c.id LIMIT 6`, uid)
	var cardIDs []int64
	gdb.Raw(`SELECT card_id FROM user_cards WHERE user_id = ?`, uid).Scan(&cardIDs)
	gdb.Exec(`DELETE FROM card_examples WHERE card_id IN ?`, cardIDs)

	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]any
		for i := 1; i <= 12; i++ {
			items = append(items, map[string]any{"i": i, "options": []string{"x"}}) // unusable: stored as []
		}
		content, _ := json.Marshal(map[string]any{"items": items})
		b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": string(content)}}}})
		w.Write(b)
	}))
	defer fake.Close()
	tok, hash, _ := auth.NewDeviceToken()
	gdb.Create(&models.Device{Name: "f", TokenHash: hash, UserID: &uid, Settings: models.JSON("{}")})
	h := New(gdb, config.Config{DeepSeekAPIKey: "k", DeepSeekBaseURL: fake.URL, DeepSeekModel: "m"})
	r := gin.New()
	r.POST("/prep", h.RequireDevice, h.DevicePrepareForms)
	req := httptest.NewRequest(http.MethodPost, "/prep", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out struct{ Queued int }
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.Queued == 0 {
		t.Fatalf("nothing queued: %d %s (cards %v)", w.Code, w.Body.String(), cardIDs)
	}
	deadline := time.Now().Add(5 * time.Second)
	var done int64
	for time.Now().Before(deadline) {
		gdb.Raw(`SELECT COUNT(*) FROM card_examples WHERE card_id IN ? AND form_options IS NOT NULL`, cardIDs).Scan(&done)
		if int(done) >= out.Queued {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if int(done) < out.Queued {
		t.Fatalf("prepared %d of %d", done, out.Queued)
	}
	// second press: all prepared, nothing to ask AI
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/prep", nil).WithContext(req.Context()))
}
