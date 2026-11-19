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
	Type        string   `json:"type"` // "choice" | "command"
	Prompt      string   `json:"prompt"`
	Choices     []string `json:"choices,omitempty"`  // choice: options
	Answer      *int     `json:"answer,omitempty"`   // choice: 0-based correct index
	Patterns    []string `json:"patterns,omitempty"` // command: accept regexes
	Explanation string   `json:"explanation"`
	Hint        string   `json:"hint,omitempty"`
	Skill       string   `json:"skill,omitempty"` // methodology phase (recon/enum/privesc/... | aws: iam/network/...), for the day-over-day report
}

// FlowOpts configures one thinking-flow generation.
type FlowOpts struct {
	Model      string
	CourseName string
	Topic      string // cheatsheet section / topic to build around ("" = whole sheet)
	Track      string // pentest: "PT1" | "WEB1" | "OPTS"; empty otherwise
	Focus      string // optional free-text the learner wants the lab to centre on
	Domain     string // "pentest" (default) | "aws"
	Language   string // "vi" | "en"
}

// trackBrief describes the kind of engagement/scenario for the chosen track,
// so the flow reads like a real THM/HTB challenge room (pentest) or a real
// client-system incident (aws).
func trackBrief(domain, track string) string {
	if domain == "aws" {
		return "an SAA-C03 style architecture scenario on a client's AWS account: a requirement (security, availability/resilience, performance, or cost) that the architect must satisfy by choosing the right services and design trade-offs, then justify — the way the Solutions Architect Associate exam frames its questions"
	}
	switch strings.ToUpper(strings.TrimSpace(track)) {
	case "WEB1":
		return "a TryHackMe WEB1 (Web Application Pentester) style target: web recon and mapping, authentication/session flaws, injection (SQLi, command injection, SSTI), broken access control / IDOR, file-upload and from a web foothold to a shell — web-application focus"
	case "CPTS":
		return "a HackTheBox CPTS (Certified Penetration Testing Specialist) style engagement that chains the full blueprint: information gathering and recon, service and web enumeration, exploitation to an initial foothold, Active Directory enumeration and attacks, Linux/Windows privilege escalation, lateral movement and pivoting/tunnelling, then post-exploitation — deeper and wider than PT1"
	case "OPTS":
		return "a mixed, open offensive engagement (CTF/boot-to-root style) that chains network/host, web and Active Directory the way a full challenge room does"
	default: // PT1
		return "a TryHackMe PT1 (Junior Penetration Tester) style box: external recon -> service enumeration -> initial foothold -> Linux/Windows privilege escalation (and basic Active Directory) — network/host focus"
	}
}

// pentestSkillTags lists the methodology phases to tag steps with, matched to
// the chosen exam's blueprint so the day-over-day report breaks competency down
// the way that exam does.
func pentestSkillTags(track string) string {
	switch strings.ToUpper(strings.TrimSpace(track)) {
	case "WEB1":
		return "recon, auth, injection, access-control, file-upload, foothold, post"
	case "CPTS":
		return "recon, enum, web, ad, privesc, lateral, pivoting, post"
	case "OPTS":
		return "recon, enum, web, ad, privesc, lateral, pivoting, post"
	default: // PT1
		return "recon, enum, foothold, privesc, ad"
	}
}

// GenerateThinkingFlow writes a methodology flow for a topic, grounded only in
// the learner's notes. opts.Domain selects pentest (default) or aws; opts.Track
// picks the pentest exam style; opts.Focus optionally steers the scenario.
func (c *Client) GenerateThinkingFlow(ctx context.Context, opts FlowOpts, notes []CheatNote) ([]FlowStep, error) {
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
	aws := opts.Domain == "aws"

	system := "You are a senior penetration-testing instructor preparing a student for the TryHackMe junior pentest path, which is open-book. You teach METHODOLOGY and decision-making, not rote command syntax. Reply with a single JSON object only."
	cmdPat := "the actual command to run (so they also build muscle memory for the few commands worth typing)"
	skillTags := pentestSkillTags(opts.Track)
	if aws {
		system = "You are a senior AWS Solutions Architect instructor preparing a student for the SAA-C03 exam. You teach how to CHOOSE the right services and designs for security, resilience, high performance and cost — architecture judgement and trade-offs, not rote syntax. Reply with a single JSON object only."
		cmdPat = "the AWS CLI command that would inspect or provision the resource central to this design decision (aws <service> <op> ...), so the choice stays concrete"
		skillTags = "secure, resilient, performance, cost"
	}

	// Lab length varies with the engagement rather than a fixed question count:
	// a full CPTS-style chain runs longer than a single PT1 box.
	loSteps, hiSteps := 10, 14
	switch strings.ToUpper(strings.TrimSpace(opts.Track)) {
	case "CPTS", "OPTS":
		loSteps, hiSteps = 16, 22
	}
	if aws {
		loSteps, hiSteps = 8, 12
	}

	focusLine := ""
	if f := strings.TrimSpace(opts.Focus); f != "" {
		focusLine = fmt.Sprintf("\n\nThe learner asked to FOCUS this lab on: %q. Centre the whole scenario on that; still ground every detail in the notes.", f)
	}

	user := fmt.Sprintf(`Scenario style: %s.
Topic to build around: %s
Course: %s

From the study notes below, build ONE coherent SITUATIONAL THINKING FLOW (methodology) — a single engagement against one named target (give it a name and an IP/host or an account/resource), so the steps CHAIN like a real challenge/incident: each step's result motivates the next. Between %d and %d ordered steps — use as many as a real engagement of this scope naturally needs (a full CPTS-style chain runs longer than a single PT1 box); never pad with filler. Each step is one question. Mixing two kinds:
- "choice": a realistic situation (2-4 sentences) ending in a decision — "what should you check or do NEXT, and why?" — with exactly 4 options, exactly 1 correct. At least 2 options must be defensible actions; the wrong ones are valid later but premature or out of methodology order (e.g. exploiting before enumerating, escalating before a foothold, remediating before confirming root cause). Never make a wrong option something no competent operator would ever do. Teach the reasoning: what to enumerate/inspect first, how to read a result, when to pivot or escalate.
- "command": at a key moment, %s.

Aim for roughly 60%% "choice" (thinking) and 40%% "command". The flow must be logically ordered. Base EVERYTHING ONLY on the notes; never invent tools, flags, services or techniques not present in them.%s

Across the whole flow, cover the exam components of this track (don't dwell on one phase) and tag each step with the phase it trains.

For EACH step provide:
- "type": "choice" or "command".
- "prompt": the situation/question (continue the same scenario).
- "skill": the methodology phase this step trains — exactly one of: %s (pick the single closest tag).
- "explanation": 1-3 sentences on why this is the right move and how to interpret the result (the key teaching point).
- choice steps also: "choices" (exactly 4 short strings) and "answer" (0-based index of the correct choice).
- command steps also: "patterns" (1-3 case-insensitive regex that accept the correct command; escape literal . / $ ( ) [ ]; use .+ for IPs/hosts/ports/ARNs/bucket names; RE2, no lookahead).
- optional "hint": a short nudge.

Return JSON: {"steps":[{"type":"choice","skill":"...","prompt":"...","choices":["...","...","...","..."],"answer":0,"explanation":"...","hint":"..."},{"type":"command","skill":"...","prompt":"...","patterns":["..."],"explanation":"...","hint":"..."}]}`,
		trackBrief(opts.Domain, opts.Track), topicOrAll(opts.Topic, aws), opts.CourseName, loSteps, hiSteps, cmdPat, focusLine, skillTags)

	if opts.Language == "en" {
		user += "\n\nIMPORTANT — LANGUAGE: Write prompt, choices, explanation and hint in ENGLISH. Keep commands/flags/service names exact."
	} else {
		user += "\n\nQUAN TRỌNG — NGÔN NGỮ: Viết prompt, choices, explanation và hint bằng TIẾNG VIỆT. Giữ nguyên lệnh, cờ, tên dịch vụ và code (không dịch)."
	}

	var out struct {
		Steps []FlowStep `json:"steps"`
	}
	if err := c.chatJSONModel(ctx, opts.Model, system, user, 0.6, &out); err != nil {
		return nil, err
	}
	return cleanFlowSteps(out.Steps, false)
}

func topicOrAll(topic string, aws bool) string {
	if strings.TrimSpace(topic) != "" {
		return topic
	}
	if aws {
		return "the whole AWS cheatsheet"
	}
	return "the whole cheatsheet"
}

// cleanFlowSteps validates/normalises steps. choiceOnly drops command steps
// (used by the PMP flow, which has no commands).
func cleanFlowSteps(steps []FlowStep, choiceOnly bool) ([]FlowStep, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	cleaned := make([]FlowStep, 0, len(steps))
	for _, s := range steps {
		s.Prompt = strings.TrimSpace(s.Prompt)
		s.Explanation = strings.TrimSpace(s.Explanation)
		s.Hint = strings.TrimSpace(s.Hint)
		s.Skill = strings.ToLower(strings.TrimSpace(s.Skill))
		if len(s.Prompt) < 8 || s.Explanation == "" {
			continue
		}
		isChoice := choiceOnly || strings.ToLower(strings.TrimSpace(s.Type)) == "choice"
		if isChoice {
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
		} else {
			var ps []string
			for _, p := range s.Patterns {
				if p = strings.TrimSpace(p); p != "" {
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
// PMI mindset. No command steps (PMP has no commands). focus optionally steers it.
func (c *Client) GeneratePmpFlow(ctx context.Context, model, topic, focus string, notes []CheatNote, language string) ([]FlowStep, error) {
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
	focusLine := ""
	if f := strings.TrimSpace(focus); f != "" {
		focusLine = fmt.Sprintf("\n\nThe learner asked to FOCUS this scenario on: %q. Centre the story on that; still ground it in the notes.", f)
	}
	system := "You are a PMP exam coach. You build SEQUENTIAL decision scenarios that teach the PMI mindset: analyse before acting, follow the plan and integrated change control, be proactive, collaborate, servant leadership in agile, escalate only when needed. Reply with a single JSON object only."
	user := fmt.Sprintf(`Topic: %s

From the PMP study notes below, build ONE evolving project scenario split into about 8 ORDERED decision points that teaches the end-to-end thinking for this topic. Each decision point is a "choice" step:
- "prompt": a short update of the SAME scenario (the situation evolves as it progresses) ending in a decision — usually "What should the project manager do NEXT/FIRST?" or "What is the BEST response?".
- "choices": exactly 4 options; exactly 1 correct per PMI thinking. Wrong options are common mistakes (acting without analysis, skipping change control, escalating too early, gold plating, etc.).
- "answer": 0-based index of the correct option.
- "explanation": 1-3 sentences naming the PMI principle/process behind the right answer and why the others are wrong.
- optional "hint".

The steps MUST build on each other (one continuous story), be realistic, and stay grounded ONLY in the notes. Do not use "all/none of the above" or refer to options by letter.%s

Return JSON: {"steps":[{"type":"choice","prompt":"...","choices":["...","...","...","..."],"answer":0,"explanation":"...","hint":"..."}]}`, topicLine, focusLine)

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
	return cleanFlowSteps(out.Steps, true)
}
