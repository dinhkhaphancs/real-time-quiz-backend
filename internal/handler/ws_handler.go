package handler

import (
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	my_ws "github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub                *my_ws.RedisHub
	quizService        service.QuizService
	userService        service.UserService
	participantService service.ParticipantService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *my_ws.RedisHub,
	quizService service.QuizService,
	userService service.UserService,
	participantService service.ParticipantService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:                hub,
		quizService:        quizService,
		userService:        userService,
		participantService: participantService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// HandleConnection upgrades an HTTP connection to WebSocket
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// Get quiz ID from the URL
	quizIDStr := c.Param("quizId")
	quizID, err := uuid.Parse(quizIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quiz ID"})
		return
	}

	// Get connection type and ID from the URL
	connectionType := c.Param("type") // "user" or "participant"
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Variable to track if this is a creator connection
	isCreator := false
	userName := ""

	// Validate the connection based on type
	if connectionType == "user" {
		// Get user to validate
		user, err := h.userService.GetUserByID(c, id)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Check if user is the creator of this quiz
		quiz, err := h.quizService.GetQuiz(c, quizID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
			return
		}

		if quiz.CreatorID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User is not the creator of this quiz"})
			return
		}

		isCreator = true
		userName = user.Name

	} else if connectionType == "participant" {
		// Get participant to validate
		participant, err := h.participantService.GetParticipantByID(c, id)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Participant not found"})
			return
		}

		if participant.QuizID != quizID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Participant not authorized for this quiz"})
			return
		}

		userName = participant.Name
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid connection type"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create a client ID
	clientID := uuid.New()

	// Create a new client
	client := &my_ws.Client{
		ID:        clientID,
		UserID:    id,
		QuizID:    quizID,
		IsCreator: isCreator,
		Conn:      conn,
		Send:      make(chan []byte, 256),
	}

	// Subscribe to Redis events for this quiz if not already subscribed
	if err := h.hub.SubscribeToQuiz(quizID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe to quiz events"})
		conn.Close()
		return
	}

	// Register client with the hub
	// Use the register channel to let the hub handle the registration
	h.hub.GetRegisterChan() <- client

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()

	// Notify other clients about the new connection (only for participants)
	if !isCreator {
		h.hub.PublishToQuiz(quizID, my_ws.Event{
			Type: my_ws.EventUserJoined,
			Payload: map[string]interface{}{
				"id":   id.String(),
				"name": userName,
				"type": connectionType,
			},
		})
	}
}
