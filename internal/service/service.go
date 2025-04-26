package service

import (
	"context"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// QuizService defines operations for quiz business logic
type QuizService interface {
	// CreateQuiz creates a new quiz
	CreateQuiz(ctx context.Context, title string, adminName string) (*model.Quiz, *model.User, error)
	
	// GetQuiz retrieves a quiz by ID
	GetQuiz(ctx context.Context, id uuid.UUID) (*model.Quiz, error)
	
	// StartQuiz starts a quiz session
	StartQuiz(ctx context.Context, quizID uuid.UUID) error
	
	// EndQuiz ends a quiz session
	EndQuiz(ctx context.Context, quizID uuid.UUID) error
	
	// JoinQuiz allows a user to join a quiz
	JoinQuiz(ctx context.Context, quizID uuid.UUID, name string) (*model.User, error)
	
	// GetQuizSession retrieves the current state of a quiz
	GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error)
}

// QuestionService defines operations for question business logic
type QuestionService interface {
	// AddQuestion adds a question to a quiz
	AddQuestion(ctx context.Context, quizID uuid.UUID, text string, options []model.Option, correctAnswer string, timeLimit int) (*model.Question, error)
	
	// GetQuestions retrieves all questions for a quiz
	GetQuestions(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error)
	
	// GetQuestion retrieves a question by ID
	GetQuestion(ctx context.Context, id uuid.UUID) (*model.Question, error)
	
	// StartQuestion starts a specific question in a quiz
	StartQuestion(ctx context.Context, quizID uuid.UUID, questionID uuid.UUID) error
	
	// EndQuestion ends the current question in a quiz
	EndQuestion(ctx context.Context, quizID uuid.UUID) error
	
	// GetNextQuestion retrieves the next question in sequence
	GetNextQuestion(ctx context.Context, quizID uuid.UUID) (*model.Question, error)
}

// AnswerService defines operations for answer business logic
type AnswerService interface {
	// SubmitAnswer records a user's answer to a question
	SubmitAnswer(ctx context.Context, userID uuid.UUID, questionID uuid.UUID, selectedOption string) (*model.Answer, error)
	
	// GetAnswerStats retrieves statistics for answers to a question
	GetAnswerStats(ctx context.Context, questionID uuid.UUID) (map[string]int, error)
	
	// GetUserAnswer retrieves a user's answer to a specific question
	GetUserAnswer(ctx context.Context, userID uuid.UUID, questionID uuid.UUID) (*model.Answer, error)
}

// LeaderboardService defines operations for leaderboard business logic
type LeaderboardService interface {
	// GetLeaderboard retrieves the top users by score
	GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.User, error)
	
	// UpdateUserScore updates a user's total score
	UpdateUserScore(ctx context.Context, userID uuid.UUID, additionalScore int) error
}

// UserService defines operations for user business logic
type UserService interface {
	// GetUserByID retrieves a user by ID and validates quiz access
	GetUserByID(ctx context.Context, userID uuid.UUID, quizID uuid.UUID) (*model.User, error)
	
	// CreateUser creates a new user
	CreateUser(ctx context.Context, quizID uuid.UUID, name string, role string) (*model.User, error)
}