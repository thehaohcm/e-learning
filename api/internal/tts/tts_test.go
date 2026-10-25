package tts

import "testing"

func TestNormalizeAndKey(t *testing.T) {
	if got := Normalize("  look   up \n"); got != "look up" {
		t.Fatalf("Normalize = %q", got)
	}
	if Normalize("   ") != "" {
		t.Fatal("blank text should not be read")
	}
	long := make([]byte, MaxText+1)
	for i := range long {
		long[i] = 'a'
	}
	if Normalize(string(long)) != "" {
		t.Fatal("too long text should not be read")
	}
	s := &Service{Voice: "en-US-AriaNeural"}
	if s.Key("Apple") != s.Key(" apple ") {
		t.Fatal("key should ignore case and spaces")
	}
	if len(s.Key("apple")) != 32 {
		t.Fatal("key length")
	}
	other := &Service{Voice: "en-GB-SoniaNeural"}
	if s.Key("apple") == other.Key("apple") {
		t.Fatal("key should depend on the voice")
	}
	if URL("abc") != "/api/audio/abc.wav" {
		t.Fatal("URL")
	}
}
