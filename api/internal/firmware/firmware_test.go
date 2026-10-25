package firmware

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.1.0", "1.0.9", 1}, {"1.2.10", "1.2.9", 1}, {"1.0.0", "1.0.0", 0}, {"0.9", "1.0.0", -1}, {"v2.0.0", "1.9.9", 1},
	}
	for _, c := range cases {
		if got := Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%s, %s) = %d", c.a, c.b, got)
		}
	}
}

func TestLatest(t *testing.T) {
	dir := t.TempDir()
	big := make([]byte, 70*1024)
	for _, n := range []string{"elearn-device-1.0.9-app.bin", "elearn-device-1.1.0-app.bin", "elearn-device-1.1.10-app.bin", "elearn-device-1.2.0.bin", "notes.txt"} {
		os.WriteFile(filepath.Join(dir, n), big, 0o600)
	}
	os.WriteFile(filepath.Join(dir, "elearn-device-9.0.0-app.bin"), []byte("tiny"), 0o600) // not a real image
	r := Latest(dir)
	if r == nil || r.Version != "1.1.10" || r.File != "elearn-device-1.1.10-app.bin" || r.Size != int64(len(big)) {
		t.Fatalf("Latest = %+v", r)
	}
	if Latest(filepath.Join(dir, "nope")) != nil {
		t.Fatal("missing dir")
	}
	if _, ok := Path(dir, "../etc/passwd"); ok {
		t.Fatal("path traversal")
	}
}
