package dto

import (
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// User DTOs

// UserRegisterRequest represents the request to register a new user
type UserRegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserLoginRequest represents the request to login a user
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

// UserResponseFromModel converts a User model to a UserResponse
func UserResponseFromModel(model *model.User) UserResponse {
	return UserResponse{
		ID:    model.ID,
		Name:  model.Name,
		Email: model.Email,
	}
}
