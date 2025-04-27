package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// AnswerSubmitRequest represents the request to submit an answer to a question
type AnswerSubmitRequest struct {
	ParticipantID   string   `json:"participantId" binding:"required"`
	QuestionID      string   `json:"questionId" binding:"required"`
	SelectedOptions []string `json:"selectedOptions" binding:"required,min=1"`
	TimeTaken       float64  `json:"timeTaken" binding:"required,min=0"`
}

// AnswerResponse represents an answer in API responses
type AnswerResponse struct {
	ID              uuid.UUID `json:"id"`
	ParticipantID   uuid.UUID `json:"participantId"`
	QuestionID      uuid.UUID `json:"questionId"`
	SelectedOptions []string  `json:"selectedOptions"`
	AnsweredAt      time.Time `json:"answeredAt"`
	TimeTaken       float64   `json:"timeTaken"`
	IsCorrect       bool      `json:"isCorrect"`
	Score           int       `json:"score"`
}

// AnswerStatsResponse represents statistics for answers to a question
type AnswerStatsResponse struct {
	OptionCounts   map[string]int `json:"optionCounts"`
	TotalAnswers   int            `json:"totalAnswers"`
	CorrectCount   int            `json:"correctCount"`
	IncorrectCount int            `json:"incorrectCount"`
}

// AnswerResponseFromModel converts an Answer model to an AnswerResponse
func AnswerResponseFromModel(answer *model.Answer) (AnswerResponse, error) {
	selectedOptions, err := answer.GetSelectedOptions()
	if err != nil {
		return AnswerResponse{}, err
	}

	return AnswerResponse{
		ID:              answer.ID,
		ParticipantID:   answer.ParticipantID,
		QuestionID:      answer.QuestionID,
		SelectedOptions: selectedOptions,
		AnsweredAt:      answer.AnsweredAt,
		TimeTaken:       answer.TimeTaken,
		IsCorrect:       answer.IsCorrect,
		Score:           answer.Score,
	}, nil
}
