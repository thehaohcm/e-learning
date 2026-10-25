// Package firmware finds the newest study-device firmware for updates over
// Wi-Fi. deploy.sh uploads firmware/prebuilt, mounted in the API container
// at FIRMWARE_DIR; the update images are the files elearn-device-X.Y.Z-app.bin.
package firmware

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var appFile = regexp.MustCompile(`^elearn-device-(\d+)\.(\d+)\.(\d+)-app\.bin$`)

type Release struct {
	Version string `json:"version"`
	Size    int64  `json:"size"`
	File    string `json:"-"` // file name in the directory
}

// Latest returns the highest version in dir, or nil if there is none.
func Latest(dir string) *Release {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var best *Release
	for _, e := range entries {
		m := appFile.FindStringSubmatch(e.Name())
		if m == nil || e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || info.Size() < 64*1024 {
			continue
		}
		v := m[1] + "." + m[2] + "." + m[3]
		if best == nil || Compare(v, best.Version) > 0 {
			best = &Release{Version: v, Size: info.Size(), File: e.Name()}
		}
	}
	return best
}

// Path of a release file (only names Latest can return are accepted).
func Path(dir, name string) (string, bool) {
	if !appFile.MatchString(name) {
		return "", false
	}
	return filepath.Join(dir, name), true
}

// Compare compares "1.2.10" and "1.2.9" as numbers: -1, 0 or 1. Missing or
// non-numeric parts count as 0.
func Compare(a, b string) int {
	pa, pb := parts(a), parts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func parts(v string) [3]int {
	var out [3]int
	for i, p := range strings.SplitN(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".", 3) {
		n, _ := strconv.Atoi(strings.TrimFunc(p, func(r rune) bool { return r < '0' || r > '9' }))
		out[i] = n
	}
	return out
}
