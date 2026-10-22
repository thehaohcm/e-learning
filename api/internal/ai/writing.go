package ai

import (
	"context"
	"fmt"
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
	Rewrites  []WritingRewrite `json:"rewrites"`
}

// EvaluateWriting scores a learner's text (0-100 per criterion), lists errors
// with Vietnamese explanations and gives rewrites at 3 levels.
func (c *Client) EvaluateWriting(ctx context.Context, prompt, text string) (*WritingFeedback, error) {
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
List at most 8 of the most important errors. Keep the learner's meaning.`, prompt, text)

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
	return &fb, nil
}
