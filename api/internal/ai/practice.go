package ai

import (
	"context"
	"fmt"
	"strings"
)

// PracticeWord is a word the learner forgot and wants example sentences for.
type PracticeWord struct {
	Word      string `json:"word"`
	MeaningVi string `json:"meaning_vi"`
	POS       string `json:"pos,omitempty"`
}

// PracticeSentence is one example sentence with its Vietnamese translation.
type PracticeSentence struct {
	En string `json:"en"`
	Vi string `json:"vi"`
}

const practiceSystem = "You are an experienced TOEIC and IELTS teacher for Vietnamese learners. Reply with a single JSON object only."

func practicePrompt(words []PracticeWord, perWord int) string {
	var list strings.Builder
	for i, w := range words {
		fmt.Fprintf(&list, "%d. %s", i+1, w.Word)
		if w.POS != "" {
			fmt.Fprintf(&list, " (%s)", w.POS)
		}
		fmt.Fprintf(&list, " = %s\n", w.MeaningVi)
	}
	return fmt.Sprintf(`A Vietnamese learner forgot these English words. For EACH word write %d example sentences
that make the meaning easy to understand from context.

Words (with the meaning to use):
%s
Rules:
- Use the word with the given meaning; any grammatical form is fine (negotiate -> negotiated).
- 8-20 words per sentence, CEFR B1-B2, natural modern English.
- The %d sentences of a word are in DIFFERENT situations: work/office, daily life, news/study...
- Each sentence has a natural Vietnamese translation in "vi" (how a Vietnamese person would say it, not word by word).
Return JSON:
{"items":[{"word":"<the word exactly as given>","sentences":[{"en":"...","vi":"..."}]}]}`, perWord, list.String(), perWord)
}

// GeneratePracticeSentences writes perWord sentences for every word in ONE
// request. The result is keyed by the lower-cased word; sentences that do not
// use the word or look wrong are dropped, so a word may get fewer.
func (c *Client) GeneratePracticeSentences(ctx context.Context, words []PracticeWord, perWord int) (map[string][]PracticeSentence, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var out struct {
		Items []struct {
			Word      string             `json:"word"`
			Sentences []PracticeSentence `json:"sentences"`
		} `json:"items"`
	}
	if err := c.chatJSON(ctx, practiceSystem, practicePrompt(words, perWord), 0.8, &out); err != nil {
		return nil, err
	}
	res := map[string][]PracticeSentence{}
	for i, it := range out.Items {
		key := strings.ToLower(strings.TrimSpace(it.Word))
		target := ""
		for _, w := range words {
			if strings.ToLower(w.Word) == key {
				target = w.Word
			}
		}
		if target == "" && i < len(words) { // model changed the spelling: fall back to the order
			target = words[i].Word
		}
		if target == "" {
			continue
		}
		res[strings.ToLower(target)] = append(res[strings.ToLower(target)], CleanSentences(target, it.Sentences)...)
	}
	return res, nil
}

// CleanSentences keeps sentences that use the word, have a translation and a
// sensible length, without duplicates.
func CleanSentences(word string, in []PracticeSentence) []PracticeSentence {
	var out []PracticeSentence
	seen := map[string]bool{}
	for _, s := range in {
		s.En, s.Vi = strings.TrimSpace(s.En), strings.TrimSpace(s.Vi)
		n := len(strings.Fields(s.En))
		key := strings.ToLower(s.En)
		if s.Vi == "" || n < 4 || n > 40 || seen[key] || !ContainsWord(s.En, word) {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	return out
}

// TranslationItem is one learner translation to check.
type TranslationItem struct {
	Word        string // the word being practised
	MeaningVi   string
	SentenceEn  string
	ReferenceVi string
	LearnerVi   string
}

// TranslationResult is the verdict on one translation.
type TranslationResult struct {
	Correct    bool   `json:"correct"`
	Score      int    `json:"score"`       // 0-100
	FeedbackVi string `json:"feedback_vi"` // short, Vietnamese
}

// GradeTranslations checks all translations in ONE request. Results are in the
// same order as items; an item the model skipped gets an empty result with
// Score -1.
func (c *Client) GradeTranslations(ctx context.Context, items []TranslationItem) ([]TranslationResult, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var b strings.Builder
	for i, it := range items {
		fmt.Fprintf(&b, "%d. Word: %s (= %s)\n   English: %s\n   Reference: %s\n   Learner: %s\n\n",
			i+1, it.Word, it.MeaningVi, it.SentenceEn, it.ReferenceVi, it.LearnerVi)
	}
	prompt := fmt.Sprintf(`A Vietnamese learner translated English sentences into Vietnamese to practise a word.
Judge each translation:
- "correct": true when the meaning of the whole sentence is right AND the practised word is understood correctly.
  Different wording from the reference is fine; small omissions of minor words are fine.
- "score": 0-100 for how accurate the meaning is.
- "feedback_vi": one short sentence in Vietnamese. If wrong, say what was misunderstood (especially the practised word).
  If correct, a short encouragement or a more natural way to say it.

%s
Return JSON: {"results":[{"i":1,"correct":true,"score":90,"feedback_vi":"..."}]} with one entry per item.`, b.String())

	var out struct {
		Results []struct {
			I int `json:"i"`
			TranslationResult
		} `json:"results"`
	}
	if err := c.chatJSON(ctx, "You are a kind but precise English-Vietnamese teacher. Reply with a single JSON object only.", prompt, 0.2, &out); err != nil {
		return nil, err
	}
	res := make([]TranslationResult, len(items))
	for i := range res {
		res[i].Score = -1
	}
	for k, r := range out.Results {
		i := r.I - 1
		if i < 0 || i >= len(items) {
			i = k // no/invalid index: use the order
		}
		if i < 0 || i >= len(items) {
			continue
		}
		r.Score = max(0, min(100, r.Score))
		res[i] = r.TranslationResult
	}
	return res, nil
}
