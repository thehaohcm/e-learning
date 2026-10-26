package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/db"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/migrations"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Needs a scratch PostgreSQL: TEST_DATABASE_DSN="host=... dbname=... sslmode=disable"
func TestWritingPhrases(t *testing.T) {
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
	gdb.Raw(`INSERT INTO users (email) VALUES ('phrases@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Exec(`DELETE FROM user_cards WHERE user_id = ?`, uid)

	h := New(gdb, config.Config{})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(ctxUserID, uid) })
	r.GET("/writing/phrases", h.WritingPhrases)
	r.POST("/writing/phrases/study", h.StudyWritingPhrases)
	r.GET("/writing/:id", func(c *gin.Context) { c.String(200, "id") }) // the routes live together
	call := func(method, path string) map[string]any {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader("")))
		if w.Code != 200 {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}

	out := call(http.MethodGet, "/writing/phrases")
	groups := out["groups"].([]any)
	if out["total"].(float64) < 80 || len(groups) != 14 || out["in_study"].(float64) != 0 {
		t.Fatalf("bank: total %v, %d groups, in study %v", out["total"], len(groups), out["in_study"])
	}
	first := groups[0].(map[string]any)
	item := first["items"].([]any)[0].(map[string]any)
	if first["name"] != "Mở bài" || !strings.HasPrefix(item["pattern"].(string), "Nowadays") || item["example_vi"] == "" {
		t.Fatalf("first group %v", first)
	}

	st := call(http.MethodPost, "/writing/phrases/study")
	if st["added"].(float64) != st["total"].(float64) {
		t.Fatalf("study: %v", st)
	}
	if again := call(http.MethodPost, "/writing/phrases/study"); again["added"].(float64) != 0 {
		t.Fatalf("second study: %v", again)
	}
	if out := call(http.MethodGet, "/writing/phrases"); out["in_study"] != out["total"] {
		t.Fatalf("after study: %v / %v", out["in_study"], out["total"])
	}

	// review options: other frames, the same group first ("Email: kết thư")
	var card models.Card
	gdb.Where("front = ?", "Trân trọng, (đã biết tên người nhận)").First(&card)
	ch, err := h.Study.MeaningChoices(card)
	if err != nil || len(ch) != 4 {
		t.Fatalf("choices %v %v", ch, err)
	}
	_, ps, _ := h.phraseCards()
	groupOf := map[string]string{}
	for _, p := range ps {
		groupOf[p.Back] = p.Group
	}
	right := false
	for _, c := range ch {
		if c.Text == "Yours sincerely," {
			right = true
		}
		if groupOf[c.Text] != "Email: kết thư" {
			t.Fatalf("option from another group: %q (%q)", c.Text, groupOf[c.Text])
		}
	}
	if !right {
		t.Fatalf("right answer missing: %v", ch)
	}
	if cc := content.ToCompact(card); cc.POS != "Email: kết thư" || cc.Back != "Yours sincerely," || cc.ExampleVi == "" {
		t.Fatalf("compact %+v", cc)
	}
	if lines := h.phraseLines(); len(lines) < 80 || !strings.HasPrefix(lines[0], "Mở bài: ") {
		t.Fatalf("lines %v", lines[:2])
	}
}
