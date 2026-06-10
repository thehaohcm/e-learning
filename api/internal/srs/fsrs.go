package srs

import (
	"math"
	"time"
)

// FSRS-5 default weights
var DefaultWeights = []float64{
	0.4872, 1.4003, 3.7145, 13.8206, 5.1618, 1.2298,
	0.8975, 0.031, 1.6474, 0.1367, 1.0461, 0.6461,
	0.0411, 1.0295, 0.1, 0.9906, 0.9714, 0.9561, 1.1687,
}

type FSRSParams struct {
	Stability   float64
	Difficulty  float64
	ElapsedDays int
	Reps        int
	Lapses      int
	State       int
}

type FSRSScheduleResult struct {
	NextStability  float64
	NextDifficulty float64
	NextInterval   int
	NextReviewAt   time.Time
	NextState      int
}

func retrievability(elapsedDays int, stability float64) float64 {
	if elapsedDays <= 0 {
		return 1.0
	}
	t := float64(elapsedDays)
	return math.Pow(1+t/(9*stability), -1)
}

func NextInterval(stability float64, retentionTarget float64) int {
	if retentionTarget <= 0 || retentionTarget >= 1 {
		retentionTarget = 0.9
	}
	return int(math.Ceil(stability * math.Log(retentionTarget) / math.Log(0.5)))
}

func Schedule(p FSRSParams, rating int) FSRSScheduleResult {
	w := DefaultWeights
	r := FSRSScheduleResult{}

	R := retrievability(p.ElapsedDays, p.Stability)

	newD := p.Difficulty + w[4]*(float64(rating-3))*(1.0/17.0)
	newD = math.Max(0.0, math.Min(1.0, newD))
	newD = w[5]*p.Difficulty + (1-w[5])*newD

	var newS float64
	if rating >= 3 {
		factor := 1 + math.Exp(w[8])*(11-newD)*math.Pow(p.Stability, -w[9])*
			(math.Exp((1-R)*w[10])-1)
		newS = p.Stability * factor
		newS = w[6]*p.Stability + (1-w[6])*newS
	} else {
		factor := w[11] * math.Pow(newD, -w[12]) * (math.Pow(1-R, w[13]) - 1)
		newS = math.Min(p.Stability, p.Stability*factor)
		newS = w[7]*p.Stability + (1-w[7])*newS
	}
	newS = math.Max(0.01, newS)

	r.NextStability = newS
	r.NextDifficulty = newD
	r.NextInterval = NextInterval(newS, 0.9)
	r.NextReviewAt = time.Now().Add(time.Duration(r.NextInterval) * 24 * time.Hour)

	if rating < 3 {
		r.NextState = 3
	} else if p.State == 0 || p.State == 1 {
		if r.NextInterval >= 7 {
			r.NextState = 2
		} else {
			r.NextState = 1
		}
	} else {
		r.NextState = 2
	}

	return r
}
