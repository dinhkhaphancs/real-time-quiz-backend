package model

import (
	"time"

	"github.com/google/uuid"
)

// Answer represents a user's answer to a question
type Answer struct {
	ID             uuid.UUID `json:"id" db:"id"`
	UserID         uuid.UUID `json:"userId" db:"user_id"`
	QuestionID     uuid.UUID `json:"questionId" db:"question_id"`
	SelectedOption string    `json:"selectedOption" db:"selected_option"` // A, B, C, or D
	AnsweredAt     time.Time `json:"answeredAt" db:"answered_at"`
	TimeTaken      float64   `json:"timeTaken" db:"time_taken"` // Time taken in seconds
	IsCorrect      bool      `json:"isCorrect" db:"is_correct"`
	Score          int       `json:"score" db:"score"`
}

// NewAnswer creates a new answer record
func NewAnswer(userID, questionID uuid.UUID, selectedOption string, timeTaken float64, isCorrect bool) *Answer {
	score := 0
	if isCorrect {
		score = 100
	}

	return &Answer{
		ID:             uuid.New(),
		UserID:         userID,
		QuestionID:     questionID,
		SelectedOption: selectedOption,
		AnsweredAt:     time.Now(),
		TimeTaken:      timeTaken,
		IsCorrect:      isCorrect,
		Score:          score,
	}
}
