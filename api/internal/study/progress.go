package study

import (
	"time"

	"mvp-lang-app/internal/content"
	"mvp-lang-app/internal/models"
	"mvp-lang-app/internal/srs"
)

// Progress: how well the cards in the user's list are known, in 5 groups that
// do not overlap (they add up to Total):
//
//   - new:        added, never studied yet;
//   - struggling: forgotten again and again (2+ lapses) or being relearnt now,
//     and not solid yet;
//   - learning:   studied, remembered for less than a week;
//   - familiar:   remembered for 1-3 weeks;
//   - mastered:   remembered for 3+ weeks (FSRS stability >= 21 days).
//
// "Remembered for N days" is the FSRS stability: the time after which the
// chance of still remembering the card falls to 90%.
type Progress struct {
	Buckets   Buckets           `json:"buckets"`
	Subjects  []SubjectProgress `json:"subjects"`
	Answers30 int               `json:"answers_30d"` // answers in the last 30 days
	Correct30 int               `json:"correct_30d"` // of them: Good / Easy (or right in a quiz)
	Weakest   []WeakCard        `json:"weakest"`     // most forgotten cards, worst first
	DueNow    int               `json:"due_now"`     // cards due to review right now (this topic)
}

type Buckets struct {
	Total      int `json:"total"`
	New        int `json:"new"`
	Struggling int `json:"struggling"`
	Learning   int `json:"learning"`
	Familiar   int `json:"familiar"`
	Mastered   int `json:"mastered"`
}

type SubjectProgress struct {
	SubjectID int64   `json:"subject_id"`
	Name      string  `json:"name"`
	Buckets   Buckets `json:"buckets"`
}

type WeakCard struct {
	CardID  int64   `json:"card_id"`
	Front   string  `json:"front"`
	Kind    string  `json:"kind"`
	Meaning string  `json:"meaning"`
	Lapses  int     `json:"lapses"`
	Reps    int     `json:"reps"`
	Days    float64 `json:"stability_days"`
}

// StrugglingLapses: forgotten this many times (after having learnt it) = struggling.
const StrugglingLapses = 2

// FamiliarStabilityDays: from here a card is "familiar" (below: still learning).
const FamiliarStabilityDays = 7.0

// bucketSQL names the group of a user_cards row "uc" (same order as the rules above).
const bucketSQL = `CASE
	WHEN uc.state = 0 THEN 'new'
	WHEN uc.stability >= @mastered THEN 'mastered'
	WHEN uc.state = 3 OR uc.lapses >= @lapses THEN 'struggling'
	WHEN uc.stability >= @familiar THEN 'familiar'
	ELSE 'learning' END`

func (b *Buckets) add(bucket string, n int) {
	b.Total += n
	switch bucket {
	case "new":
		b.New += n
	case "struggling":
		b.Struggling += n
	case "learning":
		b.Learning += n
	case "familiar":
		b.Familiar += n
	case "mastered":
		b.Mastered += n
	}
}

// Progress for the given topics (none = all).
func (s *Service) Progress(userID int64, now time.Time, subjectIDs ...int64) (Progress, error) {
	p := Progress{Subjects: []SubjectProgress{}, Weakest: []WeakCard{}}
	args := map[string]any{"user": userID, "mastered": srs.MasteredStabilityDays, "lapses": StrugglingLapses, "familiar": FamiliarStabilityDays,
		"subs": subjectIDs, "any": len(subjectIDs) == 0}
	// cards of the topics only
	const inSubjects = ` AND (@any OR uc.card_id IN (SELECT c2.id FROM cards c2 JOIN decks d2 ON d2.id = c2.deck_id WHERE d2.subject_id IN @subs))`

	var rows []struct {
		SubjectID int64
		Name      string
		Bucket    string
		N         int
	}
	if err := s.DB.Raw(`
		SELECT sub.id AS subject_id, sub.name, `+bucketSQL+` AS bucket, COUNT(*) AS n
		FROM user_cards uc
		JOIN cards c ON c.id = uc.card_id
		JOIN decks d ON d.id = c.deck_id
		JOIN subjects sub ON sub.id = d.subject_id
		WHERE uc.user_id = @user AND NOT uc.suspended`+inSubjects+`
		GROUP BY sub.id, sub.name, sub.sort_order, bucket
		ORDER BY sub.sort_order, sub.id`, args).Scan(&rows).Error; err != nil {
		return p, err
	}
	index := map[int64]int{}
	for _, r := range rows {
		i, ok := index[r.SubjectID]
		if !ok {
			i = len(p.Subjects)
			index[r.SubjectID] = i
			p.Subjects = append(p.Subjects, SubjectProgress{SubjectID: r.SubjectID, Name: r.Name})
		}
		p.Subjects[i].Buckets.add(r.Bucket, r.N)
		p.Buckets.add(r.Bucket, r.N)
	}

	args["since"] = now.AddDate(0, 0, -30)
	s.DB.Raw(`SELECT COUNT(*), COUNT(*) FILTER (WHERE rating >= 3)
	          FROM review_logs uc WHERE uc.user_id = @user AND uc.reviewed_at >= @since`+inSubjects, args).
		Row().Scan(&p.Answers30, &p.Correct30)

	var weak []struct {
		models.Card
		Lapses    int
		Reps      int
		Stability float64
	}
	if err := s.DB.Raw(`
		SELECT c.*, uc.lapses, uc.reps, uc.stability
		FROM user_cards uc JOIN cards c ON c.id = uc.card_id
		WHERE uc.user_id = @user AND NOT uc.suspended AND uc.state <> 0
		  AND (`+bucketSQL+`) = 'struggling'`+inSubjects+`
		ORDER BY uc.lapses DESC, uc.stability ASC, c.id
		LIMIT 12`, args).Scan(&weak).Error; err != nil {
		return p, err
	}
	for _, w := range weak {
		p.Weakest = append(p.Weakest, WeakCard{
			CardID: w.ID, Front: w.Front, Kind: w.Kind, Meaning: content.ChoiceText(content.ToCompact(w.Card)),
			Lapses: w.Lapses, Reps: w.Reps, Days: w.Stability,
		})
	}

	args["now"] = now
	s.DB.Raw(`SELECT COUNT(*) FROM user_cards uc
	          WHERE uc.user_id = @user AND NOT uc.suspended AND uc.state <> 0 AND uc.due <= @now`+inSubjects, args).
		Row().Scan(&p.DueNow)
	return p, nil
}
