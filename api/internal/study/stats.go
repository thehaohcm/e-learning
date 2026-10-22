package study

import (
	"database/sql"
	"time"

	"mvp-lang-app/internal/srs"
)

type SubjectStats struct {
	SubjectID int64  `json:"subject_id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Total     int    `json:"total_cards"`
	Learning  int    `json:"learning"` // added to the user's list
	Due       int    `json:"due"`
	Mastered  int    `json:"mastered"`
}

type Stats struct {
	DueNow        int            `json:"due_now"`
	NewLeftToday  int            `json:"new_left_today"`
	NewToday      int            `json:"new_today"`       // new cards started today
	NewLimitToday int            `json:"new_limit_today"` // daily setting + extra added for today
	NextDueAt     *time.Time     `json:"next_due_at"`     // next card coming back (any deck), nil if none
	ReviewsToday  int            `json:"reviews_today"`
	CorrectToday  int            `json:"correct_today"`
	StreakDays    int            `json:"streak_days"`
	Learning      int            `json:"learning"`
	Mastered      int            `json:"mastered"`
	Subjects      []SubjectStats `json:"subjects"`
	ReviewsLast14 []DayCount     `json:"reviews_last_14_days"`
}

type DayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

func (s *Service) Stats(userID int64, now time.Time) (Stats, error) {
	st := Stats{Subjects: []SubjectStats{}}
	u, settings, err := s.loadUser(userID)
	if err != nil {
		return st, err
	}
	loc := UserLocation(u.Timezone)
	today := startOfDay(now, loc)

	if err := s.DB.Raw(`
		SELECT sub.id AS subject_id, sub.slug, sub.name,
		       COUNT(c.id) AS total,
		       COUNT(uc.card_id) FILTER (WHERE NOT uc.suspended) AS learning,
		       COUNT(uc.card_id) FILTER (WHERE NOT uc.suspended AND uc.state <> 0 AND uc.due <= ?) AS due,
		       COUNT(uc.card_id) FILTER (WHERE NOT uc.suspended AND uc.stability >= ?) AS mastered
		FROM subjects sub
		JOIN decks d ON d.subject_id = sub.id
		LEFT JOIN cards c ON c.deck_id = d.id
		LEFT JOIN user_cards uc ON uc.card_id = c.id AND uc.user_id = ?
		GROUP BY sub.id ORDER BY sub.sort_order, sub.id`,
		now, srs.MasteredStabilityDays, userID).Scan(&st.Subjects).Error; err != nil {
		return st, err
	}
	for _, x := range st.Subjects {
		st.DueNow += x.Due
		st.Learning += x.Learning
		st.Mastered += x.Mastered
	}

	introduced, err := s.NewIntroducedToday(userID, loc, now)
	if err != nil {
		return st, err
	}
	st.NewToday = introduced
	st.NewLimitToday = settings.NewLimit(dayKey(now, loc))
	st.NewLeftToday = max(0, st.NewLimitToday-introduced)

	var next sql.NullTime
	_ = s.DB.Raw(`SELECT MIN(due) FROM user_cards WHERE user_id = ? AND NOT suspended AND state <> 0 AND due > ?`, userID, now).
		Row().Scan(&next)
	if next.Valid {
		t := next.Time.UTC()
		st.NextDueAt = &t
	}

	s.DB.Raw(`SELECT COUNT(*) AS n, COUNT(*) FILTER (WHERE rating >= 3) AS ok
	          FROM review_logs WHERE user_id = ? AND reviewed_at >= ?`, userID, today).
		Row().Scan(&st.ReviewsToday, &st.CorrectToday)

	// Days (in the user's time zone) with at least one review, most recent first.
	var days []time.Time
	if err := s.DB.Raw(`
		SELECT DISTINCT (reviewed_at AT TIME ZONE ?)::date AS d
		FROM review_logs WHERE user_id = ? AND reviewed_at >= ?
		ORDER BY d DESC`, loc.String(), userID, today.AddDate(0, 0, -400)).Scan(&days).Error; err != nil {
		return st, err
	}
	st.StreakDays = streak(days, today)

	var perDay []struct {
		D time.Time
		N int
	}
	s.DB.Raw(`
		SELECT (reviewed_at AT TIME ZONE ?)::date AS d, COUNT(*) AS n
		FROM review_logs WHERE user_id = ? AND reviewed_at >= ?
		GROUP BY d`, loc.String(), userID, today.AddDate(0, 0, -13)).Scan(&perDay)
	counts := map[string]int{}
	for _, p := range perDay {
		counts[p.D.Format("2006-01-02")] = p.N
	}
	for i := 13; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("2006-01-02")
		st.ReviewsLast14 = append(st.ReviewsLast14, DayCount{Day: d, Count: counts[d]})
	}
	return st, nil
}

// streak counts consecutive days ending today (or yesterday, so the streak is
// not lost before the user has studied today).
func streak(days []time.Time, today time.Time) int {
	if len(days) == 0 {
		return 0
	}
	key := func(t time.Time) string { return t.Format("2006-01-02") }
	set := map[string]bool{}
	for _, d := range days {
		set[key(d)] = true
	}
	cur := today
	if !set[key(cur)] {
		cur = cur.AddDate(0, 0, -1)
	}
	n := 0
	for set[key(cur)] {
		n++
		cur = cur.AddDate(0, 0, -1)
	}
	return n
}
