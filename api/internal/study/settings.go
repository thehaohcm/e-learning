package study

import (
	"encoding/json"
	"regexp"
	"sort"
)

// Settings are per-user study preferences stored in users.settings (JSONB).
// The device receives them on every sync.
type Settings struct {
	// New cards introduced per day, per topic: DailyNewBySubject["pmp"], or
	// DailyNew for a topic without its own number.
	DailyNew          int            `json:"daily_new"`
	DailyNewBySubject map[string]int `json:"daily_new_by_subject"`
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
	// Cards shown at each device pop-up (then "Keep studying" or "Done").
	PopupCards int `json:"popup_cards"`
	// "Learn N more today": extra new cards allowed on ExtraNewDate only
	// (YYYY-MM-DD in the user's time zone). Set via POST /api/study/extra-new.
	// Per topic (subject slug); ExtraNew is the old single value (= English).
	ExtraNewDate      string         `json:"extra_new_date,omitempty"`
	ExtraNew          int            `json:"extra_new,omitempty"`
	ExtraNewBySubject map[string]int `json:"extra_new_by_subject,omitempty"`
	// Text-to-speech speed for the web app.
	TTSSpeed float64 `json:"tts_speed"`
	// Study reminders sent as push notifications to the user's phones and
	// browsers, at these local times ("08:00", "20:30"). Empty = none.
	PushTimes []string `json:"push_times"`
	// Auto-pace: when on, each topic's daily new-card limit is raised a little
	// after a sustained "strong" memory streak and lowered on "needs focus"
	// (ordered introduction, never random). Off by default. State below is
	// managed by the server (AutoPace), not set by the user.
	AutoPace       bool           `json:"auto_pace"`
	AutoPaceDate   string         `json:"auto_pace_date,omitempty"`   // last daily evaluation (user's day)
	AutoPaceStrong map[string]int `json:"auto_pace_strong,omitempty"` // consecutive strong days, per subject slug
	AutoPaceNotes  []string       `json:"auto_pace_notes,omitempty"`  // recent changes, newest first (for the UI)
}

// setDailyNew sets a topic's daily new-card limit (used by AutoPace).
func (s *Settings) setDailyNew(slug string, n int) {
	if s.DailyNewBySubject == nil {
		s.DailyNewBySubject = map[string]int{}
	}
	s.DailyNewBySubject[slug] = n
}

// PopupCardChoices: how many cards a device pop-up can show.
var PopupCardChoices = []int{5, 10, 20}

func validPopupCards(n int) bool {
	for _, c := range PopupCardChoices {
		if n == c {
			return true
		}
	}
	return false
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
		PopupCards:       10,
		TTSSpeed:         1.0,
	}
}

// MaxExtraNew caps the extra new cards a user can add for one day ("+5 new
// words"): every new word brings several reviews in the next weeks.
const MaxExtraNew = 20

// DefaultSubject: the topic of the old single settings ("+5 new words", extra).
const DefaultSubject = "english"

// DailyNewFor: new cards a day for a topic (its own number, else DailyNew).
func (s Settings) DailyNewFor(subject string) int {
	if n, ok := s.DailyNewBySubject[subject]; ok {
		return n
	}
	return s.DailyNew
}

// extraFor: extra new cards added for a topic on day.
func (s Settings) extraFor(subject, day string) int {
	if s.ExtraNewDate != day {
		return 0
	}
	return s.ExtraNewBySubject[subject]
}

// ExtraNewLeft: how many more extra new cards of a topic can still be added on day.
func (s Settings) ExtraNewLeft(subject, day string) int {
	return max(0, MaxExtraNew-s.extraFor(subject, day))
}

// NewLimit is the number of new cards of a topic allowed on the given day
// (YYYY-MM-DD): the daily setting plus any extra added for that day.
func (s Settings) NewLimit(subject, day string) int {
	return s.DailyNewFor(subject) + s.extraFor(subject, day)
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
	if !validPopupCards(s.PopupCards) {
		s.PopupCards = d.PopupCards
	}
	if s.ExtraNew > 0 && s.ExtraNew <= MaxExtraNew && len(s.ExtraNewBySubject) == 0 { // stored before topics had their own
		s.ExtraNewBySubject = map[string]int{DefaultSubject: s.ExtraNew}
	}
	s.ExtraNew = 0
	for k, v := range s.ExtraNewBySubject {
		if v <= 0 || v > MaxExtraNew {
			delete(s.ExtraNewBySubject, k)
		}
	}
	if s.DailyNewBySubject == nil {
		s.DailyNewBySubject = map[string]int{}
	}
	for k, v := range s.DailyNewBySubject {
		if v < 0 || v > 200 || k == "" {
			delete(s.DailyNewBySubject, k)
		}
	}
	if s.ActiveDeckIDs == nil {
		s.ActiveDeckIDs = []int64{}
	}
	s.PushTimes = cleanTimes(s.PushTimes)
	for k, v := range s.AutoPaceStrong {
		if v < 0 || v > 60 || k == "" {
			delete(s.AutoPaceStrong, k)
		}
	}
	if len(s.AutoPaceNotes) > 6 {
		s.AutoPaceNotes = s.AutoPaceNotes[:6]
	}
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

// DeviceSettings are per-device overrides stored in devices.settings (JSONB).
// Zero values mean "same as the account" (Settings).
type DeviceSettings struct {
	// Topics shown on this device. Empty = all topics.
	SubjectIDs []int64 `json:"subject_ids"`
	// Pop-up / study mode on this device ("" = account setting).
	DeviceMode string `json:"device_mode"`
	// Pop-up interval and window on this device (0 / "" = account setting).
	ReminderEveryMin int    `json:"reminder_every_min"`
	ActiveFrom       string `json:"active_from"`
	ActiveTo         string `json:"active_to"`
	// Cards per pop-up on this device (0 = account setting).
	PopupCards int `json:"popup_cards"`
}

// ParseDeviceSettings reads stored device settings, dropping invalid values.
func ParseDeviceSettings(raw []byte) DeviceSettings {
	var d DeviceSettings
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &d)
	}
	return d.Sanitized()
}

func (d DeviceSettings) Sanitized() DeviceSettings {
	ids := []int64{}
	seen := map[int64]bool{}
	for _, id := range d.SubjectIDs {
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	d.SubjectIDs = ids
	switch d.DeviceMode {
	case "", DeviceModeFlashcard, DeviceModeQuiz, DeviceModeListening, DeviceModeCloze, DeviceModeMix:
	default:
		d.DeviceMode = ""
	}
	if d.ReminderEveryMin != 0 && (d.ReminderEveryMin < 5 || d.ReminderEveryMin > 24*60) {
		d.ReminderEveryMin = 0
	}
	if d.ActiveFrom != "" && !hhmm.MatchString(d.ActiveFrom) {
		d.ActiveFrom = ""
	}
	if d.ActiveTo != "" && !hhmm.MatchString(d.ActiveTo) {
		d.ActiveTo = ""
	}
	if d.PopupCards != 0 && !validPopupCards(d.PopupCards) {
		d.PopupCards = 0
	}
	return d
}

// ForDevice returns the account settings with this device's overrides applied.
func (s Settings) ForDevice(d DeviceSettings) Settings {
	if d.DeviceMode != "" {
		s.DeviceMode = d.DeviceMode
	}
	if d.ReminderEveryMin > 0 {
		s.ReminderEveryMin = d.ReminderEveryMin
	}
	if d.ActiveFrom != "" {
		s.ActiveFrom = d.ActiveFrom
	}
	if d.ActiveTo != "" {
		s.ActiveTo = d.ActiveTo
	}
	if d.PopupCards > 0 {
		s.PopupCards = d.PopupCards
	}
	return s
}
