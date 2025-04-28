package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StateHandler handles state-related requests
type StateHandler struct {
	stateService service.StateService
}

// NewStateHandler creates a new state handler
func NewStateHandler(stateService service.StateService) *StateHandler {
	return &StateHandler{
		stateService: stateService,
	}
}

// GetQuizState handles requests to get the current state of a quiz
func (h *StateHandler) GetQuizState(c *gin.Context) {
	// Parse quiz ID from path
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", err.Error())
		return
	}

	// Get the quiz state
	quizState, err := h.stateService.GetQuizState(c.Request.Context(), quizID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to get quiz state", err.Error())
		return
	}

	// Return the state
	response.WithSuccess(c, http.StatusOK, "Get quiz states successfully", quizState)
}

// GetActiveParticipants handles requests to get the active participants for a quiz
func (h *StateHandler) GetActiveParticipants(c *gin.Context) {
	// Parse quiz ID from path
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", err.Error())
		return
	}

	// Get active participants
	participants, err := h.stateService.GetActiveParticipants(c.Request.Context(), quizID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to get active participants", err.Error())
		return
	}

	// Return the participants
	response.WithSuccess(c, http.StatusOK, "Success", participants)
}
