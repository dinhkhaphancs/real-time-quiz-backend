package bootstrap

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
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, router *gin.Engine) *Server {
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
	}
}

// Start starts the server and handles graceful shutdown
func (s *Server) Start() {
	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
