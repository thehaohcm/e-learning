package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// FormItem: a sentence whose blank is Answer, a form of Word.
type FormItem struct {
	Sentence string
	Answer   string // the form used in the sentence ("fluctuated")
	Word     string // dictionary form ("fluctuate")
}

// GenerateWordForms writes TOEIC Part 5 style distractors in ONE request: for
// each sentence, 3 other forms of the same word family that do not fit the
// blank. Result i is nil when the model gave nothing usable (fewer than 3).
func (c *Client) GenerateWordForms(ctx context.Context, items []FormItem) ([][]string, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var b strings.Builder
	for i, it := range items {
		fmt.Fprintf(&b, "%d. %s\n   Answer: %s (word: %s)\n", i+1, strings.Replace(it.Sentence, it.Answer, "_____", 1), it.Answer, it.Word)
	}
	prompt := fmt.Sprintf(`Write TOEIC Part 5 "word form" questions. Each sentence below has a blank whose answer is given.
For each, give exactly 3 WRONG options that are other real forms of the SAME word family
(noun / verb / adjective / adverb / -ing / -ed / plural...), so that only the answer fits the blank
grammatically. Example: answer "decisive" → ["decide", "decision", "decisively"].
Never repeat the answer, never use words from another family. If the family has fewer than
3 real other forms, return the ones that exist (or an empty list).

%s
Return JSON: {"items":[{"i":1,"options":["...","...","..."]}]}`, b.String())

	var out struct {
		Items []struct {
			I       int      `json:"i"`
			Options []string `json:"options"`
		} `json:"items"`
	}
	if err := c.chatJSON(ctx, practiceSystem, prompt, 0.2, &out); err != nil {
		return nil, err
	}
	res := make([][]string, len(items))
	for k, it := range out.Items {
		i := it.I - 1
		if i < 0 || i >= len(items) {
			i = k
		}
		if i < 0 || i >= len(items) {
			continue
		}
		res[i] = CleanFormOptions(it.Options, items[i].Answer, items[i].Word)
	}
	return res, nil
}

var oneWord = regexp.MustCompile(`^[A-Za-z][A-Za-z'-]*$`)

// CleanFormOptions keeps 3 distinct real-looking forms of the word family
// (same beginning as the word), different from the answer, written with the
// answer's capitalisation. Returns nil if fewer than 3 are left.
func CleanFormOptions(opts []string, answer, word string) []string {
	ans := strings.ToLower(strings.TrimSpace(answer))
	base := strings.ToLower(strings.TrimSpace(word))
	stem := min(4, len(base)-1, len(ans)-1)
	if stem < 2 {
		return nil
	}
	seen := map[string]bool{ans: true}
	var out []string
	for _, o := range opts {
		o = strings.TrimSpace(o)
		lo := strings.ToLower(o)
		if !oneWord.MatchString(o) || seen[lo] || len(lo) < 3 {
			continue
		}
		if commonPrefix(lo, base) < stem && commonPrefix(lo, ans) < stem {
			continue // another family
		}
		seen[lo] = true
		out = append(out, matchCase(lo, answer))
		if len(out) == 3 {
			return out
		}
	}
	return nil
}

func commonPrefix(a, b string) int {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return n
}

// matchCase: "Decision" at the start of a sentence → options capitalised too.
func matchCase(lower, answer string) string {
	if answer != "" && answer[:1] != strings.ToLower(answer[:1]) {
		return strings.ToUpper(lower[:1]) + lower[1:]
	}
	return lower
}
