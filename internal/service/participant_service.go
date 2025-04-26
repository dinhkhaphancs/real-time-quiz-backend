package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// participantServiceImpl implements ParticipantService interface
type participantServiceImpl struct {
	participantRepo repository.ParticipantRepository
	quizRepo        repository.QuizRepository
	wsHub           *websocket.RedisHub
}

// NewParticipantService creates a new participant service
func NewParticipantService(
	participantRepo repository.ParticipantRepository,
	quizRepo repository.QuizRepository,
	wsHub *websocket.RedisHub,
) ParticipantService {
	return &participantServiceImpl{
		participantRepo: participantRepo,
		quizRepo:        quizRepo,
		wsHub:           wsHub,
	}
}

// JoinQuiz allows a user to join a quiz as a participant
func (s *participantServiceImpl) JoinQuiz(ctx context.Context, quizID uuid.UUID, name string) (*model.Participant, error) {
	// Validate inputs
	if name == "" {
		return nil, errors.New("name is required")
	}

	// Check if quiz exists and is in waiting state
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	if quiz.Status != model.QuizStatusWaiting {
		return nil, errors.New("cannot join a quiz that has already started")
	}

	// Check if name is already taken in this quiz
	participants, err := s.participantRepo.GetParticipantsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	for _, p := range participants {
		if p.Name == name {
			return nil, errors.New("name is already taken in this quiz")
		}
	}

	// Create new participant
	participant := model.NewParticipant(name, quizID)

	// Save to repository
	if err := s.participantRepo.CreateParticipant(ctx, participant); err != nil {
		return nil, err
	}

	// Broadcast participant joined event
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventUserJoined,
		Payload: map[string]interface{}{
			"participantId": participant.ID.String(),
			"name":          participant.Name,
		},
	})

	return participant, nil
}

// JoinQuizByCode allows a user to join a quiz using its code
func (s *participantServiceImpl) JoinQuizByCode(ctx context.Context, code string, name string) (*model.Participant, error) {
	// Validate inputs
	if name == "" {
		return nil, errors.New("name is required")
	}

	if code == "" {
		return nil, errors.New("code is required")
	}

	// Check if quiz exists and is in waiting state
	quiz, err := s.quizRepo.GetQuizByCode(ctx, code)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	// Use the existing JoinQuiz functionality with the retrieved quiz ID
	return s.JoinQuiz(ctx, quiz.ID, name)
}

// GetParticipantByID retrieves a participant by ID
func (s *participantServiceImpl) GetParticipantByID(ctx context.Context, id uuid.UUID) (*model.Participant, error) {
	participant, err := s.participantRepo.GetParticipantByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return participant, nil
}

// GetParticipantsByQuizID retrieves all participants for a quiz
func (s *participantServiceImpl) GetParticipantsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Participant, error) {
	participants, err := s.participantRepo.GetParticipantsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

// RemoveParticipant removes a participant from a quiz
func (s *participantServiceImpl) RemoveParticipant(ctx context.Context, id uuid.UUID) error {
	// First get the participant to broadcast the removal event
	participant, err := s.participantRepo.GetParticipantByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from repository
	if err := s.participantRepo.DeleteParticipant(ctx, id); err != nil {
		return err
	}

	// Broadcast participant left event
	s.wsHub.BroadcastToQuiz(participant.QuizID, websocket.Event{
		Type: websocket.EventUserLeft,
		Payload: map[string]interface{}{
			"participantId": id.String(),
			"name":          participant.Name,
		},
	})

	return nil
}
