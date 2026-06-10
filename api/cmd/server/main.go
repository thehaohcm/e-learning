package main

import (
	"log"
	"os"

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/db"
	"mvp-lang-app/internal/http"
	"mvp-lang-app/internal/worker"
)

func main() {
	cfg := config.Load()

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	// Start SRS background worker
	go worker.StartSRSWorker(conn)

	// Create uploads directory
	os.MkdirAll("./uploads", 0755)

	r := http.NewRouter(conn, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5173"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
