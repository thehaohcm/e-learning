package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// "Daily 20": each day the learner can generate 20 scenario practice questions
// (English) from their own notes. Originally PMP-only; now subject-aware (PMP
// uses the exam-domain weighting; English draws from the learner's vocabulary).
// More than one set can exist per (user, subject, day) so the learner can
// generate another batch, and a saved set can be re-taken later. The table is
// still called pmp_daily_sets for history compatibility.

type pmpDailySet struct {
	ID          int64       `gorm:"primaryKey"`
	UserID      int64       `gorm:"column:user_id"`
	SubjectID   int64       `gorm:"column:subject_id"`
	Seq         int         `gorm:"column:seq"`
	Day         time.Time   `gorm:"type:date"`
	Questions   models.JSON `gorm:"type:jsonb"`
	Result      models.JSON `gorm:"type:jsonb"`
	Score       *int
	CreatedAt   time.Time
	SubmittedAt *time.Time
}

func (pmpDailySet) TableName() string { return "pmp_daily_sets" }

var dayRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

var (
	errDailyNoNotes = errors.New("add cheatsheet notes first")
	errDailyNoVocab = errors.New("add some vocabulary words first")
)

// dailyDomains: question counts per PMP exam domain (sum 20) and the note tag
// that marks a cheatsheet section as belonging to it.
var dailyDomains = []struct {
	Name  string
	Tag   string
	Count int
}{
	{"People", "[People]", 7},
	{"Process", "[Process]", 8},
	{"Business", "[Business]", 5},
}

type dailyResult struct {
	Picks     [][]int           `json:"picks"`
	Flagged   []int             `json:"flagged"`
	Seconds   int               `json:"seconds"`
	ByDomain  map[string][2]int `json:"by_domain"` // domain -> [right, total]
	Correct   []bool            `json:"correct"`
	Submitted bool              `json:"submitted"`
	Reported  []int             `json:"reported"`            // questions the learner reported as possibly wrong
	AiReview  string            `json:"ai_review,omitempty"` // persisted AI review of the mistakes
}

// parseDay validates a YYYY-MM-DD date from the client: it must be a real date
// within a day of the server's clock (so a bad clock cannot create sets for
// arbitrary days).
func parseDay(s string) (time.Time, bool) {
	if !dayRe.MatchString(s) {
		return time.Time{}, false
	}
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	diff := time.Since(d)
	if diff < -48*time.Hour || diff > 72*time.Hour {
		return time.Time{}, false
	}
	return d, true
}

// dailySubject resolves the subject slug (default "pmp") to a subject row.
func (h *Handler) dailySubject(slug string) (models.Subject, bool) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		slug = "pmp"
	}
	var s models.Subject
	if err := h.DB.Where("slug = ?", slug).Limit(1).Find(&s).Error; err != nil || s.ID == 0 {
		return s, false
	}
	return s, true
}

func dailyView(s *pmpDailySet) gin.H {
	var res dailyResult
	_ = json.Unmarshal(s.Result, &res)
	score := 0
	if s.Score != nil {
		score = *s.Score
	}
	return gin.H{
		"id":        s.ID,
		"seq":       s.Seq,
		"date":      s.Day.Format("2006-01-02"),
		"questions": json.RawMessage(s.Questions),
		"submitted": s.SubmittedAt != nil,
		"score":     score,
		"result":    res,
	}
}

func (h *Handler) loadDailyByID(uid, id int64) (*pmpDailySet, error) {
	var s pmpDailySet
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).Limit(1).Find(&s).Error; err != nil {
		return nil, err
	}
	if s.ID == 0 {
		return nil, nil
	}
	return &s, nil
}

// latestDaily returns the most recent set (highest seq) for a subject on a day.
func (h *Handler) latestDaily(uid, subjectID int64, day time.Time) (*pmpDailySet, error) {
	var s pmpDailySet
	err := h.DB.Where("user_id = ? AND subject_id = ? AND day = ?", uid, subjectID, day.Format("2006-01-02")).
		Order("seq DESC").Limit(1).Find(&s).Error
	if err != nil {
		return nil, err
	}
	if s.ID == 0 {
		return nil, nil
	}
	return &s, nil
}

func (h *Handler) nextDailySeq(uid, subjectID int64, day time.Time) int {
	var maxSeq *int
	h.DB.Model(&pmpDailySet{}).
		Where("user_id = ? AND subject_id = ? AND day = ?", uid, subjectID, day.Format("2006-01-02")).
		Select("MAX(seq)").Scan(&maxSeq)
	if maxSeq == nil {
		return 1
	}
	return *maxSeq + 1
}

// PMPDailyGet: GET /api/pmp-daily?subject=<slug>&date=YYYY-MM-DD  (latest set
// that day) or GET /api/pmp-daily?id=<id> (a specific set). -> {"set": {...}|null}
func (h *Handler) PMPDailyGet(c *gin.Context) {
	uid := userID(c)
	if idStr := c.Query("id"); idStr != "" {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		s, err := h.loadDailyByID(uid, id)
		if err != nil {
			internal(c, err)
			return
		}
		if s == nil {
			c.JSON(http.StatusOK, gin.H{"set": nil})
			return
		}
		c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
		return
	}
	subj, ok := h.dailySubject(c.Query("subject"))
	if !ok {
		abort(c, http.StatusBadRequest, "subject not found")
		return
	}
	day, ok := parseDay(c.Query("date"))
	if !ok {
		abort(c, http.StatusBadRequest, "invalid date")
		return
	}
	s, err := h.latestDaily(uid, subj.ID, day)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		c.JSON(http.StatusOK, gin.H{"set": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
}

// PMPDailyHistory: GET /api/pmp-daily/history?subject=<slug>
// -> last sets {id, date, seq, score, total, submitted, has_review}
func (h *Handler) PMPDailyHistory(c *gin.Context) {
	subj, ok := h.dailySubject(c.Query("subject"))
	if !ok {
		abort(c, http.StatusBadRequest, "subject not found")
		return
	}
	var rows []pmpDailySet
	if err := h.DB.Where("user_id = ? AND subject_id = ?", userID(c), subj.ID).
		Order("day DESC, seq DESC").Limit(40).Find(&rows).Error; err != nil {
		internal(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		var qs []json.RawMessage
		_ = json.Unmarshal(r.Questions, &qs)
		var res dailyResult
		_ = json.Unmarshal(r.Result, &res)
		score := 0
		if r.Score != nil {
			score = *r.Score
		}
		out = append(out, gin.H{
			"id": r.ID, "date": r.Day.Format("2006-01-02"), "seq": r.Seq,
			"score": score, "total": len(qs), "submitted": r.SubmittedAt != nil,
			"has_review": res.AiReview != "",
		})
	}
	c.JSON(http.StatusOK, gin.H{"days": out})
}

// PMPDailyGenerate: POST /api/pmp-daily/generate {date, subject?, model?}
// Always creates a NEW set (the next batch for that subject+day).
func (h *Handler) PMPDailyGenerate(c *gin.Context) {
	var req struct {
		Date    string `json:"date"`
		Subject string `json:"subject"`
		Model   string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	day, ok := parseDay(req.Date)
	if !ok {
		abort(c, http.StatusBadRequest, "invalid date")
		return
	}
	subj, ok := h.dailySubject(req.Subject)
	if !ok {
		abort(c, http.StatusBadRequest, "subject not found")
		return
	}
	uid := userID(c)
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	model := strings.TrimSpace(req.Model)
	if model != "" && !h.chatModelAllowed(ctx, model) {
		abort(c, http.StatusBadRequest, "that model is not available")
		return
	}

	seq := h.nextDailySeq(uid, subj.ID, day)
	seed := uid*1000003 + day.Unix() + int64(seq)*7919

	var all []ai.DailyQuestion
	var genErr error
	if subj.Slug == "english" {
		all, genErr = h.generateEnglishDaily(ctx, model, uid, subj, seed)
	} else {
		all, genErr = h.generatePMPDaily(ctx, model, uid, subj, day, seed)
	}
	if genErr != nil {
		switch {
		case errors.Is(genErr, errDailyNoNotes):
			abort(c, http.StatusBadRequest, "add cheatsheet notes first")
			return
		case errors.Is(genErr, errDailyNoVocab):
			abort(c, http.StatusBadRequest, "add some vocabulary words to study first")
			return
		case errors.Is(genErr, ai.ErrNotConfigured):
			abort(c, http.StatusServiceUnavailable, genErr.Error())
			return
		}
	}
	if len(all) < 16 {
		abort(c, http.StatusBadGateway, "AI could not write enough valid questions, please try again")
		return
	}
	// mix the questions like the real exam
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(all), func(a, b int) { all[a], all[b] = all[b], all[a] })

	raw, err := json.Marshal(all)
	if err != nil {
		internal(c, err)
		return
	}
	set := pmpDailySet{UserID: uid, SubjectID: subj.ID, Seq: seq, Day: day, Questions: models.JSON(raw), Result: models.JSON("{}"), CreatedAt: time.Now().UTC()}
	if err := h.DB.Create(&set).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"set": dailyView(&set)})
}

// generatePMPDaily writes the 20 PMP questions, weighted by exam domain, from
// the learner's PMP cheatsheet notes.
func (h *Handler) generatePMPDaily(ctx context.Context, model string, uid int64, subject models.Subject, day time.Time, seed int64) ([]ai.DailyQuestion, error) {
	var notes []models.CourseNote
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, subject.ID).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes).Error; err != nil {
		return nil, err
	}
	if len(notes) == 0 {
		return nil, errDailyNoNotes
	}
	avoid := h.recentDailyStems(uid, subject.ID)

	results := make([][]ai.DailyQuestion, len(dailyDomains))
	errs := make([]error, len(dailyDomains))
	var wg sync.WaitGroup
	for i, d := range dailyDomains {
		wg.Add(1)
		go func(i int, name, tag string, count int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed + int64(i)*7919))
			picked := notesForDomain(notes, tag, name == "Process", rng)
			results[i], errs[i] = h.AI.GenerateDailyQuestions(ctx, model, name, count, picked, avoid[name], rng)
		}(i, d.Name, d.Tag, d.Count)
	}
	wg.Wait()

	var all []ai.DailyQuestion
	var firstErr error
	for i := range dailyDomains {
		if errs[i] != nil && firstErr == nil {
			firstErr = errs[i]
		}
		all = append(all, results[i]...)
	}
	if len(all) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return all, nil
}

// generateEnglishDaily writes 20 TOEIC/IELTS-style questions from a random
// sample of the learner's English vocabulary cards.
func (h *Handler) generateEnglishDaily(ctx context.Context, model string, uid int64, subject models.Subject, seed int64) ([]ai.DailyQuestion, error) {
	var cards []models.Card
	if err := h.DB.Joins("JOIN decks ON decks.id = cards.deck_id").
		Where("decks.subject_id = ? AND cards.kind = ?", subject.ID, models.KindVocab).
		Order("RANDOM()").Limit(70).Find(&cards).Error; err != nil {
		return nil, err
	}
	if len(cards) < 8 {
		return nil, errDailyNoVocab
	}
	words := make([]ai.EnglishWord, 0, len(cards))
	for _, cd := range cards {
		var p struct {
			Senses []struct {
				MeaningEN string `json:"meaning_en"`
				MeaningVI string `json:"meaning_vi"`
			} `json:"senses"`
			Topic string `json:"topic"`
		}
		_ = json.Unmarshal(cd.Payload, &p)
		me, mv := "", ""
		if len(p.Senses) > 0 {
			me, mv = p.Senses[0].MeaningEN, p.Senses[0].MeaningVI
		}
		words = append(words, ai.EnglishWord{Word: cd.Front, MeaningEN: me, MeaningVI: mv, Topic: p.Topic})
	}
	rng := rand.New(rand.NewSource(seed))
	return h.AI.GenerateEnglishDaily(ctx, model, 20, words, rng)
}

// notesForDomain orders the notes for one domain's prompt: sections tagged for
// the domain first, then general ([Chung]) ones, each group shuffled so the
// questions vary from day to day. For Process the formulas section goes first.
func notesForDomain(notes []models.CourseNote, tag string, pinFormulas bool, rng *rand.Rand) []ai.CheatNote {
	var domain, general, pinned []models.CourseNote
	for _, n := range notes {
		switch {
		case pinFormulas && strings.Contains(strings.ToLower(n.Title), "công thức"):
			pinned = append(pinned, n)
		case strings.Contains(n.Body, tag):
			domain = append(domain, n)
		case strings.Contains(n.Body, "[Chung]"):
			general = append(general, n)
		}
	}
	rng.Shuffle(len(domain), func(a, b int) { domain[a], domain[b] = domain[b], domain[a] })
	rng.Shuffle(len(general), func(a, b int) { general[a], general[b] = general[b], general[a] })
	ordered := append(append(pinned, domain...), general...)
	if len(ordered) == 0 { // untagged notes: use whatever there is
		ordered = notes
	}
	out := make([]ai.CheatNote, len(ordered))
	for i, n := range ordered {
		out[i] = ai.CheatNote{Title: n.Title, Body: n.Body}
	}
	return out
}

// recentDailyStems: short stems of the last three sets (for this subject), so
// new questions do not repeat the same scenarios.
func (h *Handler) recentDailyStems(uid, subjectID int64) map[string][]string {
	out := map[string][]string{}
	var rows []pmpDailySet
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, subjectID).
		Order("day DESC, seq DESC").Limit(3).Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		var qs []ai.DailyQuestion
		if json.Unmarshal(r.Questions, &qs) != nil {
			continue
		}
		for _, q := range qs {
			s := q.Stem
			if len(s) > 90 {
				s = s[:90]
			}
			out[q.Domain] = append(out[q.Domain], s)
		}
	}
	return out
}

// PMPDailySubmit: POST /api/pmp-daily/submit {id, picks:[[idx]...], flagged:[i...], seconds}
// Scores the set (sets can be submitted once) and returns the stored set.
func (h *Handler) PMPDailySubmit(c *gin.Context) {
	var req struct {
		ID      int64   `json:"id"`
		Picks   [][]int `json:"picks"`
		Flagged []int   `json:"flagged"`
		Seconds int     `json:"seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	uid := userID(c)
	s, err := h.loadDailyByID(uid, req.ID)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "set not found")
		return
	}
	if s.SubmittedAt != nil {
		abort(c, http.StatusConflict, "already submitted")
		return
	}
	var qs []ai.DailyQuestion
	if err := json.Unmarshal(s.Questions, &qs); err != nil {
		internal(c, err)
		return
	}
	res := dailyResult{
		Picks:     make([][]int, len(qs)),
		Seconds:   req.Seconds,
		ByDomain:  map[string][2]int{},
		Correct:   make([]bool, len(qs)),
		Submitted: true,
	}
	for _, f := range req.Flagged {
		if f >= 0 && f < len(qs) {
			res.Flagged = append(res.Flagged, f)
		}
	}
	score := 0
	for i, q := range qs {
		var pick []int
		if i < len(req.Picks) {
			seen := map[int]bool{}
			for _, p := range req.Picks[i] {
				if p >= 0 && p < len(q.Options) && !seen[p] {
					seen[p] = true
					pick = append(pick, p)
				}
			}
		}
		sort.Ints(pick)
		res.Picks[i] = pick
		right := len(pick) == len(q.Correct)
		if right {
			for k := range pick {
				if pick[k] != q.Correct[k] {
					right = false
					break
				}
			}
		}
		res.Correct[i] = right
		d := res.ByDomain[q.Domain]
		d[1]++
		if right {
			d[0]++
			score++
		}
		res.ByDomain[q.Domain] = d
	}
	raw, _ := json.Marshal(res)
	now := time.Now().UTC()
	if err := h.DB.Model(&pmpDailySet{}).Where("id = ?", s.ID).
		Updates(map[string]any{"result": models.JSON(raw), "score": score, "submitted_at": now}).Error; err != nil {
		internal(c, err)
		return
	}
	s.Result, s.Score, s.SubmittedAt = models.JSON(raw), &score, &now
	c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
}

// PMPDailyRetake: POST /api/pmp-daily/retake {id, date}
// Clones a saved set's questions into a new set for the given day, so the
// learner can take the same exam again (fresh answers and score).
func (h *Handler) PMPDailyRetake(c *gin.Context) {
	var req struct {
		ID   int64  `json:"id"`
		Date string `json:"date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	day, ok := parseDay(req.Date)
	if !ok {
		day = time.Now().UTC()
	}
	uid := userID(c)
	src, err := h.loadDailyByID(uid, req.ID)
	if err != nil {
		internal(c, err)
		return
	}
	if src == nil {
		abort(c, http.StatusNotFound, "set not found")
		return
	}
	seq := h.nextDailySeq(uid, src.SubjectID, day)
	set := pmpDailySet{UserID: uid, SubjectID: src.SubjectID, Seq: seq, Day: day, Questions: src.Questions, Result: models.JSON("{}"), CreatedAt: time.Now().UTC()}
	if err := h.DB.Create(&set).Error; err != nil {
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"set": dailyView(&set)})
}

// PMPDailyReport: POST /api/pmp-daily/report {id, index}
// The learner flags a question whose answer or explanation looks wrong.
func (h *Handler) PMPDailyReport(c *gin.Context) {
	var req struct {
		ID    int64 `json:"id"`
		Index int   `json:"index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bad(c, err)
		return
	}
	s, err := h.loadDailyByID(userID(c), req.ID)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "set not found")
		return
	}
	var qs []json.RawMessage
	if json.Unmarshal(s.Questions, &qs) != nil || req.Index < 0 || req.Index >= len(qs) {
		abort(c, http.StatusBadRequest, "invalid question")
		return
	}
	var res dailyResult
	_ = json.Unmarshal(s.Result, &res)
	for _, r := range res.Reported {
		if r == req.Index {
			c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
			return
		}
	}
	res.Reported = append(res.Reported, req.Index)
	raw, _ := json.Marshal(res)
	if err := h.DB.Model(&pmpDailySet{}).Where("id = ?", s.ID).Update("result", models.JSON(raw)).Error; err != nil {
		internal(c, err)
		return
	}
	s.Result = models.JSON(raw)
	c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
}

// PMPDailyAiReview: POST /api/pmp-daily/ai-review {id, model?, wrong:[...]}
// Returns a Vietnamese review of the concepts the learner got wrong, and
// persists it in the set so it is still there on the next load.
func (h *Handler) PMPDailyAiReview(c *gin.Context) {
	var req struct {
		ID    int64            `json:"id"`
		Model string           `json:"model"`
		Wrong []ai.MistakeItem `json:"wrong"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		abort(c, http.StatusBadRequest, "bad request")
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}
	uid := userID(c)
	s, err := h.loadDailyByID(uid, req.ID)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "set not found")
		return
	}
	// return a persisted review if one exists already
	var res dailyResult
	_ = json.Unmarshal(s.Result, &res)
	if res.AiReview != "" {
		c.JSON(http.StatusOK, gin.H{"review": res.AiReview})
		return
	}
	if len(req.Wrong) == 0 {
		c.JSON(http.StatusOK, gin.H{"review": "Bạn không sai câu nào — xuất sắc! 🎉"})
		return
	}
	if len(req.Wrong) > 30 {
		req.Wrong = req.Wrong[:30]
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()
	if req.Model != "" && !h.chatModelAllowed(ctx, req.Model) {
		abort(c, http.StatusBadRequest, "model not allowed")
		return
	}
	courseName := "PMP"
	var sj models.Subject
	if h.DB.Where("id = ?", s.SubjectID).Limit(1).Find(&sj).Error == nil && sj.Slug == "english" {
		courseName = "English (TOEIC/IELTS)"
	}
	review, err := h.AI.ReviewDailyMistakes(ctx, req.Model, courseName, req.Wrong)
	if err != nil {
		if errors.Is(err, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		abort(c, http.StatusBadGateway, "AI review failed: "+err.Error())
		return
	}
	// persist into the set's result
	res.AiReview = review
	if raw, e := json.Marshal(res); e == nil {
		h.DB.Model(&pmpDailySet{}).Where("id = ?", s.ID).Update("result", models.JSON(raw))
	}
	c.JSON(http.StatusOK, gin.H{"review": review})
}
