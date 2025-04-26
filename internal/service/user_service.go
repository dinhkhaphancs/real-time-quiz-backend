package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/google/uuid"
)

// UserServiceImpl implements UserService interface
type UserServiceImpl struct {
	userRepo repository.UserRepository
	quizRepo repository.QuizRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, quizRepo repository.QuizRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
		quizRepo: quizRepo,
	}
}

// GetUserByID retrieves a user by ID and validates quiz access
func (s *UserServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID, quizID uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Validate that the user belongs to the requested quiz
	if user.QuizID != quizID {
		return nil, errors.New("user does not belong to this quiz")
	}

	return user, nil
}

// CreateUser creates a new user
func (s *UserServiceImpl) CreateUser(ctx context.Context, quizID uuid.UUID, name string, role string) (*model.User, error) {
	// Validate quiz exists
	_, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	// Validate role
	var userRole model.UserRole
	switch role {
	case string(model.UserRoleAdmin):
		userRole = model.UserRoleAdmin
	case string(model.UserRoleJoiner):
		userRole = model.UserRoleJoiner
	default:
		return nil, errors.New("invalid user role")
	}

	// Create new user
	user := model.NewUser(name, quizID, userRole)

	// Save to repository
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}