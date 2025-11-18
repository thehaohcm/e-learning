package models

import "time"

type StudySession struct {
    ID        int64     `gorm:"primaryKey"`
    UserID    int64     `gorm:"index"`
    Topic     string    `gorm:"size:128"`
    StartedAt time.Time
    EndedAt   time.Time
}

