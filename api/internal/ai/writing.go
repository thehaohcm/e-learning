package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type WritingError struct {
	Original    string `json:"original"`
	Correction  string `json:"correction"`
	Explanation string `json:"explanation_vi"`
}

type WritingRewrite struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

type WritingFeedback struct {
	Overall   float64          `json:"overall"`
	Grammar   float64          `json:"grammar"`
	Vocab     float64          `json:"vocabulary"`
	Coherence float64          `json:"coherence"`
	Summary   string           `json:"summary_vi"`
	Errors    []WritingError   `json:"errors"`
	// PhraseTips: sentences of the learner that a frame of the phrase bank would improve.
	PhraseTips []PhraseTip `json:"phrase_tips,omitempty"`
	Rewrites  []WritingRewrite `json:"rewrites"`
}

var minWordsRe = regexp.MustCompile(`\((\d+)\s*[-–]\s*\d+ words`)

// PhraseTip: "you wrote X: with the frame P it becomes R".
type PhraseTip struct {
	Original string `json:"original"`
	Pattern  string `json:"pattern"`
	Rewrite  string `json:"rewrite"`
}

// EvaluateWriting scores a learner's text (0-100 per criterion), lists errors
// with Vietnamese explanations and gives rewrites at 3 levels.
//
// frames is the phrase bank ("group: pattern" lines): up to 3 sentences that one
// of them would improve are returned in phrase_tips.
func (c *Client) EvaluateWriting(ctx context.Context, prompt, text string, frames []string) (*WritingFeedback, error) {
	bank := ""
	if len(frames) > 0 {
		bank = `

Phrase bank the learner is studying (group: frame):
` + strings.Join(frames, "\n") + `
"phrase_tips": up to 3 of the learner's sentences that one of these frames would make
clearer or more natural: [{"original": "the learner's words", "pattern": "the frame exactly as listed (without the group)", "rewrite": "the sentence using it"}]. Empty list if none fits.`
	}
	user := fmt.Sprintf(`Evaluate this English writing by a Vietnamese learner.

Task / prompt: %s

Text:
"""
%s
"""

Return JSON:
{
  "overall": 0-100, "grammar": 0-100, "vocabulary": 0-100, "coherence": 0-100,
  "summary_vi": "2-3 sentences of feedback in Vietnamese",
  "errors": [{"original": "...", "correction": "...", "explanation_vi": "..."}],
  "rewrites": [
    {"level": 1, "text": "simple, correct version"},
    {"level": 2, "text": "natural everyday version"},
    {"level": 3, "text": "advanced IELTS 7+ version"}
  ]
}
Scoring (every score 0-100, the same scale each time):
- 80-100 good: does the whole task, almost no errors, natural and varied wording.
- 60-79 fair: does the task, some errors that do not block the meaning.
- 40-59 needs work: part of the task missing, or frequent errors / unnatural wording.
- 0-39 redo: off-task, very short, or errors that block the meaning.
"overall" is mainly task achievement. If the task asks for a number of words or for
words to use and the text has fewer words, or does not use a required word (any form),
"overall" is at most 60 and summary_vi says what is missing.
List at most 8 of the most important errors. Keep the learner's meaning.%s`, prompt, text, bank)

	var fb WritingFeedback
	if err := c.chatJSON(ctx, "You are a strict but encouraging IELTS writing examiner. Reply with a single JSON object only.", user, 0.3, &fb); err != nil {
		return nil, err
	}
	clamp := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 100 {
			return 100
		}
		return v
	}
	fb.Overall, fb.Grammar, fb.Vocab, fb.Coherence = clamp(fb.Overall), clamp(fb.Grammar), clamp(fb.Vocab), clamp(fb.Coherence)
	tips := fb.PhraseTips[:0]
	for _, t := range fb.PhraseTips {
		t.Original, t.Pattern, t.Rewrite = strings.TrimSpace(t.Original), strings.TrimSpace(t.Pattern), strings.TrimSpace(t.Rewrite)
		if t.Pattern != "" && t.Rewrite != "" && len(tips) < 3 {
			tips = append(tips, t)
		}
	}
	fb.PhraseTips = tips
	// too short for the task ("(60-100 words; ..."): the cap does not depend on the AI
	if m := minWordsRe.FindStringSubmatch(prompt); m != nil {
		if min, err := strconv.Atoi(m[1]); err == nil && len(strings.Fields(text)) < min && fb.Overall > 60 {
			fb.Overall = 60
		}
	}
	return &fb, nil
}
