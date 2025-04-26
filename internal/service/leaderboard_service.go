package service

import (
	"context"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// leaderboardService implements LeaderboardService interface
type leaderboardService struct {
	userRepo repository.UserRepository
	wsHub    *websocket.RedisHub
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(
	userRepo repository.UserRepository,
	wsHub *websocket.RedisHub,
) LeaderboardService {
	return &leaderboardService{
		userRepo: userRepo,
		wsHub:    wsHub,
	}
}

// GetLeaderboard retrieves the top users by score
func (s *leaderboardService) GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.User, error) {
	// If limit is not specified or is invalid, set a default
	if limit <= 0 {
		limit = 10
	}

	// Get users sorted by score
	users, err := s.userRepo.GetLeaderboard(ctx, quizID, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserScore updates a user's total score and broadcasts the updated leaderboard
func (s *leaderboardService) UpdateUserScore(ctx context.Context, userID uuid.UUID, additionalScore int) error {
	// Update the user's score
	if err := s.userRepo.UpdateUserScore(ctx, userID, additionalScore); err != nil {
		return err
	}

	// Get the user to determine the quiz ID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Get the updated leaderboard
	leaderboard, err := s.GetLeaderboard(ctx, user.QuizID, 10)
	if err != nil {
		return err
	}

	// Prepare leaderboard data for broadcasting
	var leaderboardData []map[string]interface{}
	for i, user := range leaderboard {
		leaderboardData = append(leaderboardData, map[string]interface{}{
			"rank":  i + 1,
			"id":    user.ID.String(),
			"name":  user.Name,
			"score": user.Score,
		})
	}

	// Broadcast updated leaderboard to all clients in the quiz
	s.wsHub.BroadcastToQuiz(user.QuizID, websocket.Event{
		Type: websocket.EventLeaderboardUpdate,
		Payload: map[string]interface{}{
			"leaderboard": leaderboardData,
		},
	})

	return nil
}