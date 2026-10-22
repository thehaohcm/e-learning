package handlers

import (
	"testing"

	"mvp-lang-app/internal/models"
)

func TestPickRound(t *testing.T) {
	ex := func(id int64, src string) models.CardExample { return models.CardExample{ID: id, Source: src} }
	ids := func(xs []models.CardExample) []int64 {
		out := []int64{}
		for _, x := range xs {
			out = append(out, x.ID)
		}
		return out
	}
	six := []models.CardExample{ex(1, "ai"), ex(2, "ai"), ex(3, "ai"), ex(4, "ai"), ex(5, "ai"), ex(6, "ai"), ex(9, "card")}
	cases := []struct {
		pool  []models.CardExample
		round int
		want  []int64
	}{
		{six, 0, []int64{1, 2, 3}},
		{six, 1, []int64{4, 5, 6}},
		{six, 2, []int64{1, 2, 3}},
		{six[:3], 1, []int64{1, 2, 3}},                       // not enough for a second set yet
		{six[:4], 1, []int64{2, 3, 4}},                       // partly new
		{[]models.CardExample{ex(9, "card")}, 0, []int64{9}}, // AI unavailable: own example
	}
	for _, c := range cases {
		got := ids(pickRound(c.pool, c.round))
		if len(got) != len(c.want) {
			t.Fatalf("round %d: got %v want %v", c.round, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("round %d: got %v want %v", c.round, got, c.want)
			}
		}
	}
}
