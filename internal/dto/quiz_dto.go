package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// Quiz DTOs

// QuizCreateRequest represents the request to create a new quiz
type QuizCreateRequest struct {
	Title       string               `json:"title" binding:"required"`
	Description string               `json:"description"`
	Questions   []QuestionCreateData `json:"questions" binding:"required"`
}

// QuizUpdateRequest represents the request to update an existing quiz
type QuizUpdateRequest struct {
	ID          string               `json:"id" binding:"required"`
	Title       string               `json:"title" binding:"required"`
	Description string               `json:"description"`
	Questions   []QuestionUpdateData `json:"questions"`
}

// QuizJoinByCodeRequest represents the request to join a quiz using a code
type QuizJoinByCodeRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// OptionData represents an option for updating a question
type OptionData struct {
	ID           *string `json:"id"`
	Text         string  `json:"text" binding:"required"`
	IsCorrect    bool    `json:"isCorrect"`
	DisplayOrder int     `json:"displayOrder"`
}

// QuizResponse represents a quiz in API responses
type QuizResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CreatorID   uuid.UUID `json:"creatorId"`
	Status      string    `json:"status"`
	Code        string    `json:"code"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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
	model.Quiz
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
		ID:          model.ID,
		Title:       model.Title,
		Description: model.Description,
		CreatorID:   model.CreatorID,
		Status:      string(model.Status),
		Code:        model.Code,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
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
