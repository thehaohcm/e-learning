package tts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=127.0.0.1 port=5432 user=postgres dbname=t sslmode=disable"
func testDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sql, err := os.ReadFile("../../migrations/0016_tts_audio.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(string(sql)).Error; err != nil {
		t.Fatal(err)
	}
	db.Exec(`DELETE FROM tts_audio`)
	return db
}

func fakeWAV() []byte {
	b := []byte("RIFF\x00\x00\x00\x00WAVEfmt ")
	return append(b, make([]byte, 40)...)
}

func TestEnsureGetPurge(t *testing.T) {
	db := testDB(t)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Query().Get("voice") != "v1" {
			t.Errorf("voice = %q", r.URL.Query().Get("voice"))
		}
		if strings.Contains(r.URL.Query().Get("text"), "slow") {
			time.Sleep(300 * time.Millisecond)
		}
		if r.URL.Query().Get("text") == "broken" {
			http.Error(w, "nope", http.StatusBadGateway)
			return
		}
		w.Write(fakeWAV())
	}))
	defer srv.Close()
	s := New(db, srv.URL, "v1", 60)

	ready := s.Ensure(context.Background(), []string{"apple", "Apple ", "look up", "broken"})
	if !ready[s.Key("apple")] || !ready[s.Key("look up")] || ready[s.Key("broken")] {
		t.Fatalf("ready = %v", ready)
	}
	if calls.Load() != 3 { // apple only once
		t.Fatalf("calls = %d", calls.Load())
	}
	if wav, ok := s.Get(s.Key("APPLE")); !ok || string(wav[:4]) != "RIFF" {
		t.Fatal("Get apple")
	}
	// the service failed: no new calls for a while
	s.Ensure(context.Background(), []string{"pear"})
	if calls.Load() != 3 {
		t.Fatalf("should back off after a failure, calls = %d", calls.Load())
	}
	s.mu.Lock()
	s.failAt = time.Time{}
	s.mu.Unlock()

	// not ready within the deadline: finished in the background
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	s.Ensure(ctx, []string{"slow one", "slow two", "slow three", "slow four", "slow five", "slow six"})
	cancel()
	deadline := time.Now().Add(5 * time.Second)
	for len(s.Ready([]string{"slow five", "slow six"})) < 2 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if len(s.Ready([]string{"slow five", "slow six"})) != 2 {
		t.Fatal("background did not finish")
	}

	st, err := s.Stats()
	if err != nil || st.Words < 4 || st.Bytes <= 0 {
		t.Fatalf("stats %+v %v", st, err)
	}
	db.Exec(`UPDATE tts_audio SET last_used_at = NOW() - INTERVAL '90 days' WHERE key = ?`, s.Key("apple"))
	if n, _ := s.Purge(60 * 24 * time.Hour); n != 1 {
		t.Fatalf("purged %d, want 1", n)
	}
	if _, ok := s.Get(s.Key("apple")); ok {
		t.Fatal("apple should be gone")
	}
	if n, _ := s.Purge(0); n < 3 {
		t.Fatalf("purge all = %d", n)
	}
}

func TestMakeFor(t *testing.T) {
	db := testDB(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(fakeWAV()) }))
	defer srv.Close()
	s := New(db, srv.URL, "v1", 60)
	if s.MakeFor(context.Background(), s.Key("pear"), "apple") {
		t.Fatal("a key must match its text")
	}
	if !s.MakeFor(context.Background(), s.Key("pear"), " Pear ") {
		t.Fatal("make pear")
	}
	if _, ok := s.Get(s.Key("pear")); !ok {
		t.Fatal("pear stored")
	}
}
