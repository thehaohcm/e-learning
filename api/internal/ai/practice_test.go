package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeModel answers every chat request with the given JSON object.
func fakeModel(t *testing.T, reply any) (*Client, func()) {
	b, _ := json.Marshal(reply)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": string(b)}}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	return NewClient("k", srv.URL, "m"), srv.Close
}

func TestGeneratePracticeSentences(t *testing.T) {
	c, done := fakeModel(t, map[string]any{"items": []any{
		map[string]any{"word": "Negotiate", "sentences": []any{
			map[string]any{"en": "We negotiated a better price with the supplier last week.", "vi": "Tuần trước chúng tôi đã thương lượng giá tốt hơn."},
			map[string]any{"en": "This sentence does not use the target word at all.", "vi": "x"}, // dropped
			map[string]any{"en": "Negotiate.", "vi": "y"}, // too short
		}},
		// misspelled by the model: matched by position
		map[string]any{"word": "carry-out", "sentences": []any{
			map[string]any{"en": "The team carried out a survey of 500 customers.", "vi": "Nhóm đã thực hiện khảo sát 500 khách hàng."},
		}},
	}})
	defer done()
	got, err := c.GeneratePracticeSentences(context.Background(),
		[]PracticeWord{{Word: "negotiate", MeaningVi: "đàm phán"}, {Word: "carry out", MeaningVi: "thực hiện"}}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got["negotiate"]) != 1 || len(got["carry out"]) != 1 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestGradeTranslationsKeepsOrder(t *testing.T) {
	c, done := fakeModel(t, map[string]any{"results": []any{
		map[string]any{"i": 2, "correct": false, "score": 140, "feedback_vi": "sai"},
		map[string]any{"i": 1, "correct": true, "score": 90, "feedback_vi": "tốt"},
	}})
	defer done()
	got, err := c.GradeTranslations(context.Background(), []TranslationItem{{}, {}, {}})
	if err != nil {
		t.Fatal(err)
	}
	if !got[0].Correct || got[0].Score != 90 || got[1].Correct || got[1].Score != 100 || got[2].Score != -1 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestGradeWrittenSentences(t *testing.T) {
	c, done := fakeModel(t, map[string]any{"results": []any{
		map[string]any{"i": 1, "correct": true, "score": 95, "feedback_vi": "Tốt", "corrected": " We negotiated a better price. "},
		map[string]any{"i": 2, "correct": false, "score": -5, "feedback_vi": "Sai nghĩa", "corrected": "x"},
	}})
	defer done()
	got, err := c.GradeWrittenSentences(context.Background(), []WrittenItem{{}, {}})
	if err != nil {
		t.Fatal(err)
	}
	if !got[0].Correct || got[0].Corrected != "We negotiated a better price." || got[1].Correct || got[1].Score != 0 {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestWritingTaskNormalize(t *testing.T) {
	vocab := []string{"negotiate", "carry out", "deadline", "budget"}
	c, done := fakeModel(t, map[string]any{"type": "weird", "instruction_en": " Reply to Anna. ", "instruction_vi": "Trả lời Anna.",
		"min_words": 5, "max_words": 9999, "required_words": []string{"Negotiate", "unknown", "negotiate"}})
	defer done()
	got, err := c.GenerateWritingTask(context.Background(), "reading", "easy", "T", "text", vocab)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "email_reply" || got.MinWords != 40 || got.MaxWords != 70 || got.InstructionEn != "Reply to Anna." || !got.Generated {
		t.Fatalf("unexpected %+v", got)
	}
	if len(got.RequiredWords) != 3 || got.RequiredWords[0] != "negotiate" || got.RequiredWords[1] != "carry out" {
		t.Fatalf("required words %v", got.RequiredWords)
	}
	tpl := TemplateWritingTask("conversation", "ielts", "Coffee chat", vocab[:2])
	if len(got.OutlineVi) == 0 || len(got.Phrases) == 0 { // the model gave none: generic help for an email reply
		t.Fatalf("no help: %+v", got)
	}
	if tpl.Type != "message" || tpl.Generated || len(tpl.RequiredWords) != 2 || tpl.MinWords != 120 || tpl.SampleAnswer != "" || len(tpl.OutlineVi) != 4 {
		t.Fatalf("template %+v", tpl)
	}
}
