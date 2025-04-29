package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

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

// StartQuestion starts a question and updates the phase
func (s *stateServiceImpl) StartQuestion(ctx context.Context, quizID uuid.UUID, questionID uuid.UUID) error {
	// Get current session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	// Check if quiz exists and is active
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}
	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Check if question exists and belongs to this quiz
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return ErrQuestionNotFound
	}
	if question.QuizID != quizID {
		return errors.New("question does not belong to this quiz")
	}

	// Load options for the question
	options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, questionID)
	if err == nil {
		question.Options = options
	}

	// Update quiz session with current question and phase
	now := time.Now()
	session.CurrentQuestionID = &questionID
	session.CurrentQuestionStartedAt = &now
	session.CurrentQuestionEndedAt = nil // Clear any previous end time
	session.CurrentPhase = model.QuizPhaseQuestionActive

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Get total question count for better UI experience
	questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		// Non-critical error, we can continue with approximate count
		questions = []*model.Question{question}
	}
	totalCount := len(questions)

	// Broadcast different payloads for creators and participants
	// For creators (quiz admins), send full question details including correct answers
	creatorOptions := make([]map[string]interface{}, len(question.Options))
	for i, opt := range question.Options {
		creatorOptions[i] = map[string]interface{}{
			"id":        opt.ID.String(),
			"text":      opt.Text,
			"isCorrect": opt.IsCorrect,
		}
	}

	creatorEvent := map[string]interface{}{
		"quizId":       quiz.ID.String(),
		"quizTitle":    quiz.Title,
		"questionId":   question.ID.String(),
		"text":         question.Text,
		"options":      creatorOptions,
		"questionType": string(question.QuestionType),
		"timeLimit":    question.TimeLimit,
		"order":        question.Order,
		"totalCount":   totalCount,
		"currentPhase": string(session.CurrentPhase),
		"startTime":    now.Format(time.RFC3339),
	}

	// Publish creator event directly to WebSocket as it's targeted only to creators
	s.wsHub.PublishToCreators(quizID, websocket.Event{
		Type:    websocket.EventQuestionStart,
		Payload: creatorEvent,
	})

	// For participants, send options without correct answer information
	participantOptions := make([]map[string]interface{}, len(question.Options))
	for i, opt := range question.Options {
		participantOptions[i] = map[string]interface{}{
			"id":   opt.ID.String(),
			"text": opt.Text,
		}
	}

	participantEvent := map[string]interface{}{
		"quizId":       quiz.ID.String(),
		"quizTitle":    quiz.Title,
		"questionId":   question.ID.String(),
		"text":         question.Text,
		"options":      participantOptions,
		"questionType": string(question.QuestionType),
		"timeLimit":    question.TimeLimit,
		"order":        question.Order,
		"totalCount":   totalCount,
		"currentPhase": string(session.CurrentPhase),
		"startTime":    now.Format(time.RFC3339),
	}

	// Publish participant event directly to WebSocket as it's targeted only to participants
	s.wsHub.PublishToParticipants(quizID, websocket.Event{
		Type:    websocket.EventQuestionStart,
		Payload: participantEvent,
	})

	// Start a timer to broadcast countdown and end the question
	go s.wsHub.StartTimerBroadcast(quizID, question.TimeLimit)

	// Start a goroutine to automatically end the question after the time limit
	go func() {
		timer := time.NewTimer(time.Duration(question.TimeLimit) * time.Second)
		<-timer.C

		// End the question automatically
		if err := s.EndQuestion(context.Background(), quizID); err != nil {
			// Log the error but don't stop execution
			// In a real application, we would use a proper logger
			// logger.Error("Error ending question", "error", err)
		}
	}()

	return nil
}

// EndQuestion ends the current question and updates the phase
func (s *stateServiceImpl) EndQuestion(ctx context.Context, quizID uuid.UUID) error {
	// Get current session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	if session.CurrentQuestionID == nil {
		return errors.New("no active question to end")
	}

	// Check if quiz exists and is active
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}
	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Get question details with options
	question, err := s.questionRepo.GetQuestionByID(ctx, *session.CurrentQuestionID)
	if err != nil {
		return ErrQuestionNotFound
	}

	// Load options for the question
	options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, question.ID)
	if err == nil {
		question.Options = options
	}

	// Update the session to record question end and change phase
	now := time.Now()
	session.CurrentQuestionEndedAt = &now
	session.CurrentPhase = model.QuizPhaseShowingResults

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Get correct options to send in the event
	var correctOptions []*model.QuestionOption
	for _, opt := range question.Options {
		if opt.IsCorrect {
			correctOptions = append(correctOptions, opt)
		}
	}

	correctOptionIds := make([]string, len(correctOptions))
	for i, opt := range correctOptions {
		correctOptionIds[i] = opt.ID.String()
	}

	// Broadcast question end event with correct answers
	return s.PublishEvent(ctx, quizID, string(websocket.EventQuestionEnd), map[string]interface{}{
		"questionId":       question.ID.String(),
		"correctOptionIds": correctOptionIds,
		"questionType":     string(question.QuestionType),
		"currentPhase":     string(session.CurrentPhase),
		"endTime":          now.Format(time.RFC3339),
	})
}

// MoveToNextQuestion prepares for the next question
func (s *stateServiceImpl) MoveToNextQuestion(ctx context.Context, quizID uuid.UUID) error {
	// Check if quiz exists and is active
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}
	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Get current session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	// Update phase to BETWEEN_QUESTIONS
	session.CurrentPhase = model.QuizPhaseBetweenQuestions

	// Try to determine the next question
	var nextQuestion *model.Question

	// If there's no current question, get the first question
	if session.CurrentQuestionID == nil {
		questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
		if err == nil && len(questions) > 0 {
			// Find question with order 1
			for _, q := range questions {
				if q.Order == 1 {
					nextQuestion = q
					break
				}
			}

			// If no question with order 1, use the first in the list
			if nextQuestion == nil {
				nextQuestion = questions[0]
			}
		}
	} else {
		// Get current question to determine order
		currentQuestion, err := s.questionRepo.GetQuestionByID(ctx, *session.CurrentQuestionID)
		if err == nil {
			// Get next question
			nextQuestion, _ = s.questionRepo.GetNextQuestion(ctx, quizID, currentQuestion.Order)
		}
	}

	// Update session with next question info
	if nextQuestion != nil {
		session.NextQuestionID = &nextQuestion.ID
	} else {
		session.NextQuestionID = nil
	}

	// Update session
	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast the phase change
	return s.PublishEvent(ctx, quizID, "PHASE_CHANGE", map[string]interface{}{
		"quizId":       quizID.String(),
		"currentPhase": string(session.CurrentPhase),
		"hasNext":      session.NextQuestionID != nil,
	})
}

// StartQuiz starts a quiz session
func (s *stateServiceImpl) StartQuiz(ctx context.Context, quizID uuid.UUID) error {
	// Get the quiz and session
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}

	if quiz.Status != model.QuizStatusWaiting {
		return ErrQuizAlreadyStarted
	}

	// Update quiz status
	if err := s.quizRepo.UpdateQuizStatus(ctx, quizID, model.QuizStatusActive); err != nil {
		return err
	}

	// Update session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.Status = model.QuizStatusActive
	session.StartedAt = &now
	// Set the initial phase to BETWEEN_QUESTIONS
	session.CurrentPhase = model.QuizPhaseBetweenQuestions

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz start event to all clients
	return s.PublishEvent(ctx, quizID, string(websocket.EventQuizStart), map[string]interface{}{
		"quizId":       quizID.String(),
		"title":        quiz.Title,
		"description":  quiz.Description,
		"startTime":    now.Format(time.RFC3339),
		"currentPhase": string(model.QuizPhaseBetweenQuestions),
	})
}

// EndQuiz ends a quiz session
func (s *stateServiceImpl) EndQuiz(ctx context.Context, quizID uuid.UUID) error {
	// Get the quiz and session
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}

	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Update quiz status
	if err := s.quizRepo.UpdateQuizStatus(ctx, quizID, model.QuizStatusCompleted); err != nil {
		return err
	}

	// Update session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.Status = model.QuizStatusCompleted
	session.EndedAt = &now
	// Phase is not relevant for completed quizzes, but for clarity set it to a final state
	session.CurrentPhase = model.QuizPhaseBetweenQuestions
	session.CurrentQuestionID = nil
	session.CurrentQuestionStartedAt = nil
	session.CurrentQuestionEndedAt = nil

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz end event to all clients
	return s.PublishEvent(ctx, quizID, string(websocket.EventQuizEnd), map[string]interface{}{
		"quizId":   quizID.String(),
		"endTime":  now.Format(time.RFC3339),
		"title":    quiz.Title,
		"duration": session.EndedAt.Sub(*session.StartedAt).Seconds(),
	})
}
