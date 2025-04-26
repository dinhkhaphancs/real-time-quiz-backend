package model

import (
	"time"

	"github.com/google/uuid"
)

// Participant represents a user who joins and participates in quizzes
type Participant struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	QuizID   uuid.UUID `json:"quizId" db:"quiz_id"`
	Score    int       `json:"score" db:"score"`
	JoinedAt time.Time `json:"joinedAt" db:"joined_at"`
}

// NewParticipant creates a new participant for a quiz
func NewParticipant(name string, quizID uuid.UUID) *Participant {
	return &Participant{
		ID:       uuid.New(),
		Name:     name,
		QuizID:   quizID,
		Score:    0,
		JoinedAt: time.Now(),
	}
}