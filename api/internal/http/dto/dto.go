package dto

// Auth
type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID          int64  `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
}

// Vocabulary
type GenerateVocabRequest struct {
	Word string `json:"word" binding:"required"`
}

type VocabResponse struct {
	ID        int64         `json:"id"`
	Word      string        `json:"word"`
	Phonetics string        `json:"phonetics"`
	Pos       string        `json:"pos"`
	TopicID   *int64        `json:"topic_id"`
	Sentences []SentenceDTO `json:"sentences"`
}

type SentenceDTO struct {
	ID          int64  `json:"id"`
	Level       int    `json:"level"`
	MeaningEn   string `json:"meaning_en"`
	MeaningVi   string `json:"meaning_vi"`
	ExampleEn   string `json:"example_en"`
	ExampleVi   string `json:"example_vi"`
	UsageNote   string `json:"usage_note"`
	GrammarNote string `json:"grammar_note"`
	AudioURL    string `json:"audio_url"`
}

// SRS
type StartLearningRequest struct {
	VocabIDs []int64 `json:"vocab_ids" binding:"required"`
}

type ReviewRequest struct {
	VocabID int64 `json:"vocab_id" binding:"required"`
	Rating  int   `json:"rating" binding:"required"` // 1=forgot, 2=hard, 3=good, 4=easy
}

type ReviewResponse struct {
	NextInterval int     `json:"next_interval_days"`
	NextReviewAt string  `json:"next_review_at"`
	Stability    float64 `json:"stability"`
	Difficulty   float64 `json:"difficulty"`
	State        int     `json:"state"`
}

type DueReviewItem struct {
	VocabID   int64  `json:"vocab_id"`
	Word      string `json:"word"`
	Phonetics string `json:"phonetics"`
	Level     int    `json:"level"`
	ExampleEn string `json:"example_en"`
	ExampleVi string `json:"example_vi"`
	MeaningEn string `json:"meaning_en"`
	MeaningVi string `json:"meaning_vi"`
	UsageNote string `json:"usage_note"`
}

// Quiz
type GenerateQuizRequest struct {
	QuizType   string `json:"quiz_type"` // vocab, toeic, ielts
	TopicID    *int64 `json:"topic_id"`
	Difficulty string `json:"difficulty"` // easy, medium, hard
	Count      int    `json:"count"`      // number of questions
}

type QuizQuestionDTO struct {
	ID        int64    `json:"id"`
	Prompt    string   `json:"prompt"`
	Choices   []string `json:"choices"`
	AnswerIdx int      `json:"-"` // hidden from response
}

type SubmitAnswerRequest struct {
	QuestionID int64 `json:"question_id" binding:"required"`
	AnswerIdx  int   `json:"answer_idx" binding:"required"`
}

type AnswerResponse struct {
	Correct          bool   `json:"correct"`
	CorrectAnswerIdx int    `json:"correct_answer_idx"`
	Explanation      string `json:"explanation"`
}

// Writing
type WritingSubmitRequest struct {
	Prompt         string  `json:"prompt" binding:"required"`
	Text           string  `json:"text" binding:"required"`
	RequiredVocabs []int64 `json:"required_vocab_ids"`
}

type WritingFeedbackResponse struct {
	ID             int64    `json:"id"`
	OverallScore   *float64 `json:"overall_score"`
	GrammarScore   *float64 `json:"grammar_score"`
	VocabScore     *float64 `json:"vocab_score"`
	CoherenceScore *float64 `json:"coherence_score"`
	Status         string   `json:"status"`
}

// Speech
type SpeechAssessResponse struct {
	OverallScore float64        `json:"overall_score"`
	Transcript   string         `json:"transcript"`
	WordScores   []WordScoreDTO `json:"word_scores"`
	Feedback     string         `json:"feedback"`
}

type WordScoreDTO struct {
	Word       string  `json:"word"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// Dashboard
type DashboardResponse struct {
	ActiveVocabs   int64    `json:"active_vocabs"`
	MasteredVocabs int64    `json:"mastered_vocabs"`
	DueReviews     int64    `json:"due_reviews"`
	Streak7d       int64    `json:"streak_7d"`
	BestWriting    *float64 `json:"best_writing_score"`
	BestQuiz       *float64 `json:"best_quiz_score"`
}

// Search
type SearchResponse struct {
	Results []VocabResponse `json:"results"`
	Query   string          `json:"query"`
}

// AI Generate
type GenerateRandomRequest struct {
	Count int `json:"count"` // number of words to generate (default 10)
}

// Preferences
type PreferencesRequest struct {
	DailyVocabGoal  *int     `json:"daily_vocab_goal"`
	ReminderEnabled *bool    `json:"reminder_enabled"`
	ReminderTimes   *string  `json:"reminder_times"`
	QuizDifficulty  *string  `json:"quiz_difficulty"`
	TTSSpeed        *float64 `json:"tts_speed"`
}
