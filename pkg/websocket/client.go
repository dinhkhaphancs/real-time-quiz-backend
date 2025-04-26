package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// EventType defines the type of WebSocket events
type EventType string

const (
	// EventQuizStart is sent when a quiz starts
	EventQuizStart EventType = "QUIZ_START"

	// EventQuestionStart is sent when a new question becomes active
	EventQuestionStart EventType = "QUESTION_START"

	// EventQuestionEnd is sent when the time for a question ends
	EventQuestionEnd EventType = "QUESTION_END"

	// EventAnswerReceived is sent to confirm an answer was received
	EventAnswerReceived EventType = "ANSWER_RECEIVED"

	// EventLeaderboardUpdate is sent when the leaderboard changes
	EventLeaderboardUpdate EventType = "LEADERBOARD_UPDATE"

	// EventQuizEnd is sent when the quiz ends
	EventQuizEnd EventType = "QUIZ_END"

	// EventUserJoined is sent when a new user joins
	EventUserJoined EventType = "USER_JOINED"

	// EventTimerUpdate is sent to update the remaining time
	EventTimerUpdate EventType = "TIMER_UPDATE"

	// EventError is sent when an error occurs
	EventError EventType = "ERROR"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// HubInterface defines the common behavior expected from any hub implementation
type HubInterface interface {
	BroadcastToQuiz(quizID uuid.UUID, event Event)
	SendToClient(userID uuid.UUID, quizID uuid.UUID, event Event)
	Run(ctx context.Context)

	// Methods to access registration channels
	GetRegisterChan() chan<- *Client
	GetUnregisterChan() chan<- *Client
}

// ClientMessage represents a message sent from client to server
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// AnswerPayload represents a client's answer to a question
type AnswerPayload struct {
	QuestionID     string `json:"questionId"`
	SelectedOption string `json:"selectedOption"`
}

// Event represents a WebSocket event message
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a WebSocket client connection
type Client struct {
	// Client identifier
	ID uuid.UUID

	// QuizID is the identifier of the quiz this client belongs to
	QuizID uuid.UUID

	// UserID is the identifier of the user this client belongs to
	UserID uuid.UUID

	// UserRole is the role of the user (e.g., "ADMIN", "JOINER")
	UserRole string

	// Hub manages the clients
	Hub HubInterface

	// WebSocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// Context for cancellation
	Ctx context.Context
}

// IncomingMessage represents a message received from the client
type IncomingMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.GetUnregisterChan() <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// Process incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle the message based on its type
		switch incomingMsg.Type {
		case "ping":
			// Send pong response back
			event := Event{
				Type:    "pong",
				Payload: map[string]interface{}{"time": time.Now().Unix()},
			}
			eventData, _ := json.Marshal(event)
			c.Send <- eventData
		case "ANSWER":
			// Process answer submission
			var answerPayload AnswerPayload
			if err := json.Unmarshal(incomingMsg.Data, &answerPayload); err != nil {
				log.Printf("Error unmarshaling answer payload: %v", err)
				continue
			}

			// Convert questionId string to UUID
			questionID, err := uuid.Parse(answerPayload.QuestionID)
			if err != nil {
				log.Printf("Invalid question ID: %v", err)
				continue
			}

			// Publish answer event to Redis
			// This would typically be processed by a message handler service
			event := Event{
				Type: "CLIENT_ANSWER",
				Payload: map[string]interface{}{
					"userId":         c.UserID.String(),
					"questionId":     questionID.String(),
					"selectedOption": answerPayload.SelectedOption,
				},
			}

			// Handle the event based on hub type
			c.Hub.BroadcastToQuiz(c.QuizID, event)

			// Also send direct confirmation to the client
			confirmEvent := Event{
				Type: EventAnswerReceived,
				Payload: map[string]interface{}{
					"questionId":     questionID.String(),
					"selectedOption": answerPayload.SelectedOption,
				},
			}
			c.Hub.SendToClient(c.UserID, c.QuizID, confirmEvent)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.Ctx.Done():
			return
		}
	}
}
