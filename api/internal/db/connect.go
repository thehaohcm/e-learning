package db

import (
    "fmt"

    "mvp-lang-app/internal/config"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

