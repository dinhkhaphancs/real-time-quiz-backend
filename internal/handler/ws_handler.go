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
	hub         *my_ws.RedisHub
	quizService service.QuizService
	userService service.UserService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *my_ws.RedisHub, quizService service.QuizService, userService service.UserService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		quizService: quizService,
		userService: userService,
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

	// Get user ID from the URL
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user to validate and get role
	user, err := h.userService.GetUserByID(c, userID, quizID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or not authorized for this quiz"})
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
		ID:       clientID,
		UserID:   userID,
		QuizID:   quizID,
		UserRole: string(user.Role),
		Conn:     conn,
		Send:     make(chan []byte, 256),
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

	// Notify other clients about the new connection
	if user.Role != "ADMIN" {
		h.hub.PublishToQuiz(quizID, my_ws.Event{
			Type: my_ws.EventUserJoined,
			Payload: map[string]interface{}{
				"userId":   userID.String(),
				"userName": user.Name,
			},
		})
	}
}
