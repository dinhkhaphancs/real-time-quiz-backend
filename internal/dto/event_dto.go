package dto

import "time"

// WebSocketEvent represents a generic websocket event structure
type WebSocketEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// EventQuizStart is sent when a quiz begins
type EventQuizStart struct {
	QuizID       uint      `json:"quizId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"startTime"`
	NumQuestions int       `json:"numQuestions"`
}

// EventQuizEnd is sent when a quiz ends
type EventQuizEnd struct {
	QuizID  uint      `json:"quizId"`
	EndTime time.Time `json:"endTime"`
}

// EventQuestionStart is sent when a new question begins
type EventQuestionStart struct {
	QuestionID     uint      `json:"questionId"`
	QuizID         uint      `json:"quizId"`
	Content        string    `json:"content"`
	Options        []string  `json:"options"`
	StartTime      time.Time `json:"startTime"`
	DurationSecs   int       `json:"durationSecs"`
	QuestionNum    int       `json:"questionNum"`
	TotalQuestions int       `json:"totalQuestions"`
}

// EventQuestionEnd is sent when a question ends
type EventQuestionEnd struct {
	QuestionID    uint      `json:"questionId"`
	QuizID        uint      `json:"quizId"`
	EndTime       time.Time `json:"endTime"`
	CorrectOption int       `json:"correctOption"`
}

// EventAnswerReceived is sent to confirm an answer was received
type EventAnswerReceived struct {
	QuestionID     uint      `json:"questionId"`
	UserID         uint      `json:"userId"`
	SelectedOption int       `json:"selectedOption"`
	AnswerTime     time.Time `json:"answerTime"`
}

// EventLeaderboard is sent to update participants about current standings
type EventLeaderboard struct {
	QuizID      uint               `json:"quizId"`
	Leaderboard []LeaderboardEntry `json:"leaderboard"`
}

// EventError is sent when an error occurs
type EventError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}
