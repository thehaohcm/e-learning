package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

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

// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestDevicePractice(t *testing.T) {
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
	gdb.Raw(`INSERT INTO users (email) VALUES ('practice@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	var ids []int64
	gdb.Raw(`SELECT c.id FROM cards c JOIN decks d ON d.id = c.deck_id WHERE d.owner_id IS NULL AND c.kind = 'vocab' ORDER BY c.id LIMIT 2`).Scan(&ids)
	if len(ids) != 2 {
		t.Fatal("seed cards missing")
	}
	gdb.Exec(`DELETE FROM word_questions WHERE card_id IN ?`, ids)

	// fake AI: 2 questions for word 1, one broken (3 choices) for word 2
	var calls atomic.Int32
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		content := `{"items":[{"i":1,"questions":[
		  {"kind":"usage","prompt":"Which sentence?","choices":["s1","s2","s3","s4"],"answer":1,"explanation":"vì"},
		  {"kind":"collocation","prompt":"____ a decision","choices":["make","do","take","get"],"answer":0,"explanation":"x"}]},
		 {"i":2,"questions":[{"kind":"usage","prompt":"bad","choices":["a","b","c"],"answer":0}]}]}`
		b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}}}})
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}))
	defer fake.Close()

	tok, hash, _ := auth.NewDeviceToken()
	gdb.Create(&models.Device{Name: "p", TokenHash: hash, UserID: &uid, Settings: models.JSON("{}")})
	h := New(gdb, config.Config{DeepSeekAPIKey: "k", DeepSeekBaseURL: fake.URL, DeepSeekModel: "m"})
	r := gin.New()
	r.POST("/practice", h.RequireDevice, h.DevicePractice)
	call := func() (int, map[string]any) {
		body, _ := json.Marshal(map[string]any{"card_ids": ids})
		req := httptest.NewRequest(http.MethodPost, "/practice", strings.NewReader(string(body)))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	code, out := call()
	qs, _ := out["questions"].([]any)
	if code != 200 || len(qs) != 1 || calls.Load() != 1 {
		t.Fatalf("first call: %d %v (AI calls %d)", code, out, calls.Load())
	}
	q := qs[0].(map[string]any)
	choices := q["choices"].([]any)
	ans := int(q["answer"].(float64))
	right := choices[ans].(string)
	if q["kind"] == "usage" && right != "s2" || q["kind"] == "collocation" && right != "make" {
		t.Fatalf("answer index lost after shuffling: %v", q)
	}
	var stored int64
	gdb.Raw(`SELECT COUNT(*) FROM word_questions WHERE card_id IN ?`, ids).Scan(&stored)
	if stored != 2 {
		t.Fatalf("stored %d questions, want 2 (the broken one dropped)", stored)
	}
	// word 1 is cached now; word 2 still has none, so AI is asked again for it only
	call()
	if calls.Load() != 2 {
		t.Fatalf("AI calls %d", calls.Load())
	}
}
