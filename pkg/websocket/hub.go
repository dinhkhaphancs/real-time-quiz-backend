package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub manages WebSocket clients
type Hub struct {
	// Registered clients mapped by quiz ID
	Clients map[uuid.UUID]map[uuid.UUID]*Client

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for safe concurrent access
	mu sync.Mutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID]map[uuid.UUID]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the WebSocket hub
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case <-ctx.Done():
			return
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	quizClients, exists := h.Clients[client.QuizID]
	if !exists {
		quizClients = make(map[uuid.UUID]*Client)
		h.Clients[client.QuizID] = quizClients
	}

	quizClients[client.ID] = client
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if quizClients, exists := h.Clients[client.QuizID]; exists {
		if _, ok := quizClients[client.ID]; ok {
			delete(quizClients, client.ID)
			close(client.Send)

			// If no more clients in the quiz, remove the quiz entry
			if len(quizClients) == 0 {
				delete(h.Clients, client.QuizID)
			}
		}
	}
}

// BroadcastToQuiz sends an event to all clients in a quiz
func (h *Hub) BroadcastToQuiz(quizID uuid.UUID, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	quizClients, exists := h.Clients[quizID]
	if !exists {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}

	for _, client := range quizClients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(quizClients, client.ID)
		}
	}
}

// BroadcastToCreators sends an event only to creator clients in a quiz
func (h *Hub) BroadcastToCreators(quizID uuid.UUID, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	quizClients, exists := h.Clients[quizID]
	if !exists {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}

	for _, client := range quizClients {
		if !client.IsCreator {
			continue
		}

		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(quizClients, client.ID)
		}
	}
}

// BroadcastToParticipants sends an event only to participant clients in a quiz
func (h *Hub) BroadcastToParticipants(quizID uuid.UUID, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	quizClients, exists := h.Clients[quizID]
	if !exists {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}

	for _, client := range quizClients {
		if client.IsCreator {
			continue
		}

		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(quizClients, client.ID)
		}
	}
}

// SendToClient sends an event to a specific client
func (h *Hub) SendToClient(userID uuid.UUID, quizID uuid.UUID, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	quizClients, exists := h.Clients[quizID]
	if !exists {
		return
	}

	// Find the client with the matching UserID
	var targetClient *Client
	for _, client := range quizClients {
		if client.UserID == userID {
			targetClient = client
			break
		}
	}

	if targetClient == nil {
		return
	}

	message, err := json.Marshal(event)
	if err != nil {
		fmt.Printf("Error marshaling event: %v\n", err)
		return
	}

	select {
	case targetClient.Send <- message:
	default:
		close(targetClient.Send)
		delete(quizClients, targetClient.ID)
	}
}

// StartTimerBroadcast starts a timer that broadcasts updates to all clients in a quiz
func (h *Hub) StartTimerBroadcast(quizID uuid.UUID, durationSeconds int) {
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationSeconds) * time.Second)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()
		if now.After(endTime) {
			// Time's up
			h.BroadcastToQuiz(quizID, Event{
				Type: EventTimerUpdate,
				Payload: map[string]interface{}{
					"remainingSeconds": 0,
				},
			})
			return
		}

		remainingSeconds := int(endTime.Sub(now).Seconds())
		h.BroadcastToQuiz(quizID, Event{
			Type: EventTimerUpdate,
			Payload: map[string]interface{}{
				"remainingSeconds": remainingSeconds,
			},
		})
	}
}

// GetRegisterChan returns the channel for registering clients
func (h *Hub) GetRegisterChan() chan<- *Client {
	return h.Register
}

// GetUnregisterChan returns the channel for unregistering clients
func (h *Hub) GetUnregisterChan() chan<- *Client {
	return h.Unregister
}
