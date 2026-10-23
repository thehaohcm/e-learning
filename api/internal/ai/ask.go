package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// TutorContext is the exercise the learner asks about (all fields optional).
type TutorContext struct {
	Task      string // the Vietnamese sentence / instruction
	Word      string // the word being practised
	Learner   string // what the learner wrote
	Corrected string // the corrected version
	Feedback  string // the feedback already given (Vietnamese)
	Sample    string // a model answer
}

// TutorTurn is an earlier question and answer of the same thread.
type TutorTurn struct {
	Q string `json:"q"`
	A string `json:"a"`
}

// TutorAnswer: a short explanation in Vietnamese, plus the English words it is
// about (offered as "+ note" buttons).
type TutorAnswer struct {
	Answer string   `json:"answer"`
	Words  []string `json:"words"`
}

// AskTutor answers a follow-up question about an exercise in ONE request.
func (c *Client) AskTutor(ctx context.Context, tc TutorContext, history []TutorTurn, question string) (TutorAnswer, error) {
	if !c.Enabled() {
		return TutorAnswer{}, ErrNotConfigured
	}
	var b strings.Builder
	line := func(label, v string, max int) {
		if v = strings.TrimSpace(v); v != "" {
			fmt.Fprintf(&b, "%s: %s\n", label, truncate(v, max))
		}
	}
	line("Task", tc.Task, 600)
	line("Word practised", tc.Word, 80)
	line("Learner's answer", tc.Learner, 1800)
	line("Corrected version", tc.Corrected, 1200)
	line("Feedback already given", tc.Feedback, 1200)
	line("Sample answer", tc.Sample, 600)
	if len(history) > 0 {
		b.WriteString("\nEarlier in this conversation:\n")
		for _, t := range history {
			fmt.Fprintf(&b, "Q: %s\nA: %s\n", truncate(t.Q, 300), truncate(t.A, 900))
		}
	}
	prompt := fmt.Sprintf(`A Vietnamese learner (TOEIC/IELTS, around B1) is doing an English writing exercise:
%s
The learner asks: %s

Answer IN VIETNAMESE, short and clear (at most about 120 words), like a friendly teacher:
- explain the difference / the rule, with 1-3 short English examples (keep English examples in English);
- a small table-like list is fine ("analyze (động từ): phân tích"), no long paragraphs;
- if the question is not about English, say briefly that you can only help with this exercise.
"words": up to 4 English words or phrases your answer is about (dictionary form), for the learner's notes.
Return JSON: {"answer":"...","words":["..."]}`, b.String(), truncate(question, 300))

	var out TutorAnswer
	if err := c.chatJSON(ctx, practiceSystem, prompt, 0.3, &out); err != nil {
		return TutorAnswer{}, err
	}
	out.Answer = strings.TrimSpace(out.Answer)
	if out.Answer == "" {
		return TutorAnswer{}, fmt.Errorf("empty answer")
	}
	out.Words = cleanTutorWords(out.Words)
	return out, nil
}

var tutorWord = regexp.MustCompile(`^[A-Za-z][A-Za-z' -]{0,40}$`)

func cleanTutorWords(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, w := range in {
		w = strings.Join(strings.Fields(w), " ")
		if !tutorWord.MatchString(w) || seen[strings.ToLower(w)] || len(strings.Fields(w)) > 4 {
			continue
		}
		seen[strings.ToLower(w)] = true
		out = append(out, w)
		if len(out) == 4 {
			break
		}
	}
	return out
}
