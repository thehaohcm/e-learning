package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PronWord is a word (or phrasal verb) to explain, with its IPA when known.
type PronWord struct {
	Word string
	IPA  string
}

// Pronunciation tells a Vietnamese learner how to say a word.
type Pronunciation struct {
	Syllables []string `json:"syllables"` // spelling split into syllables: ["hy","poth","e","sis"]
	Stress    int      `json:"stress"`    // index of the stressed syllable
	TipsVi    []string `json:"tips_vi"`   // 1-3 short tips in Vietnamese on the hard sounds
	// Sound: the vowel learners most often misread from the spelling (nil = none).
	Sound *SoundCheck `json:"sound,omitempty"`
}

// SoundCheck is a "Which sound?" question: how are Letters (in syllable
// Syllable) pronounced? Options[Answer] is right; the others are the sounds
// Vietnamese learners confuse it with.
type SoundCheck struct {
	Syllable int           `json:"syllable"`
	Letters  string        `json:"letters"`
	Options  []SoundOption `json:"options"`
	Answer   int           `json:"answer"`
	Same     []string      `json:"same"`    // common words where the same letters make the same sound
	NoteVi   string        `json:"note_vi"` // the spelling rule, one short sentence in Vietnamese
}

// SoundOption: a sound, its closest Vietnamese sound and a common example word.
type SoundOption struct {
	IPA     string `json:"ipa"`
	LikeVi  string `json:"like_vi"`
	Example string `json:"example"`
}

// GeneratePronunciations explains all words in ONE request. Results are in the
// same order; a result that does not pass the checks is nil.
func (c *Client) GeneratePronunciations(ctx context.Context, words []PronWord) ([]*Pronunciation, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	var b strings.Builder
	for i, w := range words {
		fmt.Fprintf(&b, "%d. %s", i+1, w.Word)
		if w.IPA != "" {
			fmt.Fprintf(&b, " %s", w.IPA)
		}
		b.WriteString("\n")
	}
	prompt := fmt.Sprintf(`Help a Vietnamese learner pronounce these English words (American English).
For each word:
- "syllables": the word's SPELLING split into spoken syllables, lower case, in order; joined together they must
  give back exactly the word's letters (for a phrase, split each word: "carry out" -> ["car","ry","out"]).
- "stress": the 0-based index of the syllable with the main stress.
- "tips_vi": 1-3 very short tips IN VIETNAMESE about the sounds Vietnamese speakers usually get wrong in THIS word
  (th /θ/ /ð/, final consonants and clusters like -s -ts -st -ks, weak vowels /ə/, long vs short vowels, /ʃ/ /ʒ/ /dʒ/,
  stress on an unexpected syllable...). Describe mouth position or compare with a Vietnamese sound when it helps.
  Never write the whole word with Vietnamese spelling.
- "sound": the ONE vowel (letter or letter group, e.g. "u", "ea", "ou", "i") of this word that Vietnamese learners
  most often misread from the spelling (e.g. "u" in "fluc" of fluctuate read /u/ "u" instead of /ʌ/ "â").
  {"syllable": index in "syllables", "letters": the letters as written in that syllable,
   "options": 3 sounds = the right one + the 2 sounds learners confuse it with, each
     {"ipa": "ʌ", "like_vi": closest Vietnamese sound ("â"), "example": a short common word with that sound ("cut")},
   "answer": index of the right option, "same": 2-4 common words where the same letters make the same sound,
   "note_vi": the spelling rule in ONE short Vietnamese sentence}.
  Use null when nothing in the word is tricky this way.

%s
Return JSON: {"items":[{"i":1,"syllables":["fluc","tu","ate"],"stress":0,"tips_vi":["..."],
 "sound":{"syllable":0,"letters":"u","options":[{"ipa":"u","like_vi":"u","example":"food"},{"ipa":"ʌ","like_vi":"â","example":"cut"},{"ipa":"juː","like_vi":"iu","example":"cute"}],
  "answer":1,"same":["luck","cut","structure"],"note_vi":"..."}}]}`, b.String())

	var out struct {
		Items []struct {
			I int `json:"i"`
			Pronunciation
		} `json:"items"`
	}
	if err := c.chatJSON(ctx, "You are an English pronunciation coach for Vietnamese learners. Reply with a single JSON object only.", prompt, 0.2, &out); err != nil {
		return nil, err
	}
	res := make([]*Pronunciation, len(words))
	for k, it := range out.Items {
		i := it.I - 1
		if i < 0 || i >= len(words) {
			i = k
		}
		if i < 0 || i >= len(words) {
			continue
		}
		p := it.Pronunciation
		if CleanPronunciation(&p, words[i].Word) {
			res[i] = &p
		}
	}
	return res, nil
}

// CleanPronunciation checks that the syllables spell the word and the stress
// index is valid, and tidies the tips. Returns false when unusable.
func CleanPronunciation(p *Pronunciation, word string) bool {
	letters := func(s string) string {
		var b strings.Builder
		for _, r := range strings.ToLower(s) {
			if unicode.IsLetter(r) {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	var syl []string
	for _, s := range p.Syllables {
		if s = strings.ToLower(strings.TrimSpace(s)); s != "" {
			syl = append(syl, s)
		}
	}
	if len(syl) == 0 || letters(strings.Join(syl, "")) != letters(word) {
		return false
	}
	if p.Stress < 0 || p.Stress >= len(syl) {
		p.Stress = 0
	}
	var tips []string
	for _, t := range p.TipsVi {
		if t = strings.TrimSpace(t); t != "" && len([]rune(t)) <= 200 {
			tips = append(tips, t)
		}
		if len(tips) == 3 {
			break
		}
	}
	p.Syllables, p.TipsVi = syl, tips
	if p.TipsVi == nil {
		p.TipsVi = []string{}
	}
	p.Sound = cleanSound(p.Sound, syl)
	return true
}

var plainWord = regexp.MustCompile(`^[A-Za-z][A-Za-z'-]*$`)

// cleanSound keeps a sound question only when it is consistent: the letters
// are in the syllable, 2-4 distinct sounds, a valid answer.
func cleanSound(sc *SoundCheck, syl []string) *SoundCheck {
	if sc == nil || sc.Syllable < 0 || sc.Syllable >= len(syl) {
		return nil
	}
	sc.Letters = strings.ToLower(strings.TrimSpace(sc.Letters))
	if sc.Letters == "" || !strings.Contains(syl[sc.Syllable], sc.Letters) {
		return nil
	}
	if sc.Answer < 0 || sc.Answer >= len(sc.Options) {
		return nil
	}
	seen := map[string]bool{}
	var opts []SoundOption
	answer := -1
	for i, o := range sc.Options {
		o.IPA = strings.Trim(strings.TrimSpace(o.IPA), "/[]")
		o.LikeVi = strings.Trim(strings.TrimSpace(o.LikeVi), `"“”`)
		o.Example = strings.TrimSpace(o.Example)
		if o.IPA == "" || seen[o.IPA] || len([]rune(o.LikeVi)) > 20 || !plainWord.MatchString(o.Example) {
			if i == sc.Answer {
				return nil
			}
			continue
		}
		seen[o.IPA] = true
		if i == sc.Answer {
			answer = len(opts)
		}
		opts = append(opts, o)
	}
	if answer < 0 || len(opts) < 2 || len(opts) > 4 {
		return nil
	}
	var same []string
	for _, w := range sc.Same {
		if w = strings.TrimSpace(w); plainWord.MatchString(w) && len(same) < 4 {
			same = append(same, w)
		}
	}
	sc.Options, sc.Answer, sc.Same = opts, answer, same
	if sc.Same == nil {
		sc.Same = []string{}
	}
	sc.NoteVi = strings.TrimSpace(sc.NoteVi)
	if len([]rune(sc.NoteVi)) > 220 {
		sc.NoteVi = ""
	}
	return sc
}
