package service

import (
	"context"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
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
)

// quizServiceImpl implements QuizService interface
type quizServiceImpl struct {
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
	return &quizServiceImpl{
		quizRepo:     quizRepo,
		userRepo:     userRepo,
		questionRepo: questionRepo,
		wsHub:        wsHub,
	}
}

// CreateQuiz creates a new quiz with the specified creator
func (s *quizServiceImpl) CreateQuiz(ctx context.Context, title string, description string, creatorID uuid.UUID) (*model.Quiz, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	// Verify user exists
	creator, err := s.userRepo.GetUserByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	// Create the quiz
	quiz := model.NewQuiz(title, description, creator.ID)

	// Create quiz session
	session := model.NewQuizSession(quiz.ID)

	// Save to database
	if err := s.quizRepo.CreateQuiz(ctx, quiz); err != nil {
		return nil, err
	}
	if err := s.quizRepo.CreateQuizSession(ctx, session); err != nil {
		return nil, err
	}

	return quiz, nil
}

// CreateQuizWithQuestions creates a new quiz with questions
func (s *quizServiceImpl) CreateQuizWithQuestions(ctx context.Context, title string, description string, creatorID uuid.UUID, questions []dto.QuestionCreateData) (*model.Quiz, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	if len(questions) == 0 {
		return nil, errors.New("at least one question is required")
	}

	// Verify user exists
	creator, err := s.userRepo.GetUserByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	// Create the quiz
	quiz := model.NewQuiz(title, description, creator.ID)

	// Create quiz session
	session := model.NewQuizSession(quiz.ID)

	// Save quiz to database
	if err := s.quizRepo.CreateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	// Save quiz session to database
	if err := s.quizRepo.CreateQuizSession(ctx, session); err != nil {
		return nil, err
	}

	// Create questions
	for i, q := range questions {
		// Validate question data
		if q.Text == "" {
			return nil, errors.New("question text is required")
		}
		if len(q.Options) != 4 {
			return nil, errors.New("question must have exactly 4 options")
		}
		if q.CorrectAnswer == "" || (q.CorrectAnswer != "A" && q.CorrectAnswer != "B" && q.CorrectAnswer != "C" && q.CorrectAnswer != "D") {
			return nil, errors.New("correct answer must be A, B, C, or D")
		}
		if q.TimeLimit <= 0 {
			return nil, errors.New("time limit must be positive")
		}

		// Create question with order based on array position
		question := model.NewQuestion(quiz.ID, q.Text, q.Options, q.CorrectAnswer, q.TimeLimit, i+1)

		// Save question to database
		if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
			return nil, err
		}
	}

	return quiz, nil
}

// GetQuiz retrieves a quiz by ID
func (s *quizServiceImpl) GetQuiz(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, id)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	return quiz, nil
}

// GetQuizByCode retrieves a quiz by its code
func (s *quizServiceImpl) GetQuizByCode(ctx context.Context, code string) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByCode(ctx, code)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	return quiz, nil
}

// StartQuiz starts a quiz session
func (s *quizServiceImpl) StartQuiz(ctx context.Context, quizID uuid.UUID) error {
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
func (s *quizServiceImpl) EndQuiz(ctx context.Context, quizID uuid.UUID) error {
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

// GetQuizSession retrieves the current state of a quiz
func (s *quizServiceImpl) GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error) {
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetQuizzesByCreatorID retrieves all quizzes created by a user
func (s *quizServiceImpl) GetQuizzesByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*model.Quiz, error) {
	quizzes, err := s.quizRepo.GetQuizzesByCreatorID(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	return quizzes, nil
}
