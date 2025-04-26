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

	participants, err := h.leaderboardService.GetLeaderboard(c, quizID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response
	var leaderboardEntries []map[string]interface{}
	for i, participant := range participants {
		leaderboardEntries = append(leaderboardEntries, map[string]interface{}{
			"rank":     i + 1,
			"id":       participant.ID,
			"name":     participant.Name,
			"score":    participant.Score,
			"joinedAt": participant.JoinedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": leaderboardEntries})
}