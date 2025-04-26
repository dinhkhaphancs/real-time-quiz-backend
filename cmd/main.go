package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/config"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/handler"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration from config files and environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	db, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL database")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis")

	// Initialize Redis WebSocket hub with the Redis client
	wsHub := websocket.NewRedisHub(redisClient, ctx)

	// Start Redis hub in a separate goroutine
	go wsHub.Run(ctx)
	log.Println("Started WebSocket hub")

	// Initialize repositories
	quizRepo := repository.NewPostgresQuizRepository(db)
	questionRepo := repository.NewPostgresQuestionRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)
	participantRepo := repository.NewPostgresParticipantRepository(db)
	answerRepo := repository.NewPostgresAnswerRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	participantService := service.NewParticipantService(participantRepo, quizRepo, wsHub)
	quizService := service.NewQuizService(quizRepo, userRepo, questionRepo, wsHub)
	questionService := service.NewQuestionService(quizRepo, questionRepo, wsHub)
	answerService := service.NewAnswerService(answerRepo, questionRepo, participantRepo, wsHub)
	leaderboardService := service.NewLeaderboardService(participantRepo, wsHub)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	quizHandler := handler.NewQuizHandler(quizService, questionService, userService, participantService)
	questionHandler := handler.NewQuestionHandler(questionService)
	answerHandler := handler.NewAnswerHandler(answerService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService, quizService)
	wsHandler := handler.NewWebSocketHandler(wsHub, quizService, userService, participantService)
	participantHandler := handler.NewParticipantHandler(participantService, quizService)

	// Initialize Gin router
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

	// API routes
	apiV1 := router.Group("/api/v1")
	{
		// User routes
		userRoutes := apiV1.Group("/users")
		{
			userRoutes.POST("/register", userHandler.RegisterUser)
			userRoutes.POST("/login", userHandler.LoginUser)
		}

		// Quiz routes
		quizRoutes := apiV1.Group("/quizzes")
		{
			quizRoutes.POST("", quizHandler.CreateQuiz)
			quizRoutes.GET("/:id", quizHandler.GetQuiz)
			quizRoutes.POST("/:id/start", quizHandler.StartQuiz)
			quizRoutes.POST("/:id/end", quizHandler.EndQuiz)
			quizRoutes.POST("/:id/join", quizHandler.JoinQuiz)
		}

		// Question routes
		questionRoutes := apiV1.Group("/questions")
		{
			questionRoutes.POST("", questionHandler.AddQuestion)
			questionRoutes.GET("/:id", questionHandler.GetQuestion)
			questionRoutes.POST("/:id/start", questionHandler.StartQuestion)
			questionRoutes.POST("/:id/end", questionHandler.EndQuestion)
			questionRoutes.GET("/quiz/:quizId", questionHandler.GetQuestions)
			questionRoutes.GET("/quiz/:quizId/next", questionHandler.GetNextQuestion)
		}

		// Answer routes
		answerRoutes := apiV1.Group("/answers")
		{
			answerRoutes.POST("", answerHandler.SubmitAnswer)
			answerRoutes.GET("/question/:questionId/stats", answerHandler.GetAnswerStats)
			answerRoutes.GET("/participant/:participantId/question/:questionId", answerHandler.GetParticipantAnswer)
		}

		// Leaderboard routes
		apiV1.GET("/leaderboard/quiz/:quizId", leaderboardHandler.GetLeaderboard)

		// Participant routes
		participantRoutes := apiV1.Group("/participants")
		{
			participantRoutes.GET("/:id", participantHandler.GetParticipant)
			participantRoutes.GET("/quiz/:quizId", participantHandler.GetParticipantsByQuiz)
			participantRoutes.DELETE("/:id", participantHandler.RemoveParticipant)
		}
	}

	// WebSocket route
	router.GET("/ws/:quizId/:type/:id", wsHandler.HandleConnection)

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for the shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
