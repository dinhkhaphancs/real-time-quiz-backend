package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/auth"
	"github.com/google/uuid"
)

// userServiceImpl implements UserService interface
type userServiceImpl struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) UserService {
	return &userServiceImpl{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *userServiceImpl) Register(ctx context.Context, name string, email string, password string) (*model.User, error) {
	// Validate inputs
	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	// Check if email is already registered
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email is already registered")
	}

	// Create new user
	user, err := model.NewUser(name, email, password)
	if err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	// Save to repository
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns user data
func (s *userServiceImpl) Login(ctx context.Context, email string, password string) (*model.User, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if !user.ComparePassword(password) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

// LoginWithToken authenticates a user and returns user data with JWT tokens
func (s *userServiceImpl) LoginWithToken(ctx context.Context, email string, password string) (*dto.UserLoginResponse, error) {
	// First authenticate the user using the existing login method
	user, err := s.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Create response with user data and tokens
	userResponse := dto.UserResponseFromModel(user)
	response := &dto.UserLoginResponse{
		User:         userResponse,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtManager.GetConfig().ExpirationTime.Seconds()),
	}

	return response, nil
}

// GetUserByID retrieves a user by ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
