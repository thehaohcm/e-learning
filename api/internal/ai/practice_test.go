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

func TestFillNotes(t *testing.T) {
	c, done := fakeModel(t, map[string]any{"items": []any{
		map[string]any{"i": 2, "base": "give up", "kind": "phrasal", "meaning_vi": "từ bỏ", "example_en": "She gave up smoking."},
		map[string]any{"i": 1, "base": " negotiate ", "kind": "phrasal", "pos": "verb", "meaning_vi": " đàm phán "},
		map[string]any{"i": 3, "base": "", "meaning_vi": ""},
	}})
	defer done()
	got, err := c.FillNotes(context.Background(), []NoteInput{{Text: "negotiated"}, {Text: "gave up"}, {Text: "xyzq"}})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Base != "negotiate" || got[0].Kind != "vocab" || got[0].MeaningVi != "đàm phán" {
		t.Fatalf("item 1: %+v", got[0]) // one word cannot be a phrasal verb
	}
	if got[1].Base != "give up" || got[1].Kind != "phrasal" {
		t.Fatalf("item 2: %+v", got[1])
	}
	if got[2].MeaningVi != "" || got[2].Base != "xyzq" || !got[2].Answered {
		t.Fatalf("item 3: %+v", got[2])
	}
}

func TestCleanPronunciation(t *testing.T) {
	p := Pronunciation{Syllables: []string{"Hy", "poth", " e", "sis"}, Stress: 1, TipsVi: []string{" th: lưỡi giữa răng ", ""}}
	if !CleanPronunciation(&p, "hypothesis") || p.Syllables[0] != "hy" || len(p.TipsVi) != 1 {
		t.Fatalf("got %+v", p)
	}
	bad := Pronunciation{Syllables: []string{"hy", "po", "sis"}, Stress: 1}
	if CleanPronunciation(&bad, "hypothesis") {
		t.Fatal("syllables that do not spell the word must be rejected")
	}
	phr := Pronunciation{Syllables: []string{"car", "ry", "out"}, Stress: 2}
	if !CleanPronunciation(&phr, "carry out") || phr.Stress != 2 {
		t.Fatalf("phrase: %+v", phr)
	}
	out := Pronunciation{Syllables: []string{"word"}, Stress: 5}
	if !CleanPronunciation(&out, "word") || out.Stress != 0 {
		t.Fatalf("stress out of range: %+v", out)
	}
}

func TestCleanFormOptions(t *testing.T) {
	got := CleanFormOptions([]string{"fluctuation", "Fluctuated", "fluctuating", "change", "fluctuant", "fluctuation"}, "fluctuated", "fluctuate")
	if len(got) != 3 || got[0] != "fluctuation" || got[1] != "fluctuating" || got[2] != "fluctuant" {
		t.Fatalf("%v", got)
	}
	if got := CleanFormOptions([]string{"decide", "decision", "decisively"}, "Decisive", "decisive"); got == nil || got[0] != "Decide" {
		t.Fatalf("capitalised: %v", got)
	}
	if got := CleanFormOptions([]string{"hypothetical", "two words"}, "hypothesis", "hypothesis"); got != nil {
		t.Fatalf("fewer than 3: %v", got)
	}
}

func TestCleanSound(t *testing.T) {
	syl := []string{"fluc", "tu", "ate"}
	ok := &SoundCheck{Syllable: 0, Letters: "U", Answer: 1, Same: []string{"luck", "cut", "two words"},
		Options: []SoundOption{{IPA: "/u/", LikeVi: "u", Example: "food"}, {IPA: "ʌ", LikeVi: "“â”", Example: "cut"}, {IPA: "juː", LikeVi: "iu", Example: "cute"}}}
	got := cleanSound(ok, syl)
	if got == nil || got.Letters != "u" || got.Options[0].IPA != "u" || got.Options[1].LikeVi != "â" || got.Answer != 1 || len(got.Same) != 2 {
		t.Fatalf("%+v", got)
	}
	if cleanSound(&SoundCheck{Syllable: 1, Letters: "x", Answer: 0, Options: ok.Options}, syl) != nil {
		t.Fatal("letters not in syllable")
	}
	if cleanSound(&SoundCheck{Syllable: 0, Letters: "u", Answer: 5, Options: ok.Options}, syl) != nil {
		t.Fatal("bad answer")
	}
	bad := &SoundCheck{Syllable: 0, Letters: "u", Answer: 0, Options: []SoundOption{{IPA: "", Example: "cut"}, {IPA: "u", Example: "food"}}}
	if cleanSound(bad, syl) != nil {
		t.Fatal("answer dropped")
	}
	if cleanSound(nil, syl) != nil {
		t.Fatal("nil")
	}
}

func TestCleanTutorWords(t *testing.T) {
	got := cleanTutorWords([]string{"analyze", "Analyze", "analyst", "phân tích", "  analysis ", "a b c d e", "analytical", "extra"})
	if len(got) != 4 || got[0] != "analyze" || got[1] != "analyst" || got[2] != "analysis" || got[3] != "analytical" {
		t.Fatalf("%v", got)
	}
}
