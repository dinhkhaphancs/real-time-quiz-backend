package dto

import (
	"time"

	"github.com/google/uuid"
)

// AnswerSubmitRequest represents the request payload for submitting an answer
type AnswerSubmitRequest struct {
	ParticipantID  string `json:"participantId" binding:"required"`
	QuestionID     string `json:"questionId" binding:"required"`
	SelectedOption string `json:"selectedOption" binding:"required"`
}

// AnswerResponse represents the response payload for an answer
type AnswerResponse struct {
	ID             uuid.UUID `json:"id"`
	ParticipantID  uuid.UUID `json:"participantId"`
	QuestionID     uuid.UUID `json:"questionId"`
	SelectedOption string    `json:"selectedOption"`
	IsCorrect      bool      `json:"isCorrect"`
	Score          int       `json:"score"`
	AnsweredAt     time.Time `json:"answeredAt"`
	TimeTaken      float64   `json:"timeTaken"` // in milliseconds
}

// AnswerBasicResponse represents a simplified answer response payload
type AnswerBasicResponse struct {
	ID             uuid.UUID `json:"id"`
	SelectedOption string    `json:"selectedOption"`
	IsCorrect      bool      `json:"isCorrect"`
	Score          int       `json:"score"`
}
