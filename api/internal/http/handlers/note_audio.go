package handlers

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"mvp-lang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// Learner-uploaded audio attached to a cheatsheet section (course_note).
// Reuses audioMime / cleanAudioName / size limits from daily_audio.go.

type noteAudio struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64 `gorm:"column:user_id"`
	NoteID    int64 `gorm:"column:note_id"`
	Name      string
	Mime      string
	Bytes     int
	Data      []byte `gorm:"column:data"`
	CreatedAt time.Time
}

func (noteAudio) TableName() string { return "note_audio" }

// ownsNote checks the course_note belongs to the user.
func (h *Handler) ownsNote(uid, noteID int64) bool {
	var n models.CourseNote
	if err := h.DB.Select("id", "user_id").Where("id = ? AND user_id = ?", noteID, uid).Limit(1).Find(&n).Error; err != nil {
		return false
	}
	return n.ID != 0
}

// NoteAudioUpload: POST /api/note-audio (multipart: note_id, files[])
func (h *Handler) NoteAudioUpload(c *gin.Context) {
	uid := userID(c)
	noteID, _ := strconv.ParseInt(c.PostForm("note_id"), 10, 64)
	if !h.ownsNote(uid, noteID) {
		abort(c, http.StatusNotFound, "note not found")
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
		rec := noteAudio{UserID: uid, NoteID: noteID, Name: cleanAudioName(fh.Filename), Mime: mime, Bytes: len(data), Data: data, CreatedAt: time.Now().UTC()}
		if err := h.DB.Create(&rec).Error; err != nil {
			internal(c, err)
			return
		}
		saved = append(saved, gin.H{"id": rec.ID, "name": rec.Name, "mime": rec.Mime, "bytes": rec.Bytes, "created_at": rec.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"files": saved})
}

// NoteAudioList: GET /api/note-audio?note_id=N -> metadata only
func (h *Handler) NoteAudioList(c *gin.Context) {
	uid := userID(c)
	noteID, _ := strconv.ParseInt(c.Query("note_id"), 10, 64)
	var rows []noteAudio
	if err := h.DB.Select("id", "note_id", "user_id", "name", "mime", "bytes", "created_at").
		Where("user_id = ? AND note_id = ?", uid, noteID).Order("id ASC").Find(&rows).Error; err != nil {
		internal(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{"id": r.ID, "name": r.Name, "mime": r.Mime, "bytes": r.Bytes, "created_at": r.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"files": out})
}

// NoteAudioGet: GET /api/note-audio/:id -> audio bytes (range-capable)
func (h *Handler) NoteAudioGet(c *gin.Context) {
	uid := userID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var r noteAudio
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

// NoteAudioDelete: DELETE /api/note-audio/:id
func (h *Handler) NoteAudioDelete(c *gin.Context) {
	uid := userID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res := h.DB.Where("id = ? AND user_id = ?", id, uid).Delete(&noteAudio{})
	if res.Error != nil {
		internal(c, res.Error)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": res.RowsAffected})
}
