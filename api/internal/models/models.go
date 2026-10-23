// Package models maps the database tables (see api/migrations) to Go structs.
// The schema itself is owned by the SQL migrations, never by GORM AutoMigrate.
package models

import (
	"database/sql/driver"
	"errors"
	"time"
)

// JSON is a raw JSONB column value.
type JSON []byte

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}

func (j *JSON) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*j = JSON("{}")
	case []byte:
		*j = append((*j)[:0], v...)
	case string:
		*j = JSON(v)
	default:
		return errors.New("models.JSON: unsupported scan type")
	}
	return nil
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(b []byte) error {
	*j = append((*j)[:0], b...)
	return nil
}

// ---------------------------------------------------------------- users

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Timezone     string    `json:"timezone"`
	Settings     JSON      `gorm:"type:jsonb" json:"settings"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------- content

const (
	KindVocab   = "vocab"
	KindPhrasal = "phrasal"
	KindConcept = "concept"
	KindMCQ     = "mcq"
)

type Subject struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	Decks       []Deck    `gorm:"foreignKey:SubjectID" json:"decks,omitempty"`
}

type Deck struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	SubjectID   int64     `json:"subject_id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DefaultKind string    `json:"default_kind"`
	SortOrder   int       `json:"sort_order"`
	OwnerID     *int64    `json:"owner_id,omitempty"` // personal deck ("My words"); nil = shared
	CreatedAt   time.Time `json:"created_at"`
}

type Card struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	DeckID    int64     `json:"deck_id"`
	Kind      string    `json:"kind"`
	Front     string    `json:"front"`
	Payload   JSON      `gorm:"type:jsonb" json:"payload"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------------------------------------------------------------- learning

type UserCard struct {
	UserID        int64      `gorm:"primaryKey;autoIncrement:false" json:"user_id"`
	CardID        int64      `gorm:"primaryKey;autoIncrement:false" json:"card_id"`
	State         int        `json:"state"`
	Stability     float64    `json:"stability"`
	Difficulty    float64    `json:"difficulty"`
	Due           time.Time  `json:"due"`
	LastReview    *time.Time `json:"last_review"`
	ScheduledDays int        `json:"scheduled_days"`
	Reps          int        `json:"reps"`
	Lapses        int        `json:"lapses"`
	Suspended     bool       `json:"suspended"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ReviewLog struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	UserID         int64     `json:"user_id"`
	CardID         int64     `json:"card_id"`
	Rating         int       `json:"rating"`
	StateBefore    int       `json:"state_before"`
	ElapsedDays    int       `json:"elapsed_days"`
	ScheduledDays  int       `json:"scheduled_days"`
	DurationMs     *int      `json:"duration_ms"`
	Source         string    `json:"source"`
	DeviceID       *int64    `json:"device_id"`
	ClientReviewID *string   `json:"client_review_id"`
	ReviewedAt     time.Time `json:"reviewed_at"`
}

// ---------------------------------------------------------------- devices

type Device struct {
	ID                int64      `gorm:"primaryKey" json:"id"`
	UserID            *int64     `json:"user_id"`
	Name              string     `json:"name"`
	TokenHash         string     `json:"-"`
	PairCode          *string    `json:"-"`
	PairCodeExpiresAt *time.Time `json:"-"`
	Firmware          string     `json:"firmware"`
	Settings          JSON       `gorm:"type:jsonb" json:"settings"`
	LastSeenAt        *time.Time `json:"last_seen_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ---------------------------------------------------------------- writing

type WritingSubmission struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `json:"user_id"`
	Prompt         string     `json:"prompt"`
	OriginalText   string     `json:"original_text"`
	WordCount      int        `json:"word_count"`
	OverallScore   *float64   `json:"overall_score"`
	GrammarScore   *float64   `json:"grammar_score"`
	VocabScore     *float64   `json:"vocab_score"`
	CoherenceScore *float64   `json:"coherence_score"`
	AIFeedbackJSON JSON       `gorm:"column:ai_feedback_json;type:jsonb" json:"feedback"`
	AIModelUsed    *string    `json:"ai_model_used"`
	Status         string     `json:"status"`
	ErrorMessage   *string    `json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

func (UserCard) TableName() string          { return "user_cards" }
func (ReviewLog) TableName() string         { return "review_logs" }
func (WritingSubmission) TableName() string { return "writing_submissions" }

// ---------------------------------------------------------------- reading

type Reading struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	UserID       int64      `json:"user_id"`
	Kind         string     `json:"kind"`
	Level        string     `json:"level"`
	Title        string     `json:"title"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message"`
	Content      JSON       `gorm:"type:jsonb" json:"content"`
	Answers      JSON       `gorm:"type:jsonb" json:"answers"`
	Score        *int       `json:"score"`
	Total        *int       `json:"total"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	AnsweredAt   *time.Time `json:"answered_at"`
}

func (Reading) TableName() string { return "readings" }

// ---------------------------------------------------------------- practice

// CardExample is an extra example sentence of a card (see migration 0007).
type CardExample struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	CardID        int64     `json:"card_id"`
	SentenceEn    string    `json:"sentence_en"`
	TranslationVi string    `json:"translation_vi"`
	Source        string    `json:"source"`
	Context       string    `json:"context"` // situation label (work, daily life, news...)
	CreatedAt     time.Time `json:"created_at"`
}

func (CardExample) TableName() string { return "card_examples" }

// WordNote is a word the user noted while studying (floating notes panel).
type WordNote struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `json:"-"`
	Text      string    `json:"text"`
	Note      string    `json:"note"`
	Context   string    `json:"context"`
	Source    string    `json:"source"`
	CardID    *int64    `json:"card_id"`
	Status    string    `json:"status"` // new | ready | added
	Info      JSON      `gorm:"type:jsonb" json:"info"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WordNote) TableName() string { return "word_notes" }

// CardPronunciation: syllables, stress and tips for Vietnamese learners (migration 0010).
type CardPronunciation struct {
	CardID    int64     `gorm:"primaryKey;autoIncrement:false" json:"card_id"`
	Syllables JSON      `gorm:"type:jsonb" json:"syllables"`
	Stress    int       `json:"stress"`
	TipsVi    JSON      `gorm:"column:tips_vi;type:jsonb" json:"tips_vi"`
	CreatedAt time.Time `json:"created_at"`
}

func (CardPronunciation) TableName() string { return "card_pronunciations" }
