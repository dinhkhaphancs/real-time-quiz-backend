package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// Question DTOs

// OptionCreateData represents an option for a question in API requests
type OptionCreateData struct {
	Text      string `json:"text" binding:"required"`
	IsCorrect bool   `json:"isCorrect"`
}

// QuestionCreateRequest represents the request to add a new question
type QuestionCreateRequest struct {
	QuizID       string             `json:"quizId" binding:"required"`
	Text         string             `json:"text" binding:"required"`
	Options      []OptionCreateData `json:"options" binding:"required,min=2,max=10"`
	QuestionType string             `json:"questionType" binding:"required,oneof=SINGLE_CHOICE MULTIPLE_CHOICE"`
	TimeLimit    int                `json:"timeLimit" binding:"required,min=5,max=60"`
}

// QuestionCreateData represents a question to be created as part of a quiz
type QuestionCreateData struct {
	Text         string             `json:"text" binding:"required"`
	Options      []OptionCreateData `json:"options" binding:"required,min=2,max=10"`
	QuestionType string             `json:"questionType" binding:"required,oneof=SINGLE_CHOICE MULTIPLE_CHOICE"`
	TimeLimit    int                `json:"timeLimit" binding:"required,min=5,max=60"`
}

// QuestionUpdateData represents question data for updating a quiz
type QuestionUpdateData struct {
	ID           *string      `json:"id"`
	Text         string       `json:"text" binding:"required"`
	TimeLimit    int          `json:"timeLimit" binding:"required"`
	QuestionType string       `json:"questionType" binding:"required,oneof=SINGLE_CHOICE MULTIPLE_CHOICE"`
	Options      []OptionData `json:"options" binding:"required"`
}

// OptionResponse represents an option in API responses
type OptionResponse struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	IsCorrect bool      `json:"isCorrect,omitempty"` // May be hidden for certain responses
}

// QuestionResponse represents a question in API responses
type QuestionResponse struct {
	ID           uuid.UUID        `json:"id"`
	QuizID       uuid.UUID        `json:"quizId"`
	Text         string           `json:"text"`
	Options      []OptionResponse `json:"options"`
	QuestionType string           `json:"questionType"`
	TimeLimit    int              `json:"timeLimit"`
	Order        int              `json:"order"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
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
func QuestionResponseFromModel(model *model.Question, includeCorrectAnswers bool) QuestionResponse {
	response := QuestionResponse{
		ID:           model.ID,
		QuizID:       model.QuizID,
		Text:         model.Text,
		QuestionType: string(model.QuestionType),
		TimeLimit:    model.TimeLimit,
		Order:        model.Order,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}

	// Convert options
	options := make([]OptionResponse, len(model.Options))
	for i, opt := range model.Options {
		optResponse := OptionResponse{
			ID:   opt.ID,
			Text: opt.Text,
		}

		// Only include correct answer information if requested
		if includeCorrectAnswers {
			optResponse.IsCorrect = opt.IsCorrect
		}

		options[i] = optResponse
	}

	response.Options = options
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
