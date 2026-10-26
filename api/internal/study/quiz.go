package study

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
)

// Quiz question modes.
const (
	ModeMCQ     = "mcq"     // card is itself a multiple-choice question
	ModeMeaning = "meaning" // "What does X mean?" -> pick the meaning
	ModeTerm    = "term"    // "Which word means ...?" -> pick the word
)

type Choice struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
	// Hint: the Vietnamese meaning, shown only when the learner asks for it
	// (the "?" next to the option). Quiz options themselves are English.
	Hint string `json:"hint,omitempty"`
}

type Question struct {
	CardID int64  `json:"card_id"`
	Kind   string `json:"kind"`
	Mode   string `json:"mode"`
	Prompt string `json:"prompt"`
	// PromptHint: Vietnamese of the meaning in a "Which one means: …" prompt.
	PromptHint string   `json:"prompt_hint,omitempty"`
	Speak      string   `json:"speak,omitempty"` // text the client can read aloud (listening practice)
	Choices    []Choice `json:"choices"`
	// Syllables of the word, for pronunciation questions (modes "stress", "sound").
	Syllables []string `json:"syllables,omitempty"`
	// Focus: the letters a "sound" question is about.
	Focus *SoundFocus `json:"focus,omitempty"`
}

type QuizResult struct {
	Correct       bool            `json:"correct"`
	CorrectChoice int64           `json:"correct_choice_id"`
	Explanation   string          `json:"explanation,omitempty"`
	Card          content.Compact `json:"card"`
	Scheduled     bool            `json:"scheduled"` // true if this answer updated the SRS schedule
}

// BuildQuiz picks `count` cards (preferring cards the user is learning) and
// turns each into a 4-choice question. Distractors come from the same deck.
func (s *Service) BuildQuiz(userID int64, f Filter, count int) ([]Question, error) {
	if count <= 0 || count > 50 {
		count = 10
	}
	var cards []models.Card
	q := s.DB.Table("cards c").Select("c.*").
		Joins("LEFT JOIN user_cards uc ON uc.card_id = c.id AND uc.user_id = ?", userID).
		Where("c.deck_id IN ("+VisibleDecksSQL+")", userID)
	if err := scope(q, f).
		Order("(uc.card_id IS NULL), (uc.state = 0), random()").
		Limit(count).Scan(&cards).Error; err != nil {
		return nil, err
	}

	out := make([]Question, 0, len(cards))
	for _, c := range cards {
		qq, err := s.question(c)
		if err != nil {
			return nil, err
		}
		if qq != nil {
			out = append(out, *qq)
		}
	}
	// some questions ask where the stress is (pronunciation has its own schedule)
	return s.addStressQuestions(userID, f, cards, out, time.Now())
}

func (s *Service) question(c models.Card) (*Question, error) {
	if c.Kind == models.KindMCQ {
		var p content.MCQPayload
		if json.Unmarshal(c.Payload, &p) != nil || len(p.Choices) < 2 {
			return nil, nil
		}
		choices := make([]Choice, len(p.Choices))
		for i, t := range p.Choices {
			choices[i] = Choice{ID: int64(i), Text: t}
		}
		rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		return &Question{CardID: c.ID, Kind: c.Kind, Mode: ModeMCQ, Prompt: c.Front, Choices: choices}, nil
	}

	others, err := s.distractors(c, 3)
	if err != nil {
		return nil, err
	}
	if len(others) < 3 {
		return nil, nil // not enough cards in the deck for 4 choices
	}

	me := content.ToCompact(c)
	mode := ModeMeaning
	if rand.IntN(3) == 0 && content.QuizAnswerText(me) != "" && me.MeaningEn != "" {
		mode = ModeTerm
	}
	qq := &Question{CardID: c.ID, Kind: c.Kind, Mode: mode}
	all := append([]models.Card{c}, others...)
	for _, o := range all {
		oc := content.ToCompact(o)
		text, hint := content.QuizEnglishText(oc)
		if mode == ModeTerm {
			// options are the English words: their Vietnamese would give the answer away
			text, hint = o.Front, ""
		}
		qq.Choices = append(qq.Choices, Choice{ID: o.ID, Text: text, Hint: hint})
	}
	rand.Shuffle(len(qq.Choices), func(i, j int) { qq.Choices[i], qq.Choices[j] = qq.Choices[j], qq.Choices[i] })

	if mode == ModeMeaning {
		qq.Prompt = fmt.Sprintf("What does “%s” mean?", c.Front)
		qq.Speak = c.Front
	} else {
		en, vi := content.QuizEnglishText(me)
		qq.Prompt = fmt.Sprintf("Which one means: %s", en)
		qq.PromptHint = vi
	}
	return qq, nil
}

// CheckQuizAnswer grades an answer. If the card is in the user's list and is
// due, the result also counts as a review (correct = Good, wrong = Again).
func (s *Service) CheckQuizAnswer(userID, cardID, choiceID int64) (QuizResult, error) {
	var c models.Card
	if err := s.DB.Where("deck_id IN ("+VisibleDecksSQL+")", userID).First(&c, cardID).Error; err != nil {
		return QuizResult{}, ErrCardNotFound
	}
	res := QuizResult{Card: content.ToCompact(c)}
	if c.Kind == models.KindMCQ {
		var p content.MCQPayload
		_ = json.Unmarshal(c.Payload, &p)
		res.CorrectChoice = int64(p.Answer)
		res.Explanation = p.Explanation
	} else {
		res.CorrectChoice = c.ID
	}
	res.Correct = choiceID == res.CorrectChoice

	rating := srs.Again
	if res.Correct {
		rating = srs.Good
	}
	ar, err := s.ApplyAnswer(Answer{UserID: userID, CardID: cardID, Rating: rating, At: time.Now(), Source: "quiz", OnlyIfDue: true})
	if err != nil {
		return res, err
	}
	res.Scheduled = ar.Applied
	return res, nil
}

// MeaningChoices returns 4 shuffled choices for "what does <front> mean?":
// the card's own meaning (ID = card.ID) and 3 different meanings from other
// cards of the same deck and kind. Returns nil if the deck is too small.
func (s *Service) MeaningChoices(c models.Card) ([]Choice, error) {
	own := content.ChoiceText(content.ToCompact(c))
	if own == "" {
		return nil, nil
	}
	others, err := s.distractors(c, 12)
	if err != nil {
		return nil, err
	}
	// wrong options of the same part of speech first (a verb among verbs):
	// harder to find the answer by ruling the others out
	others = samePOSFirst(content.ToCompact(c).POS, others)
	seen := map[string]bool{strings.ToLower(own): true}
	choices := []Choice{{ID: c.ID, Text: own}}
	for _, o := range others {
		t := content.ChoiceText(content.ToCompact(o))
		if t == "" || seen[strings.ToLower(t)] {
			continue // skip empty and duplicate meanings so only one answer fits
		}
		seen[strings.ToLower(t)] = true
		choices = append(choices, Choice{ID: o.ID, Text: t})
		if len(choices) == 4 {
			break
		}
	}
	if len(choices) < 4 {
		return nil, nil
	}
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	return choices, nil
}

// distractors returns up to n other cards of the same kind for wrong choices:
// from the same deck first, then (small decks, e.g. "My words") from the shared
// decks of the same subject. Never from another user's personal deck.
func (s *Service) distractors(c models.Card, n int) ([]models.Card, error) {
	var out []models.Card
	// writing phrases: frames of the same group first ("Yours sincerely" among other
	// ways to end a letter), so the answer cannot be found by its function alone
	if g := content.ToCompact(c).POS; c.Kind == models.KindConcept && g != "" {
		if err := s.DB.Where("deck_id = ? AND kind = ? AND id <> ? AND payload->>'group' = ?", c.DeckID, c.Kind, c.ID, g).
			Order("random()").Limit(n).Find(&out).Error; err != nil {
			return nil, err
		}
	}
	if missing := n - len(out); missing > 0 {
		var rest []models.Card
		q := s.DB.Where("deck_id = ? AND kind = ? AND id <> ?", c.DeckID, c.Kind, c.ID)
		if len(out) > 0 {
			ids := make([]int64, len(out))
			for i, o := range out {
				ids[i] = o.ID
			}
			q = q.Where("id NOT IN ?", ids)
		}
		if err := q.Order("random()").Limit(missing).Find(&rest).Error; err != nil {
			return nil, err
		}
		out = append(out, rest...)
	}
	if missing := n - len(out); missing > 0 {
		var more []models.Card
		if err := s.DB.Where(`kind = ? AND id <> ? AND deck_id <> ? AND deck_id IN (
				SELECT d.id FROM decks d WHERE d.owner_id IS NULL
				  AND d.subject_id = (SELECT subject_id FROM decks WHERE id = ?))`, c.Kind, c.ID, c.DeckID, c.DeckID).
			Order("random()").Limit(missing).Find(&more).Error; err != nil {
			return nil, err
		}
		out = append(out, more...)
	}
	return out, nil
}

// samePOSFirst moves the cards with the same part of speech as pos to the front
// (keeping the random order inside each group).
func samePOSFirst(pos string, cards []models.Card) []models.Card {
	pos = strings.ToLower(strings.TrimSpace(pos))
	if pos == "" {
		return cards
	}
	out := make([]models.Card, 0, len(cards))
	var rest []models.Card
	for _, o := range cards {
		if strings.ToLower(strings.TrimSpace(content.ToCompact(o).POS)) == pos {
			out = append(out, o)
		} else {
			rest = append(rest, o)
		}
	}
	return append(out, rest...)
}
