package dto

import (
	"time"

	"github.com/google/uuid"
)

// LeaderboardRequest represents the request parameters for retrieving a leaderboard
type LeaderboardRequest struct {
	QuizID string `uri:"quizId" binding:"required"`
	Limit  int    `form:"limit,default=10"`
}

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank     int       `json:"rank"`
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Score    int       `json:"score"`
	JoinedAt time.Time `json:"joinedAt"`
}

// LeaderboardResponse represents the response payload for a leaderboard request
type LeaderboardResponse struct {
	QuizID       uuid.UUID          `json:"quizId"`
	QuizTitle    string             `json:"quizTitle"`
	TotalPlayers int                `json:"totalPlayers"`
	Leaderboard  []LeaderboardEntry `json:"leaderboard"`
}
