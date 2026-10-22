package content

import (
	"encoding/json"
	"testing"

	"mvp-lang-app/internal/models"
)

func TestNormalizeMCQ(t *testing.T) {
	ok := `{"choices":[" A ","B","C"],"answer":2,"explanation":"x","junk":1}`
	out, err := Normalize(models.KindMCQ, "Q?", json.RawMessage(ok))
	if err != nil {
		t.Fatal(err)
	}
	var p MCQPayload
	_ = json.Unmarshal(out, &p)
	if p.Choices[0] != "A" || p.Answer != 2 {
		t.Fatalf("unexpected %+v", p)
	}
	for _, bad := range []string{`{"choices":["A"],"answer":0}`, `{"choices":["A","B"],"answer":2}`, `{"choices":["A",""],"answer":0}`} {
		if _, err := Normalize(models.KindMCQ, "Q?", json.RawMessage(bad)); err == nil {
			t.Fatalf("expected error for %s", bad)
		}
	}
}

func TestNormalizePhrasalSplitsFront(t *testing.T) {
	out, err := Normalize(models.KindPhrasal, "look forward to", json.RawMessage(`{"senses":[{"meaning_vi":"mong đợi","level":9}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var p PhrasalPayload
	_ = json.Unmarshal(out, &p)
	if p.Verb != "look" || p.Particle != "forward to" || p.Senses[0].Level != 1 {
		t.Fatalf("unexpected %+v", p)
	}
}

func TestNormalizeRejectsEmpty(t *testing.T) {
	if _, err := Normalize(models.KindVocab, "word", json.RawMessage(`{"senses":[{"level":1}]}`)); err == nil {
		t.Fatal("vocab without meaning must fail")
	}
	if _, err := Normalize(models.KindConcept, "  ", json.RawMessage(`{"back":"x"}`)); err == nil {
		t.Fatal("empty front must fail")
	}
}

func TestCompactPicksSimplestSense(t *testing.T) {
	c := models.Card{ID: 1, Kind: models.KindVocab, Front: "w", Payload: models.JSON(`{"senses":[
		{"level":3,"meaning_vi":"hard"},{"level":1,"meaning_vi":"easy","example_en":"ex"}]}`)}
	got := ToCompact(c)
	if got.MeaningVi != "easy" || got.ExampleEn != "ex" {
		t.Fatalf("unexpected %+v", got)
	}
}
