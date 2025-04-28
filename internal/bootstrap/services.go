package bootstrap

import (
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/auth"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
)

// Services holds all service instances
type Services struct {
	UserService        service.UserService
	ParticipantService service.ParticipantService
	QuizService        service.QuizService
	QuestionService    service.QuestionService
	AnswerService      service.AnswerService
	LeaderboardService service.LeaderboardService
	StateService       service.StateService
}

// NewServices initializes all services
func NewServices(repos *Repositories, jwtManager *auth.JWTManager, wsHub *websocket.RedisHub) *Services {
	leaderBoardSerice := service.NewLeaderboardService(repos.ParticipantRepo, wsHub)

	return &Services{
		UserService:        service.NewUserService(repos.UserRepo, jwtManager),
		ParticipantService: service.NewParticipantService(repos.ParticipantRepo, repos.QuizRepo, wsHub),
		QuizService:        service.NewQuizService(repos.QuizRepo, repos.UserRepo, repos.QuestionRepo, repos.QuestionOptionRepo, wsHub),
		QuestionService:    service.NewQuestionService(repos.QuizRepo, repos.QuestionRepo, repos.QuestionOptionRepo, wsHub),
		AnswerService:      service.NewAnswerService(repos.AnswerRepo, repos.QuestionRepo, repos.ParticipantRepo, repos.QuizRepo, leaderBoardSerice, repos.QuestionOptionRepo, wsHub),
		LeaderboardService: leaderBoardSerice,
		StateService:       service.NewStateService(repos.StateRepo, repos.QuizRepo, repos.QuestionRepo, repos.QuestionOptionRepo, repos.ParticipantRepo, wsHub),
	}
}
