// Package dto holds request/response bodies of the HTTP API.
package dto

import (
	"encoding/json"
	"time"
)

// ---------------------------------------------------------------- auth

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
	DisplayName string `json:"display_name" binding:"max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      any       `json:"user"`
}

// ---------------------------------------------------------------- content

type CreateSubjectRequest struct {
	Slug        string `json:"slug" binding:"required,max=64"`
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description"`
}

type CreateDeckRequest struct {
	SubjectID   int64  `json:"subject_id" binding:"required"`
	Slug        string `json:"slug" binding:"required,max=64"`
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description"`
	DefaultKind string `json:"default_kind" binding:"required,oneof=vocab phrasal concept mcq"`
}

type CardRequest struct {
	DeckID  int64           `json:"deck_id"`
	Kind    string          `json:"kind"`
	Front   string          `json:"front" binding:"required,max=2000"`
	Payload json.RawMessage `json:"payload"`
}

type GenerateRequest struct {
	Count int `json:"count"`
}

// ---------------------------------------------------------------- study

type EnrollRequest struct {
	CardIDs []int64 `json:"card_ids" binding:"required,min=1,max=500"`
}

type ReviewRequest struct {
	CardID     int64 `json:"card_id" binding:"required"`
	Rating     int   `json:"rating" binding:"required,min=1,max=4"` // 1 again, 2 hard, 3 good, 4 easy
	DurationMs *int  `json:"duration_ms"`
}

type SuspendRequest struct {
	CardID    int64 `json:"card_id" binding:"required"`
	Suspended bool  `json:"suspended"`
}

type QuizAnswerRequest struct {
	CardID   int64 `json:"card_id" binding:"required"`
	ChoiceID int64 `json:"choice_id"`
}

// ---------------------------------------------------------------- devices

type DeviceRegisterRequest struct {
	Name     string `json:"name" binding:"max=100"`
	Firmware string `json:"firmware" binding:"max=64"`
}

type DeviceRegisterResponse struct {
	DeviceID      int64     `json:"device_id"`
	Token         string    `json:"token"` // store in NVS; sent as "Authorization: Bearer <token>"
	PairCode      string    `json:"pair_code"`
	PairExpiresAt time.Time `json:"pair_expires_at"`
}

type PairRequest struct {
	Code string `json:"code" binding:"required,len=6,numeric"`
	Name string `json:"name" binding:"max=100"`
}

type RenameDeviceRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type DeviceReview struct {
	ClientReviewID string    `json:"client_review_id" binding:"required,max=64"`
	CardID         int64     `json:"card_id" binding:"required"`
	Rating         int       `json:"rating" binding:"required,min=1,max=4"`
	ReviewedAt     time.Time `json:"reviewed_at"`
	DurationMs     *int      `json:"duration_ms"`
	// Quiz: answered with the device's Quiz button (not Study)
	Quiz bool `json:"quiz"`
}

type DeviceReviewsRequest struct {
	Reviews []DeviceReview `json:"reviews" binding:"required,max=500,dive"`
}

// ---------------------------------------------------------------- writing

type WritingRequest struct {
	Prompt string `json:"prompt" binding:"required,max=2000"`
	Text   string `json:"text" binding:"required,min=20,max=8000"`
}
