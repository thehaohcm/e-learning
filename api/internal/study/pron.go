package study

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Pronunciation question modes (own schedule in user_pron).
const (
	ModeStress = "stress" // "Which syllable is stressed?" (tap one of the word's syllables)
	ModeSound  = "sound"  // "How do you say the “u” in FLUC?" (pick the sound)
)

// UserPron is the pronunciation schedule of one word for one user (FSRS,
// separate from the meaning schedule in user_cards).
type UserPron struct {
	UserID        int64      `gorm:"primaryKey;autoIncrement:false" json:"-"`
	CardID        int64      `gorm:"primaryKey;autoIncrement:false" json:"card_id"`
	State         int        `json:"state"`
	Stability     float64    `json:"stability"`
	Difficulty    float64    `json:"difficulty"`
	Due           time.Time  `json:"due"`
	LastReview    *time.Time `json:"last_review"`
	ScheduledDays int        `json:"scheduled_days"`
	Reps          int        `json:"reps"`
	Lapses        int        `json:"lapses"`
	CreatedAt     time.Time  `json:"-"`
	UpdatedAt     time.Time  `json:"-"`
}

func (UserPron) TableName() string { return "user_pron" }

// Pron sources (pron_logs.source).
const (
	PronStress = "stress" // tapped the stressed syllable
	PronForgot = "forgot" // "I forgot how to say it"
	PronSound  = "sound"  // picked the sound of a vowel
)

// PronResult is returned after a pronunciation answer.
type PronResult struct {
	Applied bool   `json:"applied"` // the pronunciation schedule changed
	NextIn  string `json:"next_in,omitempty"`
}

// ApplyPron records a pronunciation answer. A wrong answer (Again) always
// counts: the word comes back soon. A right answer counts only for a new word
// or one that is due, so answering the same word twice in a day does not
// stretch its interval.
func (s *Service) ApplyPron(userID, cardID int64, rating int, source string, now time.Time) (PronResult, error) {
	if rating != srs.Again && rating != srs.Good {
		return PronResult{}, errors.New("rating must be 1 or 3")
	}
	now = now.UTC()
	var out PronResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var p UserPron
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND card_id = ?", userID, cardID).First(&p).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			p = UserPron{UserID: userID, CardID: cardID, State: srs.StateNew, Due: now, CreatedAt: now}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		}
		if rating == srs.Good && p.State != srs.StateNew && p.Due.After(now) {
			return nil // already known and not due: nothing to change
		}
		at := now
		if p.LastReview != nil && at.Before(*p.LastReview) {
			at = *p.LastReview
		}
		res, err := srs.Review(srs.State{
			State: p.State, Stability: p.Stability, Difficulty: p.Difficulty, Due: p.Due, LastReview: p.LastReview,
			ScheduledDays: p.ScheduledDays, Reps: p.Reps, Lapses: p.Lapses,
		}, rating, at)
		if err != nil {
			return err
		}
		n := res.Next
		p.State, p.Stability, p.Difficulty, p.Due, p.LastReview = n.State, n.Stability, n.Difficulty, n.Due, n.LastReview
		p.ScheduledDays, p.Reps, p.Lapses, p.UpdatedAt = n.ScheduledDays, n.Reps, n.Lapses, now
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO pron_logs (user_id, card_id, rating, source, reviewed_at) VALUES (?, ?, ?, ?, ?)`,
			userID, cardID, rating, source, at).Error; err != nil {
			return err
		}
		out.Applied = true
		out.NextIn = srs.HumanInterval(p.Due.Sub(at))
		return nil
	})
	return out, err
}

// StressInfo is a word's syllables and stressed syllable.
type StressInfo struct {
	Syllables []string   `json:"syllables"`
	Stress    int        `json:"stress"`
	TipsVi    []string   `json:"tips_vi"`
	Sound     *SoundInfo `json:"sound,omitempty"`
}

// SoundInfo: how a tricky vowel of the word is said (same JSON as ai.SoundCheck).
type SoundInfo struct {
	Syllable int    `json:"syllable"`
	Letters  string `json:"letters"`
	Options  []struct {
		IPA     string `json:"ipa"`
		LikeVi  string `json:"like_vi"`
		Example string `json:"example"`
	} `json:"options"`
	Answer int      `json:"answer"`
	Same   []string `json:"same"`
	NoteVi string   `json:"note_vi"`
}

func (si *SoundInfo) usable(syllables []string) bool {
	return si != nil && si.Letters != "" && len(si.Options) >= 2 && si.Answer >= 0 && si.Answer < len(si.Options) &&
		si.Syllable >= 0 && si.Syllable < len(syllables)
}

// SoundFocus marks the letters a "Which sound?" question is about.
type SoundFocus struct {
	Syllable int    `json:"syllable"`
	Letters  string `json:"letters"`
}

// stressable: one word, at least two syllables, stress index in range.
func stressable(front string, syl []string, stress int) bool {
	return len(strings.Fields(front)) == 1 && len(syl) >= 2 && stress >= 0 && stress < len(syl)
}

// LoadStress returns the stored pronunciation of the given cards (stress
// questions only use words that have one; no AI call here).
func (s *Service) LoadStress(cardIDs []int64) (map[int64]StressInfo, error) {
	out := map[int64]StressInfo{}
	if len(cardIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		CardID    int64
		Syllables []byte
		Stress    int
		TipsVi    []byte
		Sound     []byte
	}
	if err := s.DB.Raw(`SELECT card_id, syllables, stress, tips_vi, sound FROM card_pronunciations WHERE card_id IN ?`, cardIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		v := StressInfo{Stress: r.Stress, Syllables: []string{}, TipsVi: []string{}}
		_ = json.Unmarshal(r.Syllables, &v.Syllables)
		_ = json.Unmarshal(r.TipsVi, &v.TipsVi)
		if len(r.Sound) > 2 {
			var si SoundInfo
			if json.Unmarshal(r.Sound, &si) == nil && si.usable(v.Syllables) {
				v.Sound = &si
			}
		}
		out[r.CardID] = v
	}
	return out, nil
}

// pronQuestion: a stress or a sound question for the word (random when both
// are possible), nil when neither is.
func pronQuestion(c models.Card, info StressInfo, ok bool) *Question {
	st := stressQuestion(c, info, ok)
	so := soundQuestion(c, info, ok)
	switch {
	case st != nil && so != nil:
		if rand.IntN(2) == 0 {
			return so
		}
		return st
	case st != nil:
		return st
	}
	return so
}

// soundQuestion builds "How do you say the “u” in FLUC?", nil without data.
func soundQuestion(c models.Card, info StressInfo, ok bool) *Question {
	if !ok || len(strings.Fields(c.Front)) != 1 || !info.Sound.usable(info.Syllables) {
		return nil
	}
	sd := info.Sound
	q := &Question{
		CardID: c.ID, Kind: c.Kind, Mode: ModeSound,
		Prompt:    fmt.Sprintf("How do you say the “%s” in %s?", sd.Letters, strings.ToUpper(info.Syllables[sd.Syllable])),
		Speak:     c.Front, // read only AFTER answering
		Syllables: info.Syllables,
		Focus:     &SoundFocus{Syllable: sd.Syllable, Letters: sd.Letters},
		Choices:   []Choice{},
	}
	for i, o := range sd.Options {
		text := "/" + o.IPA + "/"
		if o.Example != "" {
			text += " as in “" + o.Example + "”"
		}
		hint := ""
		if o.LikeVi != "" {
			hint = "gần “" + o.LikeVi + "”"
		}
		q.Choices = append(q.Choices, Choice{ID: int64(i), Text: text, Hint: hint})
	}
	return q
}

// StressQuestion builds "Which syllable is stressed?" for a card, or nil when
// the word has no stored pronunciation or only one syllable.
func stressQuestion(c models.Card, info StressInfo, ok bool) *Question {
	if !ok || !stressable(c.Front, info.Syllables, info.Stress) {
		return nil
	}
	return &Question{
		CardID: c.ID, Kind: c.Kind, Mode: ModeStress,
		Prompt:    "Which syllable is stressed?",
		Speak:     c.Front, // read only AFTER answering (it would give the answer away)
		Syllables: info.Syllables,
		Choices:   []Choice{},
	}
}

// StressCheck is the result of a stress answer.
type StressCheck struct {
	Correct bool       `json:"correct"`
	Word    string     `json:"word"`
	Info    StressInfo `json:"pronunciation"`
	PronResult
}

// CheckStress grades a tapped syllable and updates the pronunciation schedule.
func (s *Service) CheckStress(userID, cardID int64, syllable int, now time.Time) (StressCheck, error) {
	var c models.Card
	if err := s.DB.Where("deck_id IN ("+VisibleDecksSQL+")", userID).First(&c, cardID).Error; err != nil {
		return StressCheck{}, ErrCardNotFound
	}
	infos, err := s.LoadStress([]int64{c.ID})
	if err != nil {
		return StressCheck{}, err
	}
	info, ok := infos[c.ID]
	if !ok || !stressable(c.Front, info.Syllables, info.Stress) {
		return StressCheck{}, ErrCardNotFound
	}
	out := StressCheck{Correct: syllable == info.Stress, Word: c.Front, Info: info}
	rating := srs.Again
	if out.Correct {
		rating = srs.Good
	}
	out.PronResult, err = s.ApplyPron(userID, c.ID, rating, PronStress, now)
	return out, err
}

// SoundCheck is the result of a "Which sound?" answer.
type SoundCheck struct {
	Correct bool       `json:"correct"`
	Word    string     `json:"word"`
	Info    StressInfo `json:"pronunciation"`
	PronResult
}

// CheckSound grades a picked sound and updates the pronunciation schedule.
func (s *Service) CheckSound(userID, cardID int64, option int, now time.Time) (SoundCheck, error) {
	var c models.Card
	if err := s.DB.Where("deck_id IN ("+VisibleDecksSQL+")", userID).First(&c, cardID).Error; err != nil {
		return SoundCheck{}, ErrCardNotFound
	}
	infos, err := s.LoadStress([]int64{c.ID})
	if err != nil {
		return SoundCheck{}, err
	}
	info, ok := infos[c.ID]
	if !ok || !info.Sound.usable(info.Syllables) {
		return SoundCheck{}, ErrCardNotFound
	}
	out := SoundCheck{Correct: option == info.Sound.Answer, Word: c.Front, Info: info}
	rating := srs.Again
	if out.Correct {
		rating = srs.Good
	}
	out.PronResult, err = s.ApplyPron(userID, c.ID, rating, PronSound, now)
	return out, err
}

// ForgotPron: "I forgot how to say it" → the word's pronunciation comes back soon.
func (s *Service) ForgotPron(userID, cardID int64, now time.Time) (PronResult, error) {
	var n int64
	s.DB.Model(&models.Card{}).Where("id = ? AND deck_id IN ("+VisibleDecksSQL+")", cardID, userID).Count(&n)
	if n == 0 {
		return PronResult{}, ErrCardNotFound
	}
	return s.ApplyPron(userID, cardID, srs.Again, PronForgot, now)
}

// MaxStressShare: at most this many stress questions per 10 quiz questions.
const MaxStressShare = 3

// addStressQuestions turns some quiz questions into stress questions:
// first words whose pronunciation is due (forgotten earlier), then words of
// the quiz that have a stored pronunciation and were never stress-tested.
func (s *Service) addStressQuestions(userID int64, f Filter, cards []models.Card, qs []Question, now time.Time) ([]Question, error) {
	limit := max(1, len(qs)*MaxStressShare/10)
	if len(qs) == 0 {
		return qs, nil
	}

	// 1. pronunciation due (in this deck/subject), with a stored pronunciation
	var due []models.Card
	q := s.DB.Table("user_pron up").Select("c.*").
		Joins("JOIN cards c ON c.id = up.card_id").
		Joins("JOIN card_pronunciations cp ON cp.card_id = c.id").
		Where("up.user_id = ? AND up.due <= ? AND c.deck_id IN ("+VisibleDecksSQL+")", userID, now, userID)
	if err := scope(q, f).Order("up.due ASC").Limit(limit).Scan(&due).Error; err != nil {
		return qs, err
	}

	// 2. quiz words never stress-tested
	ids := make([]int64, 0, len(cards)+len(due))
	for _, c := range cards {
		ids = append(ids, c.ID)
	}
	for _, c := range due {
		ids = append(ids, c.ID)
	}
	infos, err := s.LoadStress(ids)
	if err != nil {
		return qs, err
	}
	var tested []int64
	s.DB.Raw(`SELECT card_id FROM user_pron WHERE user_id = ? AND card_id IN ?`, userID, ids).Scan(&tested)
	seenPron := map[int64]bool{}
	for _, id := range tested {
		seenPron[id] = true
	}

	var stress []Question
	used := map[int64]bool{}
	for _, c := range due {
		if len(stress) >= limit {
			break
		}
		info, ok := infos[c.ID]
		if sq := pronQuestion(c, info, ok); sq != nil && !used[c.ID] {
			stress = append(stress, *sq)
			used[c.ID] = true
		}
	}
	for _, c := range cards {
		if len(stress) >= limit {
			break
		}
		info, ok := infos[c.ID]
		if seenPron[c.ID] || used[c.ID] {
			continue
		}
		if sq := pronQuestion(c, info, ok); sq != nil {
			stress = append(stress, *sq)
			used[c.ID] = true
		}
	}
	if len(stress) == 0 {
		return qs, nil
	}

	// Replace: the quiz question of the same word, otherwise the last ones.
	out := make([]Question, 0, len(qs))
	for _, qq := range qs {
		if !used[qq.CardID] {
			out = append(out, qq)
		}
	}
	keep := len(qs) - len(stress)
	if keep < 0 {
		keep = 0
	}
	if len(out) > keep {
		out = out[:keep]
	}
	return interleave(out, stress), nil
}

// interleave spreads the stress questions evenly through the quiz.
func interleave(out, stress []Question) []Question {
	if len(stress) == 0 {
		return out
	}
	res := make([]Question, 0, len(out)+len(stress))
	step := max(1, (len(out)+len(stress))/len(stress))
	si := 0
	for i := 0; len(out) > 0 || si < len(stress); i++ {
		if si < len(stress) && (i%step == step-1 || len(out) == 0) {
			res = append(res, stress[si])
			si++
		} else if len(out) > 0 {
			res = append(res, out[0])
			out = out[1:]
		}
	}
	return res
}

// PronStatus: how many English words in the scope have no stored pronunciation
// yet (stress questions need one).
func (s *Service) PronMissing(userID int64, f Filter) (missing, total int64, err error) {
	q := s.DB.Table("cards c").
		Joins("LEFT JOIN card_pronunciations cp ON cp.card_id = c.id").
		Where("c.kind = ? AND c.deck_id IN ("+VisibleDecksSQL+") AND c.front NOT LIKE '% %'", models.KindVocab, userID)
	q = scope(q, f)
	err = q.Select("COUNT(*) FILTER (WHERE cp.card_id IS NULL OR cp.sound IS NULL), COUNT(*)").Row().Scan(&missing, &total)
	return
}
