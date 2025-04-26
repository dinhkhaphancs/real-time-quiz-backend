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

// SubmitAnswer handles user answer submissions
func (h *AnswerHandler) SubmitAnswer(c *gin.Context) {
	var request struct {
		UserID        string `json:"userId" binding:"required"`
		QuestionID    string `json:"questionId" binding:"required"`
		SelectedOption string `json:"selectedOption" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse UUIDs
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	questionID, err := uuid.Parse(request.QuestionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Submit the answer
	answer, err := h.answerService.SubmitAnswer(c, userID, questionID, request.SelectedOption)
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

// GetUserAnswer retrieves a specific user's answer
func (h *AnswerHandler) GetUserAnswer(c *gin.Context) {
	userIDStr := c.Param("userId")
	questionIDStr := c.Param("questionId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	questionID, err := uuid.Parse(questionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	answer, err := h.answerService.GetUserAnswer(c, userID, questionID)
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