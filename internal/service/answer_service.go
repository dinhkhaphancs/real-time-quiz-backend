package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// answerServiceImpl implements AnswerService interface
type answerServiceImpl struct {
	answerRepo      repository.AnswerRepository
	questionRepo    repository.QuestionRepository
	participantRepo repository.ParticipantRepository
	wsHub           *websocket.RedisHub
}

// NewAnswerService creates a new answer service
func NewAnswerService(
	answerRepo repository.AnswerRepository,
	questionRepo repository.QuestionRepository,
	participantRepo repository.ParticipantRepository,
	wsHub *websocket.RedisHub,
) AnswerService {
	return &answerServiceImpl{
		answerRepo:      answerRepo,
		questionRepo:    questionRepo,
		participantRepo: participantRepo,
		wsHub:           wsHub,
	}
}

// SubmitAnswer records a participant's answer to a question
func (s *answerServiceImpl) SubmitAnswer(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID, selectedOption string) (*model.Answer, error) {
	// Validate option format
	if selectedOption != "A" && selectedOption != "B" && selectedOption != "C" && selectedOption != "D" {
		return nil, errors.New("invalid option selected")
	}

	// Check if participant exists
	participant, err := s.participantRepo.GetParticipantByID(ctx, participantID)
	if err != nil {
		return nil, errors.New("participant not found")
	}

	// Get the question
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	// Check if answer already submitted
	existingAnswer, err := s.answerRepo.GetAnswerByParticipantAndQuestion(ctx, participantID, questionID)
	if err == nil && existingAnswer != nil {
		return nil, errors.New("answer already submitted")
	}

	// Calculate time taken
	// In a real system, we'd store the question start time in Redis or database
	// For now, we'll just use a simplified approach
	timeTaken := 0.0 // This would normally be calculated based on question start time

	// Check if answer is correct
	isCorrect := selectedOption == question.CorrectAnswer

	// Calculate score (simple scoring - correct answer gets full points)
	// In a more sophisticated system, score could be weighted by time taken
	score := 0
	if isCorrect {
		score = 100
	}

	// Create answer record
	answer := model.NewAnswer(participantID, questionID, selectedOption, timeTaken, isCorrect)
	answer.Score = score

	// Save to database
	if err := s.answerRepo.CreateAnswer(ctx, answer); err != nil {
		return nil, err
	}

	// Update participant's total score
	if err := s.participantRepo.UpdateParticipantScore(ctx, participantID, score); err != nil {
		return nil, err
	}

	// Notify participant of answer receipt
	s.wsHub.SendToClient(participantID, participant.QuizID, websocket.Event{
		Type: websocket.EventAnswerReceived,
		Payload: map[string]interface{}{
			"questionId":     questionID.String(),
			"selectedOption": selectedOption,
			"isCorrect":      isCorrect,
			"score":          score,
		},
	})

	return answer, nil
}

// GetAnswerStats retrieves statistics for answers to a question
func (s *answerServiceImpl) GetAnswerStats(ctx context.Context, questionID uuid.UUID) (map[string]int, error) {
	// Get all answers for the question
	answers, err := s.answerRepo.GetAnswersByQuestionID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	// Count occurrences of each option
	stats := map[string]int{
		"A": 0,
		"B": 0,
		"C": 0,
		"D": 0,
	}

	for _, answer := range answers {
		stats[answer.SelectedOption]++
	}

	return stats, nil
}

// GetParticipantAnswer retrieves a participant's answer to a specific question
func (s *answerServiceImpl) GetParticipantAnswer(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID) (*model.Answer, error) {
	answer, err := s.answerRepo.GetAnswerByParticipantAndQuestion(ctx, participantID, questionID)
	if err != nil {
		return nil, err
	}
	return answer, nil
}
