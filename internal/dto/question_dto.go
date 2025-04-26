package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// Question DTOs

// QuestionCreateRequest represents the request to add a new question
type QuestionCreateRequest struct {
	QuizID        string         `json:"quizId" binding:"required"`
	Text          string         `json:"text" binding:"required"`
	Options       []model.Option `json:"options" binding:"required,len=4"`
	CorrectAnswer string         `json:"correctAnswer" binding:"required"`
	TimeLimit     int            `json:"timeLimit" binding:"required,min=5,max=60"`
}

// QuestionResponse represents a question in API responses
type QuestionResponse struct {
	ID            uuid.UUID      `json:"id"`
	QuizID        uuid.UUID      `json:"quizId"`
	Text          string         `json:"text"`
	Options       []model.Option `json:"options"`
	CorrectAnswer string         `json:"correctAnswer,omitempty"` // May be hidden for certain responses
	TimeLimit     int            `json:"timeLimit"`
	Order         int            `json:"order"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// ParticipantResponse represents a participant in API responses
type ParticipantResponse struct {
	ID     uuid.UUID `json:"id"`
	QuizID uuid.UUID `json:"quizId"`
	Name   string    `json:"name"`
	Score  int       `json:"score"`
}

// QuestionAction represents the response for question actions (start/end)
type QuestionAction struct {
	Message string `json:"message"`
}

// QuestionResponseFromModel converts a Question model to a QuestionResponse
func QuestionResponseFromModel(model *model.Question, includeCorrectAnswer bool) QuestionResponse {
	response := QuestionResponse{
		ID:        model.ID,
		QuizID:    model.QuizID,
		Text:      model.Text,
		Options:   model.GetOptions(),
		TimeLimit: model.TimeLimit,
		Order:     model.Order,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}

	if includeCorrectAnswer {
		response.CorrectAnswer = model.CorrectAnswer
	}

	return response
}

// ParticipantResponseFromModel converts a Participant model to a ParticipantResponse
func ParticipantResponseFromModel(model *model.Participant) ParticipantResponse {
	return ParticipantResponse{
		ID:     model.ID,
		QuizID: model.QuizID,
		Name:   model.Name,
		Score:  model.Score,
	}
}
