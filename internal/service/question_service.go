package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// Error definitions for the question service
var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrNoQuestions      = errors.New("no questions available")
	ErrEmptyOptions     = errors.New("question must have options")
	ErrInvalidOption    = errors.New("invalid option selected")
)

// questionServiceImpl implements QuestionService interface
type questionServiceImpl struct {
	quizRepo           repository.QuizRepository
	questionRepo       repository.QuestionRepository
	questionOptionRepo repository.QuestionOptionRepository
	wsHub              *websocket.RedisHub
}

// NewQuestionService creates a new question service
func NewQuestionService(
	quizRepo repository.QuizRepository,
	questionRepo repository.QuestionRepository,
	questionOptionRepo repository.QuestionOptionRepository,
	wsHub *websocket.RedisHub,
) QuestionService {
	return &questionServiceImpl{
		quizRepo:           quizRepo,
		questionRepo:       questionRepo,
		questionOptionRepo: questionOptionRepo,
		wsHub:              wsHub,
	}
}

// AddQuestion adds a question to a quiz
func (s *questionServiceImpl) AddQuestion(ctx context.Context, quizID uuid.UUID, text string, options []dto.OptionCreateData, questionType string, timeLimit int) (*model.Question, error) {
	// Validate inputs
	if text == "" {
		return nil, errors.New("question text is required")
	}

	if len(options) < 2 {
		return nil, errors.New("question must have at least 2 options")
	}

	// Validate question type
	var qType model.QuestionType
	switch questionType {
	case string(model.QuestionTypeSingleChoice):
		qType = model.QuestionTypeSingleChoice
	case string(model.QuestionTypeMultipleChoice):
		qType = model.QuestionTypeMultipleChoice
	default:
		return nil, errors.New("invalid question type")
	}

	// Check if quiz exists
	_, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, errors.New("quiz not found")
	}

	// Get the current count of questions for this quiz to determine order
	existingQuestions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	order := len(existingQuestions) + 1

	// Create the question
	question := model.NewQuestion(quizID, text, qType, timeLimit, order)

	// Save to database
	if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
		return nil, err
	}

	// For single choice questions, ensure only one option is marked as correct
	if qType == model.QuestionTypeSingleChoice {
		correctCount := 0
		for _, opt := range options {
			if opt.IsCorrect {
				correctCount++
			}
		}
		if correctCount != 1 {
			return nil, errors.New("single choice questions must have exactly one correct option")
		}
	} else {
		// For multiple choice, ensure at least one option is correct
		correctCount := 0
		for _, opt := range options {
			if opt.IsCorrect {
				correctCount++
			}
		}
		if correctCount < 1 {
			return nil, errors.New("multiple choice questions must have at least one correct option")
		}
	}

	// Create options for the question
	for i, optData := range options {
		option := model.NewQuestionOption(
			question.ID,
			optData.Text,
			optData.IsCorrect,
			i+1, // display order based on position in array
		)

		if err := s.questionOptionRepo.CreateQuestionOption(ctx, option); err != nil {
			return nil, err
		}
	}

	// Fetch the complete question with options
	return s.GetQuestion(ctx, question.ID)
}

// GetQuestions retrieves all questions for a quiz
func (s *questionServiceImpl) GetQuestions(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error) {
	// Check if quiz exists
	_, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}

	// Retrieve questions
	questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Load options for each question
	for _, question := range questions {
		options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return nil, err
		}
		fmt.Println("Options: ", options)
		question.Options = options
	}

	return questions, nil
}

// GetQuestion retrieves a question by ID
func (s *questionServiceImpl) GetQuestion(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, ErrQuestionNotFound
	}

	// Load options for the question
	options, err := s.questionOptionRepo.GetQuestionOptionsByQuestionID(ctx, question.ID)
	if err != nil {
		return nil, err
	}
	question.Options = options

	return question, nil
}

// StartQuestion starts a specific question in a quiz
func (s *questionServiceImpl) StartQuestion(ctx context.Context, quizID uuid.UUID, questionID uuid.UUID) error {
	// Check if quiz exists and is active
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}
	if quiz.Status != model.QuizStatusActive {
		return ErrQuizNotActive
	}

	// Check if question exists and belongs to this quiz
	question, err := s.GetQuestion(ctx, questionID) // Use GetQuestion to load options
	if err != nil {
		return ErrQuestionNotFound
	}
	if question.QuizID != quizID {
		return errors.New("question does not belong to this quiz")
	}

	// Update quiz session with current question
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.CurrentQuestionID = &questionID
	session.CurrentQuestionStartedAt = &now

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

	// Convert options to a format suitable for broadcasting
	options := make([]map[string]interface{}, len(question.Options))
	for i, opt := range question.Options {
		options[i] = map[string]interface{}{
			"id":   opt.ID.String(),
			"text": opt.Text,
		}
	}

	// Different payloads for creators and participants
	// For creators (quiz admins), send full question details including correct answers
	creatorOptions := make([]map[string]interface{}, len(question.Options))
	for i, opt := range question.Options {
		creatorOptions[i] = map[string]interface{}{
			"id":        opt.ID.String(),
			"text":      opt.Text,
			"isCorrect": opt.IsCorrect,
		}
	}

	s.wsHub.PublishToCreators(quizID, websocket.Event{
		Type: websocket.EventQuestionStart,
		Payload: map[string]interface{}{
			"questionId":   question.ID.String(),
			"text":         question.Text,
			"options":      creatorOptions,
			"questionType": string(question.QuestionType),
			"timeLimit":    question.TimeLimit,
			"order":        question.Order,
			"totalCount":   totalCount,
		},
	})

	// For participants, send options without correct answer information
	s.wsHub.PublishToParticipants(quizID, websocket.Event{
		Type: websocket.EventQuestionStart,
		Payload: map[string]interface{}{
			"questionId":   question.ID.String(),
			"text":         question.Text,
			"options":      options,
			"questionType": string(question.QuestionType),
			"timeLimit":    question.TimeLimit,
			"order":        question.Order,
			"totalCount":   totalCount,
		},
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

// EndQuestion ends the current question in a quiz
func (s *questionServiceImpl) EndQuestion(ctx context.Context, quizID uuid.UUID) error {
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

	if session.CurrentQuestionID == nil {
		return errors.New("no active question to end")
	}

	// Get question details with options
	question, err := s.GetQuestion(ctx, *session.CurrentQuestionID)
	if err != nil {
		return ErrQuestionNotFound
	}

	// Get correct options to send in the event
	correctOptions := question.GetCorrectOptions()
	correctOptionIds := make([]string, len(correctOptions))
	for i, opt := range correctOptions {
		correctOptionIds[i] = opt.ID.String()
	}

	// Broadcast question end event with correct answers
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuestionEnd,
		Payload: map[string]interface{}{
			"questionId":       question.ID.String(),
			"correctOptionIds": correctOptionIds,
			"questionType":     string(question.QuestionType),
		},
	})

	// Clear current question in session
	session.CurrentQuestionID = nil
	session.CurrentQuestionStartedAt = nil

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	return nil
}

// GetNextQuestion retrieves the next question in sequence
func (s *questionServiceImpl) GetNextQuestion(ctx context.Context, quizID uuid.UUID) (*model.Question, error) {
	// Check if quiz exists and is active
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	if quiz.Status != model.QuizStatusActive {
		return nil, ErrQuizNotActive
	}

	// Get current session
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// If no current question, get the first question
	if session.CurrentQuestionID == nil {
		questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
		if err != nil {
			return nil, err
		}
		if len(questions) == 0 {
			return nil, ErrNoQuestions
		}

		// Find question with order 1
		var firstQuestion *model.Question
		for _, q := range questions {
			if q.Order == 1 {
				firstQuestion = q
				break
			}
		}

		// If no question with order 1, use the first in the list
		if firstQuestion == nil {
			firstQuestion = questions[0]
		}

		// Load options for the first question
		return s.GetQuestion(ctx, firstQuestion.ID)
	}

	// Get current question to determine order
	currentQuestion, err := s.questionRepo.GetQuestionByID(ctx, *session.CurrentQuestionID)
	if err != nil {
		return nil, ErrQuestionNotFound
	}

	// Get next question
	nextQuestion, err := s.questionRepo.GetNextQuestion(ctx, quizID, currentQuestion.Order)
	if err != nil {
		return nil, ErrNoQuestions
	}

	// Load options for the next question
	return s.GetQuestion(ctx, nextQuestion.ID)
}
