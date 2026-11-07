package ai

import (
	"context"
	"fmt"
)

// FlagQuizOption is one choice in a multi-select recall quiz. Correct marks
// whether it is actually a real/valid option of the tool/command in question.
type FlagQuizOption struct {
	Label   string `json:"label"`
	Correct bool   `json:"correct"`
	Explain string `json:"explain"`
}

// FlagQuiz is one multi-select question testing recall of the vocabulary
// around one command — its CLI flags (hydra -l/-L/-p/-P...), or, for a tool
// whose "options" aren't dash-flags, the equivalent thing to memorize: e.g.
// msfconsole's settable variables (RHOST, RPORT, LHOST, LPORT, PAYLOAD...).
// Unlike a normal MCQ, it is NOT fixed at 4 options and NOT single-answer:
// the learner must pick every option that is real, among a mix of real and
// plausible-but-wrong ones.
type FlagQuiz struct {
	Prompt  string           `json:"prompt"`
	Options []FlagQuizOption `json:"options"`
}

// GenerateFlagQuiz writes a multi-select quiz question about the vocabulary
// a learner needs to recall around one command: CLI flags for a tool like
// hydra/nmap (-l -L -p -P...), or, for something like msfconsole's "set
// RHOST", the settable variable/parameter names (RHOST, RPORT, LHOST,
// LPORT, PAYLOAD...) instead — the model picks whichever fits the command.
// Grounded in the command the learner just reviewed and its context.
// Ephemeral: the caller does not persist this as a card, it is answered on
// the spot.
func (c *Client) GenerateFlagQuiz(ctx context.Context, model, command, context_ string) (FlagQuiz, error) {
	var out FlagQuiz
	if !c.Enabled() {
		return out, ErrNotConfigured
	}
	ctxLine := ""
	if context_ != "" {
		ctxLine = fmt.Sprintf("\nContext (what the learner was doing with it): %s", context_)
	}
	system := "You write a multi-select quiz question testing a learner's recall of the vocabulary around a pentest command: CLI flags for a flag-style tool, or settable variables/parameters for a console-style tool (e.g. msfconsole's RHOST/RPORT/LHOST/LPORT/PAYLOAD), or named subcommands — whichever is the actual thing a learner needs to memorize for this command. Reply with a single JSON object only."
	user := fmt.Sprintf(`Command the learner just reviewed: %s%s

First decide what kind of "options" are actually worth testing for this command:
- A flag-style CLI tool (hydra, nmap, gobuster, ss, ...) -> its command-line flags (e.g. -l, -L, -p, -P).
- A console-style tool entered through "set <NAME> <value>" (msfconsole, etc.) -> the settable variable/parameter NAMES (e.g. RHOST, RPORT, LHOST, LPORT, PAYLOAD), not dashes.
- Something else with a fixed vocabulary of subcommands/keywords -> those keywords.
Pick whichever matches the command above.

Write ONE multi-select quiz question testing whether the learner remembers THIS tool/command's real options in general — not just the one instance above. List between 5 and 12 short options: a mix of ones that genuinely ARE valid for it, and a few plausible-looking but WRONG ones (an invented name, a flag/variable borrowed from a different but similar tool) so the learner has to actually know the command, not just guess. Mark each option's correctness truthfully and accurately — do not invent a fake meaning for a real one, and do not mark a fake one as real.

Return JSON: {
  "prompt": "a question asking the learner to pick every option that is real for <tool/command name> (name it explicitly, and say what kind of option — flag, variable, subcommand — this question is about)",
  "options": [
    {"label": "-l (or RHOST, etc. — match the kind chosen above)", "correct": true, "explain": "short, accurate explanation of what it does, or of why it is not a real/relevant one"}
  ]
}
The "options" array must have at least 5 items and at least 2 must be correct.`, command, ctxLine)

	if err := c.chatJSONModel(ctx, model, system, user, 0.5, &out); err != nil {
		return out, err
	}
	if len(out.Options) < 2 {
		return out, fmt.Errorf("AI returned too few options")
	}
	return out, nil
}
