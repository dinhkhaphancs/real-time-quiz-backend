# State Management Implementation Documentation

## Overview

This document outlines the state management system implementation for the Real-Time Quiz application. The system provides reliable tracking of quiz states, participant connections, and enables proper state synchronization across distributed server instances.

## Problem Statement

The original implementation had several limitations:

1. **Participant Connection Tracking**: Unable to accurately determine which participants were actively connected
2. **Quiz Phase Ambiguity**: No clear way to determine if a quiz was displaying a question or showing results
3. **State Synchronization**: New clients needed a way to get the current state when connecting mid-quiz
4. **Multi-Instance Support**: Multiple server instances needed a way to coordinate participant connections

## Solution Architecture

The implemented solution uses PostgreSQL as the source of truth for state data, with Redis used solely for real-time communication between server instances.

### Database Schema Enhancements

1. **Enhanced Quiz Sessions**
   ```sql
   ALTER TABLE quiz_sessions
   ADD COLUMN current_phase VARCHAR(30) NOT NULL DEFAULT 'BETWEEN_QUESTIONS',
   ADD COLUMN current_question_ended_at TIMESTAMP NULL,
   ADD COLUMN next_question_id UUID NULL REFERENCES questions(id);
   ```

2. **New Tables**
   ```sql
   -- Event tracking
   CREATE TABLE IF NOT EXISTS quiz_events (
       id SERIAL PRIMARY KEY,
       quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
       event_type VARCHAR(50) NOT NULL,
       payload JSONB NOT NULL,
       sequence_number BIGINT NOT NULL,
       created_at TIMESTAMP NOT NULL,
       UNIQUE (quiz_id, sequence_number)
   );
   
   -- Connection tracking
   CREATE TABLE IF NOT EXISTS participant_connections (
       participant_id UUID REFERENCES participants(id) ON DELETE CASCADE,
       quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
       is_connected BOOLEAN NOT NULL DEFAULT false,
       last_seen TIMESTAMP NOT NULL,
       instance_id VARCHAR(50),
       PRIMARY KEY (participant_id, quiz_id)
   );
   
   -- Instance tracking
   CREATE TABLE IF NOT EXISTS server_instances (
       instance_id VARCHAR(50) PRIMARY KEY,
       last_heartbeat TIMESTAMP NOT NULL
   );
   ```

### Quiz Phases

The system now tracks the specific phase of an active quiz:

```go
type QuizPhase string

const (
    // Quiz is active but between questions
    QuizPhaseBetweenQuestions QuizPhase = "BETWEEN_QUESTIONS"
    
    // Question is currently active and accepting answers
    QuizPhaseQuestionActive QuizPhase = "QUESTION_ACTIVE"
    
    // Question has ended and showing results
    QuizPhaseShowingResults QuizPhase = "SHOWING_RESULTS"
)
```

### State Management Components

1. **Model Layer**
   - Enhanced `QuizSession` with phase tracking
   - Added models for events, connections, and instances

2. **Repository Layer**
   - `StateRepository` interface for state operations
   - PostgreSQL implementation of state repository

3. **Service Layer**
   - `StateService` for centralized state management
   - Methods for tracking connections and events

4. **API Layer**
   - Endpoints for retrieving quiz state and active participants
   - WebSocket enhancements for state synchronization

5. **WebSocket Enhancements**
   - Instance identification for distributed deployment
   - Connection tracking
   - State synchronization on connection

## Key Functionalities

### Participant Connection Tracking

When a participant connects or disconnects:

1. The connection state is stored in the `participant_connections` table
2. The system records which server instance the participant is connected to
3. The connection status is visible through the API and included in state updates

### Quiz State Retrieval

The complete quiz state can be retrieved via the API:

```
GET /api/v1/states/quiz/:quizId
```

Response includes:
- Current quiz phase
- Active question details (if any)
- Timer information
- Connected participants
- Leaderboard data

### Active Participants

Active participants can be retrieved via:

```
GET /api/v1/states/quiz/:quizId/participants/active
```

Only participants with an active connection within the last 30 seconds are included.

### State Synchronization

When a client connects:
1. The WebSocket handler fetches the current quiz state
2. The state is sent to the client as a `STATE_SYNC` event
3. The client can immediately display the correct UI based on the current phase

### Multi-Instance Support

For distributed deployments:
1. Each server instance has a unique ID
2. The system tracks which instance each participant is connected to
3. Events can be routed to the correct instance when needed

## State Transitions

The state management system defines clear transitions between different quiz phases. Here's how the state transitions work at each stage of the quiz:

### Starting a Quiz

When a quiz is started:
- Set `Status` to `QuizStatusActive`
- Set `StartedAt` to current time
- Set `CurrentPhase` to `QuizPhaseBetweenQuestions`

### Starting a Question

When a question is activated:
- Set `CurrentQuestionID` to the question's ID
- Set `CurrentQuestionStartedAt` to current time
- Set `CurrentPhase` to `QuizPhaseQuestionActive`
- Clear `CurrentQuestionEndedAt`

### Ending a Question

When a question time expires or is manually ended:
- Set `CurrentQuestionEndedAt` to current time
- Set `CurrentPhase` to `QuizPhaseShowingResults`
- Calculate and update participant scores

### Moving to Next Question

When preparing for the next question:
- Set `NextQuestionID` based on the quiz flow
- When ready, transition to the next question using the Starting a Question flow

### Ending a Quiz

When all questions are completed or the quiz is manually ended:
- Set `Status` to `QuizStatusCompleted`
- Set `EndedAt` to current time
- Clear `CurrentPhase` or set to a final state

These transitions provide clear tracking of quiz status, current question, and what phase of the quiz you're in at any moment. This approach allows clients to display the appropriate UI based on the current state and phase of the quiz.

## Repository Layer

The `StateRepository` handles all state-related database operations:

```go
type StateRepository interface {
    // Quiz Events
    StoreEvent(ctx context.Context, event *model.QuizEvent) error
    GetMissedEvents(ctx context.Context, quizID uuid.UUID, lastSequence int64, limit int) ([]*model.QuizEvent, error)
    
    // Participant Connections
    UpdateParticipantConnection(ctx context.Context, conn *model.ParticipantConnection) error
    GetActiveParticipantConnections(ctx context.Context, quizID uuid.UUID, cutoffTime time.Time) ([]*model.ParticipantConnection, error)
    
    // Instance Management
    RegisterInstance(ctx context.Context, instance *model.ServerInstance) error
    UpdateInstanceHeartbeat(ctx context.Context, instanceID string) error
    GetActiveInstances(ctx context.Context, cutoffTime time.Time) ([]*model.ServerInstance, error)
    
    // Sequence Number Management
    IncrementSequenceNumber(ctx context.Context, quizID uuid.UUID) (int64, error)
}
```

## Service Layer

The `StateService` provides higher-level state management operations:

```go
type StateService interface {
    // State Management
    GetQuizState(ctx context.Context, quizID uuid.UUID) (*dto.QuizStateDTO, error)
    
    // Events
    PublishEvent(ctx context.Context, quizID uuid.UUID, eventType string, payload interface{}) error
    GetMissedEvents(ctx context.Context, quizID uuid.UUID, lastSequence int64) ([]*model.QuizEvent, error)
    
    // Participant Connection
    UpdateParticipantConnection(ctx context.Context, participantID, quizID uuid.UUID, isConnected bool, instanceID string) error
    GetActiveParticipants(ctx context.Context, quizID uuid.UUID) ([]model.Participant, error)
    
    // Instance Management
    RegisterInstance(ctx context.Context, instanceID string) error
    UpdateInstanceHeartbeat(ctx context.Context, instanceID string) error
}
```

## WebSocket Connection Flow

When a participant connects via WebSocket:

1. The handler authenticates the participant
2. The connection is upgraded to WebSocket
3. The participant's connection status is updated in the database
4. The client is registered with the hub
5. The current quiz state is sent to the client
6. When the connection closes, the participant's status is updated to disconnected

## State Management Data Flow

1. **State Update**: A change occurs (quiz starts, question activates, etc.)
2. **Database Update**: The state is updated in the database
3. **Event Publication**: An event is published to Redis
4. **Event Broadcast**: All instances broadcast the event to their connected clients
5. **State Synchronization**: New clients receive the complete state on connection

## Integration with Existing System

The state management system integrates with the existing codebase:

1. **Quiz Service**: Needs to update quiz phases during state transitions
2. **WebSocket Handler**: Now tracks connections and synchronizes state
3. **API Layer**: Provides endpoints for state retrieval

## Usage Examples

### Retrieving Quiz State

```javascript
fetch('/api/v1/states/quiz/550e8400-e29b-41d4-a716-446655440000')
  .then(response => response.json())
  .then(state => {
    // Use state to update UI
    if (state.currentPhase === 'QUESTION_ACTIVE') {
      showQuestionUI(state.activeQuestion);
    } else if (state.currentPhase === 'SHOWING_RESULTS') {
      showResultsUI(state.activeQuestion);
    }
  });
```

### WebSocket State Synchronization

```javascript
socket.onmessage = function(event) {
  const data = JSON.parse(event.data);
  
  if (data.type === 'STATE_SYNC') {
    // Update entire UI based on received state
    syncUIWithState(data.payload);
  } else if (data.type === 'QUESTION_START') {
    // Update just the question portion of the UI
    showQuestion(data.payload);
  }
};
```

## Conclusion

This state management implementation provides a robust foundation for the Real-Time Quiz application. It solves the key issues with participant tracking, quiz phase management, and state synchronization, while supporting distributed deployment for scalability.

The system is designed to be maintainable and extensible, with clear separation of concerns and well-defined interfaces at each layer of the architecture.