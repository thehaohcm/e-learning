// Package dict looks up the Vietnamese meaning of an English word in the free
// dictionary at dict.minhqnd.com (https://github.com/minhqnd/dictionary, data
// CC BY-SA 4.0: the web shows the attribution). No API key is needed.
//
// It is used for a quick meaning while a word is typed in the notes panel and
// when a note is saved; the AI ("Explain") still writes the full card.
// Answers are cached in memory (found words 7 days, unknown words 1 day).
package dict

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Meaning is one sense: part of speech (English, e.g. "noun") and the Vietnamese meaning.
type Meaning struct {
	POS string `json:"pos,omitempty"`
	Vi  string `json:"vi"`
}

// Entry is what the web shows. Found = false: the word is not in the dictionary.
type Entry struct {
	Word     string    `json:"word"`
	Found    bool      `json:"found"`
	IPA      string    `json:"ipa,omitempty"`
	Meanings []Meaning `json:"meanings"`
}

// Short joins the first meanings into one line: "miễn cưỡng; bất đắc dĩ".
func (e Entry) Short(n int) string {
	parts := []string{}
	for _, m := range e.Meanings {
		if len(parts) == n {
			break
		}
		parts = append(parts, m.Vi)
	}
	return strings.Join(parts, "; ")
}

// FirstPOS is the part of speech of the first meaning ("" if none).
func (e Entry) FirstPOS() string {
	if len(e.Meanings) > 0 {
		return e.Meanings[0].POS
	}
	return ""
}

const (
	maxMeanings = 5
	maxCache    = 5000
	keepFound   = 7 * 24 * time.Hour
	keepMissing = 24 * time.Hour
)

type cached struct {
	e  Entry
	at time.Time
}

type Client struct {
	BaseURL string // https://dict.minhqnd.com; empty = off
	HTTP    *http.Client

	mu     sync.Mutex
	cache  map[string]cached
	failAt time.Time // the dictionary did not answer: wait a minute before asking again
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 6 * time.Second},
		cache:   map[string]cached{},
	}
}

func (c *Client) Enabled() bool { return c != nil && c.BaseURL != "" }

// Normalize: lower case, single spaces; "" if it cannot be looked up.
func Normalize(word string) string {
	w := strings.ToLower(strings.Join(strings.Fields(word), " "))
	if w == "" || len([]rune(w)) > 60 {
		return ""
	}
	return w
}

// Lookup returns the entry of an English word (cached).
func (c *Client) Lookup(ctx context.Context, word string) (Entry, error) {
	w := Normalize(word)
	if w == "" {
		return Entry{Word: word}, nil
	}
	if !c.Enabled() {
		return Entry{Word: w}, fmt.Errorf("dictionary is off")
	}
	c.mu.Lock()
	hit, ok := c.cache[w]
	down := time.Since(c.failAt) < time.Minute
	c.mu.Unlock()
	if ok {
		keep := keepMissing
		if hit.e.Found {
			keep = keepFound
		}
		if time.Since(hit.at) < keep {
			return hit.e, nil
		}
	}
	if down {
		return Entry{Word: w}, fmt.Errorf("dictionary unavailable")
	}

	e, err := c.fetch(ctx, w)
	if err != nil {
		if ctx.Err() == nil {
			c.mu.Lock()
			c.failAt = time.Now()
			c.mu.Unlock()
		}
		return Entry{Word: w}, err
	}
	c.mu.Lock()
	if len(c.cache) >= maxCache { // simple: drop everything older than a day, else all
		for k, v := range c.cache {
			if time.Since(v.at) > keepMissing {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= maxCache {
			c.cache = map[string]cached{}
		}
	}
	c.cache[w] = cached{e: e, at: time.Now()}
	c.mu.Unlock()
	return e, nil
}

// the dictionary's answer (only the fields used)
type apiResponse struct {
	Exists  bool   `json:"exists"`
	Word    string `json:"word"`
	Results []struct {
		LangCode string `json:"lang_code"`
		Meanings []struct {
			Definition     string `json:"definition"`
			DefinitionLang string `json:"definition_lang"`
			POS            string `json:"pos"`
		} `json:"meanings"`
		Pronunciations []struct {
			IPA    string `json:"ipa"`
			Region string `json:"region"`
		} `json:"pronunciations"`
		Translations []struct {
			LangCode    string `json:"lang_code"`
			Translation string `json:"translation"`
		} `json:"translations"`
	} `json:"results"`
}

func (c *Client) fetch(ctx context.Context, w string) (Entry, error) {
	q := url.Values{"word": {w}, "lang": {"en"}, "def_lang": {"vi"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/v1/lookup?"+q.Encode(), nil)
	if err != nil {
		return Entry{Word: w}, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Entry{Word: w}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Entry{Word: w}, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return Entry{Word: w}, nil // not in the dictionary
	}
	if resp.StatusCode != http.StatusOK {
		return Entry{Word: w}, fmt.Errorf("dictionary: HTTP %d", resp.StatusCode)
	}
	return Parse(w, body)
}

// Parse turns the dictionary's JSON into an Entry (exported for tests).
func Parse(w string, body []byte) (Entry, error) {
	var r apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return Entry{Word: w}, fmt.Errorf("dictionary: %w", err)
	}
	e := Entry{Word: w, Meanings: []Meaning{}}
	if !r.Exists {
		return e, nil
	}
	seen := map[string]bool{}
	add := func(pos, vi string) {
		vi = clean(vi)
		key := strings.ToLower(vi)
		if vi == "" || seen[key] || len(e.Meanings) >= maxMeanings {
			return
		}
		seen[key] = true
		e.Meanings = append(e.Meanings, Meaning{POS: posName(pos), Vi: vi})
	}
	for _, res := range r.Results {
		if res.LangCode != "" && res.LangCode != "en" {
			continue
		}
		if e.IPA == "" {
			for _, p := range res.Pronunciations {
				if ipa := strings.TrimSpace(p.IPA); ipa != "" {
					e.IPA = ipa
					break
				}
			}
		}
		for _, m := range res.Meanings {
			if m.DefinitionLang == "" || m.DefinitionLang == "vi" {
				add(m.POS, m.Definition)
			}
		}
		for _, t := range res.Translations {
			if t.LangCode == "vi" {
				add("", t.Translation)
			}
		}
	}
	if e.IPA != "" && !strings.HasPrefix(e.IPA, "/") && !strings.HasPrefix(e.IPA, "[") {
		e.IPA = "/" + e.IPA + "/"
	}
	e.Found = len(e.Meanings) > 0
	return e, nil
}

// clean: one line, no trailing dot, not too long.
func clean(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimRight(s, " .;,")
	if r := []rune(s); len(r) > 140 {
		s = strings.TrimSpace(string(r[:140])) + "…"
	}
	return s
}

// posName: the dictionary's Vietnamese part of speech in English, as the cards use.
func posName(p string) string {
	k := strings.ToLower(strings.TrimSpace(p))
	switch k {
	case "danh từ", "noun", "n":
		return "noun"
	case "động từ", "verb", "v":
		return "verb"
	case "tính từ", "adjective", "adj":
		return "adjective"
	case "phó từ", "trạng từ", "adverb", "adv":
		return "adverb"
	case "giới từ", "preposition", "prep":
		return "preposition"
	case "liên từ", "conjunction", "conj":
		return "conjunction"
	case "đại từ", "pronoun", "pron":
		return "pronoun"
	case "thán từ", "interjection":
		return "interjection"
	case "ngoại động từ", "nội động từ":
		return "verb"
	case "cụm từ", "thành ngữ", "phrase", "idiom":
		return "phrase"
	}
	return strings.TrimSpace(p)
}
