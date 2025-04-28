package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/service"
	"github.com/dinhkhaphancs/real-time-quiz-backend/pkg/response"
	ws "github.com/dinhkhaphancs/real-time-quiz-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub                *ws.RedisHub
	quizService        service.QuizService
	userService        service.UserService
	participantService service.ParticipantService
	stateService       service.StateService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(
	hub *ws.RedisHub,
	quizService service.QuizService,
	userService service.UserService,
	participantService service.ParticipantService,
	stateService service.StateService,
) *WebSocketHandler {
	return &WebSocketHandler{
		hub:                hub,
		quizService:        quizService,
		userService:        userService,
		participantService: participantService,
		stateService:       stateService,
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
		log.Printf("Error parsing quiz ID: %v\n", err)
		response.WithError(c, http.StatusBadRequest, "Invalid quiz ID", "The provided quiz ID is not valid")
		return
	}

	// Get connection type and ID from the URL
	connectionType := c.Param("type") // "user" or "participant"
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Error parsing ID: %v\n", err)
		response.WithError(c, http.StatusBadRequest, "Invalid ID", "The provided user or participant ID is not valid")
		return
	}

	// Variable to track if this is a creator connection
	isCreator := false

	// Validate the connection based on type
	if connectionType == "user" {
		// Get user to validate
		user, err := h.userService.GetUserByID(c, id)
		if err != nil {
			log.Printf("Error getting user: %v\n", err)
			response.WithError(c, http.StatusUnauthorized, "Authentication failed", "User not found")
			return
		}

		// Check if user is the creator of this quiz
		quiz, err := h.quizService.GetQuiz(c, quizID)
		if err != nil {
			log.Printf("Error getting quiz: %v\n", err)
			response.WithError(c, http.StatusNotFound, "Quiz not found", "The specified quiz could not be found")
			return
		}

		if quiz.CreatorID != user.ID {
			response.WithError(c, http.StatusUnauthorized, "Authorization failed", "User is not the creator of this quiz")
			return
		}

		isCreator = true

	} else if connectionType == "participant" {
		// Get participant to validate
		participant, err := h.participantService.GetParticipantByID(c, id)
		if err != nil {
			log.Printf("Error getting participant: %v\n", err)
			response.WithError(c, http.StatusUnauthorized, "Authentication failed", "Participant not found")
			return
		}

		if participant.QuizID != quizID {
			log.Printf("Participant %s is not authorized for quiz %s\n", participant.ID, quizID)
			response.WithError(c, http.StatusUnauthorized, "Authorization failed", "Participant not authorized for this quiz")
			return
		}

	} else {
		log.Printf("Invalid connection type: %s\n", connectionType)
		response.WithError(c, http.StatusBadRequest, "Invalid connection type", "Connection type must be 'user' or 'participant'")
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v\n", err)
		response.WithError(c, http.StatusInternalServerError, "Connection error", "Failed to upgrade connection to WebSocket")
		return
	}

	// Create a client ID
	clientID := uuid.New()

	// Create a detached background context for the WebSocket connection
	wsCtx, cancel := context.WithCancel(context.Background())

	// Create a new client
	client := &ws.Client{
		ID:        clientID,
		UserID:    id,
		QuizID:    quizID,
		IsCreator: isCreator,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       h.hub,
		Ctx:       wsCtx,
		Cancel:    cancel,
	}

	// Record the connection in our state system if this is a participant
	if !isCreator {
		instanceID := h.hub.GetInstanceID()
		err = h.stateService.UpdateParticipantConnection(c, id, quizID, true, instanceID)
		if err != nil {
			log.Printf("Error recording participant connection: %v", err)
			// Continue despite error - this is not critical
		}

		// Set up a deferred cleanup to mark this participant as disconnected when the connection ends
		go func(participantID, quizID uuid.UUID, instanceID string) {
			// Wait for context cancellation (which happens when the connection closes)
			<-wsCtx.Done()

			// Mark participant as disconnected
			ctx := context.Background()
			err := h.stateService.UpdateParticipantConnection(ctx, participantID, quizID, false, instanceID)
			if err != nil {
				log.Printf("Error updating participant disconnection: %v", err)
			}
		}(id, quizID, instanceID)
	} else {
		// For creator connections, we don't need to track connections in the same way,
		// but we might want to register the instance
		instanceID := h.hub.GetInstanceID()
		err = h.stateService.RegisterInstance(c, instanceID)
		if err != nil {
			log.Printf("Error registering instance: %v", err)
			// Continue despite error
		}
	}

	// Subscribe to Redis events for this quiz if not already subscribed
	if err := h.hub.SubscribeToQuiz(quizID); err != nil {
		response.WithError(c, http.StatusInternalServerError, "Subscription error", "Failed to subscribe to quiz events")
		conn.Close()
		cancel()
		return
	}

	// Register client with the hub
	h.hub.GetRegisterChan() <- client

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()

	// Get current quiz state and send it to the client for initial synchronization
	if state, err := h.stateService.GetQuizState(c, quizID); err == nil {
		// Send full state to the client
		stateEvent := ws.Event{
			Type:    ws.EventStateSync,
			Payload: state,
		}

		// Marshal to JSON
		if jsonData, err := json.Marshal(stateEvent); err == nil {
			client.Send <- jsonData
		}
	}
}
