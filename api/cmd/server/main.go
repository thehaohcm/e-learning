package main

import (
	"log"
	"os"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/db"
	"mvp-lang-app/internal/http"
)

func main() {
	cfg := config.Load()
	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	if err := db.RunMigrations(conn, "./migrations"); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	r := http.NewRouter(conn)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5173"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
