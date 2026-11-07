package ai

import (
	"encoding/json"
	"regexp"
	"strings"
)

// CommandLine is one command deterministically pulled out of a cheatsheet
// note's fenced code block, plus whatever comment explained it (if any).
type CommandLine struct {
	Command string
	Note    string
}

var fencedBlockRe = regexp.MustCompile(`(?s)` + "```" + `[a-zA-Z0-9_-]*\r?\n(.*?)` + "```")

// inlineCommentRe splits "command            # explanation" — the comment
// must be set off by at least two spaces, so a single "#" that is part of
// the command itself (rare, but e.g. a URL fragment) is left alone.
var inlineCommentRe = regexp.MustCompile(`^(.*?)\s{2,}#\s?(.*)$`)

// ExtractCommandLines deterministically lists EVERY command line inside the
// fenced (```...```) code blocks of the given notes — no AI involved, so
// nothing here can be skipped, merged or miscounted by a model's judgment.
// A line that is only a comment (starts with "#") is not a command itself;
// it is kept as context for the next real command line instead (cheatsheets
// often write a one-line label like "# Windows (exe)" above a group of
// commands). Inline trailing comments ("cmd   # explains it") become that
// command's Note, overriding any pending label comment. Duplicate commands
// (same text, same note) are kept once.
func ExtractCommandLines(notes []CheatNote) []CommandLine {
	var out []CommandLine
	seen := map[string]bool{}
	for _, n := range notes {
		for _, block := range fencedBlockRe.FindAllStringSubmatch(n.Body, -1) {
			pending := ""
			for _, raw := range strings.Split(block[1], "\n") {
				line := strings.TrimRight(raw, " \t\r")
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				if strings.HasPrefix(trimmed, "#") {
					pending = strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
					continue
				}
				cmd := line
				note := pending
				if m := inlineCommentRe.FindStringSubmatch(line); m != nil {
					cmd = m[1]
					if c := strings.TrimSpace(m[2]); c != "" {
						note = c
					}
				}
				cmd = strings.TrimSpace(cmd)
				pending = ""
				if cmd == "" {
					continue
				}
				key := cmd + "|" + note
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, CommandLine{Command: cmd, Note: note})
			}
		}
	}
	return out
}

// BuildFallbackDrillCard makes a guaranteed, AI-free "lab" card straight from
// one extracted command line: a literal (escaped) match on the exact text.
// It is less lenient than an AI-written pattern (no equivalent phrasings,
// no order-independent flags) but it is always correct and always present —
// the safety net so a command is never silently missing just because the AI
// skipped, merged or mis-ordered it in a batch.
func BuildFallbackDrillCard(l CommandLine) GeneratedCard {
	prompt := "Gõ lại lệnh sau cho đúng:"
	if l.Note != "" {
		prompt = "Gõ lệnh để: " + l.Note
	}
	step := map[string]any{
		"prompt":      prompt,
		"patterns":    []string{regexp.QuoteMeta(l.Command)},
		"explanation": l.Command,
	}
	if l.Note != "" {
		step["note"] = l.Note
	}
	payload := map[string]any{
		"mode":  "lab",
		"steps": []any{step},
	}
	raw, _ := json.Marshal(payload)
	front := l.Command
	if len(front) > 60 {
		front = front[:60] + "…"
	}
	return GeneratedCard{Front: front, Payload: raw}
}
