package models

import "time"

type Vocab struct {
	ID        int64     `gorm:"primaryKey" json:"ID"`
	Word      string    `gorm:"index;size:128" json:"Word"`
	Phonetics string    `gorm:"size:128" json:"Phonetics"`
	MeaningEn string    `gorm:"size:512" json:"MeaningEn"`
	MeaningVi string    `gorm:"size:512" json:"MeaningVi"`
	ExampleEn string    `gorm:"size:512" json:"ExampleEn"`
	ExampleVi string    `gorm:"size:512" json:"ExampleVi"`
	AudioURL  string    `gorm:"size:512" json:"AudioURL"`
	Topic     string    `gorm:"size:128" json:"Topic"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type UserVocab struct {
	ID           int64  `gorm:"primaryKey"`
	UserID       int64  `gorm:"index"`
	VocabID      int64  `gorm:"index"`
	Difficulty   string `gorm:"size:16"` // easy | medium | hard
	LastResult   bool
	Stability    float64 // SRS factor
	IntervalDays int
	NextReviewAt time.Time `gorm:"index"`
	HistoryJSON  string    `gorm:"type:text"` // keep simple for MVP
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
