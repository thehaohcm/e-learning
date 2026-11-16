package worker

import (
	"context"
	"log"
	"slices"
	"time"

	"mvp-lang-app/internal/push"
	"mvp-lang-app/internal/study"

	"gorm.io/gorm"
)

// StartReminders sends study reminders (Web Push) at the times each user chose
// in settings.push_times, in the user's time zone. It checks every minute.
func StartReminders(ctx context.Context, db *gorm.DB, store *push.Store) {
	if store == nil {
		return
	}
	// start just after the next full minute
	now := time.Now()
	select {
	case <-ctx.Done():
		return
	case <-time.After(now.Truncate(time.Minute).Add(time.Minute + 2*time.Second).Sub(now)):
	}
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		SendDueReminders(ctx, db, store, time.Now())
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// RecentGap: no reminder for someone who answered a card (web or device) this recently.
const RecentGap = time.Hour

// RecentlyStudied reports whether the user studied within RecentGap of now.
func RecentlyStudied(st study.Stats, now time.Time) bool {
	return st.LastReviewAt != nil && now.Sub(*st.LastReviewAt) < RecentGap
}

// SendDueReminders handles reminder slots at now and one minute before (a late
// tick never skips a reminder; push_reminders_sent prevents doubles).
func SendDueReminders(ctx context.Context, db *gorm.DB, store *push.Store, now time.Time) {
	var users []struct {
		ID       int64
		Settings []byte
		Timezone string
	}
	if err := db.Raw(`SELECT u.id, u.settings, u.timezone FROM users u
	                  WHERE EXISTS (SELECT 1 FROM push_subscriptions s WHERE s.user_id = u.id)`).Scan(&users).Error; err != nil {
		log.Printf("[reminders] %v", err)
		return
	}
	svc := &study.Service{DB: db}
	for _, u := range users {
		settings := study.ParseSettings(u.Settings)
		if len(settings.PushTimes) == 0 && !settings.InactiveReminder {
			continue
		}
		loc := study.UserLocation(u.Timezone)
		for _, t := range []time.Time{now.Add(-time.Minute), now} {
			local := t.In(loc)
			// Inactivity nudge: once at the chosen hour if nothing studied today.
			maybeInactiveReminder(ctx, db, store, u.ID, settings, loc, local)
			if len(settings.PushTimes) == 0 || !slices.Contains(settings.PushTimes, local.Format("15:04")) {
				continue
			}
			slot := local.Format("2006-01-02 15:04")
			res := db.Exec(`INSERT INTO push_reminders_sent (user_id, slot) VALUES (?, ?) ON CONFLICT DO NOTHING`, u.ID, slot)
			if res.Error != nil || res.RowsAffected == 0 {
				continue // already handled
			}
			st, err := svc.Stats(u.ID, now)
			if err != nil {
				log.Printf("[reminders] stats user %d: %v", u.ID, err)
				continue
			}
			if RecentlyStudied(st, now) {
				log.Printf("[reminders] user %d slot %s: studied in the last %v, skipped", u.ID, slot, RecentGap)
				continue
			}
			msg, ok := push.StudyReminder(st)
			if !ok {
				continue // nothing to study: no notification
			}
			sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			n, err := store.SendToUser(sendCtx, u.ID, msg, "")
			cancel()
			if err != nil {
				log.Printf("[reminders] user %d: %v", u.ID, err)
			} else {
				log.Printf("[reminders] user %d slot %s: sent to %d device(s)", u.ID, slot, n)
			}
		}
	}
}

// maybeInactiveReminder sends one "you haven't studied today" push at the
// user's chosen hour, if the inactivity reminder is on and they have not
// studied at all since local midnight. push_reminders_sent keeps it to once.
func maybeInactiveReminder(ctx context.Context, db *gorm.DB, store *push.Store, userID int64, settings study.Settings, loc *time.Location, local time.Time) {
	if !settings.InactiveReminder || local.Minute() != 0 || local.Hour() != settings.InactiveReminderHour {
		return
	}
	slot := "inactive " + local.Format("2006-01-02")
	res := db.Exec(`INSERT INTO push_reminders_sent (user_id, slot) VALUES (?, ?) ON CONFLICT DO NOTHING`, userID, slot)
	if res.Error != nil || res.RowsAffected == 0 {
		return // already handled today
	}
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).UTC()
	if studiedSince(db, userID, midnight) {
		log.Printf("[reminders] inactive user %d: already studied today, skipped", userID)
		return
	}
	msg := push.Message{
		Title: "Hôm nay bạn chưa học 📚",
		Body:  "Dành vài phút ôn tập để giữ nhịp học của bạn nhé!",
		URL:   "/",
		Tag:   "inactive-reminder",
	}
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	n, err := store.SendToUser(sendCtx, userID, msg, "")
	cancel()
	if err != nil {
		log.Printf("[reminders] inactive user %d: %v", userID, err)
	} else {
		log.Printf("[reminders] inactive user %d: sent to %d device(s)", userID, n)
	}
}

// studiedSince reports whether the user has any study activity (an SRS review
// or a submitted daily set) at or after the given UTC time.
func studiedSince(db *gorm.DB, userID int64, since time.Time) bool {
	var n int64
	db.Raw(`SELECT
		(SELECT COUNT(*) FROM review_logs WHERE user_id = ? AND reviewed_at >= ?) +
		(SELECT COUNT(*) FROM pmp_daily_sets WHERE user_id = ? AND submitted_at >= ?)`,
		userID, since, userID, since).Row().Scan(&n)
	return n > 0
}
