package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // time zone database inside the binary, so LoadLocation works in any container

	"mvp-lang-app/internal/config"
	"mvp-lang-app/internal/db"
	apihttp "mvp-lang-app/internal/http"
	"mvp-lang-app/internal/push"
	"mvp-lang-app/internal/tts"
	"mvp-lang-app/internal/worker"
	"mvp-lang-app/migrations"
)

func main() {
	cfg := config.Load()

	// Retry for ~1 minute: the database may still be starting (e.g. after a reboot).
	conn, err := db.Connect(cfg)
	for i := 1; err != nil && i <= 30; i++ {
		log.Printf("database not ready (%v), retrying in 2s (%d/30)", err, i)
		time.Sleep(2 * time.Second)
		conn, err = db.Connect(cfg)
	}
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	if err := db.Migrate(conn, migrations.FS); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go worker.Start(ctx, conn)
	go tts.New(conn, cfg.TTSURL, cfg.TTSVoice, cfg.TTSKeepDays).Cleanup(ctx) // deletes audio unused for TTS_KEEP_DAYS
	if store, err := push.NewStore(conn, cfg.VAPIDSubject, cfg.PushAnyHost); err != nil {
		log.Printf("push notifications disabled: %v", err)
	} else {
		go worker.StartReminders(ctx, conn, store)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apihttp.NewRouter(conn, cfg),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      200 * time.Second, // AI generation can take a while
	}
	go func() {
		log.Printf("API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
