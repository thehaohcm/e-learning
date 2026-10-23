package study

import (
	"encoding/json"
	"strings"
	"testing"

	"mvp-lang-app/internal/models"
)

func TestInterleave(t *testing.T) {
	mk := func(mode string, n int) []Question {
		q := make([]Question, n)
		for i := range q {
			q[i].Mode = mode
		}
		return q
	}
	got := interleave(mk("m", 7), mk("s", 3))
	var b strings.Builder
	for _, q := range got {
		b.WriteString(q.Mode)
	}
	if b.String() != "mmsmmsmmsm" {
		t.Fatalf("got %s", b.String())
	}
	if len(interleave(mk("m", 0), mk("s", 2))) != 2 || len(interleave(mk("m", 3), nil)) != 3 {
		t.Fatal("edge cases")
	}
}

func TestStressQuestion(t *testing.T) {
	c := models.Card{ID: 7, Kind: models.KindVocab, Front: "hypothesis"}
	info := StressInfo{Syllables: []string{"hy", "poth", "e", "sis"}, Stress: 1}
	q := stressQuestion(c, info, true)
	if q == nil || q.Mode != ModeStress || len(q.Syllables) != 4 || q.Choices == nil {
		t.Fatalf("%+v", q)
	}
	if stressQuestion(models.Card{Front: "give up"}, StressInfo{Syllables: []string{"give", "up"}, Stress: 1}, true) != nil {
		t.Fatal("phrases are not stress-tested")
	}
	if stressQuestion(models.Card{Front: "cat"}, StressInfo{Syllables: []string{"cat"}}, true) != nil {
		t.Fatal("one syllable")
	}
	if stressQuestion(c, StressInfo{Syllables: []string{"a", "b"}, Stress: 5}, true) != nil {
		t.Fatal("bad index")
	}
	if stressQuestion(c, info, false) != nil {
		t.Fatal("no pronunciation stored")
	}
}

func TestSoundQuestion(t *testing.T) {
	var si SoundInfo
	if err := json.Unmarshal([]byte(`{"syllable":0,"letters":"u","answer":1,"same":["luck"],"note_vi":"x",
		"options":[{"ipa":"u","like_vi":"u","example":"food"},{"ipa":"ʌ","like_vi":"â","example":"cut"}]}`), &si); err != nil {
		t.Fatal(err)
	}
	info := StressInfo{Syllables: []string{"fluc", "tu", "ate"}, Stress: 0, Sound: &si}
	c := models.Card{ID: 9, Kind: models.KindVocab, Front: "fluctuate"}
	q := soundQuestion(c, info, true)
	if q == nil || q.Mode != ModeSound || q.Focus.Letters != "u" || len(q.Choices) != 2 ||
		q.Choices[1].Text != "/ʌ/ as in “cut”" || q.Choices[1].Hint != "gần “â”" || q.Prompt != "How do you say the “u” in FLUC?" {
		t.Fatalf("%+v", q)
	}
	info.Sound = nil
	if soundQuestion(c, info, true) != nil {
		t.Fatal("no sound data")
	}
	if pronQuestion(c, info, true).Mode != ModeStress {
		t.Fatal("falls back to stress")
	}
}
