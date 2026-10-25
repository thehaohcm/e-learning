package ai

import "testing"

func TestCleanWordQuestion(t *testing.T) {
	ok := WordQuestion{Kind: " Usage ", Prompt: " Which? ", Choices: []string{"a", "b", "c", "d"}, Answer: 2}
	q, valid := CleanWordQuestion(ok)
	if !valid || q.Kind != WQUsage || q.Prompt != "Which?" {
		t.Fatalf("%+v %v", q, valid)
	}
	bad := []WordQuestion{
		{Kind: "other", Prompt: "x", Choices: []string{"a", "b", "c", "d"}},
		{Kind: "usage", Prompt: "", Choices: []string{"a", "b", "c", "d"}},
		{Kind: "usage", Prompt: "x", Choices: []string{"a", "b", "c"}},
		{Kind: "usage", Prompt: "x", Choices: []string{"a", "b", "c", "A"}},
		{Kind: "usage", Prompt: "x", Choices: []string{"a", "b", "c", "d"}, Answer: 4},
	}
	for i, b := range bad {
		if _, v := CleanWordQuestion(b); v {
			t.Errorf("case %d should be rejected", i)
		}
	}
}
