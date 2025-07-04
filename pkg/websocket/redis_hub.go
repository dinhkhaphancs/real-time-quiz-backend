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
	instanceID  string // Unique identifier for this server instance
}

// NewRedisHub creates a new Redis-based WebSocket hub
func NewRedisHub(redisClient *redis.Client, ctx context.Context) *RedisHub {
	// Generate a unique instance ID for this server
	instanceID := uuid.New().String()

	return &RedisHub{
		Hub:         NewHub(),
		redisClient: redisClient,
		ctx:         ctx,
		instanceID:  instanceID,
	}
}

// GetInstanceID returns the unique identifier for this server instance
func (h *RedisHub) GetInstanceID() string {
	return h.instanceID
}

// SubscribeToQuiz subscribes to Redis events for a quiz
func (h *RedisHub) SubscribeToQuiz(quizID uuid.UUID) error {
	channel := fmt.Sprintf("quiz:%s", quizID.String())

	// Check if we already have a subscription
	if h.pubsub != nil {
		// Since Channels() method is not available, we'll just resubscribe
		// as redis client handles duplicate subscriptions gracefully
		if err := h.pubsub.Subscribe(h.ctx, channel); err != nil {
			return fmt.Errorf("error subscribing to channel: %w", err)
		}
		return nil
	}

	// Create a new subscription
	h.pubsub = h.redisClient.Subscribe(h.ctx, channel)

	// Start a goroutine to handle messages from Redis
	go func() {
		defer func() {
			if h.pubsub != nil {
				h.pubsub.Close()
			}
		}()

		for {
			select {
			case <-h.ctx.Done():
				return
			default:
				msg, err := h.pubsub.ReceiveMessage(h.ctx)
				if err != nil {
					fmt.Printf("Error receiving message: %v\n", err)
					time.Sleep(time.Second) // Add a small delay to prevent CPU spinning
					continue
				}

				// Skip empty messages
				if msg.Payload == "" {
					continue
				}

				// Skip messages with null bytes
				if len(msg.Payload) > 0 && msg.Payload[0] == 0 {
					fmt.Printf("Skipping message with null bytes\n")
					continue
				}

				var event Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					fmt.Printf("Error unmarshaling event: %v, payload: %q\n", err, msg.Payload)
					continue
				}

				// Forward the event to all WebSocket clients for this quiz
				h.BroadcastToQuiz(quizID, event)
			}
		}
	}()

	return nil
}

// PublishToQuiz publishes an event to Redis for a quiz
func (h *RedisHub) PublishToQuiz(quizID uuid.UUID, event Event) error {
	channel := fmt.Sprintf("quiz:%s", quizID.String())

	// Validate event fields to ensure we have a valid event
	if event.Type == "" {
		return fmt.Errorf("event type cannot be empty")
	}

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	// Sanity check - ensure we're not sending null bytes
	if len(message) > 0 && message[0] == 0 {
		return fmt.Errorf("invalid message format: starts with null byte")
	}

	// Log what we're publishing for debugging
	fmt.Printf("Publishing to channel %s: %s\n", channel, string(message))

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
			h.PublishToQuiz(quizID, NewEvent(EventTimerUpdate, map[string]interface{}{
				"remainingSeconds": 0,
				"totalSeconds":     durationSeconds,
				"endTime":          endTime.Format(time.RFC3339),
			}))
			return
		}

		remainingSeconds := int(endTime.Sub(now).Seconds())
		h.PublishToQuiz(quizID, NewEvent(EventTimerUpdate, map[string]interface{}{
			"remainingSeconds": remainingSeconds,
			"totalSeconds":     durationSeconds,
			"endTime":          endTime.Format(time.RFC3339),
		}))
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
