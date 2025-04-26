package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterUser handles user registration
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var request dto.UserRegisterRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	user, err := h.userService.Register(c, request.Name, request.Email, request.Password)
	if err != nil {
		// Check for common registration errors and return appropriate status
		if err.Error() == "email already in use" {
			response.WithError(c, http.StatusConflict, "Registration failed", err.Error())
			return
		}
		response.WithError(c, http.StatusInternalServerError, "Registration failed", err.Error())
		return
	}

	// Convert user model to DTO response
	userResponse := dto.UserResponseFromModel(user)
	response.WithSuccess(c, http.StatusCreated, response.MessageCreated, userResponse)
}

// LoginUser handles user login
func (h *UserHandler) LoginUser(c *gin.Context) {
	var request dto.UserLoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid request data", err.Error())
		return
	}

	user, err := h.userService.Login(c, request.Email, request.Password)
	if err != nil {
		// Return 401 for any login error
		response.WithError(c, http.StatusUnauthorized, "Authentication failed", "Invalid email or password")
		return
	}

	// Convert user model to DTO response
	userResponse := dto.UserResponseFromModel(user)
	response.WithSuccess(c, http.StatusOK, "Login successful", userResponse)
}
