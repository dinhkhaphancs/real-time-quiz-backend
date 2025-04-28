package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// StateService defines methods for managing quiz state
type StateService interface {
	// State Management
	GetQuizState(ctx context.Context, quizID uuid.UUID) (*dto.QuizStateDTO, error)

	// Events
	PublishEvent(ctx context.Context, quizID uuid.UUID, eventType string, payload interface{}) error
	GetMissedEvents(ctx context.Context, quizID uuid.UUID, lastSequence int64) ([]*model.QuizEvent, error)

	// Participant Connection
	UpdateParticipantConnection(ctx context.Context, participantID, quizID uuid.UUID, isConnected bool, instanceID string) error
	GetActiveParticipants(ctx context.Context, quizID uuid.UUID) ([]model.Participant, error)

	// Instance Management
	RegisterInstance(ctx context.Context, instanceID string) error
	UpdateInstanceHeartbeat(ctx context.Context, instanceID string) error
}

// stateServiceImpl implements StateService interface
type stateServiceImpl struct {
	stateRepo          repository.StateRepository
	quizRepo           repository.QuizRepository
	questionRepo       repository.QuestionRepository
	questionOptionRepo repository.QuestionOptionRepository
	participantRepo    repository.ParticipantRepository
	wsHub              *websocket.RedisHub
	instanceID         string
}

// NewStateService creates a new state service
func NewStateService(
	stateRepo repository.StateRepository,
	quizRepo repository.QuizRepository,
	questionRepo repository.QuestionRepository,
	questionOptionRepo repository.QuestionOptionRepository,
	participantRepo repository.ParticipantRepository,
	wsHub *websocket.RedisHub,
) StateService {
	// Generate a unique instance ID
	instanceID := uuid.New().String()

	return &stateServiceImpl{
		stateRepo:          stateRepo,
		quizRepo:           quizRepo,
		questionRepo:       questionRepo,
		questionOptionRepo: questionOptionRepo,
		participantRepo:    participantRepo,
		wsHub:              wsHub,
		instanceID:         instanceID,
	}
}

// GetQuizState retrieves the current state of a quiz
func (s *stateServiceImpl) GetQuizState(ctx context.Context, quizID uuid.UUID) (*dto.QuizStateDTO, error) {
	// Get quiz details
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Get quiz session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Get participants
	participants, err := s.participantRepo.GetParticipantsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Get active question if any
	var activeQuestion *model.Question
	var questionCount int

	// Count all questions for this quiz
	questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err == nil {
		questionCount = len(questions)
	}

	// Get active question details if there is one
	if session.CurrentQuestionID != nil {
		activeQuestion, err = s.questionRepo.GetQuestionByID(ctx, *session.CurrentQuestionID)
		if err == nil {
			// Load question options
			options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, activeQuestion.ID)
			if err == nil {
				activeQuestion.Options = options
			}
		}
	}

	// Convert to state DTO
	state := dto.ToQuizStateDTO(quiz, session, participants, activeQuestion, questionCount)

	// Update connected status from participant_connections table
	cutoffTime := time.Now().Add(-30 * time.Second) // Consider connections within last 30 seconds
	connections, err := s.stateRepo.GetActiveParticipantConnections(ctx, quizID, cutoffTime)
	if err == nil {
		activeCount := 0
		for _, conn := range connections {
			participantID := conn.ParticipantID.String()
			if participant, ok := state.Participants[participantID]; ok {
				participant.IsConnected = conn.IsConnected
				participant.LastSeen = conn.LastSeen
				state.Participants[participantID] = participant

				if conn.IsConnected {
					activeCount++
				}
			}
		}
		state.ActiveCount = activeCount
	}

	return state, nil
}

// PublishEvent publishes an event for a quiz
func (s *stateServiceImpl) PublishEvent(ctx context.Context, quizID uuid.UUID, eventType string, payload interface{}) error {
	// Generate sequence number
	seqNum, err := s.stateRepo.IncrementSequenceNumber(ctx, quizID)
	if err != nil {
		return err
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create and store the event
	event := model.NewQuizEvent(quizID, eventType, payloadJSON, seqNum)
	if err := s.stateRepo.StoreEvent(ctx, event); err != nil {
		return err
	}

	// Also broadcast via WebSocket
	wsEvent := websocket.Event{
		Type:    websocket.EventType(eventType),
		Payload: payload,
	}
	s.wsHub.BroadcastToQuiz(quizID, wsEvent)

	return nil
}

// GetMissedEvents retrieves events that a client missed
func (s *stateServiceImpl) GetMissedEvents(ctx context.Context, quizID uuid.UUID, lastSequence int64) ([]*model.QuizEvent, error) {
	// Limit to 100 most recent events
	return s.stateRepo.GetMissedEvents(ctx, quizID, lastSequence, 100)
}

// UpdateParticipantConnection updates a participant's connection status
func (s *stateServiceImpl) UpdateParticipantConnection(
	ctx context.Context,
	participantID, quizID uuid.UUID,
	isConnected bool,
	instanceID string,
) error {
	conn := model.NewParticipantConnection(participantID, quizID, instanceID)
	conn.IsConnected = isConnected

	// Update the connection in the database
	err := s.stateRepo.UpdateParticipantConnection(ctx, conn)
	if err != nil {
		return err
	}

	// If connection state changed, broadcast participant joined/left event
	if isConnected {
		// Get participant details
		participant, err := s.participantRepo.GetParticipantByID(ctx, participantID)
		if err != nil {
			return err
		}

		// Broadcast user joined event
		s.PublishEvent(ctx, quizID, string(websocket.EventUserJoined), map[string]interface{}{
			"id":       participantID.String(),
			"name":     participant.Name,
			"joinTime": time.Now().Format(time.RFC3339),
		})
	} else {
		// Broadcast user left event
		s.PublishEvent(ctx, quizID, string(websocket.EventUserLeft), map[string]interface{}{
			"id":        participantID.String(),
			"leaveTime": time.Now().Format(time.RFC3339),
		})
	}

	return nil
}

// GetActiveParticipants retrieves all active participants for a quiz
func (s *stateServiceImpl) GetActiveParticipants(ctx context.Context, quizID uuid.UUID) ([]model.Participant, error) {
	// Get all active connections
	cutoffTime := time.Now().Add(-30 * time.Second)
	connections, err := s.stateRepo.GetActiveParticipantConnections(ctx, quizID, cutoffTime)
	if err != nil {
		return nil, err
	}

	// Get participants by their IDs
	var participants []model.Participant
	for _, conn := range connections {
		if conn.IsConnected {
			participant, err := s.participantRepo.GetParticipantByID(ctx, conn.ParticipantID)
			if err == nil {
				participants = append(participants, *participant)
			}
		}
	}

	return participants, nil
}

// RegisterInstance registers a server instance
func (s *stateServiceImpl) RegisterInstance(ctx context.Context, instanceID string) error {
	instance := model.NewServerInstance(instanceID)
	return s.stateRepo.RegisterInstance(ctx, instance)
}

// UpdateInstanceHeartbeat updates a server instance's heartbeat
func (s *stateServiceImpl) UpdateInstanceHeartbeat(ctx context.Context, instanceID string) error {
	return s.stateRepo.UpdateInstanceHeartbeat(ctx, instanceID)
}
