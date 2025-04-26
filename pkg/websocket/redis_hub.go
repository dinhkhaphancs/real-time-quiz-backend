package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisHub is a WebSocket hub implementation that uses Redis for pub/sub
type RedisHub struct {
	*Hub
	redisClient *redis.Client
	pubsub      *redis.PubSub
	ctx         context.Context
}

// NewRedisHub creates a new Redis-based WebSocket hub
func NewRedisHub(redisClient *redis.Client, ctx context.Context) *RedisHub {
	return &RedisHub{
		Hub:         NewHub(),
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// SubscribeToQuiz subscribes to Redis events for a quiz
func (h *RedisHub) SubscribeToQuiz(quizID uuid.UUID) error {
	channel := fmt.Sprintf("quiz:%s", quizID.String())
	h.pubsub = h.redisClient.Subscribe(h.ctx, channel)

	// Start a goroutine to handle messages from Redis
	go func() {
		defer h.pubsub.Close()

		for {
			msg, err := h.pubsub.ReceiveMessage(h.ctx)
			if err != nil {
				fmt.Printf("Error receiving message: %v\n", err)
				return
			}

			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				fmt.Printf("Error unmarshaling event: %v\n", err)
				continue
			}

			// Forward the event to all WebSocket clients for this quiz
			h.BroadcastToQuiz(quizID, event)
		}
	}()

	return nil
}

// PublishToQuiz publishes an event to Redis for a quiz
func (h *RedisHub) PublishToQuiz(quizID uuid.UUID, event Event) error {
	channel := fmt.Sprintf("quiz:%s", quizID.String())

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	return h.redisClient.Publish(h.ctx, channel, message).Err()
}

// PublishToCreators publishes an event to Redis and broadcasts only to creator clients
func (h *RedisHub) PublishToCreators(quizID uuid.UUID, event Event) error {
	// First publish to Redis to record the event
	if err := h.PublishToQuiz(quizID, event); err != nil {
		return err
	}

	// Then filter and only broadcast to creator clients locally
	h.BroadcastToCreators(quizID, event)
	return nil
}

// PublishToParticipants publishes an event to Redis and broadcasts only to participant clients
func (h *RedisHub) PublishToParticipants(quizID uuid.UUID, event Event) error {
	// First publish to Redis to record the event
	if err := h.PublishToQuiz(quizID, event); err != nil {
		return err
	}

	// Then filter and only broadcast to participant clients locally
	h.BroadcastToParticipants(quizID, event)
	return nil
}

// StartTimerBroadcast starts a timer that broadcasts updates to all clients in a quiz
func (h *RedisHub) StartTimerBroadcast(quizID uuid.UUID, durationSeconds int) {
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationSeconds) * time.Second)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()
		if now.After(endTime) {
			// Time's up
			h.PublishToQuiz(quizID, Event{
				Type: EventTimerUpdate,
				Payload: map[string]interface{}{
					"remainingSeconds": 0,
				},
			})
			return
		}

		remainingSeconds := int(endTime.Sub(now).Seconds())
		h.PublishToQuiz(quizID, Event{
			Type: EventTimerUpdate,
			Payload: map[string]interface{}{
				"remainingSeconds": remainingSeconds,
			},
		})
	}
}

// GetRegisterChan returns the channel for registering clients
func (h *RedisHub) GetRegisterChan() chan<- *Client {
	return h.Register
}

// GetUnregisterChan returns the channel for unregistering clients
func (h *RedisHub) GetUnregisterChan() chan<- *Client {
	return h.Unregister
}
