package models

import "time"

type QuizQuestion struct {
    ID        int64     `gorm:"primaryKey"`
    VocabID   int64     `gorm:"index"`
    Prompt    string    `gorm:"size:512"`
    Choices   string    `gorm:"type:text"` // JSON array of choices
    AnswerIdx int
    CreatedAt time.Time
}

type QuizAttempt struct {
    ID         int64     `gorm:"primaryKey"`
    UserID     int64     `gorm:"index"`
    QuestionID int64     `gorm:"index"`
    Correct    bool
    CreatedAt  time.Time
}

