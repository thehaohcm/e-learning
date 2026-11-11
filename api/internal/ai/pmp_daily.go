package ai

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// PMP "Daily 20": original PMP-style practice questions in English, written by
// the AI from the learner's own cheatsheet notes (never from real exam or
// brain-dump questions). One call per exam domain (People / Process / Business
// Environment) so the set follows the exam's domain weights.

// DailyTerm is a key English term with its Vietnamese meaning.
type DailyTerm struct {
	EN string `json:"en"`
	VI string `json:"vi"`
}

// DailyOption is one answer choice with the reason it is right or wrong.
type DailyOption struct {
	Text string `json:"text"`
	Why  string `json:"why"`
}

// DailyQuestion is one practice question ("single" = 1 answer of 4,
// "multi" = choose TWO of 5).
type DailyQuestion struct {
	Domain      string        `json:"domain"`   // People | Process | Business
	Approach    string        `json:"approach"` // predictive | agile | hybrid
	Type        string        `json:"type"`     // single | multi
	Stem        string        `json:"stem"`
	Options     []DailyOption `json:"options"`
	Correct     []int         `json:"correct"` // indexes into Options
	Explanation string        `json:"explanation"`
	Section     string        `json:"section"` // cheatsheet section it is based on
	Terms       []DailyTerm   `json:"terms"`
}

const (
	dailyNoteCap   = 3500  // per note
	dailyNotesCap  = 22000 // all notes in one prompt
	dailyMaxStem   = 1200
	dailyMaxOption = 400
)

// GenerateDailyQuestions writes up to count valid questions for one domain.
// notes are the study notes to ground them in (already ordered by relevance);
// avoid lists recent scenarios not to repeat. rng must not be shared between
// goroutines (it is used to shuffle the answer options).
func (c *Client) GenerateDailyQuestions(ctx context.Context, model, domain string, count int, notes []CheatNote, avoid []string, rng *rand.Rand) ([]DailyQuestion, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var got []DailyQuestion
	var lastErr error
	for attempt := 0; attempt < 2 && len(got) < count; attempt++ {
		need := count - len(got)
		skip := append([]string{}, avoid...)
		for _, q := range got {
			skip = append(skip, truncate(q.Stem, 90))
		}
		raw, err := c.dailyBatch(ctx, model, domain, need, notes, skip)
		if err != nil {
			lastErr = err
			continue
		}
		for _, q := range raw {
			if v, ok := normalizeDaily(q, domain, rng); ok {
				got = append(got, v)
			}
		}
	}
	if len(got) == 0 && lastErr != nil {
		return nil, lastErr
	}
	if len(got) > count {
		got = got[:count]
	}
	return got, nil
}

func (c *Client) dailyBatch(ctx context.Context, model, domain string, n int, notes []CheatNote, avoid []string) ([]DailyQuestion, error) {
	var nb strings.Builder
	for _, nt := range notes {
		if nb.Len() > dailyNotesCap {
			break
		}
		fmt.Fprintf(&nb, "## %s\n%s\n\n", nt.Title, truncate(nt.Body, dailyNoteCap))
	}
	predictive := (2*n + 2) / 5 // about 40% predictive
	multi := 0
	if n >= 4 {
		multi = 1
	}
	calc := "No calculation questions."
	if domain == "Process" && n >= 6 {
		calc = "Include 2 calculation questions (EVM, PERT, float, communication channels, EMV, etc.), type single."
	}
	avoidText := ""
	if len(avoid) > 0 {
		avoidText = "\nDo not reuse these recent scenarios:\n"
		for _, a := range avoid {
			avoidText += "- " + a + "\n"
		}
	}
	system := "You are an experienced PMP exam item writer. You write original exam-style questions. Reply with a single JSON object only."
	user := fmt.Sprintf(`Write %d ORIGINAL practice questions in the style of the PMI PMP exam (2026 Exam Content Outline), domain: %s.
Base every question ONLY on facts in the study notes below (they are Vietnamese with English terms; write the questions in English). Never copy or recall real exam questions, brain dumps or copyrighted prep questions.

Requirements:
- Scenario-based: each stem is a short realistic situation (2-5 sentences) usually ending with "What should the project manager do FIRST/NEXT?", "What is the BEST ...?" or "Which TWO ...?". Vary names, industries and wording.
- Approach mix: about %d predictive, the rest agile or hybrid. Set "approach" accordingly.
- Single-answer questions: exactly 4 options, exactly 1 correct. Wrong options must be plausible (common mistakes or confusable terms from the notes).
- Exactly %d question(s) of type "multi" (choose TWO): 5 options, exactly 2 correct.
- Follow PMI thinking: analyse before acting, follow the plan and change control, collaborate, servant leadership in agile, escalate only when needed.
- Never use "all of the above", "none of the above", or refer to options by letter.
- For every option write "why": one sentence on why it is correct or wrong. "explanation": 2-4 sentences naming the principle or formula from the notes. "section": the exact title of the note the question is based on. "terms": 2-3 key English terms with Vietnamese meaning.
- %s Use ONLY formulas stated in the notes, with simple numbers, and double-check the arithmetic.
%s
Study notes:
%s
Return JSON: {"questions":[{"domain":"%s","approach":"predictive|agile|hybrid","type":"single|multi","stem":"...","options":[{"text":"...","why":"..."}],"correct":[0-based indexes],"explanation":"...","section":"...","terms":[{"en":"...","vi":"..."}]}]}`,
		n, domain, predictive, multi, calc, avoidText, nb.String(), domain)

	var out struct {
		Questions []DailyQuestion `json:"questions"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.7, &out); err != nil {
		return nil, err
	}
	return out.Questions, nil
}

// normalizeDaily validates one AI question, cleans it up and shuffles the
// options (the AI tends to put the right answer in the same place).
func normalizeDaily(q DailyQuestion, domain string, rng *rand.Rand) (DailyQuestion, bool) {
	q.Stem = strings.TrimSpace(q.Stem)
	q.Explanation = strings.TrimSpace(q.Explanation)
	q.Section = strings.TrimSpace(q.Section)
	q.Domain = domain
	switch strings.ToLower(strings.TrimSpace(q.Approach)) {
	case "agile":
		q.Approach = "agile"
	case "hybrid":
		q.Approach = "hybrid"
	default:
		q.Approach = "predictive"
	}
	q.Type = strings.ToLower(strings.TrimSpace(q.Type))
	if q.Type != "multi" {
		q.Type = "single"
	}
	if len(q.Stem) < 30 || len(q.Stem) > dailyMaxStem || q.Explanation == "" {
		return q, false
	}
	wantOpts, wantCorrect := 4, 1
	if q.Type == "multi" {
		wantOpts, wantCorrect = 5, 2
	}
	if len(q.Options) != wantOpts || len(q.Correct) != wantCorrect {
		return q, false
	}
	for i := range q.Options {
		q.Options[i].Text = strings.TrimSpace(q.Options[i].Text)
		q.Options[i].Why = strings.TrimSpace(q.Options[i].Why)
		low := strings.ToLower(q.Options[i].Text)
		if q.Options[i].Text == "" || q.Options[i].Why == "" || len(q.Options[i].Text) > dailyMaxOption ||
			strings.Contains(low, "all of the above") || strings.Contains(low, "none of the above") {
			return q, false
		}
	}
	isCorrect := make(map[int]bool, wantCorrect)
	for _, idx := range q.Correct {
		if idx < 0 || idx >= wantOpts || isCorrect[idx] {
			return q, false
		}
		isCorrect[idx] = true
	}
	perm := rng.Perm(wantOpts) // new position i holds old option perm[i]
	opts := make([]DailyOption, wantOpts)
	var correct []int
	for i, old := range perm {
		opts[i] = q.Options[old]
		if isCorrect[old] {
			correct = append(correct, i)
		}
	}
	sort.Ints(correct)
	q.Options, q.Correct = opts, correct
	if len(q.Terms) > 3 {
		q.Terms = q.Terms[:3]
	}
	return q, true
}
