package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
func TestManageCourses(t *testing.T) {
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
	gdb.Raw(`INSERT INTO users (email, role) VALUES ('course-admin@test', 'admin') ON CONFLICT (email) DO UPDATE SET role = 'admin' RETURNING id`).Scan(&uid)

	h := New(gdb, config.Config{})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(ctxUserID, uid) })
	r.POST("/subjects", h.CreateSubject)
	r.PUT("/subjects/:id", h.UpdateSubject)
	r.POST("/decks", h.CreateDeck)
	r.PUT("/decks/:id", h.UpdateDeck)
	call := func(method, path, body string) (int, map[string]any) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	gdb.Exec(`DELETE FROM subjects WHERE slug = 'german-a1'`)
	code, sub := call(http.MethodPost, "/subjects", `{"slug":"German-A1","name":"German A1","description":"Beginner German","sort_order":5}`)
	if code != http.StatusCreated || sub["slug"] != "german-a1" || sub["sort_order"].(float64) != 5 {
		t.Fatalf("create subject: %d %v", code, sub)
	}
	sid := int64(sub["id"].(float64))

	if code, _ := call(http.MethodPut, "/subjects/"+itoa(sid), `{"name":"German (A1)","sort_order":4}`); code != 200 {
		t.Fatalf("update subject: %d", code)
	}
	var s models.Subject
	gdb.First(&s, sid)
	if s.Name != "German (A1)" || s.SortOrder != 4 || s.Description != "Beginner German" {
		t.Fatalf("subject not updated: %+v", s)
	}

	code, dk := call(http.MethodPost, "/decks", `{"subject_id":`+itoa(sid)+`,"slug":"greetings","name":"Greetings","default_kind":"vocab"}`)
	if code != http.StatusCreated {
		t.Fatalf("create deck: %d %v", code, dk)
	}
	did := int64(dk["id"].(float64))
	if code, _ := call(http.MethodPut, "/decks/"+itoa(did), `{"description":"Hello & goodbye","default_kind":"concept"}`); code != 200 {
		t.Fatalf("update deck: %d", code)
	}
	var d models.Deck
	gdb.First(&d, did)
	if d.Description != "Hello & goodbye" || d.DefaultKind != "concept" || d.Name != "Greetings" {
		t.Fatalf("deck not updated: %+v", d)
	}

	// personal decks are protected
	owner := uid
	pd := models.Deck{SubjectID: sid, Slug: "mine", Name: "Mine", DefaultKind: "vocab", OwnerID: &owner}
	gdb.Create(&pd)
	if code, _ := call(http.MethodPut, "/decks/"+itoa(pd.ID), `{"name":"Hacked"}`); code != http.StatusForbidden {
		t.Fatalf("personal deck should be protected, got %d", code)
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	b := []byte{}
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}
