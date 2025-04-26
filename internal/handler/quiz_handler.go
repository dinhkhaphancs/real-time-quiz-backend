package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuizHandler handles quiz-related HTTP requests
type QuizHandler struct {
	quizService        service.QuizService
	questionService    service.QuestionService
	userService        service.UserService
	participantService service.ParticipantService
}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler(
	quizService service.QuizService,
	questionService service.QuestionService,
	userService service.UserService,
	participantService service.ParticipantService,
) *QuizHandler {
	return &QuizHandler{
		quizService:        quizService,
		questionService:    questionService,
		userService:        userService,
		participantService: participantService,
	}
}

// CreateQuiz handles quiz creation by registered users
func (h *QuizHandler) CreateQuiz(c *gin.Context) {
	var request struct {
		Title  string `json:"title" binding:"required"`
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse user ID
	creatorID, err := uuid.Parse(request.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify user exists
	creator, err := h.userService.GetUserByID(c, creatorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Create the quiz
	quiz, err := h.quizService.CreateQuiz(c, request.Title, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"quiz": quiz,
		"creator": map[string]interface{}{
			"id":    creator.ID,
			"name":  creator.Name,
			"email": creator.Email,
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

	// Get participants for the quiz
	participants, _ := h.participantService.GetParticipantsByQuizID(c, id)

	c.JSON(http.StatusOK, gin.H{
		"quiz":         quiz,
		"questions":    questions,
		"session":      session,
		"participants": participants,
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

// JoinQuiz allows a user to join a quiz as a participant
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

	participant, err := h.participantService.JoinQuiz(c, id, request.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant": map[string]interface{}{
			"id":       participant.ID,
			"name":     participant.Name,
			"quizId":   participant.QuizID,
			"score":    participant.Score,
			"joinedAt": participant.JoinedAt,
		},
	})
}
