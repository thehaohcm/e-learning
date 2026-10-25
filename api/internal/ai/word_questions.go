package ai

import (
	"context"
	"fmt"
	"strings"
)

// Word-question kinds (multiple choice, for words the learner keeps forgetting).
const (
	WQUsage       = "usage"       // which sentence uses the word correctly
	WQSynonym     = "synonym"     // closest meaning in a new sentence
	WQCollocation = "collocation" // which word goes with it ("___ a decision")
)

// WordQuestionInput: one word to write questions for.
type WordQuestionInput struct {
	Word      string
	POS       string
	MeaningVi string
	MeaningEn string
}

// WordQuestion: 4 choices, Answer = index of the right one.
type WordQuestion struct {
	Kind        string   `json:"kind"`
	Prompt      string   `json:"prompt"`
	Choices     []string `json:"choices"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"` // short, Vietnamese
}

// GenerateWordQuestions writes, in ONE request, up to 3 questions per word (one
// of each kind). Result i holds the valid questions for words[i].
func (c *Client) GenerateWordQuestions(ctx context.Context, words []WordQuestionInput) ([][]WordQuestion, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var b strings.Builder
	for i, w := range words {
		fmt.Fprintf(&b, "%d. %s", i+1, w.Word)
		if w.POS != "" {
			fmt.Fprintf(&b, " (%s)", w.POS)
		}
		fmt.Fprintf(&b, " = %s", w.MeaningVi)
		if w.MeaningEn != "" {
			fmt.Fprintf(&b, " / %s", w.MeaningEn)
		}
		b.WriteString("\n")
	}
	prompt := fmt.Sprintf(`A Vietnamese learner (TOEIC/IELTS) often recognises these words in a list but cannot really use them.
For EACH word write 3 multiple-choice questions (4 choices, exactly ONE correct), one of each kind:

- "usage": prompt "Which sentence uses \"WORD\" correctly?"; choices are 4 natural sentences of about 8-14 words,
  3 of them use the word wrongly (wrong meaning, wrong part of speech or wrong grammar) but look plausible.
- "synonym": prompt is a NEW sentence using the word followed by "— \"WORD\" is closest in meaning to:";
  choices are 4 single words or short phrases of the same part of speech.
- "collocation": a short phrase or sentence with the word and ONE blank "____" for a word that typically goes
  with it (verb + noun, adjective + preposition, phrasal verb particle...); choices are 4 single words.
  If the word has no strong collocation, write another "usage" question instead (kind "usage").

Keep it TOEIC/IELTS level, natural English, no trick questions. "explanation": one short sentence in Vietnamese
saying why the answer is right. "answer": index 0-3 of the right choice (vary the position).

Words:
%s
Return JSON: {"items":[{"i":1,"questions":[{"kind":"usage","prompt":"...","choices":["...","...","...","..."],"answer":2,"explanation":"..."}]}]}`, b.String())

	var out struct {
		Items []struct {
			I         int            `json:"i"`
			Questions []WordQuestion `json:"questions"`
		} `json:"items"`
	}
	if err := c.chatJSON(ctx, practiceSystem, prompt, 0.4, &out); err != nil {
		return nil, err
	}
	res := make([][]WordQuestion, len(words))
	for k, it := range out.Items {
		i := it.I - 1
		if i < 0 || i >= len(words) {
			i = k
		}
		if i < 0 || i >= len(words) {
			continue
		}
		for _, q := range it.Questions {
			if cq, ok := CleanWordQuestion(q); ok {
				res[i] = append(res[i], cq)
			}
		}
	}
	return res, nil
}

// CleanWordQuestion trims a generated question and checks it: known kind, a
// prompt, exactly 4 different non-empty choices, an answer index 0-3.
func CleanWordQuestion(q WordQuestion) (WordQuestion, bool) {
	q.Kind = strings.ToLower(strings.TrimSpace(q.Kind))
	switch q.Kind {
	case WQUsage, WQSynonym, WQCollocation:
	default:
		return q, false
	}
	q.Prompt = strings.TrimSpace(q.Prompt)
	q.Explanation = strings.TrimSpace(q.Explanation)
	if q.Prompt == "" || len(q.Choices) != 4 || q.Answer < 0 || q.Answer > 3 {
		return q, false
	}
	seen := map[string]bool{}
	for i, ch := range q.Choices {
		ch = strings.TrimSpace(ch)
		if ch == "" || seen[strings.ToLower(ch)] || len(ch) > 200 {
			return q, false
		}
		seen[strings.ToLower(ch)] = true
		q.Choices[i] = ch
	}
	return q, true
}
