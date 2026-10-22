package study

import "testing"

func TestDeviceCardMode(t *testing.T) {
	seen := DeviceCard{Kind: "vocab", HasOptions: true}
	withAudio := DeviceCard{Kind: "vocab", HasOptions: true, HasAudio: true}
	cases := []struct {
		name, setting string
		card          DeviceCard
		nth           int
		want          string
	}{
		{"flashcard", DeviceModeFlashcard, seen, 0, DeviceModeFlashcard},
		{"quiz", DeviceModeQuiz, seen, 0, DeviceModeQuiz},
		{"new card is always a flashcard", DeviceModeQuiz, DeviceCard{Kind: "vocab", IsNew: true}, 0, DeviceModeFlashcard},
		{"quiz without options", DeviceModeQuiz, DeviceCard{Kind: "vocab"}, 0, DeviceModeFlashcard},
		{"mcq is always a quiz", DeviceModeFlashcard, DeviceCard{Kind: "mcq"}, 0, DeviceModeQuiz},
		{"listening with audio", DeviceModeListening, withAudio, 0, DeviceModeListening},
		{"listening without audio -> quiz", DeviceModeListening, seen, 0, DeviceModeQuiz},
		{"listening on a concept -> quiz", DeviceModeListening, DeviceCard{Kind: "concept", HasOptions: true, HasAudio: true}, 0, DeviceModeQuiz},
		{"mix 0", DeviceModeMix, withAudio, 0, DeviceModeFlashcard},
		{"mix 1", DeviceModeMix, withAudio, 1, DeviceModeQuiz},
		{"mix 2", DeviceModeMix, withAudio, 2, DeviceModeListening},
		{"mix 2 without audio", DeviceModeMix, seen, 2, DeviceModeQuiz},
	}
	for _, c := range cases {
		if got := DeviceCardMode(c.setting, c.card, c.nth); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestDeviceModeSetting(t *testing.T) {
	if s := ParseSettings(nil); s.DeviceMode != DeviceModeFlashcard {
		t.Fatalf("default: %q", s.DeviceMode)
	}
	if s := ParseSettings([]byte(`{"device_mode":"karaoke"}`)); s.DeviceMode != DeviceModeFlashcard {
		t.Fatalf("invalid value kept: %q", s.DeviceMode)
	}
	s, err := ParseSettings(nil).Merge([]byte(`{"device_mode":"mix"}`))
	if err != nil || s.DeviceMode != DeviceModeMix {
		t.Fatalf("merge: %q %v", s.DeviceMode, err)
	}
}
