package service

import (
	"context"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// answerServiceImpl implements AnswerService interface
type answerServiceImpl struct {
	answerRepo         repository.AnswerRepository
	questionRepo       repository.QuestionRepository
	participantRepo    repository.ParticipantRepository
	quizRepo           repository.QuizRepository
	leaderboardService LeaderboardService
	questionOptionRepo repository.QuestionOptionRepository
	wsHub              *websocket.RedisHub
}

// NewAnswerService creates a new answer service
func NewAnswerService(
	answerRepo repository.AnswerRepository,
	questionRepo repository.QuestionRepository,
	participantRepo repository.ParticipantRepository,
	quizRepo repository.QuizRepository,
	leaderboardService LeaderboardService,
	questionOptionRepo repository.QuestionOptionRepository,
	wsHub *websocket.RedisHub,
) AnswerService {
	return &answerServiceImpl{
		answerRepo:         answerRepo,
		questionRepo:       questionRepo,
		participantRepo:    participantRepo,
		quizRepo:           quizRepo,
		leaderboardService: leaderboardService,
		questionOptionRepo: questionOptionRepo,
		wsHub:              wsHub,
	}
}

// SubmitAnswer records a participant's answer to a question
func (s *answerServiceImpl) SubmitAnswer(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID, selectedOptionIDs []string) (*model.Answer, error) {
	// Verify participant exists
	_, err := s.participantRepo.GetParticipantByID(ctx, participantID)
	if err != nil {
		return nil, errors.New("participant not found")
	}

	// Verify question exists
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	// Get quiz session
	session, err := s.quizRepo.GetQuizSession(ctx, question.QuizID)
	if err != nil {
		return nil, err
	}

	// get question options
	options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	// Check if this question is active
	// if session.CurrentQuestionID == nil || *session.CurrentQuestionID != questionID {
	// 	return nil, errors.New("question is not active")
	// }

	// Check if participant has already answered this question
	existingAnswer, err := s.answerRepo.GetAnswerByParticipantAndQuestion(ctx, participantID, questionID)
	if err == nil && existingAnswer != nil {
		return nil, errors.New("already answered this question")
	}

	// Calculate time taken
	var timeTaken float64
	if session.CurrentQuestionStartedAt != nil {
		timeTaken = time.Since(*session.CurrentQuestionStartedAt).Seconds()
	}

	// Validate selected options against question type
	if len(selectedOptionIDs) == 0 {
		return nil, errors.New("no option selected")
	}

	// For single choice questions, ensure only one option is selected
	if question.QuestionType == model.QuestionTypeSingleChoice && len(selectedOptionIDs) > 1 {
		return nil, errors.New("only one option can be selected for single choice questions")
	}

	// Check if options are valid
	optionMap := make(map[string]bool)
	for _, opt := range options {
		optionMap[opt.ID.String()] = true
	}

	for _, optID := range selectedOptionIDs {
		if !optionMap[optID] {
			return nil, errors.New("invalid option selected")
		}
	}

	// Check if answer is correct (using Question.IsCorrectAnswer method)
	isCorrect := question.IsCorrectAnswer(selectedOptionIDs)

	// Create and save the answer
	answer, err := model.NewAnswer(participantID, questionID, selectedOptionIDs, timeTaken, isCorrect)
	if err != nil {
		return nil, err
	}

	if err := s.answerRepo.CreateAnswer(ctx, answer); err != nil {
		return nil, err
	}

	// Update participant's score
	if isCorrect {
		// Calculate a time-based bonus
		timeBonus := 0
		if timeTaken < float64(question.TimeLimit)/2 {
			// If answered in less than half the time limit, award a bonus
			timeBonus = 20
		}
		totalScore := answer.Score + timeBonus

		if err := s.leaderboardService.UpdateParticipantScore(ctx, participantID, totalScore); err != nil {
			// Log the error but continue (non-critical failure)
			// In a real app, we would use a proper logger
			// logger.Error("Failed to update participant score", "error", err)
		}
	}

	return answer, nil
}

// GetAnswerStats retrieves statistics for answers to a question
func (s *answerServiceImpl) GetAnswerStats(ctx context.Context, questionID uuid.UUID) (map[string]int, error) {
	// Get the question to retrieve options
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	// We need to load the options for this question
	options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, questionID)
	if err != nil {
		return nil, err
	}
	question.Options = options

	// Get all answers for the question
	answers, err := s.answerRepo.GetAnswersByQuestionID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	// Initialize stats with all available option IDs
	stats := make(map[string]int)
	for _, option := range question.Options {
		stats[option.ID.String()] = 0
	}

	// Count occurrences of each option selection
	for _, answer := range answers {
		selectedOptions, err := answer.GetSelectedOptions()
		if err != nil {
			continue // Skip this answer if there's an error
		}

		// For multiple choice, each selected option counts as one selection
		for _, optionID := range selectedOptions {
			stats[optionID]++
		}
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
