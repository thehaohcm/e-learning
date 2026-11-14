package ai

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// English "Daily 20": original TOEIC/IELTS-style multiple-choice questions
// written from the learner's own vocabulary cards (word + meaning), so the set
// reinforces the words they are studying. Reuses the DailyQuestion shape used
// by the PMP daily set; Domain is "English" and Approach is the question style
// (vocabulary | grammar | reading).

// EnglishWord is one vocabulary item used to ground the questions.
type EnglishWord struct {
	Word      string
	MeaningEN string
	MeaningVI string
	Topic     string
}

// GenerateEnglishDaily writes up to count valid questions from the words.
func (c *Client) GenerateEnglishDaily(ctx context.Context, model string, count int, words []EnglishWord, rng *rand.Rand) ([]DailyQuestion, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var got []DailyQuestion
	var lastErr error
	for attempt := 0; attempt < 2 && len(got) < count; attempt++ {
		need := count - len(got)
		used := map[string]bool{}
		for _, q := range got {
			used[strings.ToLower(q.Section)] = true
		}
		raw, err := c.englishDailyBatch(ctx, model, need, words, used)
		if err != nil {
			lastErr = err
			continue
		}
		for _, q := range raw {
			if v, ok := normalizeEnglishDaily(q, rng); ok {
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

func (c *Client) englishDailyBatch(ctx context.Context, model string, n int, words []EnglishWord, used map[string]bool) ([]DailyQuestion, error) {
	var wb strings.Builder
	listed := 0
	for _, w := range words {
		if used[strings.ToLower(w.Word)] {
			continue
		}
		meaning := w.MeaningEN
		if w.MeaningVI != "" {
			meaning = strings.TrimSpace(meaning + " / " + w.MeaningVI)
		}
		fmt.Fprintf(&wb, "- %s: %s\n", w.Word, truncate(meaning, 180))
		listed++
		if listed >= n+10 { // give the model a few spare words to pick from
			break
		}
	}
	if listed == 0 { // all recently used: fall back to the whole list
		for _, w := range words {
			meaning := w.MeaningEN
			if w.MeaningVI != "" {
				meaning = strings.TrimSpace(meaning + " / " + w.MeaningVI)
			}
			fmt.Fprintf(&wb, "- %s: %s\n", w.Word, truncate(meaning, 180))
		}
	}

	system := "You are an experienced TOEIC and IELTS item writer. You write original, exam-style English vocabulary questions. Reply with a single JSON object only."
	user := fmt.Sprintf(`Write %d ORIGINAL multiple-choice English practice questions at TOEIC (~900) / academic IELTS (~6-7) level.
Each question tests ONE target word from the list below, in ONE of these styles (vary them):
- fill-in-the-blank: a natural sentence with the target word removed, choose the word that best completes it;
- closest meaning / synonym of the target word as used in a short sentence;
- correct usage: which sentence uses the word correctly.

Rules:
- Exactly 4 options, exactly 1 correct. Wrong options must be plausible (real English words, often confusable with the target).
- Never use "all of the above", "none of the above", or refer to options by letter.
- Write everything in English except the Vietnamese gloss. For each option write "why": one short sentence why it is right or wrong.
- "explanation": 2-3 sentences explaining the target word's meaning and why the answer fits; include its Vietnamese meaning.
- "section": the exact target word. "terms": [{"en": target word, "vi": its Vietnamese meaning}].
- Set "domain" to "English", "approach" to vocabulary|grammar|reading, "type" to "single".
- Use each target word at most once.

Target words (word: meaning):
%s
Return JSON: {"questions":[{"domain":"English","approach":"vocabulary|grammar|reading","type":"single","stem":"...","options":[{"text":"...","why":"..."}],"correct":[index],"explanation":"...","section":"word","terms":[{"en":"...","vi":"..."}]}]}`,
		n, wb.String())

	var out struct {
		Questions []DailyQuestion `json:"questions"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.7, &out); err != nil {
		return nil, err
	}
	return out.Questions, nil
}

// normalizeEnglishDaily validates one English question and shuffles the options.
func normalizeEnglishDaily(q DailyQuestion, rng *rand.Rand) (DailyQuestion, bool) {
	q.Stem = strings.TrimSpace(q.Stem)
	q.Explanation = strings.TrimSpace(q.Explanation)
	q.Section = strings.TrimSpace(q.Section)
	q.Domain = "English"
	switch strings.ToLower(strings.TrimSpace(q.Approach)) {
	case "grammar":
		q.Approach = "grammar"
	case "reading":
		q.Approach = "reading"
	default:
		q.Approach = "vocabulary"
	}
	q.Type = "single"
	if len(q.Stem) < 8 || len(q.Stem) > dailyMaxStem || q.Explanation == "" {
		return q, false
	}
	if len(q.Options) != 4 || len(q.Correct) != 1 {
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
	correctIdx := q.Correct[0]
	if correctIdx < 0 || correctIdx >= 4 {
		return q, false
	}
	perm := rng.Perm(4)
	opts := make([]DailyOption, 4)
	var correct []int
	for i, old := range perm {
		opts[i] = q.Options[old]
		if old == correctIdx {
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
