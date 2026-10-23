package study

// Device pop-up modes (Settings.DeviceMode). "mix" is only a setting: each card
// is sent to the device with one of the other three.
const (
	DeviceModeFlashcard = "flashcard" // word -> tap to reveal -> rate it
	DeviceModeQuiz      = "quiz"      // word + 4 meanings, tap one
	DeviceModeListening = "listening" // play the audio, then pick the meaning
	DeviceModeCloze     = "cloze"     // example sentence with a blank, pick the word (4 options, no typing)
	DeviceModeMix       = "mix"
)

// DeviceCard describes what is known about a card when choosing its mode.
type DeviceCard struct {
	Kind       string
	IsNew      bool
	HasOptions bool // has the 4 meaning options (vocab/phrasal/concept cards that were seen before)
	HasAudio   bool // has an audio file the device can play
	HasCloze   bool // its example sentence contains the word and 4 word options exist
}

var mixOrder = []string{DeviceModeFlashcard, DeviceModeQuiz, DeviceModeCloze, DeviceModeListening}

// DeviceCardMode picks how one card is shown on the device.
//
//   - new cards are always flashcards: there is nothing to recall yet;
//   - mcq cards are always quizzes (their choices are in the card itself);
//   - quiz needs the 4 options, otherwise the card is a flashcard;
//   - listening needs audio (the device cannot synthesise speech), otherwise
//     it becomes a quiz;
//   - cloze needs an example sentence containing the word and 3 other words,
//     otherwise it becomes a quiz. The device never types: it picks a word.
//
// For "mix", nth counts the cards that could be quizzed so far, so the modes
// alternate in the batch.
func DeviceCardMode(setting string, c DeviceCard, nth int) string {
	if c.IsNew {
		return DeviceModeFlashcard
	}
	if c.Kind == "mcq" {
		return DeviceModeQuiz
	}
	mode := setting
	if mode == DeviceModeMix {
		mode = mixOrder[nth%len(mixOrder)]
	}
	if mode == DeviceModeListening && !(c.HasAudio && (c.Kind == "vocab" || c.Kind == "phrasal")) {
		mode = DeviceModeQuiz
	}
	if mode == DeviceModeCloze {
		if c.HasCloze {
			return DeviceModeCloze
		}
		mode = DeviceModeQuiz
	}
	if mode == DeviceModeQuiz && !c.HasOptions {
		mode = DeviceModeFlashcard
	}
	if mode != DeviceModeQuiz && mode != DeviceModeListening {
		mode = DeviceModeFlashcard
	}
	return mode
}
