package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Course notes: a user's free-form notes / cheat-sheet items for one subject
// (Markdown), separate from word notes. "Make a cheat sheet" asks the AI to
// condense all the user's notes for the subject into one Markdown summary,
// cached per (user, subject) and regenerated on demand.

const (
	maxNoteTitle   = 200
	maxNoteBody    = 20000
	maxCourseNotes = 500
)

type courseNoteReq struct {
	SubjectID int64            `json:"subject_id"`
	Title     *string          `json:"title"`
	Body      *string          `json:"body"`
	Links     *json.RawMessage `json:"links"`
	SortOrder *int             `json:"sort_order"`
}

// ListCourseNotes: GET /api/course-notes?subject_id=
// {"notes":[...], "summary": {"summary_md","generated_at"} | null}
func (h *Handler) ListCourseNotes(c *gin.Context) {
	uid := userID(c)
	sid := queryInt64(c, "subject_id")
	if sid <= 0 {
		abort(c, http.StatusBadRequest, "subject_id is required")
		return
	}
	notes := []models.CourseNote{}
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, sid).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	var sum models.CourseNoteSummary
	var summary *models.CourseNoteSummary
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, sid).First(&sum).Error; err == nil {
		summary = &sum
	}
	c.JSON(http.StatusOK, gin.H{"notes": notes, "summary": summary})
}

func (h *Handler) subjectExists(id int64) bool {
	var n int64
	h.DB.Model(&models.Subject{}).Where("id = ?", id).Count(&n)
	return n > 0
}

// CreateCourseNote: POST /api/course-notes {"subject_id","title","body"}
func (h *Handler) CreateCourseNote(c *gin.Context) {
	var req courseNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if req.SubjectID <= 0 || !h.subjectExists(req.SubjectID) {
		abort(c, http.StatusBadRequest, "unknown subject")
		return
	}
	uid := userID(c)
	var n int64
	h.DB.Model(&models.CourseNote{}).Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).Count(&n)
	if n >= maxCourseNotes {
		abort(c, http.StatusBadRequest, "too many notes in this course")
		return
	}
	now := time.Now().UTC()
	note := models.CourseNote{
		UserID: uid, SubjectID: req.SubjectID,
		Title: clip(deref(req.Title), maxNoteTitle), Body: clip(deref(req.Body), maxNoteBody), Links: normLinks(req.Links),
		CreatedAt: now, UpdatedAt: now,
	}
	if req.SortOrder != nil {
		note.SortOrder = *req.SortOrder
	} else {
		// append to the end
		var maxOrder *int
		h.DB.Model(&models.CourseNote{}).Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).
			Select("MAX(sort_order)").Scan(&maxOrder)
		if maxOrder != nil {
			note.SortOrder = *maxOrder + 1
		}
	}
	if note.Title == "" && note.Body == "" {
		abort(c, http.StatusBadRequest, "an empty note cannot be saved")
		return
	}
	if err := h.DB.Create(&note).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, note)
}

// UpdateCourseNote: PATCH /api/course-notes/:id {"title","body","sort_order"}
func (h *Handler) UpdateCourseNote(c *gin.Context) {
	note, ok := h.loadCourseNote(c)
	if !ok {
		return
	}
	var req courseNoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	if req.Title != nil {
		note.Title = clip(*req.Title, maxNoteTitle)
	}
	if req.Body != nil {
		note.Body = clip(*req.Body, maxNoteBody)
	}
	if req.Links != nil {
		note.Links = normLinks(req.Links)
	}
	if req.SortOrder != nil {
		note.SortOrder = *req.SortOrder
	}
	note.UpdatedAt = time.Now().UTC()
	if err := h.DB.Save(&note).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, note)
}

// DeleteCourseNote: DELETE /api/course-notes/:id
func (h *Handler) DeleteCourseNote(c *gin.Context) {
	note, ok := h.loadCourseNote(c)
	if !ok {
		return
	}
	h.DB.Delete(&note)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) loadCourseNote(c *gin.Context) (models.CourseNote, bool) {
	id, ok := paramID(c, "id")
	if !ok {
		return models.CourseNote{}, false
	}
	var note models.CourseNote
	if err := h.DB.Where("id = ? AND user_id = ?", id, userID(c)).First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abort(c, http.StatusNotFound, "note not found")
		} else {
			internal(c, err)
		}
		return models.CourseNote{}, false
	}
	return note, true
}

// SummarizeCourseNotes: POST /api/course-notes/summarize {"subject_id"}
// Reads all the user's notes for the subject, asks the AI for one cheat-sheet
// summary (Markdown), stores it and returns it.
func (h *Handler) SummarizeCourseNotes(c *gin.Context) {
	var req struct {
		SubjectID int64 `json:"subject_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SubjectID <= 0 {
		abort(c, http.StatusBadRequest, "subject_id is required")
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	uid := userID(c)
	var notes []models.CourseNote
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	if len(notes) == 0 {
		abort(c, http.StatusBadRequest, "add some notes first")
		return
	}
	var subject models.Subject
	h.DB.First(&subject, req.SubjectID)

	items := make([]ai.CheatNote, len(notes))
	for i, n := range notes {
		items[i] = ai.CheatNote{Title: n.Title, Body: n.Body}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	md, err := h.AI.SummarizeNotes(ctx, subject.Name, items)
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("[ai] summarize notes user=%d subject=%d notes=%d: %v", uid, req.SubjectID, len(notes), err)
		abort(c, http.StatusBadGateway, "AI could not make the cheat sheet: "+err.Error())
		return
	}
	now := time.Now().UTC()
	sum := models.CourseNoteSummary{UserID: uid, SubjectID: req.SubjectID, SummaryMd: md, GeneratedAt: now}
	if err := h.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "subject_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"summary_md", "generated_at"}),
	}).Create(&sum).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, sum)
}

var noteHeadingRe = regexp.MustCompile(`^#{1,3}\s+(.*\S)\s*$`)

// splitSummary turns a Markdown cheat sheet into note items: each top-level
// heading (#, ##, ###) starts a new item (its title); text before the first
// heading becomes one untitled item. No heading at all -> a single item.
func splitSummary(md string) []models.CourseNote {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	type item struct{ title, body string }
	var items []item
	var title string
	var body []string
	flush := func() {
		b := strings.TrimSpace(strings.Join(body, "\n"))
		if title != "" || b != "" {
			items = append(items, item{title, b})
		}
		title, body = "", nil
	}
	for _, ln := range lines {
		if m := noteHeadingRe.FindStringSubmatch(ln); m != nil {
			flush()
			title = clip(strings.TrimSpace(m[1]), maxNoteTitle)
		} else {
			body = append(body, ln)
		}
	}
	flush()
	if len(items) == 0 {
		items = []item{{title: "Cheat sheet", body: strings.TrimSpace(md)}}
	}
	if len(items) > maxCourseNotes {
		items = items[:maxCourseNotes]
	}
	out := make([]models.CourseNote, len(items))
	for i, it := range items {
		out[i] = models.CourseNote{Title: it.title, Body: clip(it.body, maxNoteBody), SortOrder: i}
	}
	return out
}

// AdoptSummary: POST /api/course-notes/adopt-summary {"subject_id"}
// Replaces the user's notes for the subject with the AI cheat sheet, split into
// items by heading. Destructive (the old notes are removed) — the client asks
// the user to confirm first.
func (h *Handler) AdoptSummary(c *gin.Context) {
	var req struct {
		SubjectID int64 `json:"subject_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SubjectID <= 0 {
		abort(c, http.StatusBadRequest, "subject_id is required")
		return
	}
	uid := userID(c)
	var sum models.CourseNoteSummary
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).First(&sum).Error; err != nil || strings.TrimSpace(sum.SummaryMd) == "" {
		abort(c, http.StatusBadRequest, "make a cheat sheet first")
		return
	}
	items := splitSummary(sum.SummaryMd)
	now := time.Now().UTC()
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).Delete(&models.CourseNote{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].UserID = uid
			items[i].SubjectID = req.SubjectID
			items[i].CreatedAt = now
			items[i].UpdatedAt = now
		}
		return tx.Create(&items).Error
	})
	if err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": items})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// AskCourse: POST /api/course-notes/ask {"subject_id","question","history":[{"q","a"}]}
// A free-form chat about the course, grounded in the user's own notes.
func (h *Handler) AskCourse(c *gin.Context) {
	var req struct {
		SubjectID int64           `json:"subject_id"`
		Question  string          `json:"question"`
		Model     string          `json:"model"`
		History   []ai.CourseTurn `json:"history"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.SubjectID <= 0 || req.Question == "" {
		abort(c, http.StatusBadRequest, "subject_id and question are required")
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	if len(req.History) > 10 {
		req.History = req.History[len(req.History)-10:]
	}
	uid := userID(c)
	var subject models.Subject
	if err := h.DB.First(&subject, req.SubjectID).Error; err != nil {
		abort(c, http.StatusBadRequest, "unknown subject")
		return
	}
	// the user's own notes, as context (bounded)
	var notes []models.CourseNote
	h.DB.Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes)
	var ctxb strings.Builder
	for _, n := range notes {
		if n.Title != "" {
			ctxb.WriteString("## " + n.Title + "\n")
		}
		ctxb.WriteString(n.Body + "\n\n")
		if ctxb.Len() > 16000 {
			break
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()
	model := strings.TrimSpace(req.Model)
	if model != "" && !h.chatModelAllowed(ctx, model) {
		abort(c, http.StatusBadRequest, "that model is not available")
		return
	}
	answer, err := h.AI.AskCourse(ctx, model, subject.Name, ctxb.String(), req.History, req.Question)
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		abort(c, http.StatusBadGateway, "AI could not answer, please try again")
		return
	}
	c.JSON(http.StatusOK, gin.H{"answer": answer})
}

// chatModelList returns the models the course chat may use: the admin
// allow-list (CHAT_MODELS) if set, otherwise whatever the gateway lists.
func (h *Handler) chatModelList(ctx context.Context) []string {
	if len(h.Cfg.ChatModels) > 0 {
		return h.Cfg.ChatModels
	}
	if !h.AI.Enabled() {
		return nil
	}
	ids, err := h.AI.ListModels(ctx)
	if err != nil {
		return nil
	}
	return ids
}

func (h *Handler) chatModelAllowed(ctx context.Context, model string) bool {
	for _, m := range h.chatModelList(ctx) {
		if m == model {
			return true
		}
	}
	return false
}

// ChatModels: GET /api/chat-models -> {"models":[...]} for the chat model picker.
func (h *Handler) ChatModels(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	models := h.chatModelList(ctx)
	if models == nil {
		models = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

// GenerateCourseQuestions: POST /api/course-notes/gen-questions
// {subject_id, topic?, count?, model?} — the AI writes multiple-choice study
// questions from the learner's own notes and inserts them (kind mcq) into the
// subject's "cheatsheet" deck. Returns the created card ids so the client can
// enroll + study them. Admin only (inserts into shared cards).
func (h *Handler) GenerateCourseQuestions(c *gin.Context) {
	var req struct {
		SubjectID  int64  `json:"subject_id"`
		Topic      string `json:"topic"`
		Count      int    `json:"count"`
		Model      string `json:"model"`
		Format     string `json:"format"`     // "lab" = typed-command drills, else "mcq"
		Exhaustive bool   `json:"exhaustive"` // lab only: deterministically extract every command line from the matching notes' code blocks and make exactly one drill per line (see ai.ExtractCommandLines), instead of letting the AI decide how many/which commands to cover
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.SubjectID <= 0 {
		abort(c, http.StatusBadRequest, "subject_id is required")
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	uid := userID(c)
	var notes []models.CourseNote
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, req.SubjectID).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	if len(notes) == 0 {
		abort(c, http.StatusBadRequest, "add some notes first")
		return
	}
	// both formats study in the cheatsheet deck (SRS): lab = typed-command cards,
	// mcq = multiple-choice cards.
	isLab := req.Format == "lab"
	kind := models.KindMCQ
	if isLab {
		kind = models.KindLab
	}
	var deck models.Deck
	if err := h.DB.Where("subject_id = ? AND slug = ?", req.SubjectID, "cheatsheet").First(&deck).Error; err != nil {
		notFoundOr500(c, err, "cheatsheet deck")
		return
	}
	var subject models.Subject
	h.DB.First(&subject, req.SubjectID)

	items := make([]ai.CheatNote, len(notes))
	for i, n := range notes {
		items[i] = ai.CheatNote{Title: n.Title, Body: n.Body}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	if req.Model != "" && !h.chatModelAllowed(ctx, req.Model) {
		abort(c, http.StatusBadRequest, "model not allowed")
		return
	}

	extractedTotal := 0
	var gen []ai.GeneratedCard
	var err error
	switch {
	case isLab && req.Exhaustive:
		gen, extractedTotal, err = h.generateExhaustiveLabDrills(ctx, req.Model, req.Topic, subject.Name, notes)
	case isLab:
		gen, err = h.AI.GenerateDrillsFromNotes(ctx, req.Model, subject.Name, req.Topic, req.Count, items)
	default:
		gen, err = h.AI.GenerateQuestionsFromNotes(ctx, req.Model, subject.Name, req.Topic, req.Count, items)
	}
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		log.Printf("[ai] gen-questions (format=%s exhaustive=%v) user=%d subject=%d: %v", req.Format, req.Exhaustive, uid, req.SubjectID, err)
		abort(c, http.StatusBadGateway, "AI could not generate questions: "+err.Error())
		return
	}

	ids := []int64{}
	skipped := 0
	topic := strings.TrimSpace(req.Topic)
	for _, g := range gen {
		front := strings.TrimSpace(g.Front)
		raw := g.Payload
		if topic != "" {
			raw = withPayloadTopic(raw, topic)
		}
		payload, perr := content.Normalize(kind, front, raw)
		if perr != nil {
			skipped++
			continue
		}
		card := models.Card{DeckID: deck.ID, Kind: kind, Front: front, Payload: models.JSON(payload), Source: "ai"}
		res := h.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&card)
		if res.Error != nil || res.RowsAffected == 0 {
			skipped++
			continue
		}
		ids = append(ids, card.ID)
	}
	c.JSON(http.StatusOK, gin.H{"created": len(ids), "skipped": skipped, "card_ids": ids, "deck_id": deck.ID, "format": req.Format, "extracted": extractedTotal})
}

// generateExhaustiveLabDrills deterministically extracts every command line
// from the notes whose title matches topic (or all notes, if none match or
// topic is empty), then asks the AI to write a drill for each one in small
// batches — never letting the AI decide coverage, only wording. Any command
// a batch call fails or forgets to cover falls back to a safe, AI-free
// literal-match card (ai.BuildFallbackDrillCard), so the count returned
// always equals the number of commands extracted. Returns the generated
// cards, the total number of commands extracted (for the caller to report),
// and an error only when nothing could be extracted at all.
func (h *Handler) generateExhaustiveLabDrills(ctx context.Context, model, topic, courseName string, notes []models.CourseNote) ([]ai.GeneratedCard, int, error) {
	topic = strings.TrimSpace(topic)
	scoped := notes
	if topic != "" {
		lt := strings.ToLower(topic)
		var matched []models.CourseNote
		for _, n := range notes {
			if strings.Contains(strings.ToLower(n.Title), lt) {
				matched = append(matched, n)
			}
		}
		if len(matched) > 0 {
			scoped = matched
		}
	}
	cheatNotes := make([]ai.CheatNote, len(scoped))
	for i, n := range scoped {
		cheatNotes[i] = ai.CheatNote{Title: n.Title, Body: n.Body}
	}
	lines := ai.ExtractCommandLines(cheatNotes)
	if len(lines) == 0 {
		return nil, 0, fmt.Errorf("không tìm thấy dòng lệnh nào trong ```code block``` của ghi chú phù hợp")
	}
	const batchSize = 15
	gen := make([]ai.GeneratedCard, 0, len(lines))
	for start := 0; start < len(lines); start += batchSize {
		end := start + batchSize
		if end > len(lines) {
			end = len(lines)
		}
		batch := lines[start:end]
		cards, err := h.AI.GenerateExactDrills(ctx, model, courseName, batch)
		if err != nil {
			log.Printf("[ai] exhaustive drill batch [%d:%d]/%d: %v", start, end, len(lines), err)
			cards = nil // fall through: every line in this batch gets the AI-free fallback below
		}
		for i, l := range batch {
			if i < len(cards) && strings.TrimSpace(cards[i].Front) != "" && len(cards[i].Payload) > 0 {
				gen = append(gen, cards[i])
			} else {
				gen = append(gen, ai.BuildFallbackDrillCard(l))
			}
		}
	}
	return gen, len(lines), nil
}

// withPayloadTopic adds a "topic" field to a card payload JSON object so the
// Study "practice by section" filter can group generated drills by their
// cheatsheet section. Returns the input unchanged if it is not a JSON object.
func withPayloadTopic(raw json.RawMessage, topic string) json.RawMessage {
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return raw
	}
	tb, err := json.Marshal(topic)
	if err != nil {
		return raw
	}
	m["topic"] = tb
	out, err := json.Marshal(m)
	if err != nil {
		return raw
	}
	return out
}

// normLinks validates the resource links on a course note: keeps only entries
// with a non-empty url, trims/caps lengths, limits the count, and always
// returns a valid JSON array (never null) for the jsonb column.
func normLinks(raw *json.RawMessage) models.JSON {
	if raw == nil {
		return models.JSON("[]")
	}
	type noteLink struct {
		URL   string `json:"url"`
		Label string `json:"label,omitempty"`
	}
	var in []noteLink
	if err := json.Unmarshal(*raw, &in); err != nil {
		return models.JSON("[]")
	}
	out := make([]noteLink, 0, len(in))
	for _, l := range in {
		u := strings.TrimSpace(l.URL)
		if u == "" {
			continue
		}
		if len(u) > 500 {
			u = u[:500]
		}
		lab := strings.TrimSpace(l.Label)
		if len(lab) > 120 {
			lab = lab[:120]
		}
		out = append(out, noteLink{URL: u, Label: lab})
		if len(out) >= 20 {
			break
		}
	}
	b, err := json.Marshal(out)
	if err != nil {
		return models.JSON("[]")
	}
	return models.JSON(b)
}
