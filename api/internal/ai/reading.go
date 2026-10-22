package ai

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// TargetWord is a vocabulary item the passage must use.
type TargetWord struct {
	Word      string `json:"word"`
	MeaningVi string `json:"meaning_vi"`
	IsNew     bool   `json:"is_new"` // not learned yet: introduce it with clear context
}

type DialogueLine struct {
	Speaker string `json:"speaker"`
	Text    string `json:"text"`
	Vi      string `json:"vi,omitempty"`
}

type UsefulPhrase struct {
	Phrase    string `json:"phrase"`
	MeaningVi string `json:"meaning_vi"`
}

type ReadingQuestion struct {
	Question    string   `json:"question"`
	Choices     []string `json:"choices"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"` // Vietnamese
	Evidence    string   `json:"evidence"`    // sentence from the text that proves the answer
	Word        string   `json:"word,omitempty"`
}

// ReadingContent is what the model returns (validated before saving).
type ReadingContent struct {
	Title         string            `json:"title"`
	Situation     string            `json:"situation,omitempty"` // conversation: where/who
	Passage       string            `json:"passage,omitempty"`   // reading: paragraphs separated by blank lines
	TranslationVi string            `json:"translation_vi,omitempty"`
	Lines         []DialogueLine    `json:"lines,omitempty"`
	Phrases       []UsefulPhrase    `json:"useful_phrases,omitempty"`
	Questions     []ReadingQuestion `json:"questions"`
}

var levelText = map[string]string{
	"easy":   "CEFR A2-B1. Short simple sentences, common words besides the target words. TOEIC Part 7 single-passage style.",
	"medium": "CEFR B1-B2. Natural sentences with some linking words. TOEIC Part 7 style (email, notice, article, advertisement).",
	"ielts":  "CEFR B2-C1. Academic style like IELTS Reading, with some complex sentences and paraphrasing in the questions.",
}

func readingPrompt(kind, level string, words []TargetWord) string {
	var list strings.Builder
	for _, w := range words {
		tag := "learned"
		if w.IsNew {
			tag = "NEW - make its meaning clear from context"
		}
		fmt.Fprintf(&list, "- %s (%s) [%s]\n", w.Word, w.MeaningVi, tag)
	}
	lvl := levelText[level]

	if kind == "conversation" {
		return fmt.Sprintf(`Write a natural everyday spoken English conversation for a Vietnamese learner.
Level: %s
Use EVERY one of these target words/phrases at least once, naturally (any grammatical form is fine):
%s
Rules:
- 2 speakers with first names, 10-16 lines, a realistic daily situation (work, coffee shop, phone call, shopping, travel...).
- Sound like real people: contractions, short replies, common spoken expressions. No textbook style.
- Every line has a natural Vietnamese translation in "vi".
- 3 comprehension questions (detail, speaker's intention, meaning of a target word in context), 4 choices each, exactly one correct.
- "explanation" is in Vietnamese. "evidence" quotes the exact line that proves the answer.
- For a question about a target word, set "word" to that target word.
- 3-5 "useful_phrases": everyday spoken expressions used in the dialogue (not the target words) with Vietnamese meaning.
Return JSON:
{"title":"...","situation":"one sentence in English","lines":[{"speaker":"Anna","text":"...","vi":"..."}],
 "useful_phrases":[{"phrase":"...","meaning_vi":"..."}],
 "questions":[{"question":"...","choices":["A","B","C","D"],"answer":0,"explanation":"...","evidence":"...","word":""}]}`, lvl, list.String())
	}

	return fmt.Sprintf(`Write an English reading passage for a Vietnamese learner preparing for TOEIC/IELTS.
Level: %s
Use EVERY one of these target words/phrases at least once, naturally (any grammatical form is fine):
%s
Rules:
- 150-250 words, 2-4 paragraphs separated by a blank line, with a short title.
- A realistic text type: email, company notice, news article, product review, or short report.
- "translation_vi": a natural Vietnamese translation of the whole passage.
- 4 questions like TOEIC Part 7 / IELTS: main idea, a detail, the meaning of a target word in context, and an inference. 4 choices each, exactly one correct, wrong choices plausible.
- "explanation" is in Vietnamese. "evidence" quotes the exact sentence from the passage that proves the answer.
- For the question about a target word, set "word" to that target word.
Return JSON:
{"title":"...","passage":"...","translation_vi":"...",
 "questions":[{"question":"...","choices":["A","B","C","D"],"answer":0,"explanation":"...","evidence":"...","word":""}]}`, lvl, list.String())
}

// GenerateReading creates a passage or conversation and validates it. If the
// first answer does not pass validation, it asks once more.
func (c *Client) GenerateReading(ctx context.Context, kind, level string, words []TargetWord) (*ReadingContent, error) {
	system := "You are an experienced TOEIC and IELTS teacher writing original practice material. Reply with a single JSON object only."
	prompt := readingPrompt(kind, level, words)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		var rc ReadingContent
		if err := c.chatJSON(ctx, system, prompt, 0.7, &rc); err != nil {
			return nil, err
		}
		if err := ValidateReading(kind, &rc, words); err != nil {
			lastErr = err
			continue
		}
		return &rc, nil
	}
	return nil, fmt.Errorf("generated reading did not pass checks: %w", lastErr)
}

// ValidateReading checks structure and that most target words really appear.
func ValidateReading(kind string, rc *ReadingContent, words []TargetWord) error {
	rc.Title = strings.TrimSpace(rc.Title)
	if rc.Title == "" {
		rc.Title = "Untitled"
	}
	var text string
	if kind == "conversation" {
		var lines []DialogueLine
		for _, l := range rc.Lines {
			l.Speaker, l.Text = strings.TrimSpace(l.Speaker), strings.TrimSpace(l.Text)
			if l.Speaker != "" && l.Text != "" {
				lines = append(lines, l)
			}
		}
		if len(lines) < 6 {
			return errors.New("conversation is too short")
		}
		rc.Lines, rc.Passage = lines, ""
		for _, l := range lines {
			text += l.Text + "\n"
		}
	} else {
		rc.Passage = strings.TrimSpace(rc.Passage)
		if n := len(strings.Fields(rc.Passage)); n < 80 {
			return fmt.Errorf("passage too short (%d words)", n)
		}
		rc.Lines = nil
		text = rc.Passage
	}

	var qs []ReadingQuestion
	for _, q := range rc.Questions {
		if strings.TrimSpace(q.Question) == "" || len(q.Choices) < 3 || q.Answer < 0 || q.Answer >= len(q.Choices) {
			continue
		}
		qs = append(qs, q)
	}
	if len(qs) < 2 {
		return errors.New("not enough valid questions")
	}
	rc.Questions = qs

	found := 0
	for _, w := range words {
		if ContainsWord(text, w.Word) {
			found++
		}
	}
	if len(words) > 0 && found*10 < len(words)*7 { // at least 70% of target words used
		return fmt.Errorf("only %d of %d target words used", found, len(words))
	}
	return nil
}

// ContainsWord reports whether text uses word in some form: "negotiate" matches
// "negotiated", "carry out" matches "carried out", "give up" matches "gave ... up" loosely.
func ContainsWord(text, word string) bool {
	re := WordPattern(word)
	return re != nil && re.MatchString(text)
}

// WordPattern builds a case-insensitive regexp that matches common inflections
// of each word in a (possibly multi-word) item.
func WordPattern(word string) *regexp.Regexp {
	parts := strings.Fields(strings.ToLower(word))
	if len(parts) == 0 {
		return nil
	}
	for i, p := range parts {
		p = regexp.QuoteMeta(p)
		if i == 0 && len(p) > 3 && (strings.HasSuffix(p, "e") || strings.HasSuffix(p, "y")) {
			p = p[:len(p)-1] // negotiate -> negotiat(ed), carry -> carr(ied)
		}
		parts[i] = p + `\w*`
	}
	// allow up to 3 words between the parts of a phrasal verb ("turn it down")
	re, err := regexp.Compile(`(?i)\b` + strings.Join(parts, `(?:\W+\w+){0,3}?\W+`) + `\b`)
	if err != nil {
		return nil
	}
	return re
}
