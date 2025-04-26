package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuizHandler handles quiz-related HTTP requests
type QuizHandler struct {
	quizService     service.QuizService
	questionService service.QuestionService
}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler(quizService service.QuizService, questionService service.QuestionService) *QuizHandler {
	return &QuizHandler{
		quizService:     quizService,
		questionService: questionService,
	}
}

// CreateQuiz handles quiz creation
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var request struct {
		Title     string `json:"title" binding:"required"`
		AdminName string `json:"adminName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quiz, admin, err := h.quizService.CreateQuiz(c, request.Title, request.AdminName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"quiz": quiz,
		"admin": map[string]interface{}{
			"id":   admin.ID,
			"name": admin.Name,
			"role": admin.Role,
		},
	})
}

// GetQuiz retrieves quiz details
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Get questions for the quiz
	questions, _ := h.questionService.GetQuestions(c, id)

	// Get quiz session
	session, _ := h.quizService.GetQuizSession(c, id)

	c.JSON(http.StatusOK, gin.H{
		"quiz":      quiz,
		"questions": questions,
		"session":   session,
	})
}

// StartQuiz initiates a quiz session
func (h *QuizHandler) StartQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	if err := h.quizService.StartQuiz(c, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quiz started successfully"})
}

// EndQuiz ends a quiz session
func (h *QuizHandler) EndQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	if err := h.quizService.EndQuiz(c, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quiz ended successfully"})
}

// JoinQuiz allows a user to join a quiz
func (h *QuizHandler) JoinQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.quizService.JoinQuiz(c, id, request.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
			"role": user.Role,
		},
	})
}