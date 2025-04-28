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

// QuizPhase represents the current phase of an active quiz
type QuizPhase string

const (
	// QuizPhaseBetweenQuestions indicates the quiz is active but between questions
	QuizPhaseBetweenQuestions QuizPhase = "BETWEEN_QUESTIONS"
	// QuizPhaseQuestionActive indicates there is an active question being answered
	QuizPhaseQuestionActive QuizPhase = "QUESTION_ACTIVE"
	// QuizPhaseShowingResults indicates the question has ended and results are being shown
	QuizPhaseShowingResults QuizPhase = "SHOWING_RESULTS"
)

// Quiz represents a quiz that can be joined by users
type Quiz struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	CreatorID   uuid.UUID  `json:"creatorId" db:"creator_id"`
	Status      QuizStatus `json:"status" db:"status"`
	Code        string     `json:"code" db:"code"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

// QuizSession represents the current state of an active quiz
type QuizSession struct {
	QuizID                   uuid.UUID  `json:"quizId" db:"quiz_id"`
	CurrentQuestionID        *uuid.UUID `json:"currentQuestionId" db:"current_question_id"`
	Status                   QuizStatus `json:"status" db:"status"`
	CurrentPhase             QuizPhase  `json:"currentPhase" db:"current_phase"`
	StartedAt                *time.Time `json:"startedAt" db:"started_at"`
	EndedAt                  *time.Time `json:"endedAt" db:"ended_at"`
	CurrentQuestionStartedAt *time.Time `json:"currentQuestionStartedAt" db:"current_question_started_at"`
	CurrentQuestionEndedAt   *time.Time `json:"currentQuestionEndedAt" db:"current_question_ended_at"`
	NextQuestionID           *uuid.UUID `json:"nextQuestionId" db:"next_question_id"`
}

// NewQuiz creates a new quiz with the given title, description, and creator ID
func NewQuiz(title string, description string, creatorID uuid.UUID) *Quiz {
	return &Quiz{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		CreatorID:   creatorID,
		Status:      QuizStatusWaiting,
		Code:        generateQuizCode(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// generateQuizCode generates a random alphanumeric code for a quiz
func generateQuizCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed similar looking chars
	const codeLength = 6

	result := make([]byte, codeLength)
	for i := range result {
		u := uuid.New()
		result[i] = charset[int(u[i%16])%len(charset)]
	}
	return string(result)
}

// NewQuizSession creates a new quiz session for the given quiz ID
func NewQuizSession(quizID uuid.UUID) *QuizSession {
	return &QuizSession{
		QuizID:       quizID,
		Status:       QuizStatusWaiting,
		CurrentPhase: QuizPhaseBetweenQuestions,
	}
}
