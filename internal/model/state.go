package model

import (
	"time"

	"github.com/google/uuid"
)

// QuizEvent represents an event in a quiz for state synchronization
type QuizEvent struct {
	ID             int64     `json:"id" db:"id"`
	QuizID         uuid.UUID `json:"quizId" db:"quiz_id"`
	EventType      string    `json:"eventType" db:"event_type"`
	Payload        []byte    `json:"payload" db:"payload"`
	SequenceNumber int64     `json:"sequenceNumber" db:"sequence_number"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

// ParticipantConnection tracks the connection status of participants
type ParticipantConnection struct {
	ParticipantID uuid.UUID `json:"participantId" db:"participant_id"`
	QuizID        uuid.UUID `json:"quizId" db:"quiz_id"`
	IsConnected   bool      `json:"isConnected" db:"is_connected"`
	LastSeen      time.Time `json:"lastSeen" db:"last_seen"`
	InstanceID    string    `json:"instanceId" db:"instance_id"`
}

// ServerInstance represents a server instance in a distributed deployment
type ServerInstance struct {
	InstanceID    string    `json:"instanceId" db:"instance_id"`
	LastHeartbeat time.Time `json:"lastHeartbeat" db:"last_heartbeat"`
}

// NewQuizEvent creates a new quiz event
func NewQuizEvent(quizID uuid.UUID, eventType string, payload []byte, sequenceNumber int64) *QuizEvent {
	return &QuizEvent{
		QuizID:         quizID,
		EventType:      eventType,
		Payload:        payload,
		SequenceNumber: sequenceNumber,
		CreatedAt:      time.Now(),
	}
}

// NewParticipantConnection creates a new participant connection
func NewParticipantConnection(participantID, quizID uuid.UUID, instanceID string) *ParticipantConnection {
	return &ParticipantConnection{
		ParticipantID: participantID,
		QuizID:        quizID,
		IsConnected:   true,
		LastSeen:      time.Now(),
		InstanceID:    instanceID,
	}
}

// NewServerInstance creates a new server instance
func NewServerInstance(instanceID string) *ServerInstance {
	return &ServerInstance{
		InstanceID:    instanceID,
		LastHeartbeat: time.Now(),
	}
}
