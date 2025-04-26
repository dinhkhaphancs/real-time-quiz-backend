package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnswerHandler handles answer-related HTTP requests
type AnswerHandler struct {
	answerService service.AnswerService
}

// NewAnswerHandler creates a new answer handler
func NewAnswerHandler(answerService service.AnswerService) *AnswerHandler {
	return &AnswerHandler{
		answerService: answerService,
	}
}

// SubmitAnswer handles participant answer submissions
func (h *AnswerHandler) SubmitAnswer(c *gin.Context) {
	var request dto.AnswerSubmitRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Parse UUIDs
	participantID, err := uuid.Parse(request.ParticipantID)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid participant ID", "The provided participant ID is not valid")
		return
	}

	questionID, err := uuid.Parse(request.QuestionID)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	// Submit the answer
	answer, err := h.answerService.SubmitAnswer(c, participantID, questionID, request.SelectedOption)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to submit answer", err.Error())
		return
	}

	// Create a basic answer response that doesn't include sensitive fields
	answerResponse := dto.AnswerBasicResponse{
		ID:             answer.ID,
		SelectedOption: answer.SelectedOption,
		IsCorrect:      answer.IsCorrect,
		Score:          answer.Score,
	}

	response.WithSuccess(c, http.StatusCreated, "Answer submitted successfully", map[string]interface{}{
		"answer": answerResponse,
	})
}

// GetAnswerStats retrieves statistics for a question
func (h *AnswerHandler) GetAnswerStats(c *gin.Context) {
	questionIDStr := c.Param("questionId")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	stats, err := h.answerService.GetAnswerStats(c, questionID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to retrieve answer statistics", err.Error())
		return
	}

	response.WithSuccess(c, http.StatusOK, "Answer statistics retrieved successfully", map[string]interface{}{
		"stats": stats,
	})
}

// GetParticipantAnswer retrieves a specific participant's answer
func (h *AnswerHandler) GetParticipantAnswer(c *gin.Context) {
	participantIDStr := c.Param("participantId")
	questionIDStr := c.Param("questionId")

	participantID, err := uuid.Parse(participantIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid participant ID", "The provided participant ID is not valid")
		return
	}

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	answer, err := h.answerService.GetParticipantAnswer(c, participantID, questionID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Answer not found", err.Error())
		return
	}

	// Convert to full answer response DTO
	answerResponse := dto.AnswerResponse{
		ID:             answer.ID,
		ParticipantID:  answer.ParticipantID,
		QuestionID:     answer.QuestionID,
		SelectedOption: answer.SelectedOption,
		IsCorrect:      answer.IsCorrect,
		Score:          answer.Score,
		AnsweredAt:     answer.AnsweredAt,
		TimeTaken:      answer.TimeTaken,
	}

	response.WithSuccess(c, http.StatusOK, response.MessageFetched, map[string]interface{}{
		"answer": answerResponse,
	})
}
