package ai

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// AWS "Daily 20": original SAA-C03 (Solutions Architect – Associate) style
// multiple-choice questions written from the learner's own AWS cheatsheet notes.
// Reuses the DailyQuestion shape; Domain is one of the four SAA-C03 domains
// (Secure | Resilient | Performance | Cost) and Approach is the service area.

// GenerateAwsDaily writes up to count valid SAA-C03 questions from the notes.
// avoid lists recent archetypes not to repeat. rng shuffles the options.
func (c *Client) GenerateAwsDaily(ctx context.Context, model string, count int, notes []CheatNote, avoid []string, focus string, rng *rand.Rand) ([]DailyQuestion, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var got []DailyQuestion
	var lastErr error
	for attempt := 0; attempt < 2 && len(got) < count; attempt++ {
		need := count - len(got)
		skip := append([]string{}, avoid...)
		for _, q := range got {
			if a := strings.TrimSpace(q.Archetype); a != "" {
				skip = append(skip, a)
			}
		}
		raw, err := c.awsDailyBatch(ctx, model, need, notes, skip, focus)
		if err != nil {
			lastErr = err
			continue
		}
		for _, q := range raw {
			if v, ok := normalizeAwsDaily(q, rng); ok {
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

func (c *Client) awsDailyBatch(ctx context.Context, model string, n int, notes []CheatNote, avoid []string, focus string) ([]DailyQuestion, error) {
	var nb strings.Builder
	for _, nt := range notes {
		if nb.Len() > dailyNotesCap {
			break
		}
		fmt.Fprintf(&nb, "## %s\n%s\n\n", nt.Title, truncate(nt.Body, dailyNoteCap))
	}
	multi := 0
	if n >= 6 {
		multi = 1
	}
	avoidText := ""
	if len(avoid) > 0 {
		avoidText = "\nThese concept archetypes were already tested in recent days. Do NOT reuse any of them (changing the company, region or wording of an old one counts as reuse):\n"
		for _, a := range avoid {
			avoidText += "- " + a + "\n"
		}
	}
	system := "You are an experienced AWS Certified Solutions Architect – Associate (SAA-C03) exam item writer. You write original, exam-style questions. Reply with a single JSON object only."
	user := fmt.Sprintf(`Write %d ORIGINAL practice questions in the style of the AWS Certified Solutions Architect – Associate (SAA-C03) exam.
Base every question ONLY on services and facts in the study notes below (Vietnamese with English/AWS terms; write the questions in English). Never copy or recall real exam questions, brain dumps or copyrighted prep questions.

Requirements:
- Scenario-based: each stem is a short real-world architecture situation (2-5 sentences) ending in "Which solution is MOST cost-effective/highly available/secure/performant...?", "What should the solutions architect recommend?" or "Which TWO...?". Vary the company, industry and wording.
- Spread the questions across the four SAA-C03 domains by weight: Secure ~30%%, Resilient (reliable/highly available) ~26%%, Performance ~24%%, Cost ~20%%. Set "domain" to EXACTLY one of: Secure | Resilient | Performance | Cost — the domain the question optimises for.
- Set "approach" to the main service area: compute | storage | networking | database | security | monitoring | serverless | migration.
- Single-answer: exactly 4 options, exactly 1 correct. Every option must be a plausible AWS design; the wrong ones satisfy SOME requirements but fail the one the question optimises for (e.g. it works but is not the cheapest, or not the most available, or over-privileged). Never write an obviously wrong or joke option.
- Exactly %d question(s) of type "multi" (choose TWO): 5 options, exactly 2 correct.
- "archetype": a short lowercase slug naming the underlying concept (e.g. "s3-lifecycle-cost", "multi-az-rds-failover", "sqs-decouple-scaling", "iam-role-least-privilege"). Two questions testing the same concept must share the slug.
- Never use "all of the above", "none of the above", or refer to options by letter.
- For every option write "why": one sentence on why it is right or wrong. "explanation": 2-4 sentences naming the service/principle from the notes and why the other options lose on the optimised requirement. "section": the exact title of the note the question is based on. "terms": 2-3 key AWS terms with Vietnamese meaning.
%s
Study notes:
%s
Return JSON: {"questions":[{"domain":"Secure|Resilient|Performance|Cost","approach":"compute|storage|networking|database|security|monitoring|serverless|migration","archetype":"short-concept-slug","type":"single|multi","stem":"...","options":[{"text":"...","why":"..."}],"correct":[0-based indexes],"explanation":"...","section":"...","terms":[{"en":"...","vi":"..."}]}]}`,
		n, multi, avoidText, nb.String())

	if f := strings.TrimSpace(focus); f != "" {
		user += "\n\nIMPORTANT — the learner asked to FOCUS on: " + f + ". Weight the questions toward that, but still base everything only on the notes."
	}

	var out struct {
		Questions []DailyQuestion `json:"questions"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.7, &out); err != nil {
		return nil, err
	}
	return out.Questions, nil
}

// awsDomain maps the model's domain string to one of the four SAA-C03 domains.
func awsDomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.Contains(s, "resil") || strings.Contains(s, "reliab") || strings.Contains(s, "availab"):
		return "Resilient"
	case strings.Contains(s, "perf"):
		return "Performance"
	case strings.Contains(s, "cost"):
		return "Cost"
	default:
		return "Secure"
	}
}

// normalizeAwsDaily validates one AWS question, keeps its SAA-C03 domain, cleans
// it up and shuffles the options.
func normalizeAwsDaily(q DailyQuestion, rng *rand.Rand) (DailyQuestion, bool) {
	q.Stem = strings.TrimSpace(q.Stem)
	q.Explanation = strings.TrimSpace(q.Explanation)
	q.Section = strings.TrimSpace(q.Section)
	q.Archetype = strings.ToLower(strings.TrimSpace(q.Archetype))
	q.Domain = awsDomain(q.Domain)
	q.Approach = strings.ToLower(strings.TrimSpace(q.Approach))
	if q.Approach == "" {
		q.Approach = "design"
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
	perm := rng.Perm(wantOpts)
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
