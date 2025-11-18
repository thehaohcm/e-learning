package models

import "time"

type User struct {
    ID        int64     `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex;size:255"`
    CreatedAt time.Time
}

