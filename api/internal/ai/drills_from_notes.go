package ai

import (
	"context"
	"fmt"
	"strings"
)

// GenerateDrillsFromNotes writes typed-command practice drills (kind "lab")
// grounded in the learner's own cheatsheet notes. Each drill asks the learner to
// TYPE a command for a goal; the answer is checked against lenient regex
// patterns (matched case-insensitively anywhere in the input). topic (optional)
// focuses the drills. It never invents commands outside the notes.
func (c *Client) GenerateDrillsFromNotes(ctx context.Context, model, courseName, topic string, count int, notes []CheatNote) ([]GeneratedCard, error) {
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
		focus = fmt.Sprintf("\nFocus the drills on this topic/section: %q. If the notes have little on it, use the closest relevant notes.", truncate(t, 200))
	}

	countInstr := fmt.Sprintf("exactly %d short practice drills", count)
	if auto {
		countInstr = "short practice drills — one for each distinct command or technique in the notes, merging near-duplicates (produce as many as the notes meaningfully support: at least 3, at most 40)"
	}
	system := "You create typed-command practice drills strictly from a learner's own notes. Reply with a single JSON object only."
	user := fmt.Sprintf(`Course: %s

The learner's own cheatsheet notes:
%s
Create %s that make the learner TYPE a command from THESE notes (active recall).%s

Each drill is one card with a single step:
- "front": a short title (the goal, a few words).
- "scenario": one sentence of context (what you're trying to do).
- "prompt": the instruction telling the learner what command to type (in the notes' language).
- "patterns": 1-3 case-insensitive REGEX patterns; a learner's answer is correct if ANY pattern matches anywhere in their input.
- "explanation": ONE concrete, correct example command (plain text, with placeholders like <target> where needed) — shown as the model answer when the learner reveals it.
- "note": 1-3 short sentences IN THE NOTES' LANGUAGE explaining what the command does and what EACH argument/parameter/flag means. Example: for "unshadow passwd.txt shadow.txt", say passwd.txt is the copied /etc/passwd, shadow.txt the copied /etc/shadow, and unshadow merges them into one file John can crack.
- "lesson": 1-3 short sentences IN THE NOTES' LANGUAGE teaching WHEN to use this command/technique and WHY (the situation it fits, what problem it solves, how it fits a pentest workflow). This is the takeaway to remember, not a repeat of the note.

Rules for patterns (IMPORTANT):
- ORDER-INDEPENDENT FLAGS: when the command is a tool plus combined single-letter flags with NO values (e.g. ss, ls, ps, netstat), use the special form "@flags:<cmd>:<letters>" so the flags are accepted in ANY order or grouping. Example: "@flags:ss:tunlp" accepts "ss -tunlp", "ss -utpln", "ss -t -u -n -l -p" and "ss -tun -lp". List only the letters, no dashes.
- Otherwise use a VALID Go/RE2 regex (case-insensitive, matched anywhere): match only the essential, stable tokens — the tool name and its distinctive flags — joined with \s+ (e.g. "nmap\s+-sC\s+-sV", "hydra\s+-l\s+.+\s+-P"). RE2 has NO lookahead/backreferences.
- Do NOT bake in IP addresses, hostnames, file names or ports — use .+ or .* where the user fills those in, or leave them out.
- Escape regex metacharacters that appear literally (\. \/ \$ \( \) \[ \] \? \* \+). Prefer plain tokens over complex regex.
- Provide a second, looser fallback pattern (e.g. just the tool name "\btool\b") when helpful.
- Base everything ONLY on the notes; do not invent commands or flags not present in them.

Return JSON: {"cards": [
  {"front": "goal title", "payload": {"mode": "lab", "scenario": "...", "steps": [
    {"prompt": "Type the command to ...", "patterns": ["tool\\s+-flag", "\\btool\\b"], "hint": "optional short hint", "explanation": "tool -flag <target>", "note": "what the command does and what each argument/flag means", "lesson": "when to use it and why"}
  ]}}
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
