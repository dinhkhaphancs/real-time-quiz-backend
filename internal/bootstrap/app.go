package bootstrap

import (
	"context"
	"fmt"
	"log"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/config"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/auth"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/go-redis/redis/v8"
)

// App represents the application
type App struct {
	config      *config.Config
	server      *Server
	db          *repository.DB
	redisClient *redis.Client
}

// NewApp creates a new application instance
func NewApp() (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Setup database
	db, err := repository.NewPostgresDB(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Connected to PostgreSQL database")

	// Setup Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		db.Close() // Close DB if Redis fails
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Connected to Redis")

	// Setup WebSocket hub
	wsHub := websocket.NewRedisHub(redisClient, ctx)
	go wsHub.Run(ctx)
	log.Println("Started WebSocket hub")

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT)
	log.Println("Initialized JWT authentication manager")

	// Initialize repositories, services, and handlers
	repos := NewRepositories(db)
	services := NewServices(repos, jwtManager, wsHub)
	handlers := NewHandlers(services, wsHub)

	// Setup router
	router := SetupRouter(handlers, jwtManager)

	// Setup server
	server := NewServer(cfg, router)

	return &App{
		config:      cfg,
		server:      server,
		db:          db,
		redisClient: redisClient,
	}, nil
}

// Start starts the application
func (a *App) Start() {
	a.server.Start()
}

// Stop gracefully stops the application
func (a *App) Stop() {
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}
}
