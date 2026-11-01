// Package content defines the JSON payload of each card kind and how cards are
// validated and turned into compact views (e.g. for the ESP32 device).
package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"mvp-lang-app/internal/models"
)

// Sense is one meaning of a word/phrasal verb with an example at a given level:
// 1 = simple, 2 = everyday, 3 = academic/formal.
type Sense struct {
	Level       int    `json:"level"`
	MeaningEn   string `json:"meaning_en,omitempty"`
	MeaningVi   string `json:"meaning_vi,omitempty"`
	ExampleEn   string `json:"example_en,omitempty"`
	ExampleVi   string `json:"example_vi,omitempty"`
	UsageNote   string `json:"usage_note,omitempty"`
	GrammarNote string `json:"grammar_note,omitempty"`
}

type VocabPayload struct {
	IPA      string  `json:"ipa,omitempty"`
	POS      string  `json:"pos,omitempty"`
	Topic    string  `json:"topic,omitempty"`
	AudioURL string  `json:"audio_url,omitempty"`
	Senses   []Sense `json:"senses"`
}

type PhrasalPayload struct {
	Verb      string  `json:"verb"`
	Particle  string  `json:"particle"`
	Separable bool    `json:"separable"`
	Senses    []Sense `json:"senses"`
}

// ConceptPayload is a generic term/definition card (PMP terms, formulas, ...).
// Writing phrases are concept cards too: front = what to say (Vietnamese),
// back = the English frame, with an example and a group ("Lý do", "Email: kết thư").
type ConceptPayload struct {
	Back      string `json:"back"`
	NoteVi    string `json:"note_vi,omitempty"`
	Example   string `json:"example,omitempty"`
	ExampleVi string `json:"example_vi,omitempty"`
	Group     string `json:"group,omitempty"` // shown like a part of speech; wrong options come from the same group first
	Order     int    `json:"order,omitempty"` // position in the phrase bank
}

// MCQPayload is a multiple-choice question; the card front is the question.
type MCQPayload struct {
	Choices     []string `json:"choices"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation,omitempty"`
}

// LabStep is one step of a lab card. Two kinds of step:
//
//   - "command" (default): the learner types a shell command, checked against
//     Patterns (regexes, case-insensitive, matched anywhere in the trimmed,
//     whitespace-collapsed input) — any one matching counts as correct.
//   - "choice": the learner picks from Choices (a "what do you do next?"
//     decision point); Answer is the correct option's index into Choices.
//
// Output, when set, is simulated tool/terminal output shown above the prompt
// (e.g. a canned nmap result) so the learner reacts to it like a real
// engagement, without needing a live shell. Checking happens entirely
// client-side; this is self-practice, not a graded exam.
type LabStep struct {
	Output      string   `json:"output,omitempty"`      // simulated command output shown before the prompt
	Prompt      string   `json:"prompt"`                // scenario/instruction for this step
	Hint        string   `json:"hint,omitempty"`        // optional nudge, revealed on request
	Type        string   `json:"type,omitempty"`        // "command" (default) or "choice"
	Patterns    []string `json:"patterns,omitempty"`    // accepted command regexes (type == "command")
	Choices     []string `json:"choices,omitempty"`     // answer options (type == "choice")
	Answer      int      `json:"answer,omitempty"`      // correct index into Choices (type == "choice")
	Explanation string   `json:"explanation,omitempty"` // shown after a correct answer
}

// LabPayload is a sequential, hands-on drill (a "lab"): a short scenario plus
// an ordered list of steps the learner must get right, one at a time, to
// progress. Mode is a display hint only: "lab" (default) is a focused
// command-typing drill; "assignment" is a fuller walkthrough that mixes
// simulated tool output, typed commands and multiple-choice decision points
// to resemble a real engagement end-to-end.
type LabPayload struct {
	Mode       string    `json:"mode,omitempty"`
	Scenario   string    `json:"scenario,omitempty"`
	Steps      []LabStep `json:"steps"`
	Reference  string    `json:"reference,omitempty"`  // optional Markdown study notes shown below the runner, for review
	Difficulty string    `json:"difficulty,omitempty"` // "co_ban" | "trung_binh" | "nang_cao" -- display hint only
}

func ValidKind(k string) bool {
	switch k {
	case models.KindVocab, models.KindPhrasal, models.KindConcept, models.KindMCQ, models.KindLab:
		return true
	}
	return false
}

// Normalize validates a payload for the given kind and returns it re-encoded
// (unknown fields dropped, strings trimmed).
func Normalize(kind, front string, raw json.RawMessage) (json.RawMessage, error) {
	if strings.TrimSpace(front) == "" {
		return nil, errors.New("front is required")
	}
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	switch kind {
	case models.KindVocab:
		var p VocabPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("invalid vocab payload: %w", err)
		}
		p.Senses = cleanSenses(p.Senses)
		if len(p.Senses) == 0 {
			return nil, errors.New("vocab card needs at least one sense with a meaning")
		}
		return json.Marshal(p)
	case models.KindPhrasal:
		var p PhrasalPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("invalid phrasal payload: %w", err)
		}
		p.Verb, p.Particle = strings.TrimSpace(p.Verb), strings.TrimSpace(p.Particle)
		if p.Verb == "" || p.Particle == "" {
			parts := strings.SplitN(strings.TrimSpace(front), " ", 2)
			if len(parts) == 2 {
				p.Verb, p.Particle = parts[0], parts[1]
			}
		}
		p.Senses = cleanSenses(p.Senses)
		if len(p.Senses) == 0 {
			return nil, errors.New("phrasal card needs at least one sense with a meaning")
		}
		return json.Marshal(p)
	case models.KindConcept:
		var p ConceptPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("invalid concept payload: %w", err)
		}
		p.Back = strings.TrimSpace(p.Back)
		if p.Back == "" {
			return nil, errors.New("concept card needs a back side")
		}
		return json.Marshal(p)
	case models.KindMCQ:
		var p MCQPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("invalid mcq payload: %w", err)
		}
		var choices []string
		for _, c := range p.Choices {
			if c = strings.TrimSpace(c); c != "" {
				choices = append(choices, c)
			}
		}
		if len(choices) < 2 || len(choices) != len(p.Choices) {
			return nil, errors.New("mcq needs at least 2 non-empty choices")
		}
		if p.Answer < 0 || p.Answer >= len(choices) {
			return nil, errors.New("mcq answer index out of range")
		}
		p.Choices = choices
		return json.Marshal(p)
	case models.KindLab:
		var p LabPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("invalid lab payload: %w", err)
		}
		p.Scenario = strings.TrimSpace(p.Scenario)
		p.Mode = strings.TrimSpace(p.Mode)
		p.Reference = strings.TrimSpace(p.Reference)
		p.Difficulty = strings.TrimSpace(p.Difficulty)
		var steps []LabStep
		for _, st := range p.Steps {
			st.Output = strings.TrimSpace(st.Output)
			st.Prompt = strings.TrimSpace(st.Prompt)
			st.Hint = strings.TrimSpace(st.Hint)
			st.Explanation = strings.TrimSpace(st.Explanation)
			st.Type = strings.TrimSpace(st.Type)
			if st.Prompt == "" {
				continue
			}
			if st.Type == "choice" {
				var choices []string
				for _, c := range st.Choices {
					if c = strings.TrimSpace(c); c != "" {
						choices = append(choices, c)
					}
				}
				if len(choices) < 2 {
					return nil, fmt.Errorf("lab step %q (type choice) needs at least 2 non-empty choices", st.Prompt)
				}
				if st.Answer < 0 || st.Answer >= len(choices) {
					return nil, fmt.Errorf("lab step %q (type choice) answer index out of range", st.Prompt)
				}
				st.Choices, st.Patterns = choices, nil
			} else {
				st.Type = "command"
				var pats []string
				for _, pat := range st.Patterns {
					if pat = strings.TrimSpace(pat); pat != "" {
						pats = append(pats, pat)
					}
				}
				if len(pats) == 0 {
					continue
				}
				for _, pat := range pats {
					if _, err := regexp.Compile(pat); err != nil {
						return nil, fmt.Errorf("lab step %q has invalid pattern %q: %w", st.Prompt, pat, err)
					}
				}
				st.Patterns, st.Choices = pats, nil
			}
			steps = append(steps, st)
		}
		if len(steps) == 0 {
			return nil, errors.New("lab card needs at least one step with a prompt and (patterns or choices)")
		}
		p.Steps = steps
		return json.Marshal(p)
	}
	return nil, fmt.Errorf("unknown card kind %q", kind)
}

func cleanSenses(in []Sense) []Sense {
	var out []Sense
	for _, s := range in {
		s.MeaningEn, s.MeaningVi = strings.TrimSpace(s.MeaningEn), strings.TrimSpace(s.MeaningVi)
		s.ExampleEn, s.ExampleVi = strings.TrimSpace(s.ExampleEn), strings.TrimSpace(s.ExampleVi)
		if s.MeaningEn == "" && s.MeaningVi == "" {
			continue
		}
		if s.Level < 1 || s.Level > 3 {
			s.Level = 1
		}
		out = append(out, s)
	}
	return out
}

// Compact is a flat, small view of a card used by the ESP32 device and by
// quiz generation. Only the simplest sense is included.
type Compact struct {
	ID        int64    `json:"id"`
	Kind      string   `json:"kind"`
	Front     string   `json:"front"`
	IPA       string   `json:"ipa,omitempty"`
	POS       string   `json:"pos,omitempty"`
	MeaningVi string   `json:"meaning_vi,omitempty"`
	MeaningEn string   `json:"meaning_en,omitempty"`
	ExampleEn string   `json:"example_en,omitempty"`
	ExampleVi string   `json:"example_vi,omitempty"`
	Back      string   `json:"back,omitempty"`
	Choices   []string `json:"choices,omitempty"`
	Answer    *int     `json:"answer,omitempty"`
	AudioURL  string   `json:"audio_url,omitempty"`
}

func ToCompact(c models.Card) Compact {
	out := Compact{ID: c.ID, Kind: c.Kind, Front: c.Front}
	switch c.Kind {
	case models.KindVocab:
		var p VocabPayload
		_ = json.Unmarshal(c.Payload, &p)
		out.IPA, out.POS, out.AudioURL = p.IPA, p.POS, p.AudioURL
		fillSense(&out, p.Senses)
	case models.KindPhrasal:
		var p PhrasalPayload
		_ = json.Unmarshal(c.Payload, &p)
		fillSense(&out, p.Senses)
	case models.KindConcept:
		var p ConceptPayload
		_ = json.Unmarshal(c.Payload, &p)
		out.Back, out.MeaningVi, out.ExampleEn, out.ExampleVi, out.POS = p.Back, p.NoteVi, p.Example, p.ExampleVi, p.Group
	case models.KindMCQ:
		var p MCQPayload
		_ = json.Unmarshal(c.Payload, &p)
		out.Choices = p.Choices
		a := p.Answer
		out.Answer = &a
		out.Back = p.Explanation
	case models.KindLab:
		var p LabPayload
		_ = json.Unmarshal(c.Payload, &p)
		out.Back = p.Scenario
	}
	return out
}

func fillSense(out *Compact, senses []Sense) {
	if len(senses) == 0 {
		return
	}
	best := senses[0]
	for _, s := range senses {
		if s.Level < best.Level {
			best = s
		}
	}
	out.MeaningVi, out.MeaningEn = best.MeaningVi, best.MeaningEn
	out.ExampleEn, out.ExampleVi = best.ExampleEn, best.ExampleVi
}

// Answer text used by auto-generated quiz questions ("what does X mean?").
func QuizAnswerText(c Compact) string {
	switch {
	case c.MeaningVi != "" && c.MeaningEn != "":
		return c.MeaningEn + " (" + c.MeaningVi + ")"
	case c.MeaningEn != "":
		return c.MeaningEn
	case c.MeaningVi != "":
		return c.MeaningVi
	default:
		return c.Back
	}
}

// QuizEnglishText is a quiz option in English plus its Vietnamese meaning as a
// separate hint (shown only on request). Cards without an English meaning fall
// back to QuizAnswerText with no hint.
func QuizEnglishText(c Compact) (text, hintVi string) {
	if c.MeaningEn != "" {
		return c.MeaningEn, c.MeaningVi
	}
	return QuizAnswerText(c), ""
}

// ChoiceText is the short answer shown in review multiple-choice questions:
// the Vietnamese meaning when available (clear for Vietnamese learners and
// short enough for the device screen), otherwise the English meaning/back.
func ChoiceText(c Compact) string {
	switch {
	case c.MeaningVi != "" && c.Kind != models.KindConcept:
		return c.MeaningVi
	case c.MeaningEn != "":
		return c.MeaningEn
	default:
		return c.Back
	}
}
