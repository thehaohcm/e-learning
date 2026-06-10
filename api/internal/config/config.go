package config

import "os"

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPass          string
	DBName          string
	RedisURL        string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	JWTSecret       string
	AppEnv          string
	FrontendURL     string
}

func Load() Config {
	return Config{
		DBHost:          requireEnv("DB_HOST"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          requireEnv("DB_USER"),
		DBPass:          requireEnv("DB_PASSWORD"),
		DBName:          requireEnv("DB_NAME"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret"),
		AppEnv:          getEnv("APP_ENV", "development"),
		FrontendURL:     getEnv("FRONTEND_URL", ""),
	}
}

func (c Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPass +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("Missing required environment variable: " + key)
	}
	return v
}
