package srs

import (
    "time"
)

// Simple SM2-like update (MVP)
// rating: 5=easy, 4=good, 3=pass, 2=hard, 1=fail
func Update(interval int, stability float64, rating int) (nextInterval int, nextStability float64, nextReview time.Time) {
    if rating < 3 {
        nextInterval = 1
        nextStability = max(1.3, stability-0.3)
    } else {
        if interval == 1 {
            nextInterval = 2
        } else {
            nextInterval = int(float64(interval) * stability)
            if nextInterval < interval+1 {
                nextInterval = interval + 1
            }
        }
        // adjust stability based on rating
        nextStability = stability + 0.1*float64(rating-3)
        if nextStability < 1.3 {
            nextStability = 1.3
        }
    }
    nextReview = time.Now().Add(time.Duration(nextInterval) * 24 * time.Hour)
    return
}

func max(a, b float64) float64 {
    if a > b { return a }
    return b
}

