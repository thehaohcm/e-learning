package dto

type QuizAttemptRequest struct {
    UserID     int64 `json:"user_id"`
    QuestionID int64 `json:"question_id"`
    AnswerIdx  int   `json:"answer_idx"`
}

