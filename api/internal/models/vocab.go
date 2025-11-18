package models

import "time"

type Vocab struct {
    ID          int64     `gorm:"primaryKey"`
    Word        string    `gorm:"index;size:128"`
    Phonetics   string    `gorm:"size:128"`
    MeaningEn   string    `gorm:"size:512"`
    MeaningVi   string    `gorm:"size:512"`
    ExampleEn   string    `gorm:"size:512"`
    ExampleVi   string    `gorm:"size:512"`
    AudioURL    string    `gorm:"size:512"`
    Topic       string    `gorm:"size:128"`
    CreatedAt   time.Time
}

type UserVocab struct {
    ID            int64     `gorm:"primaryKey"`
    UserID        int64     `gorm:"index"`
    VocabID       int64     `gorm:"index"`
    Difficulty    string    `gorm:"size:16"` // easy | medium | hard
    LastResult    bool
    Stability     float64   // SRS factor
    IntervalDays  int
    NextReviewAt  time.Time `gorm:"index"`
    HistoryJSON   string    `gorm:"type:text"` // keep simple for MVP
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

