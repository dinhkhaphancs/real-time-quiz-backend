package bootstrap

import (
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/handler"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
)

// Handlers holds all handler instances
type Handlers struct {
	UserHandler        *handler.UserHandler
	QuizHandler        *handler.QuizHandler
	QuestionHandler    *handler.QuestionHandler
	AnswerHandler      *handler.AnswerHandler
	LeaderboardHandler *handler.LeaderboardHandler
	WSHandler          *handler.WebSocketHandler
	ParticipantHandler *handler.ParticipantHandler
}

// NewHandlers initializes all handlers
func NewHandlers(services *Services, wsHub *websocket.RedisHub) *Handlers {
	return &Handlers{
		UserHandler:        handler.NewUserHandler(services.UserService),
		QuizHandler:        handler.NewQuizHandler(services.QuizService, services.QuestionService, services.UserService, services.ParticipantService),
		QuestionHandler:    handler.NewQuestionHandler(services.QuestionService, services.QuizService),
		AnswerHandler:      handler.NewAnswerHandler(services.AnswerService),
		LeaderboardHandler: handler.NewLeaderboardHandler(services.LeaderboardService, services.QuizService),
		WSHandler:          handler.NewWebSocketHandler(wsHub, services.QuizService, services.UserService, services.ParticipantService),
		ParticipantHandler: handler.NewParticipantHandler(services.ParticipantService, services.QuizService),
	}
}
