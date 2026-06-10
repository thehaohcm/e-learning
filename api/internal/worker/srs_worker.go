package worker

import (
	"log"
	"time"

	"mvp-lang-app/internal/models"

	"gorm.io/gorm"
)

func StartSRSWorker(db *gorm.DB) {
	log.Println("[SRS Worker] Background worker started - scanning every 5 minutes")
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		processDueReviews(db)
	}
}

func processDueReviews(db *gorm.DB) {
	var count int64
	db.Model(&models.UserVocabProgress{}).
		Where("next_review_at <= NOW() AND is_active = true").
		Count(&count)

	if count > 0 {
		log.Printf("[SRS Worker] Found %d items due for review", count)

		// Update elapsed days for due items
		db.Exec(`
			UPDATE user_vocab_progress
			SET elapsed_days = GREATEST(0, EXTRACT(DAY FROM NOW() - last_review_at)::int)
			WHERE next_review_at <= NOW() AND is_active = true AND last_review_at IS NOT NULL
		`)
	}

	// Log review sessions for tracking
	db.Exec(`
		INSERT INTO review_sessions (user_id, vocab_count, duration_ms, session_type, created_at)
		SELECT user_id, COUNT(*), 0, 'auto', NOW()
		FROM user_vocab_progress
		WHERE next_review_at <= NOW() AND is_active = true
		GROUP BY user_id
		HAVING COUNT(*) > 0
	`)
}
