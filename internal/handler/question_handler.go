package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/middleware"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// QuestionHandler handles question-related HTTP requests
type QuestionHandler struct {
	questionService service.QuestionService
	quizService     service.QuizService
}

// NewQuestionHandler creates a new question handler
func NewQuestionHandler(questionService service.QuestionService, quizService service.QuizService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
		quizService:     quizService,
	}
}

// AddQuestion adds a new question to a quiz
func (h *QuestionHandler) AddQuestion(c *gin.Context) {
	var request dto.QuestionCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Parse quiz ID
	quizID, err := uuid.Parse(request.QuizID)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	// Get authenticated user ID from JWT context
	userID := middleware.GetAuthUserID(c)
	if userID == uuid.Nil {
		response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	// Verify quiz ownership
	quiz, err := h.quizService.GetQuiz(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can add questions")
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
		response.WithError(c, http.StatusBadRequest, "Failed to create question", err.Error())
		return
	}

	questionResponse := dto.QuestionResponseFromModel(question, true)
	response.WithSuccess(c, http.StatusCreated, response.MessageCreated, map[string]interface{}{
		"question": questionResponse,
	})
}

// GetQuestions retrieves all questions for a quiz
func (h *QuestionHandler) GetQuestions(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	questions, err := h.questionService.GetQuestions(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to retrieve questions", err.Error())
		return
	}

	var responseQuestions []dto.QuestionResponse
	for _, q := range questions {
		responseQuestions = append(responseQuestions, dto.QuestionResponseFromModel(q, true))
	}

	response.WithSuccess(c, http.StatusOK, response.MessageListFetched, map[string]interface{}{
		"questions": responseQuestions,
	})
}

// GetQuestion retrieves a specific question
func (h *QuestionHandler) GetQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Question not found", err.Error())
		return
	}

	questionResponse := dto.QuestionResponseFromModel(question, true)
	response.WithSuccess(c, http.StatusOK, response.MessageFetched, map[string]interface{}{
		"question": questionResponse,
	})
}

// StartQuestion begins a question in a quiz
func (h *QuestionHandler) StartQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	// Get authenticated user ID from JWT context
	userID := middleware.GetAuthUserID(c)
	if userID == uuid.Nil {
		response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	// Get the question to determine quiz ID
	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Question not found", err.Error())
		return
	}

	// Verify quiz ownership
	quiz, err := h.quizService.GetQuiz(c, question.QuizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can start questions")
		return
	}

	// Start the question
	if err := h.questionService.StartQuestion(c, question.QuizID, id); err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to start question", err.Error())
		return
	}

	questionAction := dto.QuestionAction{
		Message: "Question started successfully",
	}
	response.WithSuccess(c, http.StatusOK, "Question started successfully", questionAction)
}

// EndQuestion ends the current question in a quiz
func (h *QuestionHandler) EndQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid question ID", "The provided question ID is not valid")
		return
	}

	// Get authenticated user ID from JWT context
	userID := middleware.GetAuthUserID(c)
	if userID == uuid.Nil {
		response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	// Get the question to determine quiz ID
	question, err := h.questionService.GetQuestion(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Question not found", err.Error())
		return
	}

	// Verify quiz ownership
	quiz, err := h.quizService.GetQuiz(c, question.QuizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can end questions")
		return
	}

	// End the question
	if err := h.questionService.EndQuestion(c, question.QuizID); err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to end question", err.Error())
		return
	}

	questionAction := dto.QuestionAction{
		Message: "Question ended successfully",
	}
	response.WithSuccess(c, http.StatusOK, "Question ended successfully", questionAction)
}

// GetNextQuestion retrieves the next question in sequence
func (h *QuestionHandler) GetNextQuestion(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	question, err := h.questionService.GetNextQuestion(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "No next question available", err.Error())
		return
	}

	// Don't include the correct answer when returning the next question
	questionResponse := dto.QuestionResponseFromModel(question, false)
	response.WithSuccess(c, http.StatusOK, "Next question fetched successfully", map[string]interface{}{
		"question": questionResponse,
	})
}
