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
	var request dto.QuizCreateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Get the authenticated user ID from the JWT context
	creatorID := middleware.GetAuthUserID(c)
	if creatorID == uuid.Nil {
		response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	// Verify user exists (optional since JWT middleware already validated the token)
	creator, err := h.userService.GetUserByID(c, creatorID)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "User not found", "The authenticated user could not be found")
		return
	}

	// Create the quiz with questions
	quiz, err := h.quizService.CreateQuizWithQuestions(c, request.Title, request.Description, creatorID, request.Questions)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to create quiz", err.Error())
		return
	}

	// Get the created questions for the response
	questions, _ := h.questionService.GetQuestions(c, quiz.ID)

	// Convert to DTOs
	quizResponse := dto.QuizResponseFromModel(quiz)
	creatorResponse := dto.CreatorResponseFromModel(creator)

	var questionResponses []dto.QuestionResponse
	for _, q := range questions {
		questionResponses = append(questionResponses, dto.QuestionResponseFromModel(q, true))
	}

	// Create response data
	data := map[string]interface{}{
		"quiz":      quizResponse,
		"creator":   creatorResponse,
		"questions": questionResponses,
	}

	response.WithSuccess(c, http.StatusCreated, response.MessageCreated, data)
}

// GetQuiz retrieves quiz details
func (h *QuizHandler) GetQuiz(c *gin.Context) {
	// Get auth user ID from JWT context
	authUserId := middleware.GetAuthUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Get creator
	creator, err := h.userService.GetUserByID(c, quiz.CreatorID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to get quiz creator", "Could not retrieve quiz creator information")
		return
	}

	// Get questions for the quiz
	questions, _ := h.questionService.GetQuestions(c, id)

	// Get quiz session
	session, _ := h.quizService.GetQuizSession(c, id)

	// Get participants for the quiz
	participants, _ := h.participantService.GetParticipantsByQuizID(c, id)

	// Convert to DTOs
	creatorResponse := dto.CreatorResponseFromModel(creator)

	var questionResponses []dto.QuestionResponse
	for _, q := range questions {
		// Include correct answers for authenticated users, or for non-authenticated users only if the quiz has ended
		includeAnswer := true
		if authUserId == uuid.Nil {
			includeAnswer = session != nil && session.EndedAt != nil
		}
		questionResponses = append(questionResponses, dto.QuestionResponseFromModel(q, includeAnswer))
	}

	var participantResponses []dto.ParticipantResponse
	for _, p := range participants {
		participantResponses = append(participantResponses, dto.ParticipantResponseFromModel(p))
	}

	// Create quiz details response
	quizDetails := dto.QuizDetails{
		Quiz:         *quiz,
		Creator:      creatorResponse,
		Questions:    questionResponses,
		Participants: participantResponses,
	}

	if session != nil {
		sessionResponse := dto.QuizSession{
			QuizID:            session.QuizID,
			Status:            string(session.Status),
			CurrentQuestionID: session.CurrentQuestionID,
			StartedAt:         session.StartedAt,
			EndedAt:           session.EndedAt,
		}
		quizDetails.Session = &sessionResponse
	}

	response.WithSuccess(c, http.StatusOK, response.MessageFetched, quizDetails)
}

// StartQuiz initiates a quiz session
func (h *QuizHandler) StartQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	// Verify ownership by getting the quiz first
	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can start this quiz")
		return
	}

	if err := h.quizService.StartQuiz(c, id); err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to start quiz", err.Error())
		return
	}

	quizAction := dto.QuizAction{
		Message: "Quiz started successfully",
	}
	response.WithSuccess(c, http.StatusOK, "Quiz started successfully", quizAction)
}

// EndQuiz ends a quiz session
func (h *QuizHandler) EndQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	// Verify ownership by getting the quiz first
	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can end this quiz")
		return
	}

	if err := h.quizService.EndQuiz(c, id); err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to end quiz", err.Error())
		return
	}

	quizAction := dto.QuizAction{
		Message: "Quiz ended successfully",
	}
	response.WithSuccess(c, http.StatusOK, "Quiz ended successfully", quizAction)
}

// JoinQuiz allows a user to join a quiz as a participant
func (h *QuizHandler) JoinQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	var request dto.QuizJoinRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	participant, err := h.participantService.JoinQuiz(c, id, request.Name)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to join quiz", err.Error())
		return
	}

	participantResponse := dto.ParticipantResponseFromModel(participant)
	response.WithSuccess(c, http.StatusOK, "Successfully joined quiz", participantResponse)
}

// JoinQuizByCode allows a user to join a quiz using a code
func (h *QuizHandler) JoinQuizByCode(c *gin.Context) {
	var request struct {
		Code string `json:"code" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	participant, err := h.participantService.JoinQuizByCode(c, request.Code, request.Name)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to join quiz", err.Error())
		return
	}

	participantResponse := dto.ParticipantResponseFromModel(participant)
	response.WithSuccess(c, http.StatusOK, "Successfully joined quiz", participantResponse)
}

// GetCurrentUserQuizzes retrieves all quizzes created by the authenticated user
func (h *QuizHandler) GetCurrentUserQuizzes(c *gin.Context) {
	// Get authenticated user ID from JWT context
	userID := middleware.GetAuthUserID(c)
	if userID == uuid.Nil {
		response.WithError(c, http.StatusUnauthorized, "Unauthorized", "Authentication required")
		return
	}

	// Get quizzes created by the user
	quizzes, err := h.quizService.GetQuizzesByCreatorID(c, userID)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to get quizzes", err.Error())
		return
	}

	// Convert to DTOs
	var quizResponses []dto.QuizResponse
	for _, quiz := range quizzes {
		quizResponses = append(quizResponses, dto.QuizResponseFromModel(quiz))
	}

	response.WithSuccess(c, http.StatusOK, "Quizzes retrieved successfully", quizResponses)
}

// UpdateQuiz handles updating an existing quiz
func (h *QuizHandler) UpdateQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	// Get the quiz to verify ownership
	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can update this quiz")
		return
	}

	// Parse request body
	var request dto.QuizUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	// Update the quiz with questions
	updatedQuiz, err := h.quizService.UpdateQuizWithQuestions(c, id, request.Title, request.Description, request.Questions)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to update quiz", err.Error())
		return
	}

	// Get the updated questions for the response
	questions, _ := h.questionService.GetQuestions(c, id)

	// Convert to DTOs
	quizResponse := dto.QuizResponseFromModel(updatedQuiz)

	var questionResponses []dto.QuestionResponse
	for _, q := range questions {
		questionResponses = append(questionResponses, dto.QuestionResponseFromModel(q, true))
	}

	// Create response data
	data := map[string]interface{}{
		"quiz":      quizResponse,
		"questions": questionResponses,
	}

	response.WithSuccess(c, http.StatusOK, "Quiz updated successfully", data)
}

// DeleteQuiz handles deleting a quiz
func (h *QuizHandler) DeleteQuiz(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	// Get the quiz to verify ownership
	quiz, err := h.quizService.GetQuiz(c, id)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", err.Error())
		return
	}

	// Check if the authenticated user is the quiz creator
	if quiz.CreatorID != userID {
		response.WithError(c, http.StatusForbidden, "Access denied", "Only the quiz creator can delete this quiz")
		return
	}

	// Delete the quiz
	if err := h.quizService.DeleteQuiz(c, id); err != nil {
		response.WithError(c, http.StatusBadRequest, "Failed to delete quiz", err.Error())
		return
	}

	response.WithSuccess(c, http.StatusOK, "Quiz deleted successfully", nil)
}
