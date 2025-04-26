package model

import (
	"time"

	"github.com/google/uuid"
)

// QuizStatus represents the status of a quiz
type QuizStatus string

const (
	// QuizStatusWaiting indicates the quiz is created but not started
	QuizStatusWaiting QuizStatus = "WAITING"
	// QuizStatusActive indicates the quiz is ongoing
	QuizStatusActive QuizStatus = "ACTIVE"
	// QuizStatusCompleted indicates the quiz has finished
	QuizStatusCompleted QuizStatus = "COMPLETED"
)

// Quiz represents a quiz that can be joined by users
type Quiz struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	CreatorID uuid.UUID  `json:"creatorId" db:"creator_id"`
	Status    QuizStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// QuizSession represents the current state of an active quiz
type QuizSession struct {
	QuizID                   uuid.UUID  `json:"quizId" db:"quiz_id"`
	CurrentQuestionID        *uuid.UUID `json:"currentQuestionId" db:"current_question_id"`
	Status                   QuizStatus `json:"status" db:"status"`
	StartedAt                *time.Time `json:"startedAt" db:"started_at"`
	EndedAt                  *time.Time `json:"endedAt" db:"ended_at"`
	CurrentQuestionStartedAt *time.Time `json:"currentQuestionStartedAt" db:"current_question_started_at"`
}

// NewQuiz creates a new quiz with the given title and creator ID
func NewQuiz(title string, creatorID uuid.UUID) *Quiz {
	return &Quiz{
		ID:        uuid.New(),
		Title:     title,
		CreatorID: creatorID,
		Status:    QuizStatusWaiting,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewQuizSession creates a new quiz session for the given quiz ID
func NewQuizSession(quizID uuid.UUID) *QuizSession {
	return &QuizSession{
		QuizID: quizID,
		Status: QuizStatusWaiting,
	}
}
