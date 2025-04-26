package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ParticipantHandler handles participant-related HTTP requests
type ParticipantHandler struct {
	participantService service.ParticipantService
	quizService        service.QuizService
}

// NewParticipantHandler creates a new participant handler
func NewParticipantHandler(
	participantService service.ParticipantService,
	quizService service.QuizService,
) *ParticipantHandler {
	return &ParticipantHandler{
		participantService: participantService,
		quizService:        quizService,
	}
}

// GetParticipant retrieves a specific participant by ID
func (h *ParticipantHandler) GetParticipant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid participant ID", "The provided participant ID is not valid")
		return
	}

	participant, err := h.participantService.GetParticipantByID(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Participant not found", err.Error())
		return
	}

	participantResponse := dto.ParticipantResponseFromModel(participant)
	response.WithSuccess(c, http.StatusOK, response.MessageFetched, map[string]interface{}{
		"participant": participantResponse,
	})
}

// GetParticipantsByQuiz retrieves all participants for a specific quiz
func (h *ParticipantHandler) GetParticipantsByQuiz(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	// Check if quiz exists
	_, err = h.quizService.GetQuiz(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", "The specified quiz could not be found")
		return
	}

	participants, err := h.participantService.GetParticipantsByQuizID(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to retrieve participants", err.Error())
		return
	}

	var participantResponses []dto.ParticipantResponse
	for _, p := range participants {
		participantResponses = append(participantResponses, dto.ParticipantResponseFromModel(p))
	}

	response.WithSuccess(c, http.StatusOK, response.MessageListFetched, map[string]interface{}{
		"participants": participantResponses,
	})
}

// RemoveParticipant removes a participant from a quiz
func (h *ParticipantHandler) RemoveParticipant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid participant ID", "The provided participant ID is not valid")
		return
	}

	// First get the participant to make sure it exists and to get the quiz ID
	participant, err := h.participantService.GetParticipantByID(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Participant not found", err.Error())
		return
	}

	// Now check if the quiz has already started
	quiz, err := h.quizService.GetQuiz(c, participant.QuizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", "The quiz for this participant could not be found")
		return
	}

	if quiz.Status != "waiting" {
		response.WithError(c, http.StatusBadRequest, "Cannot remove participant", "Participants cannot be removed once the quiz has started")
		return
	}

	// This will be implemented in the next step in participant_service.go
	err = h.participantService.RemoveParticipant(c, id)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to remove participant", err.Error())
		return
	}

	response.WithSuccess(c, http.StatusOK, "Participant successfully removed", nil)
}
