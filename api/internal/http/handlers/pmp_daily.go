package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// PMP "Daily 20": each day the learner can generate 20 PMP-style practice
// questions (English, scenario-based), written by the AI from their own PMP
// cheatsheet notes and weighted like the exam domains (People 33%, Process 41%,
// Business Environment 26%). The set is stored per (user, day) so reopening the
// page shows the same questions; the score is computed here on submit.

type pmpDailySet struct {
	ID          int64       `gorm:"primaryKey"`
	UserID      int64       `gorm:"column:user_id"`
	Day         time.Time   `gorm:"type:date"`
	Questions   models.JSON `gorm:"type:jsonb"`
	Result      models.JSON `gorm:"type:jsonb"`
	Score       *int
	CreatedAt   time.Time
	SubmittedAt *time.Time
}

func (pmpDailySet) TableName() string { return "pmp_daily_sets" }

var dayRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// dailyDomains: question counts per exam domain (sum 20) and the note tag that
// marks a cheatsheet section as belonging to it.
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
	Reported  []int             `json:"reported"` // questions the learner reported as possibly wrong
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

func dailyView(s *pmpDailySet) gin.H {
	var res dailyResult
	_ = json.Unmarshal(s.Result, &res)
	score := 0
	if s.Score != nil {
		score = *s.Score
	}
	return gin.H{
		"date":      s.Day.Format("2006-01-02"),
		"questions": json.RawMessage(s.Questions),
		"submitted": s.SubmittedAt != nil,
		"score":     score,
		"result":    res,
	}
}

func (h *Handler) loadDaily(uid int64, day time.Time) (*pmpDailySet, error) {
	var s pmpDailySet
	err := h.DB.Where("user_id = ? AND day = ?", uid, day.Format("2006-01-02")).Limit(1).Find(&s).Error
	if err != nil {
		return nil, err
	}
	if s.ID == 0 {
		return nil, nil
	}
	return &s, nil
}

// PMPDailyGet: GET /api/pmp-daily?date=YYYY-MM-DD -> {"set": {...} | null}
func (h *Handler) PMPDailyGet(c *gin.Context) {
	day, ok := parseDay(c.Query("date"))
	if !ok {
		abort(c, http.StatusBadRequest, "invalid date")
		return
	}
	s, err := h.loadDaily(userID(c), day)
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

// PMPDailyHistory: GET /api/pmp-daily/history -> last 30 days {date, score, total, submitted}
func (h *Handler) PMPDailyHistory(c *gin.Context) {
	var rows []pmpDailySet
	if err := h.DB.Where("user_id = ?", userID(c)).Order("day DESC").Limit(30).Find(&rows).Error; err != nil {
		internal(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		var qs []json.RawMessage
		_ = json.Unmarshal(r.Questions, &qs)
		score := 0
		if r.Score != nil {
			score = *r.Score
		}
		out = append(out, gin.H{"date": r.Day.Format("2006-01-02"), "score": score, "total": len(qs), "submitted": r.SubmittedAt != nil})
	}
	c.JSON(http.StatusOK, gin.H{"days": out})
}

// PMPDailyGenerate: POST /api/pmp-daily/generate {date, model?}
// Idempotent: returns today's set if it exists, otherwise has the AI write it.
func (h *Handler) PMPDailyGenerate(c *gin.Context) {
	var req struct {
		Date  string `json:"date"`
		Model string `json:"model"`
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
	uid := userID(c)
	if s, err := h.loadDaily(uid, day); err != nil {
		internal(c, err)
		return
	} else if s != nil {
		c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
		return
	}
	if !h.AI.Enabled() {
		abort(c, http.StatusServiceUnavailable, ai.ErrNotConfigured.Error())
		return
	}

	// the learner's PMP cheatsheet notes
	var subject models.Subject
	if err := h.DB.Where("slug = ?", "pmp").Limit(1).Find(&subject).Error; err != nil || subject.ID == 0 {
		abort(c, http.StatusBadRequest, "PMP subject not found")
		return
	}
	var notes []models.CourseNote
	if err := h.DB.Where("user_id = ? AND subject_id = ?", uid, subject.ID).
		Order("sort_order ASC, id ASC").Limit(maxCourseNotes).Find(&notes).Error; err != nil {
		internal(c, err)
		return
	}
	if len(notes) == 0 {
		abort(c, http.StatusBadRequest, "add PMP cheatsheet notes first")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 170*time.Second)
	defer cancel()
	model := strings.TrimSpace(req.Model)
	if model != "" && !h.chatModelAllowed(ctx, model) {
		abort(c, http.StatusBadRequest, "that model is not available")
		return
	}

	avoid := h.recentDailyStems(uid)
	seed := uid*1000003 + day.Unix()

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
	if len(all) < 16 {
		if firstErr != nil && errors.Is(firstErr, ai.ErrNotConfigured) {
			abort(c, http.StatusServiceUnavailable, firstErr.Error())
			return
		}
		abort(c, http.StatusBadGateway, "AI could not write enough valid questions, please try again")
		return
	}
	// mix the domains like the real exam
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(all), func(a, b int) { all[a], all[b] = all[b], all[a] })

	raw, err := json.Marshal(all)
	if err != nil {
		internal(c, err)
		return
	}
	set := pmpDailySet{UserID: uid, Day: day, Questions: models.JSON(raw), Result: models.JSON("{}"), CreatedAt: time.Now().UTC()}
	if err := h.DB.Create(&set).Error; err != nil {
		// a parallel request may have created it first
		if s, e2 := h.loadDaily(uid, day); e2 == nil && s != nil {
			c.JSON(http.StatusOK, gin.H{"set": dailyView(s)})
			return
		}
		internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"set": dailyView(&set)})
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

// recentDailyStems: short stems of the last three sets per domain, so new
// questions do not repeat the same scenarios.
func (h *Handler) recentDailyStems(uid int64) map[string][]string {
	out := map[string][]string{}
	var rows []pmpDailySet
	if err := h.DB.Where("user_id = ?", uid).Order("day DESC").Limit(3).Find(&rows).Error; err != nil {
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

// PMPDailySubmit: POST /api/pmp-daily/submit {date, picks:[[idx]...], flagged:[i...], seconds}
// Scores the set (sets can be submitted once) and returns the stored set.
func (h *Handler) PMPDailySubmit(c *gin.Context) {
	var req struct {
		Date    string  `json:"date"`
		Picks   [][]int `json:"picks"`
		Flagged []int   `json:"flagged"`
		Seconds int     `json:"seconds"`
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
	uid := userID(c)
	s, err := h.loadDaily(uid, day)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "no set for that day")
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

// PMPDailyReport: POST /api/pmp-daily/report {date, index}
// The learner flags a question whose answer or explanation looks wrong (AI
// written questions can be wrong). Stored in the set's result so it can be
// reviewed later.
func (h *Handler) PMPDailyReport(c *gin.Context) {
	var req struct {
		Date  string `json:"date"`
		Index int    `json:"index"`
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
	s, err := h.loadDaily(userID(c), day)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "no set for that day")
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
