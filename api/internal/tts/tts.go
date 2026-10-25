// Package tts keeps English words read aloud for the study device.
//
// The audio is made once by the "tts" container (Microsoft Edge voices, see
// /tts) and stored in the tts_audio table as 16 kHz mono WAV (~20-40 KB a
// word). The device downloads it from /api/audio/<key>.wav. Rows that were not
// played for KeepDays are deleted by Cleanup; they are simply made again the
// next time the word is studied.
package tts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MaxText is the longest text that is read aloud (words and short phrases).
const MaxText = 120

type Service struct {
	DB       *gorm.DB
	BaseURL  string // http://tts:5050; empty = text-to-speech off
	Voice    string
	KeepDays int
	HTTP     *http.Client

	mu      sync.Mutex
	pending map[string]bool // keys being made in the background
	failAt  time.Time       // last time the tts service failed (back off for a while)
}

func New(db *gorm.DB, baseURL, voice string, keepDays int) *Service {
	return &Service{
		DB: db, BaseURL: strings.TrimRight(baseURL, "/"), Voice: voice, KeepDays: keepDays,
		HTTP:    &http.Client{Timeout: 25 * time.Second},
		pending: map[string]bool{},
	}
}

// Enabled reports whether audio can be made.
func (s *Service) Enabled() bool { return s != nil && s.BaseURL != "" }

// Normalize: the text as it is read (trimmed, single spaces). "" = not readable.
func Normalize(text string) string {
	t := strings.Join(strings.Fields(text), " ")
	if t == "" || len([]rune(t)) > MaxText {
		return ""
	}
	return t
}

// Key identifies the audio of a text in a voice (lower case: "Apple" = "apple").
func (s *Service) Key(text string) string {
	sum := sha256.Sum256([]byte(s.Voice + "|" + strings.ToLower(Normalize(text))))
	return hex.EncodeToString(sum[:16])
}

// URL is the path the device downloads the audio from.
func URL(key string) string { return "/api/audio/" + key + ".wav" }

// Ready returns the keys (of these texts) whose audio is already stored.
func (s *Service) Ready(texts []string) map[string]bool {
	out := map[string]bool{}
	keys := []string{}
	for _, t := range texts {
		if Normalize(t) != "" {
			keys = append(keys, s.Key(t))
		}
	}
	if len(keys) == 0 {
		return out
	}
	var have []string
	s.DB.Raw(`SELECT key FROM tts_audio WHERE key IN ?`, keys).Scan(&have)
	for _, k := range have {
		out[k] = true
	}
	return out
}

// Ensure makes the missing audio of these texts, a few at a time, until ctx
// ends; what is left continues in the background (ready at the next sync).
// It returns the keys that are stored.
func (s *Service) Ensure(ctx context.Context, texts []string) map[string]bool {
	ready := s.Ready(texts)
	if !s.Enabled() {
		return ready
	}
	var missing []string
	seen := map[string]bool{}
	for _, t := range texts {
		t = Normalize(t)
		k := s.Key(t)
		if t != "" && !ready[k] && !seen[k] {
			seen[k] = true
			missing = append(missing, t)
		}
	}
	if len(missing) == 0 || s.backingOff() {
		return ready
	}

	var mu sync.Mutex
	work := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range work {
				if err := s.make(ctx, t); err == nil {
					mu.Lock()
					ready[s.Key(t)] = true
					mu.Unlock()
				}
			}
		}()
	}
	left := []string{}
	for i, t := range missing {
		select {
		case work <- t:
			continue
		case <-ctx.Done():
			left = missing[i:]
		}
		break
	}
	close(work)
	wg.Wait()
	if len(left) > 0 {
		s.background(left)
	}
	return ready
}

// background makes audio after the request has been answered.
func (s *Service) background(texts []string) {
	s.mu.Lock()
	todo := []string{}
	for _, t := range texts {
		if k := s.Key(t); !s.pending[k] {
			s.pending[k] = true
			todo = append(todo, t)
		}
	}
	s.mu.Unlock()
	if len(todo) == 0 {
		return
	}
	go func() {
		for _, t := range todo {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := s.make(ctx, t)
			cancel()
			s.mu.Lock()
			delete(s.pending, s.Key(t))
			s.mu.Unlock()
			if err != nil && s.backingOff() {
				// the service is down: forget the rest, the next sync asks again
				s.mu.Lock()
				for _, r := range todo {
					delete(s.pending, s.Key(r))
				}
				s.mu.Unlock()
				return
			}
		}
	}()
}

func (s *Service) backingOff() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.failAt) < 2*time.Minute
}

// make asks the tts container for one text and stores the WAV.
func (s *Service) make(ctx context.Context, text string) error {
	q := url.Values{"text": {text}, "voice": {s.Voice}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/speak?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := s.HTTP.Do(req)
	if err != nil {
		s.failed(text, err)
		return err
	}
	defer resp.Body.Close()
	wav, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("tts: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(wav[:min(len(wav), 200)])))
		s.failed(text, err)
		return err
	}
	if len(wav) < 44 || string(wav[:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		return errors.New("tts: not a WAV file")
	}
	return s.DB.Exec(`INSERT INTO tts_audio (key, text, voice, wav, bytes) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (key) DO NOTHING`, s.Key(text), text, s.Voice, wav, len(wav)).Error
}

func (s *Service) failed(text string, err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	s.mu.Lock()
	first := time.Since(s.failAt) >= 2*time.Minute
	s.failAt = time.Now()
	s.mu.Unlock()
	if first {
		log.Printf("tts: %q: %v (retrying in 2 minutes)", text, err)
	}
}

// Test makes (or finds) the audio of a text now, ignoring the back-off, and
// returns its key: an admin check that the tts container works.
func (s *Service) Test(ctx context.Context, text string) (string, error) {
	text = Normalize(text)
	if text == "" {
		return "", errors.New("empty text")
	}
	if !s.Enabled() {
		return "", errors.New("text-to-speech is off (TTS_URL is empty)")
	}
	k := s.Key(text)
	if s.Ready([]string{text})[k] {
		return k, nil
	}
	if err := s.make(ctx, text); err != nil {
		return "", err
	}
	s.mu.Lock()
	s.failAt = time.Time{} // it works again: no need to wait
	s.mu.Unlock()
	return k, nil
}

// MakeFor makes the audio of text now if key is really its key (a word the
// device got before its audio was ready). false if it could not be made.
func (s *Service) MakeFor(ctx context.Context, key, text string) bool {
	text = Normalize(text)
	if !s.Enabled() || text == "" || s.Key(text) != key || s.backingOff() {
		return false
	}
	return s.make(ctx, text) == nil
}

// Get returns the stored WAV and marks it as used (at most once a day, to
// keep writes low).
func (s *Service) Get(key string) ([]byte, bool) {
	var row struct {
		Wav      []byte
		LastUsed time.Time `gorm:"column:last_used_at"`
	}
	if err := s.DB.Raw(`SELECT wav, last_used_at FROM tts_audio WHERE key = ?`, key).Scan(&row).Error; err != nil || len(row.Wav) == 0 {
		return nil, false
	}
	if time.Since(row.LastUsed) > 24*time.Hour {
		s.DB.Exec(`UPDATE tts_audio SET last_used_at = NOW() WHERE key = ?`, key)
	}
	return row.Wav, true
}

// Touch marks these texts' audio as used (the device was sent their URLs).
func (s *Service) Touch(keys []string) {
	if len(keys) > 0 {
		s.DB.Exec(`UPDATE tts_audio SET last_used_at = NOW() WHERE key IN ? AND last_used_at < NOW() - INTERVAL '1 day'`, keys)
	}
}

type Stats struct {
	Words    int64      `json:"words"`
	Bytes    int64      `json:"bytes"`
	Oldest   *time.Time `json:"oldest_used_at"`
	KeepDays int        `json:"keep_days"`
	Enabled  bool       `json:"enabled"`
	Voice    string     `json:"voice"`
}

func (s *Service) Stats() (Stats, error) {
	st := Stats{KeepDays: s.KeepDays, Enabled: s.Enabled(), Voice: s.Voice}
	err := s.DB.Raw(`SELECT COUNT(*) AS words, COALESCE(SUM(bytes), 0) AS bytes, MIN(last_used_at) AS oldest
		FROM tts_audio`).Row().Scan(&st.Words, &st.Bytes, &st.Oldest)
	return st, err
}

// Purge deletes audio not used for olderThan (0 = all). Returns the rows deleted.
func (s *Service) Purge(olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		res := s.DB.Exec(`DELETE FROM tts_audio`)
		return res.RowsAffected, res.Error
	}
	res := s.DB.Exec(`DELETE FROM tts_audio WHERE last_used_at < ?`, time.Now().Add(-olderThan))
	return res.RowsAffected, res.Error
}

// Cleanup deletes audio unused for KeepDays, once at start and then daily.
func (s *Service) Cleanup(ctx context.Context) {
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		if n, err := s.Purge(time.Duration(s.KeepDays) * 24 * time.Hour); err != nil {
			log.Printf("tts cleanup: %v", err)
		} else if n > 0 {
			log.Printf("tts cleanup: deleted %d unused words", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
