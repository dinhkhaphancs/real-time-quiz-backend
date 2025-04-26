package service

import (
	"context"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// Errors
var (
	ErrQuizNotFound       = errors.New("quiz not found")
	ErrQuizAlreadyStarted = errors.New("quiz has already started")
	ErrQuizNotActive      = errors.New("quiz is not active")
	ErrNameTaken          = errors.New("name is already taken")
	ErrNameRequired       = errors.New("name is required")
)

// quizService implements QuizService interface
type quizService struct {
	quizRepo     repository.QuizRepository
	userRepo     repository.UserRepository
	questionRepo repository.QuestionRepository
	wsHub        *websocket.RedisHub
}

// NewQuizService creates a new quiz service
func NewQuizService(
	quizRepo repository.QuizRepository,
	userRepo repository.UserRepository,
	questionRepo repository.QuestionRepository,
	wsHub *websocket.RedisHub,
) QuizService {
	return &quizService{
		quizRepo:     quizRepo,
		userRepo:     userRepo,
		questionRepo: questionRepo,
		wsHub:        wsHub,
	}
}

// CreateQuiz creates a new quiz with an admin user
func (s *quizService) CreateQuiz(ctx context.Context, title string, adminName string) (*model.Quiz, *model.User, error) {
	if title == "" {
		return nil, nil, errors.New("title is required")
	}
	if adminName == "" {
		return nil, nil, ErrNameRequired
	}

	// Create a new admin user first with a temporary quiz ID
	tempID := uuid.New()
	admin := model.NewAdmin(adminName, tempID)

	// Create the quiz
	quiz := model.NewQuiz(title, admin.ID)

	// Update the admin's quiz ID to the actual quiz ID
	admin.QuizID = quiz.ID

	// Create quiz session
	session := model.NewQuizSession(quiz.ID)

	// Save to database
	if err := s.quizRepo.CreateQuiz(ctx, quiz); err != nil {
		return nil, nil, err
	}
	if err := s.quizRepo.CreateQuizSession(ctx, session); err != nil {
		return nil, nil, err
	}
	if err := s.userRepo.CreateUser(ctx, admin); err != nil {
		return nil, nil, err
	}

	return quiz, admin, nil
}

// GetQuiz retrieves a quiz by ID
func (s *quizService) GetQuiz(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, id)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	return quiz, nil
}

// StartQuiz starts a quiz session
func (s *quizService) StartQuiz(ctx context.Context, quizID uuid.UUID) error {
	// Get the quiz and session
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}

	if quiz.Status != model.QuizStatusWaiting {
		return ErrQuizAlreadyStarted
	}

	// Update quiz status
	if err := s.quizRepo.UpdateQuizStatus(ctx, quizID, model.QuizStatusActive); err != nil {
		return err
	}

	// Update session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.Status = model.QuizStatusActive
	session.StartedAt = &now

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz start event to all clients
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuizStart,
		Payload: map[string]interface{}{
			"quizId": quizID.String(),
			"title":  quiz.Title,
		},
	})

	return nil
}

// EndQuiz ends a quiz session
func (s *quizService) EndQuiz(ctx context.Context, quizID uuid.UUID) error {
	// Get the quiz and session
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}

	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Update quiz status
	if err := s.quizRepo.UpdateQuizStatus(ctx, quizID, model.QuizStatusCompleted); err != nil {
		return err
	}

	// Update session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.Status = model.QuizStatusCompleted
	session.EndedAt = &now

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz end event to all clients
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuizEnd,
		Payload: map[string]interface{}{
			"quizId": quizID.String(),
		},
	})

	return nil
}

// JoinQuiz allows a user to join a quiz
func (s *quizService) JoinQuiz(ctx context.Context, quizID uuid.UUID, name string) (*model.User, error) {
	if name == "" {
		return nil, ErrNameRequired
	}

	// Get the quiz
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}

	// Ensure quiz is in WAITING state
	if quiz.Status != model.QuizStatusWaiting {
		return nil, errors.New("cannot join a quiz that has already started")
	}

	// Check if name is already taken in this quiz
	users, err := s.userRepo.GetUsersByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Name == name {
			return nil, ErrNameTaken
		}
	}

	// Create new joiner
	joiner := model.NewJoiner(name, quizID)

	// Save to database
	if err := s.userRepo.CreateUser(ctx, joiner); err != nil {
		return nil, err
	}

	// Broadcast user joined event
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventUserJoined,
		Payload: map[string]interface{}{
			"userId": joiner.ID.String(),
			"name":   joiner.Name,
		},
	})

	return joiner, nil
}

// GetQuizSession retrieves the current state of a quiz
func (s *quizService) GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error) {
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return nil, err
	}
	return session, nil
}
