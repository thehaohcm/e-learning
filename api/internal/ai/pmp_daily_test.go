package ai

import (
	"math/rand"
	"strings"
	"testing"
)

func TestNormalizeDailyShuffle(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		q := DailyQuestion{Type: "single", Approach: "Agile", Stem: strings.Repeat("A realistic scenario. ", 3), Explanation: "x",
			Options: []DailyOption{{"right", "r"}, {"w1", "a"}, {"w2", "b"}, {"w3", "c"}}, Correct: []int{0}}
		v, ok := normalizeDaily(q, "People", rand.New(rand.NewSource(seed)))
		if !ok || len(v.Correct) != 1 || v.Options[v.Correct[0]].Text != "right" || v.Approach != "agile" {
			t.Fatalf("seed %d: %+v ok=%v", seed, v, ok)
		}
		m := DailyQuestion{Type: "multi", Stem: q.Stem, Explanation: "x",
			Options: []DailyOption{{"r1", "r"}, {"w1", "a"}, {"r2", "b"}, {"w3", "c"}, {"w4", "d"}}, Correct: []int{0, 2}}
		mv, ok := normalizeDaily(m, "Process", rand.New(rand.NewSource(seed)))
		if !ok || len(mv.Correct) != 2 {
			t.Fatalf("multi seed %d", seed)
		}
		for _, i := range mv.Correct {
			if !strings.HasPrefix(mv.Options[i].Text, "r") {
				t.Fatalf("multi mapping wrong seed %d", seed)
			}
		}
	}
	bad := DailyQuestion{Type: "single", Stem: strings.Repeat("x", 40), Explanation: "x", Options: []DailyOption{{"a", "1"}, {"b", "1"}, {"c", "1"}}, Correct: []int{0}}
	if _, ok := normalizeDaily(bad, "People", rand.New(rand.NewSource(1))); ok {
		t.Fatal("3 options should be rejected")
	}
}
