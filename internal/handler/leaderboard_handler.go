package handler

import (
	"net/http"
	"strconv"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/dto"
	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LeaderboardHandler handles leaderboard-related HTTP requests
type LeaderboardHandler struct {
	leaderboardService service.LeaderboardService
	quizService        service.QuizService
}

// NewLeaderboardHandler creates a new leaderboard handler
func NewLeaderboardHandler(
	leaderboardService service.LeaderboardService,
	quizService service.QuizService,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
		quizService:        quizService,
	}
}

// GetLeaderboard retrieves the leaderboard for a quiz
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	var request dto.LeaderboardRequest

	request.QuizID = c.Param("quizId")
	quizID, err := uuid.Parse(request.QuizID)
	if err != nil {
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	// Get limit parameter if provided, default to 10
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	request.Limit = limit

	participants, err := h.leaderboardService.GetLeaderboard(c, quizID, limit)
	if err != nil {
		response.WithError(c, http.StatusInternalServerError, "Failed to get leaderboard", err.Error())
		return
	}

	// Get quiz information for the response
	quiz, err := h.quizService.GetQuiz(c, quizID)
	if err != nil {
		response.WithError(c, http.StatusNotFound, "Quiz not found", "The specified quiz could not be found")
		return
	}

	// Format response using the DTO
	var entries []dto.LeaderboardEntry
	for i, participant := range participants {
		entries = append(entries, dto.LeaderboardEntry{
			Rank:     i + 1,
			ID:       participant.ID,
			Name:     participant.Name,
			Score:    participant.Score,
			JoinedAt: participant.JoinedAt,
		})
	}

	leaderboardResponse := dto.LeaderboardResponse{
		QuizID:       quizID,
		QuizTitle:    quiz.Title,
		TotalPlayers: len(participants),
		Leaderboard:  entries,
	}

	response.WithSuccess(c, http.StatusOK, "Leaderboard fetched successfully", leaderboardResponse)
}
