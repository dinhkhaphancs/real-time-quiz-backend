package service

import (
	"context"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// answerService implements AnswerService interface
type answerService struct {
	answerRepo   repository.AnswerRepository
	questionRepo repository.QuestionRepository
	userRepo     repository.UserRepository
	wsHub        *websocket.RedisHub
}

// NewAnswerService creates a new answer service
func NewAnswerService(
	answerRepo repository.AnswerRepository,
	questionRepo repository.QuestionRepository,
	userRepo repository.UserRepository,
	wsHub *websocket.RedisHub,
) AnswerService {
	return &answerService{
		answerRepo:   answerRepo,
		questionRepo: questionRepo,
		userRepo:     userRepo,
		wsHub:        wsHub,
	}
}

// SubmitAnswer records a user's answer to a question
func (s *answerService) SubmitAnswer(ctx context.Context, userID uuid.UUID, questionID uuid.UUID, selectedOption string) (*model.Answer, error) {
	// Validate option format
	if selectedOption != "A" && selectedOption != "B" && selectedOption != "C" && selectedOption != "D" {
		return nil, ErrInvalidOption
	}

	// Check if user exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get the question
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	// Check if answer already submitted
	existingAnswer, err := s.answerRepo.GetAnswerByUserAndQuestion(ctx, userID, questionID)
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
	answer := model.NewAnswer(userID, questionID, selectedOption, timeTaken, isCorrect)
	answer.Score = score

	// Save to database
	if err := s.answerRepo.CreateAnswer(ctx, answer); err != nil {
		return nil, err
	}

	// Update user's total score
	if err := s.userRepo.UpdateUserScore(ctx, userID, score); err != nil {
		return nil, err
	}

	// Notify user of answer receipt
	s.wsHub.SendToClient(uuid.New(), user.QuizID, websocket.Event{
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
func (s *answerService) GetAnswerStats(ctx context.Context, questionID uuid.UUID) (map[string]int, error) {
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

// GetUserAnswer retrieves a user's answer to a specific question
func (s *answerService) GetUserAnswer(ctx context.Context, userID uuid.UUID, questionID uuid.UUID) (*model.Answer, error) {
	answer, err := s.answerRepo.GetAnswerByUserAndQuestion(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}
	return answer, nil
}
