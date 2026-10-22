package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// GeneratedCard is one AI-generated card; Payload follows content.*Payload.
type GeneratedCard struct {
	Front   string          `json:"front"`
	Payload json.RawMessage `json:"payload"`
}

// DeckContext tells the model what the deck is about.
type DeckContext struct {
	SubjectName string
	DeckName    string
	Description string
	Kind        string
}

const senseSchema = `"senses": [
        {"level": 1, "meaning_en": "...", "meaning_vi": "...", "example_en": "simple everyday sentence", "example_vi": "...", "usage_note": "..."},
        {"level": 2, "meaning_en": "...", "meaning_vi": "...", "example_en": "natural work/daily-life sentence", "example_vi": "...", "usage_note": "..."},
        {"level": 3, "meaning_en": "...", "meaning_vi": "...", "example_en": "formal/academic sentence", "example_vi": "...", "usage_note": "..."}
      ]`

func generationPrompt(d DeckContext, count int, avoid []string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Subject: %s\nDeck: %s\nDeck description: %s\n\n", d.SubjectName, d.DeckName, d.Description)

	switch d.Kind {
	case "vocab":
		fmt.Fprintf(&b, `Generate exactly %d English words useful for TOEIC/IELTS at B1-C1 level for Vietnamese learners.
Prefer words used in work and daily life. Examples must be short, natural and easy to understand.
Return JSON: {"cards": [
  {"front": "word", "payload": {"ipa": "/.../", "pos": "noun|verb|adjective|...", "topic": "business",
      %s}}
]}`, count, senseSchema)
	case "phrasal":
		fmt.Fprintf(&b, `Generate exactly %d common English phrasal verbs (B1-C1) for Vietnamese learners.
Return JSON: {"cards": [
  {"front": "give up", "payload": {"verb": "give", "particle": "up", "separable": true,
      %s}}
]}`, count, senseSchema)
	case "concept":
		fmt.Fprintf(&b, `Generate exactly %d flashcards of key terms, definitions or formulas for this deck.
Each back side must be short (fits a 2.8-inch screen, max ~120 characters) and factually correct.
note_vi is a short Vietnamese explanation.
Return JSON: {"cards": [
  {"front": "term", "payload": {"back": "short definition or formula", "note_vi": "...", "example": "optional short example"}}
]}`, count)
	case "mcq":
		fmt.Fprintf(&b, `Write exactly %d ORIGINAL multiple-choice practice questions for this deck.
Do not copy questions from real exams or published question banks.
Each question has 4 choices, exactly one correct answer, and an explanation of why it is correct.
Return JSON: {"cards": [
  {"front": "question text", "payload": {"choices": ["A", "B", "C", "D"], "answer": 0, "explanation": "..."}}
]}`, count)
	default:
		return "", fmt.Errorf("cannot generate cards of kind %q", d.Kind)
	}

	if len(avoid) > 0 {
		fmt.Fprintf(&b, "\n\nDo NOT include any of these existing items: %s", strings.Join(avoid, ", "))
	}
	return b.String(), nil
}

// GenerateCards asks the model for new cards for a deck. The caller validates
// each payload (content.Normalize) and skips duplicates.
func (c *Client) GenerateCards(ctx context.Context, d DeckContext, count int, avoid []string) ([]GeneratedCard, error) {
	prompt, err := generationPrompt(d, count, avoid)
	if err != nil {
		return nil, err
	}
	var out struct {
		Cards []GeneratedCard `json:"cards"`
	}
	system := "You create accurate study material for Vietnamese learners. Reply with a single JSON object only."
	if err := c.chatJSON(ctx, system, prompt, 0.8, &out); err != nil {
		return nil, err
	}
	return out.Cards, nil
}
