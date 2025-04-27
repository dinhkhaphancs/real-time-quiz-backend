package bootstrap

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/middleware"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the HTTP router
func SetupRouter(handlers *Handlers, jwtManager *auth.JWTManager) *gin.Engine {
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes
	setupRoutes(router, handlers, jwtManager)

	return router
}

// setupRoutes configures all API routes
func setupRoutes(router *gin.Engine, handlers *Handlers, jwtManager *auth.JWTManager) {
	// API routes base group
	apiV1 := router.Group("/api/v1")

	// Create auth middleware
	authMiddleware := middleware.JWTAuthMiddleware(jwtManager)

	// ========== User Module ==========
	userRoutes := apiV1.Group("/users")
	{
		// Public user routes
		userRoutes.POST("/register", handlers.UserHandler.RegisterUser)
		userRoutes.POST("/login", handlers.UserHandler.LoginUser)

		// Private user routes
		userPrivate := userRoutes.Group("")
		userPrivate.Use(authMiddleware)
		{
			// Add protected user endpoints here if needed
			// Example: userPrivate.GET("/profile", handlers.UserHandler.GetProfile)
		}
	}

	// ========== Quiz Module ==========
	quizRoutes := apiV1.Group("/quizzes")
	{
		// Public quiz routes
		quizRoutes.POST("/:id/join", handlers.QuizHandler.JoinQuiz)
		quizRoutes.POST("/join", handlers.QuizHandler.JoinQuizByCode)

		// Private quiz routes
		quizPrivate := quizRoutes.Group("")
		quizPrivate.Use(authMiddleware)
		{
			quizPrivate.GET("/my", handlers.QuizHandler.GetCurrentUserQuizzes)
			quizPrivate.GET("/:id", handlers.QuizHandler.GetQuiz)
			quizPrivate.POST("", handlers.QuizHandler.CreateQuiz)
			quizPrivate.PUT("/:id", handlers.QuizHandler.UpdateQuiz)
			quizPrivate.DELETE("/:id", handlers.QuizHandler.DeleteQuiz)
			quizPrivate.POST("/:id/start", handlers.QuizHandler.StartQuiz)
			quizPrivate.POST("/:id/end", handlers.QuizHandler.EndQuiz)
		}
	}

	// ========== Question Module ==========
	questionRoutes := apiV1.Group("/questions")
	{
		// Public question routes
		questionRoutes.GET("/:id", handlers.QuestionHandler.GetQuestion)
		questionRoutes.GET("/quiz/:quizId", handlers.QuestionHandler.GetQuestions)
		questionRoutes.GET("/quiz/:quizId/next", handlers.QuestionHandler.GetNextQuestion)

		// Private question routes
		questionPrivate := questionRoutes.Group("")
		questionPrivate.Use(authMiddleware)
		{
			questionPrivate.POST("", handlers.QuestionHandler.AddQuestion)
			questionPrivate.POST("/:id/start", handlers.QuestionHandler.StartQuestion)
			questionPrivate.POST("/:id/end", handlers.QuestionHandler.EndQuestion)
		}
	}

	// ========== Answer Module ==========
	answerRoutes := apiV1.Group("/answers")
	{
		// All answer routes are currently public
		answerRoutes.POST("", handlers.AnswerHandler.SubmitAnswer)
		answerRoutes.GET("/question/:questionId/stats", handlers.AnswerHandler.GetAnswerStats)
		answerRoutes.GET("/participant/:participantId/question/:questionId", handlers.AnswerHandler.GetParticipantAnswer)

		// Private answer routes if needed
		answerPrivate := answerRoutes.Group("")
		answerPrivate.Use(authMiddleware)
		{
			// Add protected answer endpoints here if needed
		}
	}

	// ========== Participant Module ==========
	participantRoutes := apiV1.Group("/participants")
	{
		// Public participant routes
		participantRoutes.GET("/:id", handlers.ParticipantHandler.GetParticipant)
		participantRoutes.GET("/quiz/:quizId", handlers.ParticipantHandler.GetParticipantsByQuiz)

		// Private participant routes
		participantPrivate := participantRoutes.Group("")
		participantPrivate.Use(authMiddleware)
		{
			participantPrivate.DELETE("/:id", handlers.ParticipantHandler.RemoveParticipant)
		}
	}

	// ========== Leaderboard Module ==========
	leaderboardRoutes := apiV1.Group("/leaderboard")
	{
		// Public leaderboard routes
		leaderboardRoutes.GET("/quiz/:quizId", handlers.LeaderboardHandler.GetLeaderboard)

		// Private leaderboard routes if needed
		leaderboardPrivate := leaderboardRoutes.Group("")
		leaderboardPrivate.Use(authMiddleware)
		{
			// Add protected leaderboard endpoints here if needed
		}
	}

	// ========== WebSocket ==========
	// WebSocket route (outside API versioning)
	router.GET("/ws/:quizId/:type/:id", handlers.WSHandler.HandleConnection)
}
