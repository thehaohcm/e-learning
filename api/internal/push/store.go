package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"mvp-lang-app/internal/study"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Message is the JSON the service worker (web/public/sw.js) shows.
type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url,omitempty"` // opened when the notification is tapped
	Tag   string `json:"tag,omitempty"` // a newer notification with the same tag replaces the old one
}

// Store keeps subscriptions in the database and sends to them.
type Store struct {
	DB     *gorm.DB
	Sender *Sender
}

const vapidSecretName = "vapid_private_key"

// NewStore loads the VAPID key from app_secrets, creating it on first use.
func NewStore(db *gorm.DB, subject string, anyHost bool) (*Store, error) {
	key, err := loadOrCreateKey(db)
	if err != nil {
		return nil, err
	}
	v, err := ParseVAPID(key, subject)
	if err != nil {
		return nil, err
	}
	allow := KnownPushHost
	if anyHost {
		allow = func(string) bool { return true }
	}
	return &Store{DB: db, Sender: &Sender{VAPID: v, AllowHost: allow}}, nil
}

func loadOrCreateKey(db *gorm.DB) (string, error) {
	db = db.Session(&gorm.Session{Logger: logger.Discard}) // never print the private key in SQL logs
	var value string
	read := func() error {
		return db.Raw(`SELECT value FROM app_secrets WHERE name = ?`, vapidSecretName).Row().Scan(&value)
	}
	if err := read(); err == nil && value != "" {
		return value, nil
	}
	gen, err := GenerateKey()
	if err != nil {
		return "", err
	}
	// Two processes starting at once: the first insert wins, both read it back.
	if err := db.Exec(`INSERT INTO app_secrets (name, value) VALUES (?, ?) ON CONFLICT (name) DO NOTHING`, vapidSecretName, gen).Error; err != nil {
		return "", err
	}
	if err := read(); err != nil {
		return "", err
	}
	return value, nil
}

// PublicKey for PushManager.subscribe({applicationServerKey}).
func (s *Store) PublicKey() string { return s.Sender.VAPID.PublicKey() }

// MaxPerUser: phones/browsers per account; the oldest is dropped beyond this.
const MaxPerUser = 10

// Validate checks a subscription before it is stored.
func (s *Store) Validate(sub Subscription) error {
	sub.Endpoint = strings.TrimSpace(sub.Endpoint)
	if len(sub.Endpoint) > 1000 || !strings.HasPrefix(sub.Endpoint, "https://") {
		return errors.New("invalid endpoint")
	}
	host := strings.SplitN(strings.TrimPrefix(sub.Endpoint, "https://"), "/", 2)[0]
	host = strings.SplitN(host, ":", 2)[0]
	if !s.Sender.AllowHost(host) {
		return fmt.Errorf("unsupported push service: %s", host)
	}
	if raw, err := unb64(sub.P256dh); err != nil || len(raw) != 65 || raw[0] != 4 {
		return errors.New("invalid p256dh key")
	}
	if raw, err := unb64(sub.Auth); err != nil || len(raw) < 16 {
		return errors.New("invalid auth secret")
	}
	return nil
}

// Save stores (or moves to this user) a subscription.
func (s *Store) Save(userID int64, sub Subscription, userAgent string) error {
	if err := s.Validate(sub); err != nil {
		return err
	}
	if len(userAgent) > 300 {
		userAgent = userAgent[:300]
	}
	err := s.DB.Exec(`
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (endpoint) DO UPDATE SET user_id = EXCLUDED.user_id, p256dh = EXCLUDED.p256dh,
		    auth = EXCLUDED.auth, user_agent = EXCLUDED.user_agent, fail_count = 0`,
		userID, strings.TrimSpace(sub.Endpoint), sub.P256dh, sub.Auth, userAgent).Error
	if err != nil {
		return err
	}
	return s.DB.Exec(`
		DELETE FROM push_subscriptions WHERE user_id = ? AND id NOT IN (
		    SELECT id FROM push_subscriptions WHERE user_id = ? ORDER BY created_at DESC, id DESC LIMIT ?)`,
		userID, userID, MaxPerUser).Error
}

// Delete removes one subscription of the user.
func (s *Store) Delete(userID int64, endpoint string) error {
	return s.DB.Exec(`DELETE FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`, userID, endpoint).Error
}

// Has reports whether the endpoint is stored for the user.
func (s *Store) Has(userID int64, endpoint string) bool {
	var n int64
	s.DB.Raw(`SELECT COUNT(*) FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`, userID, endpoint).Row().Scan(&n)
	return n > 0
}

type subRow struct {
	ID       int64
	Endpoint string
	P256dh   string
	Auth     string
}

// SendToUser sends msg to the user's subscriptions (only to endpoint, if set).
// Expired subscriptions are deleted. Returns how many were delivered.
func (s *Store) SendToUser(ctx context.Context, userID int64, msg Message, endpoint string) (int, error) {
	var subs []subRow
	q := s.DB.Raw(`SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = ?`, userID)
	if endpoint != "" {
		q = s.DB.Raw(`SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE user_id = ? AND endpoint = ?`, userID, endpoint)
	}
	if err := q.Scan(&subs).Error; err != nil {
		return 0, err
	}
	payload, _ := json.Marshal(msg)
	sent := 0
	var lastErr error
	for _, sub := range subs {
		err := s.Sender.Send(ctx, Subscription{Endpoint: sub.Endpoint, P256dh: sub.P256dh, Auth: sub.Auth}, payload, 6*time.Hour, "study")
		switch {
		case err == nil:
			sent++
			s.DB.Exec(`UPDATE push_subscriptions SET last_sent_at = NOW(), fail_count = 0 WHERE id = ?`, sub.ID)
		case errors.Is(err, ErrGone):
			s.DB.Exec(`DELETE FROM push_subscriptions WHERE id = ?`, sub.ID)
		default:
			lastErr = err
			log.Printf("[push] user %d: %v", userID, err)
			// a subscription that keeps failing is dropped after ~10 attempts
			s.DB.Exec(`UPDATE push_subscriptions SET fail_count = fail_count + 1 WHERE id = ?`, sub.ID)
			s.DB.Exec(`DELETE FROM push_subscriptions WHERE id = ? AND fail_count >= 10`, sub.ID)
		}
	}
	if sent == 0 && lastErr != nil {
		return 0, lastErr
	}
	return sent, nil
}

// StudyReminder builds the reminder text from the user's study state.
// ok=false: nothing to study right now, no reminder needed.
func StudyReminder(st study.Stats) (msg Message, ok bool) {
	var parts []string
	if st.DueNow > 0 {
		parts = append(parts, plural(st.DueNow, "card", "cards")+" to review")
	}
	if st.NewLeftToday > 0 {
		parts = append(parts, plural(st.NewLeftToday, "new word", "new words")+" left today")
	}
	if len(parts) == 0 {
		return Message{
			Title: "All done for now ✓",
			Body:  "Nothing to review right now. Notifications are working.",
			URL:   "/",
			Tag:   "study-reminder",
		}, false
	}
	body := strings.Join(parts, " · ")
	if st.ReviewsToday == 0 && st.StreakDays >= 2 {
		body += fmt.Sprintf(" — keep your %d-day streak!", st.StreakDays)
	}
	return Message{Title: "Time to study 📚", Body: body, URL: "/review", Tag: "study-reminder"}, true
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}
