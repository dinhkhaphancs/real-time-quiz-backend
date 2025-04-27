package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Answer represents a participant's answer to a question
type Answer struct {
	ID              uuid.UUID `json:"id" db:"id"`
	ParticipantID   uuid.UUID `json:"participantId" db:"participant_id"`
	QuestionID      uuid.UUID `json:"questionId" db:"question_id"`
	SelectedOptions []string  `json:"selectedOptions" db:"-"`       // Array of selected option IDs
	SelectedJSON    string    `json:"-" db:"selected_options_json"` // JSON string of selected options for DB storage
	AnsweredAt      time.Time `json:"answeredAt" db:"answered_at"`
	TimeTaken       float64   `json:"timeTaken" db:"time_taken"` // Time taken in seconds
	IsCorrect       bool      `json:"isCorrect" db:"is_correct"`
	Score           int       `json:"score" db:"score"`
}

// SetSelectedOptions sets the selected options and updates the JSON representation
func (a *Answer) SetSelectedOptions(options []string) error {
	a.SelectedOptions = options

	// Convert to JSON for database storage
	jsonData, err := json.Marshal(options)
	if err != nil {
		return err
	}

	a.SelectedJSON = string(jsonData)
	return nil
}

// GetSelectedOptions loads the selected options from the JSON string
func (a *Answer) GetSelectedOptions() ([]string, error) {
	if len(a.SelectedOptions) > 0 {
		return a.SelectedOptions, nil
	}

	// If SelectedOptions is empty but we have JSON, parse it
	if a.SelectedJSON != "" {
		var options []string
		err := json.Unmarshal([]byte(a.SelectedJSON), &options)
		if err != nil {
			return nil, err
		}
		a.SelectedOptions = options
		return options, nil
	}

	return []string{}, nil
}

// NewAnswer creates a new answer record
func NewAnswer(participantID, questionID uuid.UUID, selectedOptions []string, timeTaken float64, isCorrect bool) (*Answer, error) {
	score := 0
	if isCorrect {
		score = 100
	}

	answer := &Answer{
		ID:            uuid.New(),
		ParticipantID: participantID,
		QuestionID:    questionID,
		AnsweredAt:    time.Now(),
		TimeTaken:     timeTaken,
		IsCorrect:     isCorrect,
		Score:         score,
	}

	// Set the selected options
	if err := answer.SetSelectedOptions(selectedOptions); err != nil {
		return nil, err
	}

	return answer, nil
}
