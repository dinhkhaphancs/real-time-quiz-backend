package dto

import (
	"github.com/google/uuid"
)

// ClientMessage represents a message sent from a client to the server via WebSocket
type ClientMessage struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

// ClientJoinQuizRequest represents a request to join a quiz
type ClientJoinQuizRequest struct {
	QuizID uuid.UUID `json:"quizId" binding:"required"`
	Name   string    `json:"name" binding:"required"`
}

// ClientSubmitAnswerRequest represents a request to submit an answer to a quiz question
type ClientSubmitAnswerRequest struct {
	QuestionID      uuid.UUID `json:"questionId" binding:"required"`
	SelectedOptions []string  `json:"selectedOptions" binding:"required,min=1"`
	TimeTaken       float64   `json:"timeTaken" binding:"required,min=0"`
}

// ClientStartQuizRequest represents a request to start a quiz
type ClientStartQuizRequest struct {
	QuizID uuid.UUID `json:"quizId" binding:"required"`
}

// ClientNextQuestionRequest represents a request to move to the next question in a quiz
type ClientNextQuestionRequest struct {
	QuizID uuid.UUID `json:"quizId" binding:"required"`
}

// ClientEndQuizRequest represents a request to end a quiz
type ClientEndQuizRequest struct {
	QuizID uuid.UUID `json:"quizId" binding:"required"`
}

// WebSocketConnectionRequest represents a request to establish a WebSocket connection
type WebSocketConnectionRequest struct {
	UserID string `uri:"userId" binding:"required"`
}
