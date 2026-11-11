package ai

import (
	"context"
	"fmt"
	"strings"
)

// GenerateQuestionsFromNotes writes multiple-choice study questions grounded in
// the learner's own cheatsheet notes. topic (optional) focuses the questions on
// one part of the notes. The returned cards are MCQ (payload: choices, answer,
// explanation); the caller validates each with content.Normalize and inserts
// them. It never invents facts outside the notes.
func (c *Client) GenerateQuestionsFromNotes(ctx context.Context, model, courseName, topic string, count int, notes []CheatNote) ([]GeneratedCard, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	auto := count <= 0
	if count > 40 {
		count = 40
	}
	var b strings.Builder
	total := 0
	for i, n := range notes {
		title := strings.TrimSpace(n.Title)
		body := strings.TrimSpace(n.Body)
		if title == "" && body == "" {
			continue
		}
		chunk := fmt.Sprintf("### Note %d: %s\n%s\n\n", i+1, truncate(title, 200), truncate(body, 4000))
		if total+len(chunk) > 24000 {
			b.WriteString("\n(more notes omitted)\n")
			break
		}
		b.WriteString(chunk)
		total += len(chunk)
	}
	if strings.TrimSpace(b.String()) == "" {
		return nil, fmt.Errorf("no notes to generate from")
	}

	focus := ""
	if t := strings.TrimSpace(topic); t != "" {
		focus = fmt.Sprintf("\nFocus the questions on this topic/section: %q. If the notes have little on it, use the closest relevant notes.", truncate(t, 200))
	}

	countInstr := fmt.Sprintf("exactly %d ORIGINAL multiple-choice questions", count)
	if auto {
		countInstr = "ORIGINAL multiple-choice questions — one for each distinct fact, command or technique in the notes, merging near-duplicates (produce as many as the notes meaningfully support: at least 3, at most 40)"
	}
	system := "You create original study questions strictly from a learner's own notes. Reply with a single JSON object only."
	user := fmt.Sprintf(`Course: %s

The learner's own cheatsheet notes:
%s
Write %s to help the learner study THESE notes.%s

Rules:
- Base every question ONLY on the notes above. Do NOT invent commands, flags, numbers or facts that are not in the notes.
- For command notes, test when/why to use a command, what a flag does, or the expected output.
- For concept, process or formula notes (e.g. project management), mix definition questions, situational questions ("what should the project manager do first/next?") and calculation questions. For calculations use ONLY formulas stated in the notes, pick simple numbers and double-check the arithmetic so exactly one choice is correct.
- Vary which choice index is correct; the 3 wrong choices must be plausible (common confusions between similar terms in the notes), never obviously absurd.
- Each question has exactly 4 choices, exactly one correct answer (index into choices), and a short explanation.
- Keep commands, code and technical terms EXACT. Keep the learner's language (Vietnamese or English) as in the notes.
Return JSON: {"cards": [
  {"front": "question text", "payload": {"choices": ["A","B","C","D"], "answer": 0, "explanation": "why it is correct"}}
]}`, courseName, b.String(), countInstr, focus)

	var out struct {
		Cards []GeneratedCard `json:"cards"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.5, &out); err != nil {
		return nil, err
	}
	if len(out.Cards) > 40 {
		out.Cards = out.Cards[:40]
	}
	return out.Cards, nil
}
