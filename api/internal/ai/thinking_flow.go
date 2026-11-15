package ai

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// shuffleChoices randomizes the options and returns the new correct index, so
// the right answer is not always the one the model wrote first (usually A).
func shuffleChoices(ch []string, answer int, rng *rand.Rand) ([]string, int) {
	perm := rng.Perm(len(ch))
	out := make([]string, len(ch))
	na := 0
	for ni, oi := range perm {
		out[ni] = ch[oi]
		if oi == answer {
			na = ni
		}
	}
	return out, na
}

// Thinking-flow: a situational methodology drill for a topic. Instead of rote
// command syntax (PT1 is open-book), it walks the learner through how to THINK
// on a box — a mix of "choice" decision steps (what to check / do NEXT, with
// the reasoning) and a few "command" steps for the handful of commands worth
// typing. Reuses the lab step shape used by the drill runner.

// FlowStep is one step of a thinking flow (choice or command).
type FlowStep struct {
	Type        string   `json:"type"`              // "choice" | "command"
	Prompt      string   `json:"prompt"`
	Choices     []string `json:"choices,omitempty"` // choice: options
	Answer      *int     `json:"answer,omitempty"`  // choice: 0-based correct index
	Patterns    []string `json:"patterns,omitempty"`// command: accept regexes
	Explanation string   `json:"explanation"`
	Hint        string   `json:"hint,omitempty"`
}

// GenerateThinkingFlow writes a methodology flow for topic, grounded in notes.
func (c *Client) GenerateThinkingFlow(ctx context.Context, model, courseName, topic string, notes []CheatNote, language string) ([]FlowStep, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var nb strings.Builder
	for _, nt := range notes {
		if nb.Len() > 16000 {
			break
		}
		fmt.Fprintf(&nb, "## %s\n%s\n\n", nt.Title, truncate(nt.Body, 3500))
	}
	const count = 10
	system := "You are a senior penetration-testing instructor preparing a student for the TryHackMe PT1 (Junior Penetration Tester) exam, which is open-book. You teach METHODOLOGY and decision-making, not rote command syntax. Reply with a single JSON object only."
	user := fmt.Sprintf(`Topic: %s
Course: %s

From the study notes below, build a SITUATIONAL THINKING FLOW (methodology) for this topic: an ordered sequence of about %d steps that walks the learner through how to THINK on a real box, mixing two kinds of steps:
- "choice": a realistic situation (2-4 sentences) ending in a decision — "what should you check or do NEXT, and why?" — with exactly 4 options, exactly 1 correct. The wrong options should be plausible mistakes. Teach the reasoning: what to enumerate first, how to read a result, when to pivot.
- "command": at a key moment, the actual command to run (so they also build muscle memory for the few commands worth typing).

Aim for roughly 60%% "choice" (thinking) and 40%% "command". The flow must be logically ordered (enumerate -> interpret -> decide -> exploit). Base EVERYTHING ONLY on the notes; never invent tools, flags or techniques not present in them.

For EACH step provide:
- "type": "choice" or "command".
- "prompt": the situation/question.
- "explanation": 1-3 sentences on why this is the right move and how to interpret the result (the key teaching point).
- choice steps also: "choices" (exactly 4 short strings) and "answer" (0-based index of the correct choice).
- command steps also: "patterns" (1-3 case-insensitive regex that accept the correct command; escape literal . / $ ( ) [ ]; use .+ for IPs/hosts/ports; RE2, no lookahead).
- optional "hint": a short nudge.

Return JSON: {"steps":[{"type":"choice","prompt":"...","choices":["...","...","...","..."],"answer":0,"explanation":"...","hint":"..."},{"type":"command","prompt":"...","patterns":["..."],"explanation":"...","hint":"..."}]}`, topic, courseName, count, nb.String())

	if language == "en" {
		user += "\n\nIMPORTANT — LANGUAGE: Write prompt, choices, explanation and hint in ENGLISH. Keep commands/flags exact."
	} else {
		user += "\n\nQUAN TRỌNG — NGÔN NGỮ: Viết prompt, choices, explanation và hint bằng TIẾNG VIỆT. Giữ nguyên lệnh, cờ và code (không dịch)."
	}

	var out struct {
		Steps []FlowStep `json:"steps"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.6, &out); err != nil {
		return nil, err
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	cleaned := make([]FlowStep, 0, len(out.Steps))
	for _, s := range out.Steps {
		s.Prompt = strings.TrimSpace(s.Prompt)
		s.Explanation = strings.TrimSpace(s.Explanation)
		s.Hint = strings.TrimSpace(s.Hint)
		if len(s.Prompt) < 8 || s.Explanation == "" {
			continue
		}
		if strings.ToLower(strings.TrimSpace(s.Type)) == "choice" {
			var ch []string
			for _, o := range s.Choices {
				o = strings.TrimSpace(o)
				if o != "" {
					ch = append(ch, o)
				}
			}
			if len(ch) < 3 || len(ch) > 6 || s.Answer == nil || *s.Answer < 0 || *s.Answer >= len(ch) {
				continue
			}
			nc, na := shuffleChoices(ch, *s.Answer, rng)
			s.Type, s.Choices, s.Patterns, s.Answer = "choice", nc, nil, &na
		} else {
			var ps []string
			for _, p := range s.Patterns {
				p = strings.TrimSpace(p)
				if p != "" {
					ps = append(ps, p)
				}
			}
			if len(ps) == 0 {
				continue
			}
			s.Type, s.Patterns, s.Choices, s.Answer = "command", ps, nil, nil
		}
		cleaned = append(cleaned, s)
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("AI không tạo được bước nào hợp lệ")
	}
	return cleaned, nil
}

// GeneratePmpFlow builds a SEQUENTIAL, choice-only decision scenario for PMP: one
// evolving project situation split into ordered decision points, teaching the
// PMI mindset. No command steps (PMP has no commands).
func (c *Client) GeneratePmpFlow(ctx context.Context, model, topic string, notes []CheatNote, language string) ([]FlowStep, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var nb strings.Builder
	for _, nt := range notes {
		if nb.Len() > 16000 {
			break
		}
		fmt.Fprintf(&nb, "## %s\n%s\n\n", nt.Title, truncate(nt.Body, 3500))
	}
	topicLine := "the whole PMP cheatsheet"
	if strings.TrimSpace(topic) != "" {
		topicLine = topic
	}
	system := "You are a PMP exam coach. You build SEQUENTIAL decision scenarios that teach the PMI mindset: analyse before acting, follow the plan and integrated change control, be proactive, collaborate, servant leadership in agile, escalate only when needed. Reply with a single JSON object only."
	user := fmt.Sprintf(`Topic: %s

From the PMP study notes below, build ONE evolving project scenario split into about 8 ORDERED decision points that teaches the end-to-end thinking for this topic. Each decision point is a "choice" step:
- "prompt": a short update of the SAME scenario (the situation evolves as it progresses) ending in a decision — usually "What should the project manager do NEXT/FIRST?" or "What is the BEST response?".
- "choices": exactly 4 options; exactly 1 correct per PMI thinking. Wrong options are common mistakes (acting without analysis, skipping change control, escalating too early, gold plating, etc.).
- "answer": 0-based index of the correct option.
- "explanation": 1-3 sentences naming the PMI principle/process behind the right answer and why the others are wrong.
- optional "hint".

The steps MUST build on each other (one continuous story), be realistic, and stay grounded ONLY in the notes. Do not use "all/none of the above" or refer to options by letter.

Return JSON: {"steps":[{"type":"choice","prompt":"...","choices":["...","...","...","..."],"answer":0,"explanation":"...","hint":"..."}]}`, topicLine, nb.String())

	if language == "en" {
		user += "\n\nIMPORTANT — LANGUAGE: Write prompt, choices, explanation and hint in ENGLISH (PMP exam language)."
	} else {
		user += "\n\nQUAN TRỌNG — NGÔN NGỮ: Viết prompt, choices, explanation và hint bằng TIẾNG VIỆT; giữ nguyên thuật ngữ PMI tiếng Anh trong ngoặc nếu cần."
	}

	var out struct {
		Steps []FlowStep `json:"steps"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.6, &out); err != nil {
		return nil, err
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	cleaned := make([]FlowStep, 0, len(out.Steps))
	for _, s := range out.Steps {
		s.Prompt = strings.TrimSpace(s.Prompt)
		s.Explanation = strings.TrimSpace(s.Explanation)
		s.Hint = strings.TrimSpace(s.Hint)
		if len(s.Prompt) < 8 || s.Explanation == "" {
			continue
		}
		var ch []string
		for _, o := range s.Choices {
			if o = strings.TrimSpace(o); o != "" {
				ch = append(ch, o)
			}
		}
		if len(ch) < 3 || len(ch) > 6 || s.Answer == nil || *s.Answer < 0 || *s.Answer >= len(ch) {
			continue
		}
		nc, na := shuffleChoices(ch, *s.Answer, rng)
		s.Type, s.Choices, s.Patterns, s.Answer = "choice", nc, nil, &na
		cleaned = append(cleaned, s)
	}
	if len(cleaned) == 0 {
		return nil, fmt.Errorf("AI không tạo được bước nào hợp lệ")
	}
	return cleaned, nil
}
