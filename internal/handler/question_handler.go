package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuestionHandler handles question-related HTTP requests
type QuestionHandler struct {
	questionService service.QuestionService
}

// NewQuestionHandler creates a new question handler
func NewQuestionHandler(questionService service.QuestionService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
	}
}

// AddQuestion adds a new question to a quiz
func (h *QuestionHandler) AddQuestion(c *gin.Context) {
	var request struct {
		QuizID       string        `json:"quizId" binding:"required"`
		Text         string        `json:"text" binding:"required"`
		Options      []model.Option `json:"options" binding:"required,len=4"`
		CorrectAnswer string        `json:"correctAnswer" binding:"required"`
		TimeLimit    int           `json:"timeLimit" binding:"required,min=5,max=60"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse quiz ID
	quizID, err := uuid.Parse(request.QuizID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	// Create the question
	question, err := h.questionService.AddQuestion(
		c,
		quizID,
		request.Text,
		request.Options,
		request.CorrectAnswer,
		request.TimeLimit,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"question": question})
}

// GetQuestions retrieves all questions for a quiz
func (h *QuestionHandler) GetQuestions(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	questions, err := h.questionService.GetQuestions(c, quizID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

// GetQuestion retrieves a specific question
func (h *QuestionHandler) GetQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"question": question})
}

// StartQuestion begins a question in a quiz
func (h *QuestionHandler) StartQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Get the question to determine quiz ID
	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Start the question
	if err := h.questionService.StartQuestion(c, question.QuizID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question started successfully"})
}

// EndQuestion ends the current question in a quiz
func (h *QuestionHandler) EndQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	// Get the question to determine quiz ID
	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// End the question
	if err := h.questionService.EndQuestion(c, question.QuizID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Question ended successfully"})
}

// GetNextQuestion retrieves the next question in sequence
func (h *QuestionHandler) GetNextQuestion(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	question, err := h.questionService.GetNextQuestion(c, quizID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"question": question})
}