package dto

import (
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// QuizStateDTO represents the complete state of a quiz at any point in time
type QuizStateDTO struct {
	QuizID         uuid.UUID                      `json:"quizId"`
	Title          string                         `json:"title"`
	Status         string                         `json:"status"`
	CurrentPhase   model.QuizPhase                `json:"currentPhase"`
	ActiveQuestion *ActiveQuestionStateDTO        `json:"activeQuestion,omitempty"`
	Timer          *TimerStateDTO                 `json:"timer,omitempty"`
	Participants   map[string]ParticipantStateDTO `json:"participants"`
	ActiveCount    int                            `json:"activeCount"`
	SequenceNumber int64                          `json:"sequenceNumber"`
	StartTime      *time.Time                     `json:"startTime,omitempty"`
	EndTime        *time.Time                     `json:"endTime,omitempty"`
}

// ActiveQuestionStateDTO represents the state of the currently active question
type ActiveQuestionStateDTO struct {
	QuestionID     uuid.UUID                `json:"questionId"`
	QuestionText   string                   `json:"questionText"`
	Options        []QuestionOptionStateDTO `json:"options"`
	QuestionType   string                   `json:"questionType"`
	TimeLimit      int                      `json:"timeLimit"`
	StartTime      time.Time                `json:"startTime"`
	EndTime        *time.Time               `json:"endTime,omitempty"`
	Order          int                      `json:"order"`
	TotalQuestions int                      `json:"totalQuestions"`
}

// QuestionOptionStateDTO represents an option for a question
type QuestionOptionStateDTO struct {
	ID        uuid.UUID `json:"id"`
	Text      string    `json:"text"`
	IsCorrect bool      `json:"isCorrect,omitempty"` // Only visible to creators
}

// TimerStateDTO represents the current timer state for a quiz or question
type TimerStateDTO struct {
	StartTime        time.Time `json:"startTime"`
	DurationSeconds  int       `json:"durationSeconds"`
	RemainingSeconds int       `json:"remainingSeconds"`
	IsRunning        bool      `json:"isRunning"`
}

// ParticipantStateDTO represents the state of a participant in a quiz
type ParticipantStateDTO struct {
	ParticipantID     uuid.UUID   `json:"participantId"`
	Nickname          string      `json:"nickname"`
	IsConnected       bool        `json:"isConnected"`
	LastSeen          time.Time   `json:"lastSeen"`
	Score             int         `json:"score"`
	AnsweredQuestions []uuid.UUID `json:"answeredQuestions"`
	Position          int         `json:"position,omitempty"`
}

// ToQuizStateDTO converts a quiz model and session to a QuizStateDTO
func ToQuizStateDTO(quiz *model.Quiz, session *model.QuizSession, participants []*model.Participant, activeQuestion *model.Question, questionCount int) *QuizStateDTO {
	state := &QuizStateDTO{
		QuizID:         quiz.ID,
		Title:          quiz.Title,
		Status:         string(quiz.Status),
		CurrentPhase:   model.QuizPhase(session.CurrentPhase),
		SequenceNumber: 1, // Default value, should be updated
		StartTime:      session.StartedAt,
		EndTime:        session.EndedAt,
		Participants:   make(map[string]ParticipantStateDTO),
		ActiveCount:    0,
	}

	// Add participants
	for _, p := range participants {
		state.Participants[p.ID.String()] = ParticipantStateDTO{
			ParticipantID: p.ID,
			Nickname:      p.Name,
			IsConnected:   false, // Default to not connected
			LastSeen:      time.Now(),
			Score:         p.Score,
		}
	}

	// Add active question if exists
	if activeQuestion != nil && session.CurrentQuestionID != nil {
		options := make([]QuestionOptionStateDTO, len(activeQuestion.Options))
		for i, opt := range activeQuestion.Options {
			options[i] = QuestionOptionStateDTO{
				ID:        opt.ID,
				Text:      opt.Text,
				IsCorrect: opt.IsCorrect,
			}
		}

		state.ActiveQuestion = &ActiveQuestionStateDTO{
			QuestionID:     activeQuestion.ID,
			QuestionText:   activeQuestion.Text,
			Options:        options,
			QuestionType:   string(activeQuestion.QuestionType),
			TimeLimit:      activeQuestion.TimeLimit,
			StartTime:      *session.CurrentQuestionStartedAt,
			Order:          activeQuestion.Order,
			TotalQuestions: questionCount,
		}

		// Add timer if question is active
		if session.CurrentQuestionStartedAt != nil && session.CurrentPhase == model.QuizPhaseQuestionActive {
			elapsed := time.Since(*session.CurrentQuestionStartedAt).Seconds()
			remaining := float64(activeQuestion.TimeLimit) - elapsed
			if remaining < 0 {
				remaining = 0
			}

			state.Timer = &TimerStateDTO{
				StartTime:        *session.CurrentQuestionStartedAt,
				DurationSeconds:  activeQuestion.TimeLimit,
				RemainingSeconds: int(remaining),
				IsRunning:        remaining > 0,
			}
		}
	}

	return state
}
