package ai

import (
	"context"
	"fmt"
	"strings"
)

// CheatNote is one of the learner's own notes, to be summarized.
type CheatNote struct {
	Title string
	Body  string
}

// SummarizeNotes reads the learner's own notes for a course and returns a
// concise, well-organized cheat sheet in Markdown. It never invents facts:
// it only reorganizes and condenses what the notes contain.
func (c *Client) SummarizeNotes(ctx context.Context, courseName string, notes []CheatNote) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
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
		if total+len(chunk) > 24000 { // keep the prompt bounded
			b.WriteString("\n(more notes omitted)\n")
			break
		}
		b.WriteString(chunk)
		total += len(chunk)
	}
	if strings.TrimSpace(b.String()) == "" {
		return "", fmt.Errorf("no notes to summarize")
	}

	system := "You turn a learner's own study notes into a clean cheat sheet written in Markdown."
	user := fmt.Sprintf(`Course: %s

Below are the learner's own notes for this course. Write a concise, well-organized
cheat sheet in Markdown that a learner can review quickly before an exam.

Rules:
- Only use information present in the notes. Do NOT invent facts, commands or numbers.
- Group related points under short headings; use bullet points.
- Keep technical terms, commands and code EXACT; put commands/code in fenced code blocks.
- Be compact: cut filler, merge duplicates. Keep the learner's language (Vietnamese or English) as written.
- No preamble or sign-off, just the cheat sheet itself.
- Output ONLY the Markdown of the cheat sheet (do not wrap the whole thing in a code block).

Notes:
%s`, courseName, b.String())

	// A long Markdown body (with code fences, quotes, newlines) is fragile inside
	// a JSON string, so we ask for Markdown directly instead of JSON.
	md, err := c.chatText(ctx, "", system, user, 0.3)
	if err != nil {
		return "", err
	}
	md = stripOuterFence(md)
	if md == "" {
		return "", fmt.Errorf("AI returned an empty summary")
	}
	return md, nil
}

// stripOuterFence removes a single ``` or ```markdown fence that wraps the whole
// answer, which some models add around the entire output.
func stripOuterFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	nl := strings.IndexByte(s, '\n')
	if nl < 0 || !strings.HasSuffix(s, "```") {
		return s
	}
	inner := strings.TrimSuffix(s[nl+1:], "```")
	return strings.TrimSpace(inner)
}
