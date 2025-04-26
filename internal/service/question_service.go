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

// Error definitions for the question service
var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrNoQuestions      = errors.New("no questions available")
	ErrEmptyOptions     = errors.New("question must have options")
	ErrInvalidOption    = errors.New("invalid option selected")
)

// questionService implements QuestionService interface
type questionService struct {
	quizRepo     repository.QuizRepository
	questionRepo repository.QuestionRepository
	wsHub        *websocket.RedisHub
}

// NewQuestionService creates a new question service
func NewQuestionService(
	quizRepo repository.QuizRepository,
	questionRepo repository.QuestionRepository,
	wsHub *websocket.RedisHub,
) QuestionService {
	return &questionService{
		quizRepo:     quizRepo,
		questionRepo: questionRepo,
		wsHub:        wsHub,
	}
}

// AddQuestion adds a question to a quiz
func (s *questionService) AddQuestion(ctx context.Context, quizID uuid.UUID, text string, options []model.Option, correctAnswer string, timeLimit int) (*model.Question, error) {
	// Validate inputs
	if text == "" {
		return nil, errors.New("question text is required")
	}
	if len(options) != 4 {
		return nil, errors.New("question must have exactly 4 options")
	}
	if correctAnswer == "" || (correctAnswer != "A" && correctAnswer != "B" && correctAnswer != "C" && correctAnswer != "D") {
		return nil, errors.New("correct answer must be A, B, C, or D")
	}
	if timeLimit <= 0 {
		return nil, errors.New("time limit must be positive")
	}

	// Get the quiz to check existence and status
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}

	// Only allow adding questions to quizzes in WAITING state
	if quiz.Status != model.QuizStatusWaiting {
		return nil, errors.New("cannot add questions to a quiz that has already started")
	}

	// Get existing questions to determine order
	questions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Create new question with next order number
	question := model.NewQuestion(quizID, text, options, correctAnswer, timeLimit, len(questions)+1)

	// Save to database
	if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
		return nil, err
	}

	return question, nil
}

// GetQuestions retrieves all questions for a quiz
func (s *questionService) GetQuestions(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error) {
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

	return questions, nil
}

// GetQuestion retrieves a question by ID
func (s *questionService) GetQuestion(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	question, err := s.questionRepo.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, ErrQuestionNotFound
	}
	return question, nil
}

// StartQuestion starts a specific question in a quiz
func (s *questionService) StartQuestion(ctx context.Context, quizID uuid.UUID, questionID uuid.UUID) error {
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

	// Different payloads for creators and participants
	// For creators (quiz admins), send full question details
	s.wsHub.PublishToCreators(quizID, websocket.Event{
		Type: websocket.EventQuestionStart,
		Payload: map[string]interface{}{
			"questionId": question.ID.String(),
			"text":       question.Text,
			"options":    question.GetOptions(),
			"timeLimit":  question.TimeLimit,
			"order":      question.Order,
			"totalCount": totalCount,
		},
	})

	// For participants, send options without text and correct answer
	s.wsHub.PublishToParticipants(quizID, websocket.Event{
		Type: websocket.EventQuestionStart,
		Payload: map[string]interface{}{
			"questionId": question.ID.String(),
			"options": []map[string]string{
				{"key": "A"},
				{"key": "B"},
				{"key": "C"},
				{"key": "D"},
			},
			"timeLimit":  question.TimeLimit,
			"order":      question.Order,
			"totalCount": totalCount,
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
func (s *questionService) EndQuestion(ctx context.Context, quizID uuid.UUID) error {
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

	// Get question details
	question, err := s.questionRepo.GetQuestionByID(ctx, *session.CurrentQuestionID)
	if err != nil {
		return ErrQuestionNotFound
	}

	// Broadcast question end event with correct answer
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuestionEnd,
		Payload: map[string]interface{}{
			"questionId":    question.ID.String(),
			"correctAnswer": question.CorrectAnswer,
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
func (s *questionService) GetNextQuestion(ctx context.Context, quizID uuid.UUID) (*model.Question, error) {
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
		for _, q := range questions {
			if q.Order == 1 {
				return q, nil
			}
		}

		// If no question with order 1, return the first in the list
		return questions[0], nil
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

	return nextQuestion, nil
}
