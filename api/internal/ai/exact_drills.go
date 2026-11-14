package ai

import (
	"context"
	"fmt"
	"strings"
)

// GenerateExactDrills turns an EXPLICIT, pre-extracted list of commands into
// typed-command practice drills, one drill per command, in the same order.
// Unlike GenerateDrillsFromNotes, the AI has no discretion over which or how
// many commands become drills — that was already decided deterministically
// by ExtractCommandLines; the AI only writes the learner-facing wording and
// match patterns for each one. The caller must still treat the result as
// advisory and validate the count: if the model returns fewer cards than
// requested, fill any missing index with BuildFallbackDrillCard so no
// command is ever silently dropped.
func (c *Client) GenerateExactDrills(ctx context.Context, model, courseName string, lines []CommandLine, language string) ([]GeneratedCard, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no commands to generate from")
	}
	var b strings.Builder
	for i, l := range lines {
		note := l.Note
		if note == "" {
			note = "(không có ghi chú gốc)"
		}
		fmt.Fprintf(&b, "%d. Command: %s\n   Context: %s\n", i+1, l.Command, note)
	}
	system := "You write typed-command practice drills from an EXACT given list of commands. Reply with a single JSON object only."
	user := fmt.Sprintf(`Course: %s

Here is the EXACT, numbered list of %d commands. You MUST produce EXACTLY %d drills, one per command, IN THE SAME ORDER (item i's drill is for command i). Do not add, remove, merge, split or reorder any — this list is final and already complete.

%s
For EACH command, write one drill:
- "front": a short title (the goal, a few words).
- "scenario": one sentence of context.
- "prompt": instruction telling the learner what to type (in the notes' language).
- "patterns": 1-3 case-insensitive REGEX patterns (or "@flags:<cmd>:<letters>" for order-independent short flags like ss/ls/ps — letters only, no dashes) that accept this exact command and reasonable equivalent phrasings; a learner's input is correct if ANY pattern matches anywhere in it.
- "explanation": this exact command (plain text), generalized with placeholders like <target> if it has a literal IP/host/port that should generalize.
- "note": 1-3 short sentences IN THE NOTES' LANGUAGE on what the command/each flag means — use the given Context if it explains that, else infer from the command itself.
- "lesson": 1-3 short sentences IN THE NOTES' LANGUAGE on when/why to use it in a pentest.

Regex rules: escape literal metacharacters (\. \/ \$ \( \) \[ \] \? \* \+), join stable tokens with \s+, never bake in specific IPs/hosts/ports (use .+ or .* there), prefer a looser fallback pattern (e.g. just \btool\b) as the 2nd/3rd pattern. RE2 has no lookahead/backreferences. Base everything ONLY on the given command/context; do not invent commands or flags not present in it.

Return JSON: {"cards": [
  {"front": "goal title", "payload": {"mode": "lab", "scenario": "...", "steps": [
    {"prompt": "...", "patterns": ["..."], "explanation": "...", "note": "...", "lesson": "..."}
  ]}}
]} — the "cards" array must have exactly %d items, in command order.`, courseName, len(lines), len(lines), b.String(), len(lines))

	if language == "en" {
		user += "\n\nIMPORTANT — LANGUAGE: Write prompt, note and lesson in ENGLISH. Keep commands, flags and code exact."
	} else {
		user += "\n\nQUAN TRỌNG — NGÔN NGỮ: Viết prompt, note và lesson bằng TIẾNG VIỆT. Giữ nguyên lệnh, cờ và code (không dịch)."
	}

	var out struct {
		Cards []GeneratedCard `json:"cards"`
	}
	if err := c.chatJSONModel(ctx, model, system, user, 0.4, &out); err != nil {
		return nil, err
	}
	return out.Cards, nil
}
