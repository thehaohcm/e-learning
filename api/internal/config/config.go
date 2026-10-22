package config

import (
	"os"
	"strings"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	DBSSL  string

	Port        string
	AppEnv      string
	JWTSecret   string
	CORSOrigins []string
	AdminEmails []string
	OpenSignup  bool
	DefaultTZ   string

	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string
}

func Load() Config {
	cfg := Config{
		DBHost: requireEnv("DB_HOST"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: requireEnv("DB_USER"),
		DBPass: requireEnv("DB_PASSWORD"),
		DBName: requireEnv("DB_NAME"),
		DBSSL:  getEnv("DB_SSLMODE", "disable"),

		Port:        getEnv("PORT", "8000"),
		AppEnv:      getEnv("APP_ENV", "development"),
		JWTSecret:   getEnv("JWT_SECRET", ""),
		CORSOrigins: splitList(getEnv("CORS_ORIGINS", "http://localhost:8085,http://localhost:5173")),
		AdminEmails: splitList(strings.ToLower(getEnv("ADMIN_EMAILS", ""))),
		OpenSignup:  getEnv("OPEN_SIGNUP", "true") == "true",
		DefaultTZ:   getEnv("DEFAULT_TIMEZONE", "Asia/Ho_Chi_Minh"),

		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),
		DeepSeekModel:   getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
	}

	if strings.Contains(cfg.JWTSecret, "change-this") {
		cfg.JWTSecret = "" // the placeholder from .env.example is not a secret
	}
	if len(cfg.JWTSecret) < 32 {
		if cfg.AppEnv == "production" {
			panic("JWT_SECRET must be set to at least 32 random characters in production")
		}
		cfg.JWTSecret = "dev-only-insecure-secret-change-me-please-0000"
	}
	return cfg
}

func (c Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPass +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSL +
		" TimeZone=UTC"
}

func (c Config) IsAdminEmail(email string) bool {
	email = strings.ToLower(strings.TrimSpace(email))
	for _, e := range c.AdminEmails {
		if e == email {
			return true
		}
	}
	return false
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

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
