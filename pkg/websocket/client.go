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

	// EventUserLeft is sent when a user leaves the quiz
	EventUserLeft EventType = "USER_LEFT"

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
	QuestionID      string   `json:"questionId"`
	SelectedOptions []string `json:"selectedOptions"`
	TimeTaken       float64  `json:"timeTaken"`
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

	// UserID is the identifier of the user or participant this client belongs to
	UserID uuid.UUID

	// IsCreator indicates if this client is connected as a quiz creator (user) or participant
	IsCreator bool

	// Hub manages the clients
	Hub HubInterface

	// WebSocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan []byte

	// Context for cancellation
	Ctx context.Context

	// Cancel function to clean up the context
	Cancel context.CancelFunc
}

// IncomingMessage represents a message received from the client
type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.GetUnregisterChan() <- c
		c.Conn.Close()
		if c.Cancel != nil {
			c.Cancel()
		}
		log.Printf("Client disconnected: %s for quiz %s", c.UserID, c.QuizID)
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		log.Printf("Received pong from client: %s", c.UserID)
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
			// Only participants can submit answers
			if c.IsCreator {
				log.Printf("Creator attempted to submit answer: %s", c.UserID)
				continue
			}

			// Process answer submission
			var answerPayload AnswerPayload
			if err := json.Unmarshal(incomingMsg.Payload, &answerPayload); err != nil {
				log.Printf("Error unmarshaling answer payload: %v", err)
				continue
			}

			// Validate the answer payload
			if len(answerPayload.SelectedOptions) == 0 {
				log.Printf("No options selected in answer: %s", c.UserID)
				continue
			}

			// Convert questionId string to UUID
			questionID, err := uuid.Parse(answerPayload.QuestionID)
			if err != nil {
				log.Printf("Invalid question ID: %v", err)
				continue
			}

			// Publish answer event to Redis
			event := Event{
				Type: "CLIENT_ANSWER",
				Payload: map[string]interface{}{
					"participantId":   c.UserID.String(),
					"questionId":      questionID.String(),
					"selectedOptions": answerPayload.SelectedOptions,
					"timeTaken":       answerPayload.TimeTaken,
				},
			}

			// Handle the event based on hub type
			c.Hub.BroadcastToQuiz(c.QuizID, event)

			// Also send direct confirmation to the client
			confirmEvent := Event{
				Type: EventAnswerReceived,
				Payload: map[string]interface{}{
					"questionId":      questionID.String(),
					"selectedOptions": answerPayload.SelectedOptions,
					"timeTaken":       answerPayload.TimeTaken,
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
		if c.Cancel != nil {
			c.Cancel()
		}
		log.Printf("WritePump terminated for client: %s", c.UserID)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// Check if the hub is still sending messages
			// log.Printf("Sending message to client: %s", c.UserID)
			// log.Printf("Message: %s", string(message))

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				log.Printf("Send channel closed for client: %s", c.UserID)
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error getting writer for client %s: %v", c.UserID, err)
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
				log.Printf("Error closing writer for client %s: %v", c.UserID, err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.UserID, err)
				return
			}
			log.Printf("Sent ping to client: %s", c.UserID)
		case <-c.Ctx.Done():
			log.Printf("Context done for client: %s", c.UserID)
			return
		}
	}
}
