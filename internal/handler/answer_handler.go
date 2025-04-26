package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
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
	var request struct {
		ParticipantID  string `json:"participantId" binding:"required"`
		QuestionID     string `json:"questionId" binding:"required"`
		SelectedOption string `json:"selectedOption" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse UUIDs
	participantID, err := uuid.Parse(request.ParticipantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID"})
		return
	}

	questionID, err := uuid.Parse(request.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Submit the answer
	answer, err := h.answerService.SubmitAnswer(c, participantID, questionID, request.SelectedOption)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"answer": map[string]interface{}{
			"id":             answer.ID,
			"selectedOption": answer.SelectedOption,
			"isCorrect":      answer.IsCorrect,
			"score":          answer.Score,
		},
	})
}

// GetAnswerStats retrieves statistics for a question
func (h *AnswerHandler) GetAnswerStats(c *gin.Context) {
	questionIDStr := c.Param("questionId")
	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	stats, err := h.answerService.GetAnswerStats(c, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetParticipantAnswer retrieves a specific participant's answer
func (h *AnswerHandler) GetParticipantAnswer(c *gin.Context) {
	participantIDStr := c.Param("participantId")
	questionIDStr := c.Param("questionId")

	participantID, err := uuid.Parse(participantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID"})
		return
	}

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	answer, err := h.answerService.GetParticipantAnswer(c, participantID, questionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer": map[string]interface{}{
			"id":             answer.ID,
			"selectedOption": answer.SelectedOption,
			"isCorrect":      answer.IsCorrect,
			"score":          answer.Score,
			"answeredAt":     answer.AnsweredAt,
			"timeTaken":      answer.TimeTaken,
		},
	})
}