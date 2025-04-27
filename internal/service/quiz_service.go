package service

import (
	"context"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/repository"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/google/uuid"
)

// Errors
var (
	ErrQuizNotFound       = errors.New("quiz not found")
	ErrQuizAlreadyStarted = errors.New("quiz has already started")
	ErrQuizNotActive      = errors.New("quiz is not active")
)

// quizServiceImpl implements QuizService interface
type quizServiceImpl struct {
	quizRepo           repository.QuizRepository
	userRepo           repository.UserRepository
	questionRepo       repository.QuestionRepository
	questionOptionRepo repository.QuestionOptionRepository
	wsHub              *websocket.RedisHub
}

// NewQuizService creates a new quiz service
func NewQuizService(
	quizRepo repository.QuizRepository,
	userRepo repository.UserRepository,
	questionRepo repository.QuestionRepository,
	questionOptionRepo repository.QuestionOptionRepository,
	wsHub *websocket.RedisHub,
) QuizService {
	return &quizServiceImpl{
		quizRepo:           quizRepo,
		userRepo:           userRepo,
		questionRepo:       questionRepo,
		questionOptionRepo: questionOptionRepo,
		wsHub:              wsHub,
	}
}

// CreateQuiz creates a new quiz with the specified creator
func (s *quizServiceImpl) CreateQuiz(ctx context.Context, title string, description string, creatorID uuid.UUID) (*model.Quiz, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	// Verify user exists
	creator, err := s.userRepo.GetUserByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	// Create the quiz
	quiz := model.NewQuiz(title, description, creator.ID)

	// Create quiz session
	session := model.NewQuizSession(quiz.ID)

	// Save to database
	if err := s.quizRepo.CreateQuiz(ctx, quiz); err != nil {
		return nil, err
	}
	if err := s.quizRepo.CreateQuizSession(ctx, session); err != nil {
		return nil, err
	}

	return quiz, nil
}

// CreateQuizWithQuestions creates a new quiz with questions
func (s *quizServiceImpl) CreateQuizWithQuestions(ctx context.Context, title string, description string, creatorID uuid.UUID, questions []dto.QuestionCreateData) (*model.Quiz, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	if len(questions) == 0 {
		return nil, errors.New("at least one question is required")
	}

	// Verify user exists
	creator, err := s.userRepo.GetUserByID(ctx, creatorID)
	if err != nil {
		return nil, errors.New("creator not found")
	}

	// Create the quiz
	quiz := model.NewQuiz(title, description, creator.ID)

	// Create quiz session
	session := model.NewQuizSession(quiz.ID)

	// Save quiz to database
	if err := s.quizRepo.CreateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	// Save quiz session to database
	if err := s.quizRepo.CreateQuizSession(ctx, session); err != nil {
		return nil, err
	}

	// Create questions
	for i, q := range questions {
		// Validate question data
		if q.Text == "" {
			return nil, errors.New("question text is required")
		}

		if len(q.Options) < 2 {
			return nil, errors.New("question must have at least 2 options")
		}

		// Validate at least one option is marked as correct
		hasCorrectOption := false
		for _, opt := range q.Options {
			if opt.IsCorrect {
				hasCorrectOption = true
				break
			}
		}

		if !hasCorrectOption {
			return nil, errors.New("question must have at least one correct option")
		}

		// Parse question type
		questionType := model.QuestionTypeSingleChoice
		if q.QuestionType == string(model.QuestionTypeMultipleChoice) {
			questionType = model.QuestionTypeMultipleChoice
		}

		if q.TimeLimit <= 0 {
			return nil, errors.New("time limit must be positive")
		}

		// Create question with order based on array position
		question := model.NewQuestion(quiz.ID, q.Text, questionType, q.TimeLimit, i+1)

		// Save question to database
		if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
			return nil, err
		}

		// Add options for the question
		for idx, optData := range q.Options {
			option := model.NewQuestionOption(question.ID, optData.Text, optData.IsCorrect, idx+1)
			// Save option to database
			if err := s.questionOptionRepo.CreateQuestionOption(ctx, option); err != nil {
				return nil, err
			}
		}
	}

	return quiz, nil
}

// GetQuiz retrieves a quiz by ID
func (s *quizServiceImpl) GetQuiz(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByID(ctx, id)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	return quiz, nil
}

// GetQuizByCode retrieves a quiz by its code
func (s *quizServiceImpl) GetQuizByCode(ctx context.Context, code string) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByCode(ctx, code)
	if err != nil {
		return nil, ErrQuizNotFound
	}
	return quiz, nil
}

// StartQuiz starts a quiz session
func (s *quizServiceImpl) StartQuiz(ctx context.Context, quizID uuid.UUID) error {
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

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz start event to all clients
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuizStart,
		Payload: map[string]interface{}{
			"quizId": quizID.String(),
			"title":  quiz.Title,
		},
	})

	return nil
}

// EndQuiz ends a quiz session
func (s *quizServiceImpl) EndQuiz(ctx context.Context, quizID uuid.UUID) error {
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

	if err := s.quizRepo.UpdateQuizSession(ctx, session); err != nil {
		return err
	}

	// Broadcast quiz end event to all clients
	s.wsHub.BroadcastToQuiz(quizID, websocket.Event{
		Type: websocket.EventQuizEnd,
		Payload: map[string]interface{}{
			"quizId": quizID.String(),
		},
	})

	return nil
}

// GetQuizSession retrieves the current state of a quiz
func (s *quizServiceImpl) GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error) {
	session, err := s.quizRepo.GetQuizSession(ctx, quizID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetQuizzesByCreatorID retrieves all quizzes created by a user
func (s *quizServiceImpl) GetQuizzesByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*model.Quiz, error) {
	quizzes, err := s.quizRepo.GetQuizzesByCreatorID(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	return quizzes, nil
}

// UpdateQuiz updates an existing quiz
func (s *quizServiceImpl) UpdateQuiz(ctx context.Context, quizID uuid.UUID, title string, description string) (*model.Quiz, error) {
	// Validate inputs
	if title == "" {
		return nil, errors.New("title is required")
	}

	// Check if quiz exists and get the current data
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}

	// Update fields
	quiz.Title = title
	quiz.Description = description
	quiz.UpdatedAt = time.Now()

	// Save to database
	if err := s.quizRepo.UpdateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	return quiz, nil
}

// UpdateQuizWithQuestions updates an existing quiz with its questions
func (s *quizServiceImpl) UpdateQuizWithQuestions(ctx context.Context, quizID uuid.UUID, title string, description string, questions []dto.QuestionUpdateData) (*model.Quiz, error) {
	// Validate inputs
	if title == "" {
		return nil, errors.New("title is required")
	}

	// Check if quiz exists and get the current data
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, ErrQuizNotFound
	}

	// Only allow updating quizzes in WAITING state
	if quiz.Status != model.QuizStatusWaiting {
		return nil, errors.New("cannot update a quiz that has already started or completed")
	}

	// Update quiz basic fields
	quiz.Title = title
	quiz.Description = description
	quiz.UpdatedAt = time.Now()

	// Save quiz changes to database
	if err := s.quizRepo.UpdateQuiz(ctx, quiz); err != nil {
		return nil, err
	}

	// Get existing questions to track which ones to keep, update, or delete
	existingQuestions, err := s.questionRepo.GetQuestionsByQuizID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Create a map of existing question IDs for easy lookup
	existingQuestionMap := make(map[string]*model.Question)
	for _, q := range existingQuestions {
		existingQuestionMap[q.ID.String()] = q
	}

	// Track questions that are kept in the update to determine which ones to delete
	updatedQuestionIDs := make(map[string]struct{})

	// Process each question in the update
	for i, questionData := range questions {
		// If the question has an ID and it exists in our database, update it
		if questionData.ID != nil && *questionData.ID != "" {
			questionID, err := uuid.Parse(*questionData.ID)
			if err != nil {
				return nil, errors.New("invalid question ID format")
			}

			// Check if this question belongs to the quiz
			existingQuestion, exists := existingQuestionMap[questionID.String()]
			if !exists || existingQuestion.QuizID != quizID {
				return nil, errors.New("question does not belong to this quiz")
			}

			// Mark this question as updated
			updatedQuestionIDs[questionID.String()] = struct{}{}

			// Validate options
			if len(questionData.Options) < 2 {
				return nil, errors.New("question must have at least 2 options")
			}

			// Validate at least one option is marked as correct
			hasCorrectOption := false
			for _, opt := range questionData.Options {
				if opt.IsCorrect {
					hasCorrectOption = true
					break
				}
			}

			if !hasCorrectOption {
				return nil, errors.New("question must have at least one correct option")
			}

			// Parse question type
			questionType := model.QuestionTypeSingleChoice
			if questionData.QuestionType == string(model.QuestionTypeMultipleChoice) {
				questionType = model.QuestionTypeMultipleChoice
			}

			// Update the question
			existingQuestion.Text = questionData.Text
			existingQuestion.TimeLimit = questionData.TimeLimit
			existingQuestion.QuestionType = questionType
			existingQuestion.Order = i + 1 // Update order based on position in array
			existingQuestion.UpdatedAt = time.Now()

			// Save the question first to ensure it exists
			if err := s.questionRepo.UpdateQuestion(ctx, existingQuestion); err != nil {
				return nil, err
			}

			// Handle options - first delete all existing options
			if err := s.questionRepo.DeleteQuestion(ctx, questionID); err != nil {
				return nil, err
			}

			// Create new options
			for _, optData := range questionData.Options {
				option := model.NewQuestionOption(questionID, optData.Text, optData.IsCorrect, optData.DisplayOrder)

				// Save option to database
				if err := s.questionOptionRepo.CreateQuestionOption(ctx, option); err != nil {
					return nil, err
				}
			}
		} else {
			// This is a new question, add it
			if len(questionData.Options) < 2 {
				return nil, errors.New("question must have at least 2 options")
			}

			// Validate at least one option is marked as correct
			hasCorrectOption := false
			for _, opt := range questionData.Options {
				if opt.IsCorrect {
					hasCorrectOption = true
					break
				}
			}

			if !hasCorrectOption {
				return nil, errors.New("question must have at least one correct option")
			}

			// Parse question type
			questionType := model.QuestionTypeSingleChoice
			if questionData.QuestionType == string(model.QuestionTypeMultipleChoice) {
				questionType = model.QuestionTypeMultipleChoice
			}

			// Create question with order based on array position
			question := model.NewQuestion(quizID, questionData.Text, questionType, questionData.TimeLimit, i+1)

			// Save the question first to ensure it has an ID
			if err := s.questionRepo.CreateQuestion(ctx, question); err != nil {
				return nil, err
			}

			// Create options for the question
			for _, optData := range questionData.Options {
				option := model.NewQuestionOption(question.ID, optData.Text, optData.IsCorrect, optData.DisplayOrder)

				// Save option to database
				if err := s.questionOptionRepo.CreateQuestionOption(ctx, option); err != nil {
					return nil, err
				}
			}
		}
	}

	// Delete questions that were not included in the update
	for _, existingQuestion := range existingQuestions {
		if _, exists := updatedQuestionIDs[existingQuestion.ID.String()]; !exists {
			// Question was not in the update, delete it (this should cascade and delete its options)
			if err := s.questionRepo.DeleteQuestion(ctx, existingQuestion.ID); err != nil {
				return nil, err
			}
		}
	}

	return quiz, nil
}

// DeleteQuiz deletes a quiz and all its related data
func (s *quizServiceImpl) DeleteQuiz(ctx context.Context, quizID uuid.UUID) error {
	// Check if quiz exists
	quiz, err := s.quizRepo.GetQuizByID(ctx, quizID)
	if err != nil {
		return ErrQuizNotFound
	}

	// Only allow deleting quizzes in WAITING state
	if quiz.Status != model.QuizStatusWaiting {
		return errors.New("cannot delete a quiz that has already started or completed")
	}

	// Delete the quiz (this will cascade to delete related questions, participants, etc.)
	if err := s.quizRepo.DeleteQuiz(ctx, quizID); err != nil {
		return err
	}

	return nil
}
