// Package bootstrap handles the initialization and wiring of application components
package bootstrap

import (
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
)

// Repositories holds all repository instances
type Repositories struct {
	QuizRepo           repository.QuizRepository
	QuestionRepo       repository.QuestionRepository
	QuestionOptionRepo repository.QuestionOptionRepository
	UserRepo           repository.UserRepository
	ParticipantRepo    repository.ParticipantRepository
	AnswerRepo         repository.AnswerRepository
	StateRepo          repository.StateRepository
}

// NewRepositories initializes all repositories
func NewRepositories(db *repository.DB) *Repositories {
	return &Repositories{
		QuizRepo:           repository.NewPostgresQuizRepository(db),
		QuestionRepo:       repository.NewPostgresQuestionRepository(db),
		QuestionOptionRepo: repository.NewPostgresQuestionOptionRepository(db),
		UserRepo:           repository.NewPostgresUserRepository(db),
		ParticipantRepo:    repository.NewPostgresParticipantRepository(db),
		AnswerRepo:         repository.NewPostgresAnswerRepository(db),
		StateRepo:          repository.NewStateRepository(db.DB),
	}
}
