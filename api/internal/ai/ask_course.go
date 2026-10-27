package ai

import (
	"context"
	"fmt"
	"strings"
)

// CourseTurn is an earlier question and answer in the same course chat.
type CourseTurn struct {
	Q string `json:"q"`
	A string `json:"a"`
}

// AskCourse answers a learner's free question about a course in one request.
// courseName is the subject (e.g. "Pentesting"); notes is the learner's own
// cheat-sheet text for the course (optional context); history is earlier turns.
// The answer is Markdown in Vietnamese.
func (c *Client) AskCourse(ctx context.Context, model, courseName, notes string, history []CourseTurn, question string) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Course: %s\n", courseName)
	if n := strings.TrimSpace(notes); n != "" {
		fmt.Fprintf(&b, "\nThe learner's own notes for this course (use them when relevant, but you may also add correct general knowledge):\n%s\n", truncate(n, 8000))
	}
	if len(history) > 0 {
		b.WriteString("\nEarlier in this chat:\n")
		for _, t := range history {
			fmt.Fprintf(&b, "Q: %s\nA: %s\n", truncate(t.Q, 400), truncate(t.A, 1200))
		}
	}

	prompt := fmt.Sprintf(`You are a knowledgeable tutor helping a learner with the course above.
%s
The learner asks: %s

Answer IN VIETNAMESE (keep technical terms, commands and code in their original language).
- Be accurate and concrete; if you are not sure, say so rather than inventing.
- Keep it focused and not too long; use Markdown: short headings, bullet points, and fenced code blocks for commands/code.
- If the question is clearly unrelated to the course, answer briefly and say it is outside this course.
Return JSON: {"answer_md":"your answer in Markdown"}`, b.String(), truncate(question, 1500))

	var out struct {
		Answer string `json:"answer_md"`
	}
	if err := c.chatJSONModel(ctx, model, "You are a helpful course tutor. Reply with a single JSON object only.", prompt, 0.4, &out); err != nil {
		return "", err
	}
	a := strings.TrimSpace(out.Answer)
	if a == "" {
		return "", fmt.Errorf("empty answer")
	}
	return a, nil
}
