package study

import (
	"testing"

	"mvp-lang-app/internal/models"
)

func TestSamePOSFirst(t *testing.T) {
	cards := []models.Card{
		{ID: 1, Kind: models.KindVocab, Payload: models.JSON(`{"pos":"noun"}`)},
		{ID: 2, Kind: models.KindVocab, Payload: models.JSON(`{"pos":"verb"}`)},
		{ID: 3, Kind: models.KindVocab, Payload: models.JSON(`{"pos":"Verb "}`)},
		{ID: 4, Kind: models.KindVocab, Payload: models.JSON(`{}`)},
	}
	got := samePOSFirst("verb", cards)
	if got[0].ID != 2 || got[1].ID != 3 || got[2].ID != 1 || got[3].ID != 4 {
		t.Fatalf("order %v %v %v %v", got[0].ID, got[1].ID, got[2].ID, got[3].ID)
	}
	if len(samePOSFirst("", cards)) != 4 {
		t.Fatal("no pos: unchanged")
	}
}
