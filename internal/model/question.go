package model

import (
	"time"

	"github.com/google/uuid"
)

// Option represents a multiple-choice option
type Option struct {
	Key   string `json:"key"`   // A, B, C, or D
	Value string `json:"value"` // The content of the option
}

// Question represents a quiz question
type Question struct {
	ID            uuid.UUID `json:"id" db:"id"`
	QuizID        uuid.UUID `json:"quizId" db:"quiz_id"`
	Text          string    `json:"text" db:"text"`
	Options       []Option  `json:"options"`
	OptionA       string    `json:"-" db:"option_a"`
	OptionB       string    `json:"-" db:"option_b"`
	OptionC       string    `json:"-" db:"option_c"`
	OptionD       string    `json:"-" db:"option_d"`
	CorrectAnswer string    `json:"correctAnswer" db:"correct_answer"` // A, B, C, or D
	TimeLimit     int       `json:"timeLimit" db:"time_limit"`         // Time limit in seconds
	Order         int       `json:"order" db:"order"`                  // Question sequence in the quiz
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// GetOptions returns the question options as a slice
func (q *Question) GetOptions() []Option {
	return []Option{
		{Key: "A", Value: q.OptionA},
		{Key: "B", Value: q.OptionB},
		{Key: "C", Value: q.OptionC},
		{Key: "D", Value: q.OptionD},
	}
}

// SetOptions converts the options slice to individual fields for DB storage
func (q *Question) SetOptions(options []Option) {
	for _, opt := range options {
		switch opt.Key {
		case "A":
			q.OptionA = opt.Value
		case "B":
			q.OptionB = opt.Value
		case "C":
			q.OptionC = opt.Value
		case "D":
			q.OptionD = opt.Value
		}
	}
}

// NewQuestion creates a new question for a quiz
func NewQuestion(quizID uuid.UUID, text string, options []Option, correctAnswer string, timeLimit int, order int) *Question {
	q := &Question{
		ID:            uuid.New(),
		QuizID:        quizID,
		Text:          text,
		CorrectAnswer: correctAnswer,
		TimeLimit:     timeLimit,
		Order:         order,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	q.SetOptions(options)
	return q
}
