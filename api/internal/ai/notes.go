package ai

import (
	"context"
	"fmt"
	"strings"
)

// NoteInput is a word/phrase the learner noted, with the sentence it came from.
type NoteInput struct {
	Text    string
	Context string
}

// NoteInfo describes a noted word, enough to create a vocabulary card.
type NoteInfo struct {
	Base      string `json:"base"` // dictionary form: "negotiated" -> "negotiate", "gave up" -> "give up"
	Kind      string `json:"kind"` // vocab | phrasal
	IPA       string `json:"ipa,omitempty"`
	POS       string `json:"pos,omitempty"`
	MeaningVi string `json:"meaning_vi"`
	MeaningEn string `json:"meaning_en,omitempty"`
	ExampleEn string `json:"example_en,omitempty"`
	ExampleVi string `json:"example_vi,omitempty"`
	// Answered: the model returned this item. Answered with an empty MeaningVi =
	// not an English word (typo, other language): the note can be dropped.
	Answered bool `json:"-"`
}

// FillNotes explains all noted words in ONE request. Results are in the same
// order as items; an item the model said it could not explain has Answered
// set and an empty MeaningVi; an item missing from the reply is not Answered.
func (c *Client) FillNotes(ctx context.Context, items []NoteInput) ([]NoteInfo, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var b strings.Builder
	for i, it := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, it.Text)
		if it.Context != "" {
			fmt.Fprintf(&b, "   Context: %s\n", truncate(it.Context, 300))
		}
	}
	prompt := fmt.Sprintf(`A Vietnamese learner (TOEIC/IELTS) noted these English words or phrases while studying.
For each one:
- "base": the dictionary form (negotiated -> negotiate, gave up -> give up, companies -> company).
- "kind": "phrasal" for a phrasal verb, otherwise "vocab".
- "ipa": US pronunciation in slashes; "pos": part of speech (noun, verb, adjective, phrasal verb...).
- "meaning_vi": short Vietnamese meaning, IN THE SENSE USED IN THE CONTEXT when a context is given.
- "meaning_en": a short simple English definition.
- "example_en": if a context sentence is given and uses the word, that sentence (one sentence only),
  otherwise a natural B1-B2 example; "example_vi": its natural Vietnamese translation.
If an item is not an English word/phrase (typo, Vietnamese...), set "meaning_vi" to "".

%s
Return JSON: {"items":[{"i":1,"base":"...","kind":"vocab","ipa":"/.../","pos":"...","meaning_vi":"...","meaning_en":"...","example_en":"...","example_vi":"..."}]}`, b.String())

	var out struct {
		Items []struct {
			I int `json:"i"`
			NoteInfo
		} `json:"items"`
	}
	if err := c.chatJSON(ctx, practiceSystem, prompt, 0.3, &out); err != nil {
		return nil, err
	}
	res := make([]NoteInfo, len(items))
	for k, it := range out.Items {
		i := it.I - 1
		if i < 0 || i >= len(items) {
			i = k
		}
		if i < 0 || i >= len(items) {
			continue
		}
		res[i] = cleanNoteInfo(it.NoteInfo, items[i].Text)
		res[i].Answered = true
	}
	return res, nil
}

func cleanNoteInfo(n NoteInfo, text string) NoteInfo {
	n.Base = strings.Join(strings.Fields(strings.TrimSpace(n.Base)), " ")
	if n.Base == "" || len(n.Base) > 60 {
		n.Base = text
	}
	if n.Kind != "phrasal" || len(strings.Fields(n.Base)) < 2 {
		n.Kind = "vocab"
	}
	n.MeaningVi = strings.TrimSpace(n.MeaningVi)
	n.MeaningEn = strings.TrimSpace(n.MeaningEn)
	n.ExampleEn = strings.TrimSpace(n.ExampleEn)
	n.ExampleVi = strings.TrimSpace(n.ExampleVi)
	n.IPA, n.POS = strings.TrimSpace(n.IPA), strings.TrimSpace(n.POS)
	return n
}
