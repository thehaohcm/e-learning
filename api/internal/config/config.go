package config

import "os"

type Config struct {
    DBHost string
    DBPort string
    DBUser string
    DBPass string
    DBName string
}

func Load() Config {
    return Config{
        DBHost: os.Getenv("DB_HOST"),
        DBPort: os.Getenv("DB_PORT"),
        DBUser: os.Getenv("DB_USER"),
        DBPass: os.Getenv("DB_PASSWORD"),
        DBName: os.Getenv("DB_NAME"),
    }
}

