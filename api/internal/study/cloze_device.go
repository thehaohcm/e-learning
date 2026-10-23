package study

import (
	"regexp"
	"strings"
)

// Blank is shown in place of the word in device cloze sentences.
const Blank = "____"

// BlankWord replaces the first occurrence of word (or a common form of it:
// negotiate → negotiated, carry out → carried it out) in sentence with Blank.
// ok=false if the sentence does not contain it.
func BlankWord(sentence, word string) (string, bool) {
	re := wordPattern(word)
	if re == nil {
		return sentence, false
	}
	loc := re.FindStringIndex(sentence)
	if loc == nil {
		return sentence, false
	}
	return sentence[:loc[0]] + Blank + sentence[loc[1]:], true
}

// wordPattern matches a word/phrase in common inflected forms. The first word
// drops a final e/y (negotiate → negotiat\w*), words of a phrase may have up to
// 3 words between them (give it up).
func wordPattern(word string) *regexp.Regexp {
	parts := strings.Fields(strings.ToLower(word))
	if len(parts) == 0 {
		return nil
	}
	for i, p := range parts {
		x := regexp.QuoteMeta(p)
		if i == 0 && len(p) > 3 && strings.ContainsAny(p[len(p)-1:], "ey") {
			x = regexp.QuoteMeta(p[:len(p)-1])
		}
		parts[i] = x + `\w*`
	}
	re, err := regexp.Compile(`(?i)\b` + strings.Join(parts, `(?:\W+\w+){0,3}?\W+`) + `\b`)
	if err != nil {
		return nil
	}
	return re
}

// FindWord returns the form of word used in sentence ("fluctuated" for
// "fluctuate"), or ok=false.
func FindWord(sentence, word string) (string, bool) {
	re := wordPattern(word)
	if re == nil {
		return "", false
	}
	m := re.FindString(sentence)
	return m, m != ""
}
