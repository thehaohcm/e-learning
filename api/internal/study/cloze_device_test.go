package study

import "testing"

func TestBlankWord(t *testing.T) {
	cases := []struct {
		s, w, want string
		ok         bool
	}{
		{"Vegetable prices fluctuated wildly.", "fluctuate", "Vegetable prices ____ wildly.", true},
		{"They negotiated a better deal.", "negotiate", "They ____ a better deal.", true},
		{"She gave it up last year.", "give up", "She gave it up last year.", false}, // irregular past: not matched
		{"Don't give it up now.", "give up", "Don't ____ now.", true},
		{"The Hypothesis needs testing.", "hypothesis", "The ____ needs testing.", true},
		{"Nothing here.", "legitimate", "Nothing here.", false},
	}
	for _, c := range cases {
		got, ok := BlankWord(c.s, c.w)
		if got != c.want || ok != c.ok {
			t.Errorf("%q/%q: got %q %v", c.s, c.w, got, ok)
		}
	}
}

func TestDeviceCardModeCloze(t *testing.T) {
	c := DeviceCard{Kind: "vocab", HasOptions: true, HasCloze: true}
	if m := DeviceCardMode(DeviceModeCloze, c, 0); m != DeviceModeCloze {
		t.Fatal(m)
	}
	c.HasCloze = false
	if m := DeviceCardMode(DeviceModeCloze, c, 0); m != DeviceModeQuiz {
		t.Fatal(m)
	}
	c.IsNew, c.HasCloze = true, true
	if m := DeviceCardMode(DeviceModeCloze, c, 0); m != DeviceModeFlashcard {
		t.Fatal("new cards stay flashcards:", m)
	}
}
