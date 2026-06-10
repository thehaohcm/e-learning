package db

import (
	"mvp-lang-app/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := cfg.DSN()
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
