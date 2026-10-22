package ai

import (
	"context"
	"fmt"
	"strings"
)

// WritingTask is a short writing exercise based on a reading or conversation.
type WritingTask struct {
	Type          string   `json:"type"`           // email_reply | message | opinion | summary
	InstructionEn string   `json:"instruction_en"` // what to write, in English (sent to the evaluator)
	InstructionVi string   `json:"instruction_vi"`
	MinWords      int      `json:"min_words"`
	MaxWords      int      `json:"max_words"`
	RequiredWords []string `json:"required_words"` // from the text's vocabulary
	// Help for when the learner is stuck, revealed step by step in the web app:
	OutlineVi    []string       `json:"outline_vi"`     // 1. the points to write, in Vietnamese
	Phrases      []UsefulPhrase `json:"useful_phrases"` // 2. ready-made phrases (opening, linking, closing)
	SampleAnswer string         `json:"sample_answer"`  // 3. a model answer using the required words
	Generated    bool           `json:"generated"`      // false = template (no AI)
}

var taskWords = map[string][2]int{"easy": {40, 70}, "medium": {60, 100}, "ielts": {120, 180}}

// GenerateWritingTask asks the model for a TOEIC/IELTS-style task that answers
// the text (reply to the email, message after the conversation, opinion on the
// article...). vocab are the words of the text; 3 of them become required.
func (c *Client) GenerateWritingTask(ctx context.Context, kind, level, title, text string, vocab []string) (*WritingTask, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	wr := taskWords[level]
	if wr[0] == 0 {
		wr = taskWords["medium"]
	}
	prompt := fmt.Sprintf(`Here is a %s a Vietnamese learner has just read (level: %s).
Title: %s
"""
%s
"""
Vocabulary of the text: %s

Write ONE short writing task that makes the learner respond to this text, like TOEIC Writing or IELTS:
- an email/notice/ad -> reply to it ("email_reply"); a conversation -> a message as one speaker, e.g. confirming what was agreed ("message");
  an article/report -> give and support an opinion ("opinion") or summarise for a colleague ("summary").
- Clear and concrete: who writes to whom and 2-3 points to include.
- Length: %d-%d words.
- "required_words": 3 words from the vocabulary list (exactly as listed) that fit the task naturally.
- Help for a learner who is stuck:
  "outline_vi": 3-5 short points in Vietnamese, in order (what each part/paragraph should say);
  "useful_phrases": 4-6 phrases for this kind of text (opening, linking, closing) with Vietnamese meaning;
  "sample_answer": a model answer within the word limit, using all required words, at the learner's level.
Return JSON: {"type":"email_reply","instruction_en":"...","instruction_vi":"(the same in Vietnamese)","min_words":%d,"max_words":%d,
"required_words":["...","...","..."],"outline_vi":["..."],"useful_phrases":[{"phrase":"...","meaning_vi":"..."}],"sample_answer":"..."}`,
		kind, level, title, truncate(text, 4000), strings.Join(vocab, ", "), wr[0], wr[1], wr[0], wr[1])

	var t WritingTask
	if err := c.chatJSON(ctx, "You are an experienced TOEIC and IELTS writing teacher. Reply with a single JSON object only.", prompt, 0.6, &t); err != nil {
		return nil, err
	}
	t.InstructionEn, t.InstructionVi = strings.TrimSpace(t.InstructionEn), strings.TrimSpace(t.InstructionVi)
	t.SampleAnswer = strings.TrimSpace(t.SampleAnswer)
	if t.InstructionEn == "" {
		return nil, fmt.Errorf("empty writing task")
	}
	t.Generated = true
	t.normalize(kind, level, vocab)
	return &t, nil
}

// TemplateWritingTask is used when AI is not available.
func TemplateWritingTask(kind, level, title string, vocab []string) *WritingTask {
	t := &WritingTask{}
	if kind == "conversation" {
		t.Type = "message"
		t.InstructionEn = fmt.Sprintf("You took part in the conversation \"%s\". Write a short message to the other person: say what you agreed on and what happens next.", title)
		t.InstructionVi = fmt.Sprintf("Bạn là một người trong đoạn hội thoại \"%s\". Viết tin nhắn ngắn cho người kia: tóm tắt điều đã thống nhất và bước tiếp theo.", title)
	} else {
		t.Type = "summary"
		t.InstructionEn = fmt.Sprintf("Write a short email to a colleague summarising \"%s\" and giving your opinion about it.", title)
		t.InstructionVi = fmt.Sprintf("Viết email ngắn cho đồng nghiệp tóm tắt bài \"%s\" và nêu ý kiến của bạn.", title)
	}
	t.normalize(kind, level, vocab)
	return t
}

// generic help by text type, used by the template and when the model gave none
var templateHelp = map[string]struct {
	outline []string
	phrases []UsefulPhrase
}{
	"message": {
		[]string{"Chào và nhắc lại cuộc trò chuyện", "Tóm tắt điều hai bên đã thống nhất", "Nêu bước tiếp theo / thời gian", "Kết thúc thân thiện"},
		[]UsefulPhrase{{"Just a quick message about…", "Nhắn nhanh về…"}, {"As we agreed,…", "Như đã thống nhất,…"}, {"The next step is to…", "Bước tiếp theo là…"}, {"Let me know if…", "Báo mình nếu…"}, {"See you on…", "Hẹn gặp vào…"}},
	},
	"email_reply": {
		[]string{"Cảm ơn / nhắc đến email hoặc thông báo", "Trả lời ý chính (đồng ý, xác nhận…)", "Hỏi thêm hoặc đề xuất 1 điều", "Kết thư lịch sự"},
		[]UsefulPhrase{{"Thank you for your email about…", "Cảm ơn email của bạn về…"}, {"I am writing to confirm that…", "Tôi viết để xác nhận rằng…"}, {"Could you let me know…?", "Bạn có thể cho tôi biết…?"}, {"I would also like to suggest…", "Tôi cũng muốn đề xuất…"}, {"I look forward to hearing from you.", "Mong sớm nhận phản hồi."}},
	},
	"summary": {
		[]string{"Giới thiệu bài đọc nói về gì", "Tóm tắt 2-3 ý chính", "Nêu ý kiến của bạn", "Kết luận / đề xuất"},
		[]UsefulPhrase{{"I read an article about…", "Tôi đã đọc một bài về…"}, {"The main point is that…", "Ý chính là…"}, {"In addition,…", "Ngoài ra,…"}, {"In my opinion,…", "Theo tôi,…"}, {"Overall,…", "Nhìn chung,…"}},
	},
	"opinion": {
		[]string{"Nêu quan điểm của bạn", "Lý do 1 + ví dụ", "Lý do 2 + ví dụ", "Kết luận, nhắc lại quan điểm"},
		[]UsefulPhrase{{"In my view,…", "Theo quan điểm của tôi,…"}, {"The main reason is that…", "Lý do chính là…"}, {"For example,…", "Ví dụ,…"}, {"On the other hand,…", "Mặt khác,…"}, {"To sum up,…", "Tóm lại,…"}},
	},
}

// normalize keeps sensible word limits and required words taken from the vocabulary.
func (t *WritingTask) normalize(kind, level string, vocab []string) {
	wr := taskWords[level]
	if wr[0] == 0 {
		wr = taskWords["medium"]
	}
	if t.MinWords < 20 || t.MinWords > 300 {
		t.MinWords = wr[0]
	}
	if t.MaxWords < t.MinWords || t.MaxWords > 400 {
		t.MaxWords = max(wr[1], t.MinWords+20)
	}
	inVocab := map[string]string{}
	for _, v := range vocab {
		inVocab[strings.ToLower(v)] = v
	}
	var req []string
	seen := map[string]bool{}
	for _, w := range t.RequiredWords {
		if v, ok := inVocab[strings.ToLower(strings.TrimSpace(w))]; ok && !seen[v] {
			seen[v] = true
			req = append(req, v)
		}
	}
	for _, v := range vocab { // top up to 3 from the vocabulary
		if len(req) >= 3 {
			break
		}
		if !seen[v] {
			seen[v] = true
			req = append(req, v)
		}
	}
	if len(req) > 4 {
		req = req[:4]
	}
	t.RequiredWords = req
	switch t.Type {
	case "email_reply", "message", "opinion", "summary":
	default:
		if kind == "conversation" {
			t.Type = "message"
		} else {
			t.Type = "email_reply"
		}
	}
	// generic help when the model gave none (or for the template task)
	help := templateHelp[t.Type]
	if len(t.OutlineVi) == 0 {
		t.OutlineVi = help.outline
	}
	if len(t.Phrases) == 0 {
		t.Phrases = help.phrases
	}
}
