package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// Quiz DTOs

// QuizCreateRequest represents the request to create a new quiz
type QuizCreateRequest struct {
	Title string `json:"title" binding:"required"`
}

// QuizResponse represents a quiz in API responses
type QuizResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatorID uuid.UUID `json:"creatorId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreatorResponse represents a quiz creator in API responses
type CreatorResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// QuizSession represents the state of a quiz session
type QuizSession struct {
	QuizID            uuid.UUID  `json:"quizId"`
	Status            string     `json:"status"`
	CurrentQuestionID *uuid.UUID `json:"currentQuestionId,omitempty"`
	StartedAt         *time.Time `json:"startedAt,omitempty"`
	EndedAt           *time.Time `json:"endedAt,omitempty"`
}

// QuizDetails represents complete quiz details including questions and participants
type QuizDetails struct {
	Quiz         QuizResponse          `json:"quiz"`
	Creator      CreatorResponse       `json:"creator,omitempty"`
	Questions    []QuestionResponse    `json:"questions,omitempty"`
	Session      *QuizSession          `json:"session,omitempty"`
	Participants []ParticipantResponse `json:"participants,omitempty"`
}

// QuizAction represents the response after a quiz action (start/end)
type QuizAction struct {
	Message string `json:"message"`
}

// QuizJoinRequest represents the request to join a quiz
type QuizJoinRequest struct {
	Name string `json:"name" binding:"required"`
}

// QuizResponseFromModel converts a Quiz model to a QuizResponse
func QuizResponseFromModel(model *model.Quiz) QuizResponse {
	return QuizResponse{
		ID:        model.ID,
		Title:     model.Title,
		CreatorID: model.CreatorID,
		Status:    string(model.Status),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// CreatorResponseFromModel converts a User model to a CreatorResponse
func CreatorResponseFromModel(model *model.User) CreatorResponse {
	return CreatorResponse{
		ID:    model.ID,
		Name:  model.Name,
		Email: model.Email,
	}
}
