package study

import (
	"encoding/json"
	"regexp"
	"sort"
)

// Settings are per-user study preferences stored in users.settings (JSONB).
// The device receives them on every sync.
type Settings struct {
	// New cards introduced per day across all active decks.
	DailyNew int `json:"daily_new"`
	// Decks that automatically feed new cards into the study queue.
	// Empty = use the default English core vocabulary deck.
	ActiveDeckIDs []int64 `json:"active_deck_ids"`
	// Device "pop-up" reminder: show one card every N minutes inside the window.
	ReminderEveryMin int    `json:"reminder_every_min"`
	ActiveFrom       string `json:"active_from"` // "07:30"
	ActiveTo         string `json:"active_to"`   // "22:00"
	// What the device shows at each pop-up: "flashcard", "quiz", "listening" or
	// "mix" (rotates the three). See DeviceCardMode.
	DeviceMode string `json:"device_mode"`
	// "Learn N more today": extra new cards allowed on ExtraNewDate only
	// (YYYY-MM-DD in the user's time zone). Set via POST /api/study/extra-new.
	ExtraNewDate string `json:"extra_new_date,omitempty"`
	ExtraNew     int    `json:"extra_new,omitempty"`
	// Text-to-speech speed for the web app.
	TTSSpeed float64 `json:"tts_speed"`
	// Study reminders sent as push notifications to the user's phones and
	// browsers, at these local times ("08:00", "20:30"). Empty = none.
	PushTimes []string `json:"push_times"`
}

// MaxPushTimes caps the number of daily reminder times.
const MaxPushTimes = 6

func DefaultSettings() Settings {
	return Settings{
		DailyNew:         10,
		ReminderEveryMin: 45,
		ActiveFrom:       "07:30",
		ActiveTo:         "22:00",
		DeviceMode:       DeviceModeFlashcard,
		TTSSpeed:         1.0,
	}
}

// MaxExtraNew caps the extra new cards a user can add for one day.
const MaxExtraNew = 100

// NewLimit is the number of new cards allowed on the given day (YYYY-MM-DD):
// the daily setting plus any extra added for that day.
func (s Settings) NewLimit(day string) int {
	if s.ExtraNewDate == day {
		return s.DailyNew + s.ExtraNew
	}
	return s.DailyNew
}

var hhmm = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// ParseSettings reads stored settings and fills missing values with defaults.
func ParseSettings(raw []byte) Settings {
	s := DefaultSettings()
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &s)
	}
	return s.sanitized()
}

func (s Settings) sanitized() Settings {
	d := DefaultSettings()
	if s.DailyNew < 0 || s.DailyNew > 200 {
		s.DailyNew = d.DailyNew
	}
	if s.ReminderEveryMin < 5 || s.ReminderEveryMin > 24*60 {
		s.ReminderEveryMin = d.ReminderEveryMin
	}
	if !hhmm.MatchString(s.ActiveFrom) {
		s.ActiveFrom = d.ActiveFrom
	}
	if !hhmm.MatchString(s.ActiveTo) {
		s.ActiveTo = d.ActiveTo
	}
	if s.TTSSpeed < 0.5 || s.TTSSpeed > 2 {
		s.TTSSpeed = d.TTSSpeed
	}
	switch s.DeviceMode {
	case DeviceModeFlashcard, DeviceModeQuiz, DeviceModeListening, DeviceModeCloze, DeviceModeMix:
	default:
		s.DeviceMode = d.DeviceMode
	}
	if s.ExtraNew < 0 || s.ExtraNew > MaxExtraNew {
		s.ExtraNew = 0
	}
	if s.ActiveDeckIDs == nil {
		s.ActiveDeckIDs = []int64{}
	}
	s.PushTimes = cleanTimes(s.PushTimes)
	return s
}

// Merge applies a partial JSON update on top of the current settings.
func (s Settings) Merge(patch []byte) (Settings, error) {
	if err := json.Unmarshal(patch, &s); err != nil {
		return s, err
	}
	return s.sanitized(), nil
}

// cleanTimes keeps valid, distinct HH:MM times, sorted, at most MaxPushTimes.
func cleanTimes(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, t := range in {
		if hhmm.MatchString(t) && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	sort.Strings(out)
	if len(out) > MaxPushTimes {
		out = out[:MaxPushTimes]
	}
	return out
}
