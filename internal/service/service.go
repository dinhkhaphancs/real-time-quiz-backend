package service

import (
	"context"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// QuizService defines operations for quiz business logic
type QuizService interface {
	// CreateQuiz creates a new quiz
	CreateQuiz(ctx context.Context, title string, description string, creatorID uuid.UUID) (*model.Quiz, error)

	// CreateQuizWithQuestions creates a new quiz with questions
	CreateQuizWithQuestions(ctx context.Context, title string, description string, creatorID uuid.UUID, questions []dto.QuestionCreateData) (*model.Quiz, error)

	// GetQuiz retrieves a quiz by ID
	GetQuiz(ctx context.Context, id uuid.UUID) (*model.Quiz, error)

	// GetQuizByCode retrieves a quiz by its code
	GetQuizByCode(ctx context.Context, code string) (*model.Quiz, error)

	// GetQuizzesByCreatorID retrieves all quizzes created by a user
	GetQuizzesByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*model.Quiz, error)

	// StartQuiz starts a quiz session
	StartQuiz(ctx context.Context, quizID uuid.UUID) error

	// EndQuiz ends a quiz session
	EndQuiz(ctx context.Context, quizID uuid.UUID) error

	// GetQuizSession retrieves the current state of a quiz
	GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error)

	// UpdateQuiz updates an existing quiz's basic info
	UpdateQuiz(ctx context.Context, quizID uuid.UUID, title string, description string) (*model.Quiz, error)

	// UpdateQuizWithQuestions updates an existing quiz with its questions
	UpdateQuizWithQuestions(ctx context.Context, quizID uuid.UUID, title string, description string, questions []dto.QuestionUpdateData) (*model.Quiz, error)

	// DeleteQuiz deletes a quiz and all its related data
	DeleteQuiz(ctx context.Context, quizID uuid.UUID) error
}

// QuestionService defines operations for question business logic
type QuestionService interface {
	// AddQuestion adds a question to a quiz
	AddQuestion(ctx context.Context, quizID uuid.UUID, text string, options []dto.OptionCreateData, questionType string, timeLimit int) (*model.Question, error)

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
	// SubmitAnswer records a participant's answer to a question
	SubmitAnswer(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID, selectedOptions []string) (*model.Answer, error)

	// GetAnswerStats retrieves statistics for answers to a question
	GetAnswerStats(ctx context.Context, questionID uuid.UUID) (map[string]int, error)

	// GetParticipantAnswer retrieves a participant's answer to a specific question
	GetParticipantAnswer(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID) (*model.Answer, error)
}

// LeaderboardService defines operations for leaderboard business logic
type LeaderboardService interface {
	// GetLeaderboard retrieves the top participants by score
	GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.Participant, error)

	// UpdateParticipantScore updates a participant's total score
	UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, additionalScore int) error
}

// UserService defines operations for user business logic
type UserService interface {
	// Register creates a new user account
	Register(ctx context.Context, name string, email string, password string) (*model.User, error)

	// Login authenticates a user and returns user data
	Login(ctx context.Context, email string, password string) (*model.User, error)

	// LoginWithToken authenticates a user and returns user data with JWT tokens
	LoginWithToken(ctx context.Context, email string, password string) (*dto.UserLoginResponse, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// ParticipantService defines operations for participant business logic
type ParticipantService interface {
	// JoinQuiz allows a user to join a quiz as a participant
	JoinQuiz(ctx context.Context, quizID uuid.UUID, name string) (*model.Participant, error)

	// JoinQuizByCode allows a user to join a quiz using its code
	JoinQuizByCode(ctx context.Context, code string, name string) (*model.Participant, error)

	// GetParticipantByID retrieves a participant by ID
	GetParticipantByID(ctx context.Context, id uuid.UUID) (*model.Participant, error)

	// GetParticipantsByQuizID retrieves all participants for a quiz
	GetParticipantsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Participant, error)

	// RemoveParticipant removes a participant from a quiz
	RemoveParticipant(ctx context.Context, id uuid.UUID) error
}
