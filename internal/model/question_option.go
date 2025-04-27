package model

import (
	"time"

	"github.com/google/uuid"
)

// QuestionOption represents an option for a quiz question with database mapping
type QuestionOption struct {
	ID           uuid.UUID `json:"id" db:"id"`
	QuestionID   uuid.UUID `json:"questionId" db:"question_id"`
	Text         string    `json:"text" db:"text"`
	IsCorrect    bool      `json:"isCorrect" db:"is_correct"`
	DisplayOrder int       `json:"displayOrder" db:"display_order"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// NewQuestionOption creates a new question option
func NewQuestionOption(questionID uuid.UUID, text string, isCorrect bool, displayOrder int) *QuestionOption {
	return &QuestionOption{
		ID:           uuid.New(),
		QuestionID:   questionID,
		Text:         text,
		IsCorrect:    isCorrect,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
