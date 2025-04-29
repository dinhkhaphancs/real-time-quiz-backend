package dto

import (
	"encoding/json"
	"time"
)

// StandardMessageDTO provides a consistent structure for all websocket messages
type StandardMessageDTO struct {
	Type      string      `json:"type"`      // Message event type (e.g., QUIZ_START, USER_JOINED)
	Payload   interface{} `json:"payload"`   // The actual message data
	Timestamp time.Time   `json:"timestamp"` // When the message was created
}

// NewStandardMessage creates a standard message with the current timestamp
func NewStandardMessage(eventType string, payload interface{}) StandardMessageDTO {
	return StandardMessageDTO{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// ToJSON converts the standardized message to JSON
func (m StandardMessageDTO) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
