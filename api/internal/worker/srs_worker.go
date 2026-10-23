// Package worker runs small periodic maintenance jobs.
//
// Note: no worker is needed to "find due cards" any more. Due cards are found
// at query time with the (user_id, due) index on user_cards.
package worker

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// Start runs maintenance every hour until ctx is cancelled.
func Start(ctx context.Context, db *gorm.DB) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		cleanup(db)
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

func cleanup(db *gorm.DB) {
	// Devices that registered but were never paired within 7 days.
	res := db.Exec(`DELETE FROM devices WHERE user_id IS NULL AND created_at < NOW() - INTERVAL '7 days'`)
	if res.Error != nil {
		log.Printf("[worker] cleanup devices: %v", res.Error)
	} else if res.RowsAffected > 0 {
		log.Printf("[worker] removed %d unpaired devices", res.RowsAffected)
	}
	// Writing evaluations stuck in "processing" (e.g. server restarted mid-request).
	db.Exec(`UPDATE writing_submissions SET status = 'failed', error_message = 'evaluation interrupted, please submit again'
	         WHERE status IN ('pending', 'processing') AND created_at < NOW() - INTERVAL '15 minutes'`)
	db.Exec(`DELETE FROM push_reminders_sent WHERE sent_at < NOW() - INTERVAL '3 days'`)
	db.Exec(`UPDATE readings SET status = 'failed', error_message = 'generation interrupted, please try again'
	         WHERE status = 'processing' AND created_at < NOW() - INTERVAL '15 minutes'`)
}
