package ai

import "testing"

func TestContainsWord(t *testing.T) {
	cases := []struct {
		text, word string
		want       bool
	}{
		{"They negotiated a better deal.", "negotiate", true},
		{"The team carried out a survey.", "carry out", true},
		{"He turned the offer down.", "turn down", true},
		{"Prices fluctuate daily.", "fluctuate", true},
		{"Nothing here.", "substantial", false},
		{"She gave up smoking.", "give up", false}, // irregular past: not detected (acceptable)
	}
	for _, c := range cases {
		if got := ContainsWord(c.text, c.word); got != c.want {
			t.Errorf("ContainsWord(%q, %q) = %v, want %v", c.text, c.word, got, c.want)
		}
	}
}

func TestValidateReadingRejectsMissingWords(t *testing.T) {
	passage := "This is a simple passage about work and life. "
	for i := 0; i < 20; i++ {
		passage += "We meet every week to talk about plans. "
	}
	rc := &ReadingContent{Passage: passage, Questions: []ReadingQuestion{
		{Question: "Q1", Choices: []string{"a", "b", "c", "d"}, Answer: 0},
		{Question: "Q2", Choices: []string{"a", "b", "c", "d"}, Answer: 9}, // invalid answer: dropped
		{Question: "Q3", Choices: []string{"a", "b", "c", "d"}, Answer: 1},
	}}
	words := []TargetWord{{Word: "negotiate"}, {Word: "substantial"}}
	if err := ValidateReading("reading", rc, words); err == nil {
		t.Fatal("expected error: target words missing")
	}
	rc.Passage += " They negotiated a substantial raise."
	if err := ValidateReading("reading", rc, words); err != nil {
		t.Fatal(err)
	}
	if len(rc.Questions) != 2 {
		t.Fatalf("invalid question should be dropped, got %d", len(rc.Questions))
	}
}
