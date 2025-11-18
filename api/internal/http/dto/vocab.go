package dto

type SelectVocabRequest struct {
    UserID  int64   `json:"user_id"`
    VocabIDs []int64 `json:"vocab_ids"`
}

type FeedbackRequest struct {
    UserID   int64  `json:"user_id"`
    VocabID  int64  `json:"vocab_id"`
    Difficulty string `json:"difficulty"` // easy|medium|hard
    Correct  bool   `json:"correct"`
}

