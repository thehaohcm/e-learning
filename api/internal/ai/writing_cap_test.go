package ai

import "testing"

func TestMinWordsRe(t *testing.T) {
	m := minWordsRe.FindStringSubmatch("Write to Mr Davis (60-100 words; use these words: forefront)")
	if m == nil || m[1] != "60" {
		t.Fatalf("%v", m)
	}
	if minWordsRe.FindStringSubmatch("Describe your town.") != nil {
		t.Fatal("matched a prompt without a word range")
	}
}
