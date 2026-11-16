package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AudioLibrary: GET /api/audio-library -> every audio the user has uploaded,
// grouped by its source (a cheatsheet section, or a Daily set), so the global
// listening panel can build a cross-page playlist. Metadata only (no bytes);
// the player fetches each track from its own endpoint by kind.
//
// Group kinds: "cheatsheet" (note audio), "daily" (Daily 20 — pmp/english) and
// "lab" (Daily Lab — pentest/aws). Audio kinds: "note" -> /note-audio/:id,
// "daily" -> /pmp-daily/audio/:id.
func (h *Handler) AudioLibrary(c *gin.Context) {
	uid := userID(c)

	type audioOut struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Kind string `json:"kind"` // "note" | "daily"
	}
	type group struct {
		Key         string      `json:"key"`
		Kind        string      `json:"kind"` // "cheatsheet" | "daily" | "lab"
		Subject     string      `json:"subject"`
		SubjectName string      `json:"subject_name"`
		Label       string      `json:"label"`
		Audios      []audioOut  `json:"audios"`
	}

	groups := []group{}
	index := map[string]int{} // key -> position in groups

	add := func(key, kind, subject, subjectName, label string, a audioOut) {
		if i, ok := index[key]; ok {
			groups[i].Audios = append(groups[i].Audios, a)
			return
		}
		index[key] = len(groups)
		groups = append(groups, group{Key: key, Kind: kind, Subject: subject, SubjectName: subjectName, Label: label, Audios: []audioOut{a}})
	}

	// cheatsheet section audio
	var notes []struct {
		ID      int64
		Name    string
		NoteID  int64
		Title   string
		Slug    string
		SName   string
	}
	h.DB.Raw(`SELECT na.id, na.name, na.note_id, cn.title, s.slug, s.name AS s_name
		FROM note_audio na
		JOIN course_notes cn ON cn.id = na.note_id
		JOIN subjects s ON s.id = cn.subject_id
		WHERE na.user_id = ?
		ORDER BY s.sort_order, s.id, cn.sort_order, cn.id, na.id`, uid).Scan(&notes)
	for _, n := range notes {
		add("note-"+strconv.FormatInt(n.NoteID, 10), "cheatsheet", n.Slug, n.SName, n.Title, audioOut{ID: n.ID, Name: n.Name, Kind: "note"})
	}

	// Daily set audio (Daily 20 or Daily Lab, by subject)
	var daily []struct {
		ID    int64
		Name  string
		SetID int64
		Day   time.Time
		Seq   int
		Slug  string
		SName string
	}
	h.DB.Raw(`SELECT da.id, da.name, da.set_id, ds.day, ds.seq, s.slug, s.name AS s_name
		FROM daily_audio da
		JOIN pmp_daily_sets ds ON ds.id = da.set_id
		JOIN subjects s ON s.id = ds.subject_id
		WHERE da.user_id = ?
		ORDER BY s.sort_order, s.id, ds.day DESC, ds.seq, da.id`, uid).Scan(&daily)
	for _, d := range daily {
		kind := "daily"
		if d.Slug == "pentest" || d.Slug == "aws" {
			kind = "lab"
		}
		label := d.Day.Format("2006-01-02")
		if d.Seq > 1 {
			label += " #" + strconv.Itoa(d.Seq)
		}
		add("set-"+strconv.FormatInt(d.SetID, 10), kind, d.Slug, d.SName, label, audioOut{ID: d.ID, Name: d.Name, Kind: "daily"})
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}
