package service

import (
	"context"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// leaderboardServiceImpl implements LeaderboardService interface
type leaderboardServiceImpl struct {
	participantRepo repository.ParticipantRepository
	wsHub           *websocket.RedisHub
}

// NewLeaderboardService creates a new leaderboard service
func NewLeaderboardService(
	participantRepo repository.ParticipantRepository,
	wsHub *websocket.RedisHub,
) LeaderboardService {
	return &leaderboardServiceImpl{
		participantRepo: participantRepo,
		wsHub:           wsHub,
	}
}

// GetLeaderboard retrieves the top participants by score
func (s *leaderboardServiceImpl) GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.Participant, error) {
	// If limit is not specified or is invalid, set a default
	if limit <= 0 {
		limit = 10
	}

	// Get participants sorted by score
	participants, err := s.participantRepo.GetLeaderboard(ctx, quizID, limit)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

// UpdateParticipantScore updates a participant's total score and broadcasts the updated leaderboard
func (s *leaderboardServiceImpl) UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, additionalScore int) error {
	// Update the participant's score
	if err := s.participantRepo.UpdateParticipantScore(ctx, participantID, additionalScore); err != nil {
		return err
	}

	// Get the participant to determine the quiz ID
	participant, err := s.participantRepo.GetParticipantByID(ctx, participantID)
	if err != nil {
		return err
	}

	// Get the updated leaderboard
	leaderboard, err := s.GetLeaderboard(ctx, participant.QuizID, 10)
	if err != nil {
		return err
	}

	// Prepare leaderboard data for broadcasting
	var leaderboardData []map[string]interface{}
	for i, participant := range leaderboard {
		leaderboardData = append(leaderboardData, map[string]interface{}{
			"rank":  i + 1,
			"id":    participant.ID.String(),
			"name":  participant.Name,
			"score": participant.Score,
		})
	}

	// Broadcast updated leaderboard to all clients in the quiz
	s.wsHub.BroadcastToQuiz(participant.QuizID, websocket.Event{
		Type: websocket.EventLeaderboardUpdate,
		Payload: map[string]interface{}{
			"leaderboard": leaderboardData,
		},
	})

	return nil
}