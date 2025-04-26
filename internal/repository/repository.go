package repository

import (
	"context"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// QuizRepository defines operations for quiz management
type QuizRepository interface {
	// CreateQuiz creates a new quiz
	CreateQuiz(ctx context.Context, quiz *model.Quiz) error
	
	// GetQuizByID retrieves a quiz by its ID
	GetQuizByID(ctx context.Context, id uuid.UUID) (*model.Quiz, error)
	
	// UpdateQuizStatus updates the status of a quiz
	UpdateQuizStatus(ctx context.Context, id uuid.UUID, status model.QuizStatus) error
	
	// CreateQuizSession creates a new quiz session
	CreateQuizSession(ctx context.Context, session *model.QuizSession) error
	
	// GetQuizSession retrieves a quiz session
	GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error)
	
	// UpdateQuizSession updates a quiz session
	UpdateQuizSession(ctx context.Context, session *model.QuizSession) error
}

// QuestionRepository defines operations for question management
type QuestionRepository interface {
	// CreateQuestion creates a new question
	CreateQuestion(ctx context.Context, question *model.Question) error
	
	// GetQuestionsByQuizID retrieves all questions for a quiz
	GetQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error)
	
	// GetQuestionByID retrieves a question by its ID
	GetQuestionByID(ctx context.Context, id uuid.UUID) (*model.Question, error)
	
	// GetNextQuestion retrieves the next question after the current one
	GetNextQuestion(ctx context.Context, quizID uuid.UUID, currentOrder int) (*model.Question, error)
}

// UserRepository defines operations for user management
type UserRepository interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *model.User) error
	
	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	
	// GetUsersByQuizID retrieves all users for a quiz
	GetUsersByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.User, error)
	
	// UpdateUserScore updates a user's score
	UpdateUserScore(ctx context.Context, userID uuid.UUID, score int) error
	
	// GetLeaderboard retrieves the top users by score for a quiz
	GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.User, error)
}

// AnswerRepository defines operations for answer management
type AnswerRepository interface {
	// CreateAnswer creates a new answer
	CreateAnswer(ctx context.Context, answer *model.Answer) error
	
	// GetAnswersByQuestionID retrieves all answers for a question
	GetAnswersByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*model.Answer, error)
	
	// GetAnswersByUserID retrieves all answers for a user
	GetAnswersByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Answer, error)
	
	// GetAnswerByUserAndQuestion retrieves a user's answer for a specific question
	GetAnswerByUserAndQuestion(ctx context.Context, userID uuid.UUID, questionID uuid.UUID) (*model.Answer, error)
}