package handlers

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Learner-uploaded audio notes attached to a Daily set (an "exam"): after
// taking a set the learner can upload mp3 / m4a files (e.g. a NotebookLM
// summary), then list, play and delete them. Stored as BYTEA in Postgres.

const (
	maxAudioBytes = 25 * 1024 * 1024 // 25 MB per file
	maxAudioFiles = 8                // per upload request
)

type dailyAudio struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"column:user_id"`
	SetID     int64 `gorm:"column:set_id"`
	Name      string
	Mime      string
	Bytes     int
	Data      []byte `gorm:"column:data"`
	CreatedAt time.Time
}

func (dailyAudio) TableName() string { return "daily_audio" }

func audioMime(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3":
		return "audio/mpeg"
	case ".m4a", ".mp4":
		return "audio/mp4"
	case ".aac":
		return "audio/aac"
	case ".wav":
		return "audio/wav"
	case ".ogg", ".oga":
		return "audio/ogg"
	case ".webm":
		return "audio/webm"
	}
	return ""
}

func cleanAudioName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == "/" {
		name = "audio"
	}
	if len(name) > 120 {
		name = name[:120]
	}
	return name
}

// DailyAudioUpload: POST /api/pmp-daily/audio (multipart: set_id, files[])
func (h *Handler) DailyAudioUpload(c *gin.Context) {
	uid := userID(c)
	setID, _ := strconv.ParseInt(c.PostForm("set_id"), 10, 64)
	s, err := h.loadDailyByID(uid, setID)
	if err != nil {
		internal(c, err)
		return
	}
	if s == nil {
		abort(c, http.StatusNotFound, "set not found")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		abort(c, http.StatusBadRequest, "no files uploaded")
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		abort(c, http.StatusBadRequest, "no files uploaded")
		return
	}
	if len(files) > maxAudioFiles {
		abort(c, http.StatusBadRequest, "too many files (max 8 per upload)")
		return
	}
	saved := make([]gin.H, 0, len(files))
	for _, fh := range files {
		if fh.Size > maxAudioBytes {
			abort(c, http.StatusRequestEntityTooLarge, "file too large (max 25MB): "+fh.Filename)
			return
		}
		mime := audioMime(filepath.Ext(fh.Filename))
		if mime == "" {
			abort(c, http.StatusBadRequest, "unsupported audio type (use mp3/m4a): "+fh.Filename)
			return
		}
		f, err := fh.Open()
		if err != nil {
			internal(c, err)
			return
		}
		data, err := io.ReadAll(io.LimitReader(f, maxAudioBytes+1))
		f.Close()
		if err != nil {
			internal(c, err)
			return
		}
		if len(data) > maxAudioBytes {
			abort(c, http.StatusRequestEntityTooLarge, "file too large (max 25MB): "+fh.Filename)
			return
		}
		rec := dailyAudio{UserID: uid, SetID: setID, Name: cleanAudioName(fh.Filename), Mime: mime, Bytes: len(data), Data: data, CreatedAt: time.Now().UTC()}
		if err := h.DB.Create(&rec).Error; err != nil {
			internal(c, err)
			return
		}
		saved = append(saved, gin.H{"id": rec.ID, "name": rec.Name, "mime": rec.Mime, "bytes": rec.Bytes, "created_at": rec.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"files": saved})
}

// DailyAudioList: GET /api/pmp-daily/audio?set_id=N -> metadata only (no blob)
func (h *Handler) DailyAudioList(c *gin.Context) {
	uid := userID(c)
	setID, _ := strconv.ParseInt(c.Query("set_id"), 10, 64)
	var rows []dailyAudio
	if err := h.DB.Select("id", "set_id", "user_id", "name", "mime", "bytes", "created_at").
		Where("user_id = ? AND set_id = ?", uid, setID).Order("id ASC").Find(&rows).Error; err != nil {
		internal(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{"id": r.ID, "name": r.Name, "mime": r.Mime, "bytes": r.Bytes, "created_at": r.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"files": out})
}

// DailyAudioGet: GET /api/pmp-daily/audio/:id -> the audio bytes (range-capable)
func (h *Handler) DailyAudioGet(c *gin.Context) {
	uid := userID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r dailyAudio
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).Limit(1).Find(&r).Error; err != nil {
		internal(c, err)
		return
	}
	if r.ID == 0 {
		abort(c, http.StatusNotFound, "not found")
		return
	}
	c.Header("Content-Type", r.Mime)
	c.Header("Cache-Control", "private, max-age=3600")
	http.ServeContent(c.Writer, c.Request, r.Name, r.CreatedAt, bytes.NewReader(r.Data))
}

// DailyAudioDelete: DELETE /api/pmp-daily/audio/:id
func (h *Handler) DailyAudioDelete(c *gin.Context) {
	uid := userID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res := h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&dailyAudio{})
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": res.RowsAffected})
}
