package handler

import (
	"net/http"
	"strconv"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LeaderboardHandler handles leaderboard-related HTTP requests
type LeaderboardHandler struct {
	leaderboardService service.LeaderboardService
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(leaderboardService service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

// GetLeaderboard retrieves the leaderboard for a quiz
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	// Get limit parameter if provided, default to 10
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	users, err := h.leaderboardService.GetLeaderboard(c, quizID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	var leaderboardEntries []map[string]interface{}
	for i, user := range users {
		leaderboardEntries = append(leaderboardEntries, map[string]interface{}{
			"rank":     i + 1,
			"id":       user.ID,
			"name":     user.Name,
			"score":    user.Score,
			"joinedAt": user.JoinedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboardEntries})
}