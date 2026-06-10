package models

import "time"

// ---- Core ----

type User struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Email       string    `gorm:"uniqueIndex;size:255" json:"email"`
	DisplayName string    `gorm:"size:255" json:"display_name"`
	FirebaseUID string    `gorm:"uniqueIndex;size:255" json:"firebase_uid"`
	AvatarURL   string    `gorm:"size:512" json:"avatar_url"`
	Timezone    string    `gorm:"size:64;default:UTC" json:"timezone"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TopicCategory struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Slug        string    `gorm:"uniqueIndex;size:128" json:"slug"`
	NameVi      string    `gorm:"size:256" json:"name_vi"`
	NameEn      string    `gorm:"size:256" json:"name_en"`
	Description string    `json:"description"`
	ParentID    *int64    `json:"parent_id"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Vocabulary struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Word      string    `gorm:"size:256;not null" json:"word"`
	Phonetics string    `gorm:"size:256" json:"phonetics"`
	Pos       string    `gorm:"size:64" json:"pos"`
	TopicID   *int64    `json:"topic_id"`
	AudioURL  string    `gorm:"size:512" json:"audio_url"`
	ImageURL  string    `gorm:"size:512" json:"image_url"`
	Source    string    `gorm:"size:64" json:"source"`
	MetaJSON  string    `gorm:"type:jsonb;default:'{}'" json:"meta_json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Sentences []Sentence `gorm:"foreignKey:VocabID" json:"sentences,omitempty"`
}

type Sentence struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	VocabID     int64     `gorm:"index;not null" json:"vocab_id"`
	Level       int       `gorm:"not null;check:level IN (1,2,3)" json:"level"`
	MeaningEn   string    `gorm:"not null" json:"meaning_en"`
	MeaningVi   string    `gorm:"not null" json:"meaning_vi"`
	ExampleEn   string    `gorm:"not null" json:"example_en"`
	ExampleVi   string    `gorm:"not null" json:"example_vi"`
	UsageNote   string    `json:"usage_note"`
	GrammarNote string    `json:"grammar_note"`
	AudioURL    string    `gorm:"size:512" json:"audio_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type VocabTag struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:128" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ---- User Progress ----

type UserVocabProgress struct {
	ID            int64      `gorm:"primaryKey" json:"id"`
	UserID        int64      `gorm:"index;not null" json:"user_id"`
	VocabID       int64      `gorm:"index;not null" json:"vocab_id"`
	State         int        `gorm:"default:0" json:"state"`
	Stability     float64    `gorm:"default:2.5" json:"stability"`
	Difficulty    float64    `gorm:"default:0.3" json:"difficulty"`
	ElapsedDays   int        `gorm:"default:0" json:"elapsed_days"`
	ScheduledDays int        `gorm:"default:0" json:"scheduled_days"`
	Reps          int        `gorm:"default:0" json:"reps"`
	Lapses        int        `gorm:"default:0" json:"lapses"`
	LastReviewAt  *time.Time `json:"last_review_at"`
	NextReviewAt  time.Time  `gorm:"default:NOW()" json:"next_review_at"`
	LevelFocus    int        `gorm:"default:1;check:level_focus IN (1,2,3)" json:"level_focus"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	MasteryScore  float64    `gorm:"default:0" json:"mastery_score"`
	HistoryJSON   string     `gorm:"type:jsonb;default:'[]'" json:"history_json"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// Relations
	Vocabulary *Vocabulary `gorm:"foreignKey:VocabID" json:"vocabulary,omitempty"`
}

type ReviewSession struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"index" json:"user_id"`
	VocabCount  int       `json:"vocab_count"`
	DurationMs  int64     `json:"duration_ms"`
	AvgRating   float64   `json:"avg_rating"`
	SessionType string    `gorm:"size:32;default:full" json:"session_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserStreak struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	UserID        int64     `gorm:"index" json:"user_id"`
	StreakDate    string    `gorm:"type:date" json:"streak_date"`
	VocabReviewed int       `gorm:"default:0" json:"vocab_reviewed"`
	WritingDone   bool      `json:"writing_done"`
	QuizDone      bool      `json:"quiz_done"`
	CreatedAt     time.Time `json:"created_at"`
}

// ---- Writing ----

type WritingSubmission struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `gorm:"index" json:"user_id"`
	TopicID        *int64     `json:"topic_id"`
	Prompt         string     `gorm:"not null" json:"prompt"`
	OriginalText   string     `gorm:"not null" json:"original_text"`
	WordCount      int        `json:"word_count"`
	RequiredVocabs string     `gorm:"type:jsonb" json:"required_vocabs"`
	UsedVocabs     string     `gorm:"type:jsonb" json:"used_vocabs"`
	OverallScore   *float64   `json:"overall_score"`
	GrammarScore   *float64   `json:"grammar_score"`
	VocabScore     *float64   `json:"vocab_score"`
	CoherenceScore *float64   `json:"coherence_score"`
	AIFeedbackJSON string     `gorm:"type:jsonb;default:'{}'" json:"ai_feedback_json"`
	AIModelUsed    string     `gorm:"size:64" json:"ai_model_used"`
	AITokensUsed   int        `json:"ai_tokens_used"`
	Status         string     `gorm:"size:32;default:pending" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// ---- Quiz ----

type Quiz struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	QuizType      string         `gorm:"size:32;default:vocab" json:"quiz_type"`
	TopicID       *int64         `json:"topic_id"`
	Difficulty    string         `gorm:"size:16;default:medium" json:"difficulty"`
	QuestionCount int            `gorm:"default:10" json:"question_count"`
	TimeLimitMin  *int           `json:"time_limit_min"`
	IsAIGenerated bool           `gorm:"default:true" json:"is_ai_generated"`
	AIPromptUsed  string         `json:"ai_prompt_used"`
	CreatedBy     *int64         `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	Questions     []QuizQuestion `gorm:"foreignKey:QuizID" json:"questions,omitempty"`
}

type QuizQuestion struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	QuizID       int64     `gorm:"index" json:"quiz_id"`
	VocabID      *int64    `json:"vocab_id"`
	QuestionType string    `gorm:"size:32;default:multiple_choice" json:"question_type"`
	Prompt       string    `gorm:"not null" json:"prompt"`
	Choices      string    `gorm:"type:jsonb;not null" json:"choices"`
	Explanation  string    `json:"explanation"`
	SortOrder    int       `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type QuizAttempt struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	UserID         int64      `gorm:"index" json:"user_id"`
	QuizID         int64      `gorm:"index" json:"quiz_id"`
	Score          *int       `json:"score"`
	TotalQuestions int        `json:"total_questions"`
	DurationMs     *int64     `json:"duration_ms"`
	Status         string     `gorm:"size:32;default:started" json:"status"`
	StartedAt      time.Time  `gorm:"default:NOW()" json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
}

// ---- Notification ----

type NotificationLog struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `gorm:"index" json:"user_id"`
	VocabID      int64     `json:"vocab_id"`
	Level        *int      `json:"level"`
	Channel      string    `gorm:"size:16;default:push" json:"channel"`
	Status       string    `gorm:"size:16;default:sent" json:"status"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserPreference struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	UserID          int64     `gorm:"uniqueIndex" json:"user_id"`
	DailyVocabGoal  int       `gorm:"default:10" json:"daily_vocab_goal"`
	ReminderEnabled bool      `gorm:"default:true" json:"reminder_enabled"`
	ReminderTimes   string    `gorm:"type:jsonb;default:'[\"09:00\",\"13:00\",\"19:00\"]'" json:"reminder_times"`
	QuizDifficulty  string    `gorm:"size:16;default:medium" json:"quiz_difficulty"`
	TTSSpeed        float64   `gorm:"default:1.0" json:"tts_speed"`
	UILanguage      string    `gorm:"size:8;default:vi" json:"ui_language"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Table names override
func (UserVocabProgress) TableName() string { return "user_vocab_progress" }
func (TopicCategory) TableName() string     { return "topic_categories" }
func (VocabTag) TableName() string          { return "vocab_tags" }
func (UserStreak) TableName() string        { return "user_streaks" }
func (ReviewSession) TableName() string     { return "review_sessions" }
func (WritingSubmission) TableName() string { return "writing_submissions" }
func (QuizQuestion) TableName() string      { return "quiz_questions" }
func (QuizAttempt) TableName() string       { return "quiz_attempts" }
func (NotificationLog) TableName() string   { return "notification_log" }
func (UserPreference) TableName() string    { return "user_preferences" }
func (Sentence) TableName() string          { return "sentences" }
func (Vocabulary) TableName() string        { return "vocabularies" }
