# WebSocket Events Documentation

## Overview

This document outlines the WebSocket communication protocol used in the Real-Time Quiz application. The WebSocket connection allows for real-time bidirectional communication between the server and clients (both quiz creators and participants).

### Connection Endpoint

```
GET /ws/:quizId/:type/:id
```

Where:
- `:quizId` - UUID of the quiz to connect to
- `:type` - Either "creator" or "participant"
- `:id` - UUID of the user or participant

### Authentication

Connections require a valid JWT token provided in the Authorization header or as a query parameter.

## Event Structure

All WebSocket events follow this JSON structure:

```json
{
  "type": "EVENT_TYPE",
  "payload": {
    // Event-specific data
  }
}
```

## WebSocket Event Types

The following event types are used in the application for real-time communication:

### Core Event Types
These are the officially defined event types in the codebase:

- `QUIZ_START` - Sent when a quiz begins
- `QUESTION_START` - Sent when a new question becomes active
- `QUESTION_END` - Sent when a question ends
- `ANSWER_RECEIVED` - Confirmation that a participant's answer was received
- `LEADERBOARD_UPDATE` - Sent when the leaderboard changes
- `QUIZ_END` - Sent when a quiz ends
- `USER_JOINED` - Sent when a new participant joins
- `USER_LEFT` - Sent when a participant leaves
- `TIMER_UPDATE` - Sent periodically to update the timer countdown
- `ERROR` - Sent when an error occurs

### System Events
- `ping`/`pong` - Used for connection health checks

## Server-to-Client Events

These events are sent from the server to connected clients.

### QUIZ_START

Sent when a quiz begins.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| quizId | string (UUID) | Quiz identifier |
| title | string | Quiz title |
| description | string | Quiz description |
| startTime | string (ISO timestamp) | When the quiz started |

#### Example

```json
{
  "type": "QUIZ_START",
  "payload": {
    "quizId": "550e8400-e29b-41d4-a716-446655440000",
    "title": "General Knowledge Quiz",
    "description": "Test your knowledge on various topics",
    "startTime": "2025-04-28T14:30:00Z"
  }
}
```

### QUESTION_START

Sent when a new question becomes active.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| questionId | string (UUID) | Question identifier |
| questionText | string | The question text |
| options | array | List of answer options |
| timeLimit | integer | Time limit in seconds |
| allowMultipleAnswers | boolean | Whether multiple options can be selected |

#### Example

```json
{
  "type": "QUESTION_START",
  "payload": {
    "questionId": "550e8400-e29b-41d4-a716-446655440000",
    "questionText": "What is the capital of France?",
    "options": [
      { "id": "opt1", "text": "London" },
      { "id": "opt2", "text": "Paris" },
      { "id": "opt3", "text": "Berlin" },
      { "id": "opt4", "text": "Rome" }
    ],
    "timeLimit": 30,
    "allowMultipleAnswers": false
  }
}
```

### QUESTION_END

Sent when a question's time limit is reached or the creator manually ends the question.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| questionId | string (UUID) | Question identifier |
| correctOptions | array of strings | IDs of the correct answer options |
| statistics | object | Statistics about answers received |

#### Example

```json
{
  "type": "QUESTION_END",
  "payload": {
    "questionId": "550e8400-e29b-41d4-a716-446655440000",
    "correctOptions": ["opt2"],
    "statistics": {
      "totalAnswers": 42,
      "optionCounts": {
        "opt1": 5,
        "opt2": 30,
        "opt3": 4,
        "opt4": 3
      }
    }
  }
}
```

### ANSWER_RECEIVED

Sent to a participant to confirm their answer was received.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| questionId | string (UUID) | Question identifier |
| selectedOptions | array of strings | IDs of options selected by the participant |
| timeTaken | number | Time taken to answer in seconds |

#### Example

```json
{
  "type": "ANSWER_RECEIVED",
  "payload": {
    "questionId": "550e8400-e29b-41d4-a716-446655440000",
    "selectedOptions": ["opt2"],
    "timeTaken": 12.5
  }
}
```

### LEADERBOARD_UPDATE

Sent when the leaderboard changes (typically after each question ends).

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| entries | array | List of leaderboard entries |

#### Example

```json
{
  "type": "LEADERBOARD_UPDATE",
  "payload": {
    "entries": [
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440001",
        "nickname": "QuizWhiz",
        "score": 1200,
        "position": 1
      },
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440002",
        "nickname": "BrainBox",
        "score": 900,
        "position": 2
      }
    ]
  }
}
```

### QUIZ_END

Sent when the quiz ends.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| quizId | string (UUID) | Quiz identifier |
| endTime | string (ISO timestamp) | When the quiz ended |
| finalLeaderboard | array | Final leaderboard data |

#### Example

```json
{
  "type": "QUIZ_END",
  "payload": {
    "quizId": "550e8400-e29b-41d4-a716-446655440000",
    "endTime": "2025-04-28T15:00:00Z",
    "finalLeaderboard": [
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440001",
        "nickname": "QuizWhiz",
        "score": 2500,
        "position": 1
      }
    ]
  }
}
```

### USER_JOINED

Sent when a new participant joins the quiz.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| participantId | string (UUID) | Participant identifier |
| nickname | string | Participant's display name |
| joinTime | string (ISO timestamp) | When they joined |

#### Example

```json
{
  "type": "USER_JOINED",
  "payload": {
    "participantId": "550e8400-e29b-41d4-a716-446655440001",
    "nickname": "QuizWhiz",
    "joinTime": "2025-04-28T14:25:30Z"
  }
}
```

### USER_LEFT

Sent when a participant leaves the quiz.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| participantId | string (UUID) | Participant identifier |
| nickname | string | Participant's display name |
| leaveTime | string (ISO timestamp) | When they left |

#### Example

```json
{
  "type": "USER_LEFT",
  "payload": {
    "participantId": "550e8400-e29b-41d4-a716-446655440001",
    "nickname": "QuizWhiz",
    "leaveTime": "2025-04-28T14:45:30Z"
  }
}
```

### TIMER_UPDATE

Sent periodically to update clients about remaining time for the current question.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| remainingSeconds | integer | Seconds remaining |

#### Example

```json
{
  "type": "TIMER_UPDATE",
  "payload": {
    "remainingSeconds": 15
  }
}
```

### STATE_SYNC

Sent to clients when they connect or reconnect to provide the complete current state of the quiz.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| quizId | string (UUID) | Quiz identifier |
| currentPhase | string | Current quiz phase (BETWEEN_QUESTIONS, QUESTION_ACTIVE, SHOWING_RESULTS) |
| activeQuestion | object (optional) | Details of the currently active question (if any) |
| activeParticipants | array | List of currently connected participants |
| leaderboard | array | Current leaderboard data |
| timerInfo | object (optional) | Information about any active timers |

#### Example

```json
{
  "type": "STATE_SYNC",
  "payload": {
    "quizId": "550e8400-e29b-41d4-a716-446655440000",
    "currentPhase": "QUESTION_ACTIVE",
    "activeQuestion": {
      "questionId": "550e8400-e29b-41d4-a716-446655440000",
      "questionText": "What is the capital of France?",
      "options": [
        { "id": "opt1", "text": "London" },
        { "id": "opt2", "text": "Paris" },
        { "id": "opt3", "text": "Berlin" },
        { "id": "opt4", "text": "Rome" }
      ],
      "timeLimit": 30,
      "allowMultipleAnswers": false,
      "startedAt": "2025-04-28T14:40:00Z"
    },
    "activeParticipants": [
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440001",
        "nickname": "QuizWhiz",
        "isConnected": true
      },
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440002",
        "nickname": "BrainBox",
        "isConnected": true
      }
    ],
    "leaderboard": [
      {
        "participantId": "550e8400-e29b-41d4-a716-446655440001",
        "nickname": "QuizWhiz",
        "score": 1200,
        "position": 1
      }
    ],
    "timerInfo": {
      "remainingSeconds": 15,
      "totalSeconds": 30
    }
  }
}
```

### ERROR

Sent when an error occurs.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| code | string | Error code |
| message | string | Human-readable error message |

#### Example

```json
{
  "type": "ERROR",
  "payload": {
    "code": "INVALID_ANSWER",
    "message": "Answer submitted for an inactive question"
  }
}
```

## Client-to-Server Events

These events are sent from clients to the server.

### ANSWER

Sent by participants to submit an answer to the current question.

#### Payload

| Field | Type | Description |
|-------|------|-------------|
| questionId | string (UUID) | Question identifier |
| selectedOptions | array of strings | IDs of selected options |
| timeTaken | number | Time taken to answer in seconds |

#### Example

```json
{
  "type": "ANSWER",
  "payload": {
    "questionId": "550e8400-e29b-41d4-a716-446655440000",
    "selectedOptions": ["opt2"],
    "timeTaken": 12.5
  }
}
```

### ping

Sent by clients to keep the connection alive.

#### Example

```json
{
  "type": "ping"
}
```

## Connection Management

### Connection States

1. **Connected**: WebSocket connection established
2. **Quiz Active**: Connected during an active quiz
3. **Disconnected**: Connection lost

### Reconnection Strategy

If a client loses connection:

1. Attempt to reconnect immediately
2. If unsuccessful, retry with exponential backoff (1s, 2s, 4s, etc.)
3. On reconnection, server will send the current quiz state

### State Synchronization

When a client reconnects during an active quiz, they should receive:

1. Current quiz state (active question, remaining time)
2. Current leaderboard
3. Any event they missed during disconnection (if available)

## Security Considerations

### Token Authentication

- All WebSocket connections require a valid JWT token
- Tokens expire after a specified period
- Tokens include user role information (creator/participant)

### Rate Limiting

- Answer submissions are rate-limited to prevent spam
- Connection attempts are limited to prevent DoS attacks

### Data Validation

- All incoming messages are validated against schema
- Malformed messages result in ERROR events being sent back

## Error Handling

### Common Errors

| Error Code | Description |
|------------|-------------|
| INVALID_TOKEN | Authentication token is invalid or expired |
| SESSION_EXPIRED | Quiz session has expired |
| INVALID_ANSWER | Answer submission is invalid |
| RATE_LIMITED | Too many requests from client |
| SERVER_ERROR | Internal server error |

### Error Recovery

1. For authentication errors, client should refresh token and reconnect
2. For temporary errors, follow reconnection strategy
3. For permanent errors, show appropriate message to user