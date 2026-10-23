package handlers

import "testing"

func TestCleanNoteText(t *testing.T) {
	cases := map[string]string{
		"  “negotiate,”  ": "negotiate",
		"carry   out.":     "carry out",
		"well-known":       "well-known",
		"don't":            "don't",
		"...":              "",
		"":                 "",
		"thương lượng":     "",
		"café":             "",
		"COVID-19":         "COVID-19",
		"e.g.":             "e.g",
		"a very long phrase that is clearly not a single word or a short phrase at all": "",
	}
	for in, want := range cases {
		if got := cleanNoteText(in); got != want {
			t.Errorf("cleanNoteText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPosKey(t *testing.T) {
	for in, want := range map[string]string{"noun": "noun", "Noun (countable)": "noun", "v.": "verb", "verb": "verb",
		"adj": "adjective", "adjective": "adjective", "adv.": "adverb", "phrasal verb": "phrasal verb", "": ""} {
		if got := posKey(in, "vocab"); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
	if posKey("noun", "phrasal") != "phrasal verb" {
		t.Error("phrasal kind")
	}
}
