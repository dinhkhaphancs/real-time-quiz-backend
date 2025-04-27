package model

import (
	"time"

	"github.com/google/uuid"
)

// QuestionType represents the type of question
type QuestionType string

const (
	// QuestionTypeSingleChoice represents a single choice question
	QuestionTypeSingleChoice QuestionType = "SINGLE_CHOICE"
	// QuestionTypeMultipleChoice represents a multiple choice question
	QuestionTypeMultipleChoice QuestionType = "MULTIPLE_CHOICE"
)

// Question represents a quiz question
type Question struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	QuizID       uuid.UUID         `json:"quizId" db:"quiz_id"`
	Text         string            `json:"text" db:"text"`
	QuestionType QuestionType      `json:"questionType" db:"question_type"`
	TimeLimit    int               `json:"timeLimit" db:"time_limit"`
	Order        int               `json:"order" db:"order"`
	CreatedAt    time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time         `json:"updatedAt" db:"updated_at"`
	Options      []*QuestionOption `json:"options" db:"-"` // Will be loaded separately from DB
}

// GetCorrectOptions returns all correct options for the question
func (q *Question) GetCorrectOptions() []*QuestionOption {
	var correctOptions []*QuestionOption
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correctOptions = append(correctOptions, opt)
		}
	}
	return correctOptions
}

// IsCorrectAnswer checks if the provided option IDs represent a correct answer
func (q *Question) IsCorrectAnswer(selectedOptionIDs []string) bool {
	// We need options to be loaded
	if len(q.Options) == 0 {
		return false
	}

	// Convert string IDs to a map for easy checking
	selectedMap := make(map[string]bool)
	for _, optID := range selectedOptionIDs {
		selectedMap[optID] = true
	}

	// For single choice questions, ensure exactly one option is selected and it's correct
	if q.QuestionType == QuestionTypeSingleChoice {
		if len(selectedOptionIDs) != 1 {
			return false
		}

		// The single selected option must be correct
		for _, opt := range q.Options {
			if selectedMap[opt.ID.String()] && opt.IsCorrect {
				return true
			}
		}
		return false
	}

	// For multiple choice questions
	// Get all correct options
	correctOptions := q.GetCorrectOptions()

	// For multiple choice, all correct options must be selected, and no incorrect options can be selected
	// First, check if all correct options were selected
	for _, opt := range correctOptions {
		if !selectedMap[opt.ID.String()] {
			return false // Missing a correct option
		}
	}

	// Then check if any incorrect options were selected
	for _, opt := range q.Options {
		if selectedMap[opt.ID.String()] && !opt.IsCorrect {
			return false // An incorrect option was selected
		}
	}

	// All correct options were selected and no incorrect options were selected
	return true
}

// NewQuestion creates a new question with dynamic options
func NewQuestion(quizID uuid.UUID, text string, questionType QuestionType, timeLimit int, order int) *Question {
	return &Question{
		ID:           uuid.New(),
		QuizID:       quizID,
		Text:         text,
		QuestionType: questionType,
		TimeLimit:    timeLimit,
		Order:        order,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Options:      []*QuestionOption{},
	}
}
