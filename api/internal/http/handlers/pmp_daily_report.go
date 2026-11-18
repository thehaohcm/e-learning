package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"mvp-lang-app/internal/ai"
	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// "Daily report": day-over-day competency across every SUBMITTED daily set for a
// subject, so the learner sees which exam components they are strong/weak in and
// gets a cautious readiness signal. Everything here is computed from data already
// stored on each set (per-question domain/approach/section/archetype for MCQ, or
// per-step skill for labs, plus correctness) — nothing is regenerated.

const (
	rptPassPct   = 75 // per-component accuracy (recent window) to look exam-ready
	rptMinSets   = 8  // submitted sets before a "ready" verdict is credible
	rptCovPct    = 70 // share of cheatsheet sections the learner has been tested on
	rptRecentWin = 8  // how many recent sets feed the readiness check
	rptMinForBar = 3  // ignore a component/topic with too few questions to be meaningful
)

type rptGroup struct {
	Key   string `json:"key"`
	Right int    `json:"right"`
	Total int    `json:"total"`
	Pct   int    `json:"pct"`
}

type rptTrendPoint struct {
	Date string `json:"date"`
	Pct  int    `json:"pct"`
}

type rptCoverage struct {
	Seen  int `json:"seen"`
	Total int `json:"total"`
	Pct   int `json:"pct"`
}

type rptReadiness struct {
	Level   string   `json:"level"` // ready | almost | not_ready
	Label   string   `json:"label"`
	Reasons []string `json:"reasons"`
}

type rptOut struct {
	Subject        string          `json:"subject"`
	GroupLabel     string          `json:"group_label"`
	SetsCounted    int             `json:"sets_counted"`
	QuestionsCount int             `json:"questions_counted"`
	Overall        rptGroup        `json:"overall"`
	ByGroup        []rptGroup      `json:"by_group"`
	WeakTopics     []rptGroup      `json:"weak_topics"`
	Trend          []rptTrendPoint `json:"trend"`
	Coverage       *rptCoverage    `json:"coverage,omitempty"`
	Readiness      rptReadiness    `json:"readiness"`
	Disclaimer     string          `json:"disclaimer"`
}

type rptAcc struct{ right, total int }

func (a *rptAcc) add(right bool) {
	a.total++
	if right {
		a.right++
	}
}

func rptPct(right, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(100 * float64(right) / float64(total)))
}

// PMPDailyCompetency: GET /api/pmp-daily/competency?subject=<slug>
// Aggregated strong/weak report over all submitted sets of the subject.
func (h *Handler) PMPDailyCompetency(c *gin.Context) {
	subj, ok := h.dailySubject(c.Query("subject"))
	if !ok {
		abort(c, http.StatusBadRequest, "subject not found")
		return
	}
	uid := userID(c)

	var rows []pmpDailySet
	if err := h.DB.Where("user_id = ? AND subject_id = ? AND submitted_at IS NOT NULL", uid, subj.ID).
		Order("day ASC, seq ASC").Limit(200).Find(&rows).Error; err != nil {
		internal(c, err)
		return
	}

	isLab := subj.Slug == "pentest"
	groupLabel := "Domain"
	var rank map[string]int
	switch {
	case isLab:
		groupLabel = "Kỹ năng"
	case subj.Slug == "english":
		groupLabel, rank = "Kiểu câu", map[string]int{"vocabulary": 0, "grammar": 1, "reading": 2}
	case subj.Slug == "aws":
		groupLabel, rank = "Lĩnh vực SAA-C03", map[string]int{"Secure": 0, "Resilient": 1, "Performance": 2, "Cost": 3}
	default:
		rank = map[string]int{"People": 0, "Process": 1, "Business": 2}
	}

	groups := map[string]*rptAcc{}
	recentGroups := map[string]*rptAcc{}
	topics := map[string]*rptAcc{}
	seenSections := map[string]bool{}
	byDay := map[string]*rptAcc{}
	overall, recentOverall := rptAcc{}, rptAcc{}
	questions := 0

	recentFrom := 0
	if len(rows) > rptRecentWin {
		recentFrom = len(rows) - rptRecentWin
	}
	bump := func(m map[string]*rptAcc, key string, right bool) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if m[key] == nil {
			m[key] = &rptAcc{}
		}
		m[key].add(right)
	}

	for idx, r := range rows {
		var res dailyResult
		_ = json.Unmarshal(r.Result, &res)
		correct := res.Correct
		recent := idx >= recentFrom
		day := r.Day.Format("2006-01-02")
		if byDay[day] == nil {
			byDay[day] = &rptAcc{}
		}

		if isLab {
			var steps []ai.FlowStep
			if json.Unmarshal(r.Questions, &steps) != nil {
				continue
			}
			for i, st := range steps {
				if i >= len(correct) {
					break
				}
				okAns := correct[i]
				skill := strings.TrimSpace(st.Skill)
				if skill == "" {
					skill = "khác"
				}
				bump(groups, skill, okAns)
				if recent {
					bump(recentGroups, skill, okAns)
				}
				byDay[day].add(okAns)
				overall.add(okAns)
				if recent {
					recentOverall.add(okAns)
				}
				questions++
			}
			continue
		}

		var qs []ai.DailyQuestion
		if json.Unmarshal(r.Questions, &qs) != nil {
			continue
		}
		for i, q := range qs {
			if i >= len(correct) {
				break
			}
			okAns := correct[i]
			g := q.Domain
			if subj.Slug == "english" {
				g = q.Approach
			}
			bump(groups, g, okAns)
			if recent {
				bump(recentGroups, g, okAns)
			}
			topic := strings.TrimSpace(q.Section)
			if topic == "" {
				topic = strings.TrimSpace(q.Archetype)
			}
			bump(topics, topic, okAns)
			if s := strings.ToLower(strings.TrimSpace(q.Section)); s != "" {
				seenSections[s] = true
			}
			byDay[day].add(okAns)
			overall.add(okAns)
			if recent {
				recentOverall.add(okAns)
			}
			questions++
		}
	}

	out := rptOut{
		Subject:        subj.Slug,
		GroupLabel:     groupLabel,
		SetsCounted:    len(rows),
		QuestionsCount: questions,
		Overall:        rptGroup{Right: overall.right, Total: overall.total, Pct: rptPct(overall.right, overall.total)},
		ByGroup:        rptGroupsSlice(groups, rank),
		Trend:          rptTrendSlice(byDay),
		Disclaimer:     rptDisclaimer(subj.Slug),
	}
	if !isLab {
		out.WeakTopics = rptWeakTopics(topics)
		var total64 int64
		h.DB.Model(&models.CourseNote{}).Where("user_id = ? AND subject_id = ?", uid, subj.ID).Count(&total64)
		seen := len(seenSections)
		if total64 > 0 && seen > int(total64) {
			seen = int(total64)
		}
		out.Coverage = &rptCoverage{Seen: seen, Total: int(total64), Pct: rptPct(seen, int(total64))}
	}
	out.Readiness = rptReadinessOf(len(rows), recentOverall, rptGroupsSlice(recentGroups, rank), out.Coverage)

	c.JSON(http.StatusOK, out)
}

// rptGroupsSlice turns an accumulator map into a sorted slice: known keys in
// their exam order first (via rank), then the rest by volume so the most-
// practised components lead.
func rptGroupsSlice(m map[string]*rptAcc, rank map[string]int) []rptGroup {
	out := make([]rptGroup, 0, len(m))
	for k, a := range m {
		out = append(out, rptGroup{Key: k, Right: a.right, Total: a.total, Pct: rptPct(a.right, a.total)})
	}
	sort.Slice(out, func(i, j int) bool {
		ri, oki := rank[out[i].Key]
		rj, okj := rank[out[j].Key]
		if oki && okj {
			return ri < rj
		}
		if oki != okj {
			return oki
		}
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// rptWeakTopics ranks sections/archetypes by accuracy (worst first), ignoring
// ones with too few questions to be meaningful. Up to 8.
func rptWeakTopics(m map[string]*rptAcc) []rptGroup {
	out := make([]rptGroup, 0, len(m))
	for k, a := range m {
		if a.total < rptMinForBar {
			continue
		}
		out = append(out, rptGroup{Key: k, Right: a.right, Total: a.total, Pct: rptPct(a.right, a.total)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pct != out[j].Pct {
			return out[i].Pct < out[j].Pct
		}
		return out[i].Total > out[j].Total
	})
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

func rptTrendSlice(byDay map[string]*rptAcc) []rptTrendPoint {
	out := make([]rptTrendPoint, 0, len(byDay))
	for d, a := range byDay {
		out = append(out, rptTrendPoint{Date: d, Pct: rptPct(a.right, a.total)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	if len(out) > 30 {
		out = out[len(out)-30:]
	}
	return out
}

// rptReadinessOf is a cautious composite, never a promise: every recent
// component at or above the bar, enough coverage, and enough sets. Reasons list
// whatever is holding it back.
func rptReadinessOf(sets int, recent rptAcc, recentGroups []rptGroup, cov *rptCoverage) rptReadiness {
	reasons := []string{}
	if sets < rptMinSets {
		reasons = append(reasons, fmt.Sprintf("Mới có %d/%d bộ đã nộp", sets, rptMinSets))
	}
	for _, g := range recentGroups {
		if g.Total >= rptMinForBar && g.Pct < rptPassPct {
			reasons = append(reasons, fmt.Sprintf("%s còn %d%% (cần ≥%d%%)", g.Key, g.Pct, rptPassPct))
		}
	}
	if cov != nil && cov.Total > 0 && cov.Pct < rptCovPct {
		reasons = append(reasons, fmt.Sprintf("Độ phủ chủ đề %d%% (cần ≥%d%%)", cov.Pct, rptCovPct))
	}
	recentPct := rptPct(recent.right, recent.total)
	switch {
	case sets >= rptMinSets && len(reasons) == 0:
		return rptReadiness{Level: "ready", Label: "Đủ tự tin để thi thử sát thật", Reasons: reasons}
	case sets >= rptMinForBar && recentPct >= 65:
		return rptReadiness{Level: "almost", Label: "Gần sẵn sàng", Reasons: reasons}
	default:
		return rptReadiness{Level: "not_ready", Label: "Chưa sẵn sàng", Reasons: reasons}
	}
}

func rptDisclaimer(slug string) string {
	ref := "điểm thi thật"
	switch slug {
	case "pmp":
		ref = "điểm PMI thật"
	case "english":
		ref = "điểm TOEIC/IELTS thật"
	case "pentest":
		ref = "kết quả bài thi PT1/WEB1/CPTS thật"
	case "aws":
		ref = "kết quả thi AWS SAA-C03 thật"
	}
	return "Báo cáo phản ánh độ thành thạo trên chính cheatsheet của bạn — câu hỏi do AI viết từ note của bạn, nên điểm cao chưa tương đương " + ref + " (đề thật phủ rộng hơn). Hãy xem đây là tín hiệu tương đối, và mở rộng cheatsheet để tăng độ phủ."
}
