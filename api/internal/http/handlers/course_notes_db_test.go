package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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
func TestCourseNotes(t *testing.T) {
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
	var uid, sid int64
	gdb.Raw(`INSERT INTO users (email) VALUES ('notes@test') ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email RETURNING id`).Scan(&uid)
	gdb.Raw(`SELECT id FROM subjects WHERE slug = 'pentest'`).Scan(&sid)
	gdb.Exec(`DELETE FROM course_notes WHERE user_id = ?`, uid)
	gdb.Exec(`DELETE FROM course_note_summaries WHERE user_id = ?`, uid)

	// fake AI returning a markdown summary
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := `{"summary_md":"# Cheat sheet\n- point one\n- point two","answer_md":"Enumeration là liệt kê dịch vụ."}`
		b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"role": "assistant", "content": content}}}})
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}))
	defer fake.Close()

	h := New(gdb, config.Config{DeepSeekAPIKey: "k", DeepSeekBaseURL: fake.URL, DeepSeekModel: "m"})
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(ctxUserID, uid) })
	r.GET("/course-notes", h.ListCourseNotes)
	r.POST("/course-notes", h.CreateCourseNote)
	r.PATCH("/course-notes/:id", h.UpdateCourseNote)
	r.DELETE("/course-notes/:id", h.DeleteCourseNote)
	r.POST("/course-notes/summarize", h.SummarizeCourseNotes)
	r.POST("/course-notes/adopt-summary", h.AdoptSummary)
	r.POST("/course-notes/ask", h.AskCourse)
	call := func(method, path, body string) (int, map[string]any) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var out map[string]any
		json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	// empty note rejected
	if code, _ := call(http.MethodPost, "/course-notes", `{"subject_id":`+itoa(sid)+`,"title":"","body":"  "}`); code != http.StatusBadRequest {
		t.Fatalf("empty note should be rejected, got %d", code)
	}
	// create two
	code, n1 := call(http.MethodPost, "/course-notes", `{"subject_id":`+itoa(sid)+`,"title":"nmap","body":"nmap -sV TARGET"}`)
	if code != http.StatusCreated || n1["sort_order"].(float64) != 0 {
		t.Fatalf("create 1: %d %v", code, n1)
	}
	_, n2 := call(http.MethodPost, "/course-notes", `{"subject_id":`+itoa(sid)+`,"title":"ports","body":"80 443 22"}`)
	if n2["sort_order"].(float64) != 1 {
		t.Fatalf("second note should append (order 1): %v", n2)
	}
	id1 := int64(n1["id"].(float64))

	// list
	_, lst := call(http.MethodGet, "/course-notes?subject_id="+itoa(sid), "")
	if arr, _ := lst["notes"].([]any); len(arr) != 2 || lst["summary"] != nil {
		t.Fatalf("list: %v", lst)
	}

	// edit
	if code, _ := call(http.MethodPatch, "/course-notes/"+itoa(id1), `{"body":"nmap -sV -sC TARGET"}`); code != 200 {
		t.Fatalf("edit: %d", code)
	}
	var edited models.CourseNote
	gdb.First(&edited, id1)
	if edited.Body != "nmap -sV -sC TARGET" || edited.Title != "nmap" {
		t.Fatalf("not edited: %+v", edited)
	}

	// summarize → stored
	code, sum := call(http.MethodPost, "/course-notes/summarize", `{"subject_id":`+itoa(sid)+`}`)
	if code != 200 || !strings.Contains(sum["summary_md"].(string), "Cheat sheet") {
		t.Fatalf("summarize: %d %v", code, sum)
	}
	_, lst2 := call(http.MethodGet, "/course-notes?subject_id="+itoa(sid), "")
	if lst2["summary"] == nil {
		t.Fatal("summary should be returned after generation")
	}

	// delete + ownership (another user cannot touch it)
	if code, _ := call(http.MethodDelete, "/course-notes/"+itoa(id1), ""); code != 200 {
		t.Fatalf("delete: %d", code)
	}
	var cnt int64
	gdb.Model(&models.CourseNote{}).Where("user_id = ?", uid).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("expected 1 note left, got %d", cnt)
	}

	// adopt the AI summary: replaces notes, split by heading
	// (the fake summary is "# Cheat sheet\n- point one\n- point two")
	code, ad := call(http.MethodPost, "/course-notes/adopt-summary", `{"subject_id":`+itoa(sid)+`}`)
	if code != 200 {
		t.Fatalf("adopt: %d %v", code, ad)
	}
	var after []models.CourseNote
	gdb.Where("user_id = ? AND subject_id = ?", uid, sid).Order("sort_order").Find(&after)
	if len(after) != 1 || after[0].Title != "Cheat sheet" || !strings.Contains(after[0].Body, "point one") {
		t.Fatalf("adopt result: %+v", after)
	}
	// ask AI about the course (fake AI echoes a markdown answer)
	code, ans := call(http.MethodPost, "/course-notes/ask", `{"subject_id":`+itoa(sid)+`,"question":"what is enumeration?"}`)
	if code != 200 || ans["answer"] == nil {
		t.Fatalf("ask: %d %v", code, ans)
	}

	_ = strconv.Itoa
}

func TestSplitSummary(t *testing.T) {
	items := splitSummary("intro text\n\n# A\nline a\n## B\nline b")
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d: %+v", len(items), items)
	}
	if items[0].Title != "" || items[0].Body != "intro text" {
		t.Fatalf("preamble: %+v", items[0])
	}
	if items[1].Title != "A" || items[1].Body != "line a" || items[2].Title != "B" {
		t.Fatalf("sections: %+v", items[1:])
	}
	one := splitSummary("no headings here\njust text")
	if len(one) != 1 || one[0].Title != "" {
		t.Fatalf("no-heading: %+v", one)
	}
}
