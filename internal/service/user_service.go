package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/google/uuid"
)

// userServiceImpl implements UserService interface
type userServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
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

// GetUserByID retrieves a user by ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
