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
	"github.com/go-redis/redis/v8" // Add Redis client import
)

func main() {
	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg := config.NewConfig()

	// Override config with environment variables if provided
	if os.Getenv("POSTGRES_HOST") != "" {
		cfg.Postgres.Host = os.Getenv("POSTGRES_HOST")
	}
	if os.Getenv("POSTGRES_PORT") != "" {
		fmt.Sscanf(os.Getenv("POSTGRES_PORT"), "%d", &cfg.Postgres.Port)
	}
	if os.Getenv("POSTGRES_USER") != "" {
		cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	}
	if os.Getenv("POSTGRES_PASSWORD") != "" {
		cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	}
	if os.Getenv("POSTGRES_DB") != "" {
		cfg.Postgres.Database = os.Getenv("POSTGRES_DB")
	}
	if os.Getenv("POSTGRES_SSLMODE") != "" {
		cfg.Postgres.SSLMode = os.Getenv("POSTGRES_SSLMODE")
	}
	if os.Getenv("REDIS_HOST") != "" {
		cfg.Redis.Host = os.Getenv("REDIS_HOST")
	}
	if os.Getenv("REDIS_PORT") != "" {
		fmt.Sscanf(os.Getenv("REDIS_PORT"), "%d", &cfg.Redis.Port)
	}
	if os.Getenv("REDIS_PASSWORD") != "" {
		cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}
	if os.Getenv("REDIS_DB") != "" {
		fmt.Sscanf(os.Getenv("REDIS_DB"), "%d", &cfg.Redis.DB)
	}

	// Initialize database connection
	db, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL database")

	// Initialize Redis client
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
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
	answerRepo := repository.NewPostgresAnswerRepository(db)

	// Initialize services
	quizService := service.NewQuizService(quizRepo, userRepo, questionRepo, wsHub)
	questionService := service.NewQuestionService(quizRepo, questionRepo, wsHub)
	answerService := service.NewAnswerService(answerRepo, questionRepo, userRepo, wsHub)
	leaderboardService := service.NewLeaderboardService(userRepo, wsHub)
	userService := service.NewUserService(userRepo, quizRepo)

	// Initialize handlers
	quizHandler := handler.NewQuizHandler(quizService, questionService)
	questionHandler := handler.NewQuestionHandler(questionService)
	answerHandler := handler.NewAnswerHandler(answerService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)
	wsHandler := handler.NewWebSocketHandler(wsHub, quizService, userService)

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
			answerRoutes.GET("/user/:userId/question/:questionId", answerHandler.GetUserAnswer)
		}

		// Leaderboard routes
		apiV1.GET("/leaderboard/quiz/:quizId", leaderboardHandler.GetLeaderboard)
	}

	// WebSocket route
	router.GET("/ws/:quizId/:userId", wsHandler.HandleConnection)

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
