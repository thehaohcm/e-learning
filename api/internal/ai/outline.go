package ai

import (
	"context"
	"fmt"
	"strings"
)

// forTermsSchema is the per-card payload schema the model must follow when
// filling a specific list of terms, by deck kind.
func forTermsSchema(kind string) (string, error) {
	switch kind {
	case "vocab":
		return `{"front": "the exact word", "payload": {"ipa": "/.../", "pos": "noun|verb|...", "topic": "...",
      ` + senseSchema + `}}`, nil
	case "phrasal":
		return `{"front": "the exact phrasal verb", "payload": {"verb": "...", "particle": "...", "separable": true,
      ` + senseSchema + `}}`, nil
	case "concept":
		return `{"front": "the exact term", "payload": {"back": "short definition or formula (max ~120 chars)", "note_vi": "short Vietnamese explanation", "example": "optional short example"}}`, nil
	case "mcq":
		return `{"front": "an ORIGINAL multiple-choice question about the term", "payload": {"choices": ["A","B","C","D"], "answer": 0, "explanation": "why it is correct"}}`, nil
	default:
		return "", fmt.Errorf("cannot generate cards of kind %q", kind)
	}
}

// OutlineDeck asks the model for the full list of concept titles that cover the
// whole scope of a deck (a finite syllabus), excluding the ones that already
// exist. It returns at most maxTerms titles. Best for bounded subjects such as
// PMP or pentesting, where a deck has a knowable, finite set of concepts.
func (c *Client) OutlineDeck(ctx context.Context, model string, d DeckContext, maxTerms int, existing []string) ([]string, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	if maxTerms <= 0 {
		maxTerms = 150
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Subject: %s\nDeck: %s\nDeck description: %s\n\n", d.SubjectName, d.DeckName, d.Description)
	fmt.Fprintf(&b, `List the key %s that a learner must know to fully cover the scope of THIS deck — a comprehensive syllabus, not just a few examples.
- Return short titles only (the term/concept name, or for a question deck a short topic label), not definitions.
- Be exhaustive but stay strictly within this deck's topic; do not drift into other decks.
- Order them from foundational to advanced.
- Return at most %d items; if the real scope is smaller, return only what genuinely belongs.`, termWord(d.Kind), maxTerms)
	if len(existing) > 0 {
		fmt.Fprintf(&b, "\n\nThese already exist — do NOT repeat them: %s", strings.Join(existing, ", "))
	}
	b.WriteString(`

Return JSON: {"terms": ["title 1", "title 2", ...]}`)

	var out struct {
		Terms []string `json:"terms"`
	}
	system := "You are a curriculum expert. You lay out the complete, finite scope of a study topic. Reply with a single JSON object only."
	if err := c.chatJSONModel(ctx, model, system, b.String(), 0.3, &out); err != nil {
		return nil, err
	}
	// tidy: trim, drop empties and anything already present (case-insensitive)
	have := make(map[string]bool, len(existing))
	for _, e := range existing {
		have[strings.ToLower(strings.TrimSpace(e))] = true
	}
	seen := make(map[string]bool, len(out.Terms))
	terms := make([]string, 0, len(out.Terms))
	for _, t := range out.Terms {
		t = strings.TrimSpace(t)
		k := strings.ToLower(t)
		if t == "" || have[k] || seen[k] {
			continue
		}
		seen[k] = true
		terms = append(terms, t)
		if len(terms) >= maxTerms {
			break
		}
	}
	return terms, nil
}

func termWord(kind string) string {
	switch kind {
	case "vocab":
		return "words"
	case "phrasal":
		return "phrasal verbs"
	case "mcq":
		return "question topics"
	default:
		return "concepts / terms / formulas"
	}
}

// GenerateCardsFor builds one full card for each of the given terms. The caller
// validates each payload (content.Normalize) and skips duplicates.
func (c *Client) GenerateCardsFor(ctx context.Context, model string, d DeckContext, terms []string) ([]GeneratedCard, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	if len(terms) == 0 {
		return nil, nil
	}
	schema, err := forTermsSchema(d.Kind)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Subject: %s\nDeck: %s\nDeck description: %s\n\n", d.SubjectName, d.DeckName, d.Description)
	fmt.Fprintf(&b, `Create exactly one accurate flashcard for EACH of the following items, keeping the same order.
The "front" must correspond to the item. Each back/definition must be short (fits a 2.8-inch screen, max ~120 characters) and factually correct.

Items:
%s

Return JSON: {"cards": [
  %s
]}`, "- "+strings.Join(terms, "\n- "), schema)

	var out struct {
		Cards []GeneratedCard `json:"cards"`
	}
	system := "You create accurate study material for Vietnamese learners. Reply with a single JSON object only."
	if err := c.chatJSONModel(ctx, model, system, b.String(), 0.6, &out); err != nil {
		return nil, err
	}
	return out.Cards, nil
}
