# Real-Time Quiz Application (Kahoot-like)

A real-time interactive quiz application that allows registered users to create and manage quizzes, while others can join and participate without registration.

## Requirements

### User Roles
- **Registered Users**: Can create and manage multiple quizzes
- **Joiners**: Can participate in quizzes without registration by simply entering their name

### Quiz Flow
1. A registered user creates a quiz with multiple questions
2. They share a unique link with potential participants
3. Participants join by entering their name
4. The quiz creator starts the quiz
5. Questions are presented one at a time with configurable time limits
6. Participants answer within the time limit
7. After all questions, final results are shown

### Real-Time Features
- Live leaderboard updates
- Synchronized question timing across all participants
- Real-time answer submission and scoring

## System Architecture

### High-Level Architecture

```mermaid
graph TD
    Client[Frontend Clients] --> API[REST API Layer]
    Client <--> WS[WebSocket Connection]
    API --> Services[Service Layer]
    WS --> Hub[WebSocket Hub]
    Hub <--> Redis[Redis Pub/Sub]
    Services --> Repositories[Repository Layer]
    Repositories --> DB[(PostgreSQL)]
    Services <--> Redis
    
    subgraph "Core Components"
        API
        Services
        Repositories
    end
    
    subgraph "Real-time Communication"
        WS
        Hub
        Redis
    end
```

Our architecture follows a layered approach with clear separation of concerns:

1. **Client Layer**: Frontend applications (web, mobile) that interact with our backend through both REST API and WebSocket connections.

2. **API Layer**: RESTful API endpoints built with Gin framework handling HTTP requests for:
   - User authentication and management
   - Quiz creation and configuration
   - Question management
   - Answer submission and validation
   - Leaderboard retrieval

3. **WebSocket Layer**: Real-time communication using Gorilla WebSockets for:
   - Live quiz updates to all participants
   - Question timing synchronization
   - Immediate answer feedback
   - Real-time leaderboard updates

4. **Service Layer**: Business logic encapsulation with:
   - User service: Authentication, registration, profile management
   - Quiz service: Quiz lifecycle management (create, start, end)
   - Question service: Question management and sequencing
   - Participant service: Handling participant connections and state
   - Answer service: Processing and scoring participant answers
   - Leaderboard service: Calculating and maintaining real-time standings

5. **Repository Layer**: Data access abstraction with:
   - PostgreSQL repositories for persistent data storage
   - Clean separation between business logic and data access

6. **Data Storage**:
   - PostgreSQL: Primary relational database for all persistent data
   - Redis: For WebSocket pub/sub, caching, and session management

7. **Communication Patterns**:
   - Request/Response: For REST API calls
   - Pub/Sub: For real-time event broadcasting via Redis
   - WebSockets: For bidirectional client-server communication

This architecture provides:
- **Scalability**: Services can be scaled independently
- **Maintainability**: Clear separation of concerns
- **Reliability**: Robust error handling and state management
- **Performance**: Efficient real-time communication for interactive quiz experience

### Database Schema

```mermaid
erDiagram
    USER {
        uuid id PK
        string email
        string password_hash
        string name
        timestamp created_at
    }
    
    QUIZ {
        uuid id PK
        string title
        string description
        uuid creator_id FK
        string status
        string code UK
        timestamp created_at
        timestamp updated_at
    }
    
    QUESTION {
        uuid id PK
        uuid quiz_id FK
        string content
        string question_type
        int order
        int time_limit
        timestamp created_at
    }
    
    QUESTION_OPTION {
        uuid id PK
        uuid question_id FK
        string text
        boolean is_correct
        int display_order
        timestamp created_at
    }
    
    PARTICIPANT {
        uuid id PK
        string name
        uuid quiz_id FK
        int score
        timestamp joined_at
    }
    
    ANSWER {
        uuid id PK
        uuid question_id FK
        uuid participant_id FK
        uuid option_id FK
        int score
        timestamp submitted_at
    }
    
    MULTIPLE_ANSWER {
        uuid id PK
        uuid answer_id FK
        uuid option_id FK
    }
    
    QUIZ_SESSION {
        uuid id PK
        uuid quiz_id FK
        uuid current_question_id FK
        string status
        timestamp started_at
        timestamp ended_at
        timestamp current_question_started_at
    }
    
    USER ||--o{ QUIZ : creates
    QUIZ ||--o{ QUESTION : contains
    QUESTION ||--o{ QUESTION_OPTION : has
    QUIZ ||--o{ PARTICIPANT : joins
    PARTICIPANT ||--o{ ANSWER : submits
    QUESTION ||--o{ ANSWER : has
    ANSWER ||--o{ MULTIPLE_ANSWER : contains
    QUESTION_OPTION ||--o{ MULTIPLE_ANSWER : references
    QUESTION_OPTION ||--o{ ANSWER : references
    QUIZ ||--|| QUIZ_SESSION : tracks
```

### API Flow

```mermaid
sequenceDiagram
    participant RU as Registered User
    participant J as Joiner
    participant API as Backend API
    participant WS as WebSocket
    participant DB as Database
    
    RU->>API: Create Quiz
    API->>DB: Store Quiz
    API->>RU: Return Quiz ID
    
    RU->>API: Add Questions
    API->>DB: Store Questions
    
    J->>API: Join Quiz (quiz ID + name)
    API->>DB: Create Participant
    API->>J: Return Participant ID
    
    J->>WS: Connect (quiz ID + participant ID)
    RU->>WS: Connect (quiz ID + user ID)
    
    RU->>API: Start Quiz
    API->>DB: Update Quiz Status
    API->>WS: Broadcast Quiz Started
    
    loop For each question
        API->>WS: Send Question to All
        WS->>J: Display Question (options only)
        WS->>RU: Display Question (with correct answer)
        
        J->>API: Submit Answer
        API->>DB: Store Answer & Update Score
        
        API->>WS: Broadcast Leaderboard Update
        WS->>J: Show Leaderboard (top 10)
        WS->>RU: Show Leaderboard (all)
        
        RU->>API: Next Question
    end
    
    API->>WS: Send Final Results
    WS->>J: Show Final Results (top 3)
    WS->>RU: Show Final Results (all)
```

## Tech Stack

- **Backend**: Go (Golang)
  - Gin framework for HTTP API
  - Gorilla WebSockets for real-time communication
  - Redis for pub/sub and caching
  - PostgreSQL for data persistence
  
- **Data Storage**:
  - PostgreSQL: Primary database for storing quiz data, questions, users, and answers
  - Redis: For real-time communication, leaderboard caching, and session management

## JWT Authentication

The application now implements JWT (JSON Web Token) authentication to secure protected endpoints and verify user identity.

### Authentication Flow

1. **Registration & Login**:
   - Users register with email, name, and password
   - Upon login, the system generates access and refresh tokens
   - The access token is short-lived (24h in development, 12h in production)
   - The refresh token has a longer lifespan (7 days)

2. **Protected Endpoints**:
   - Creating quizzes
   - Managing quizzes (start/end)
   - Adding and managing questions
   - Certain WebSocket connections

3. **Token Usage**:
   - Include the JWT token in the `Authorization` header
   - Format: `Bearer <your_jwt_token>`
   - The token contains user identity information

4. **Ownership Verification**:
   - The system verifies quiz ownership for operations like:
     - Starting/ending quizzes
     - Adding/managing questions
   - Only quiz creators can perform these actions on their quizzes

### Implementation Details

1. **Configuration**: 
   - JWT settings in `config.yaml` and `config.production.yaml`
   - Environment variables for secret keys (in production)
   - Configurable expiration times for both token types

2. **Components**:
   - `pkg/auth/jwt.go`: Core JWT generation and validation
   - `internal/middleware/jwt_middleware.go`: Authentication middleware for routes
   - `internal/service/user_service.go`: Token generation during login

3. **Protected Routes**:
   - Quiz creation & management: `/api/v1/quizzes`
   - Question management: `/api/v1/questions`

### Security Considerations

1. **Token Storage**:
   - Store tokens securely on the client side
   - Refresh tokens require additional security measures
   - Use HTTPS in production environments

2. **Token Validation**:
   - Signatures are validated on every request
   - Expired tokens are rejected
   - Tokens contain only necessary user information

3. **Authorization**:
   - Role-based authorization using token claims
   - Resource ownership verification
   - Proper error messages without leaking sensitive information

The JWT implementation improves security by eliminating the need to pass user IDs in request bodies, preventing impersonation attacks. It also enables stateless authentication that scales well in distributed environments.

## Quiz Code Feature

The application now allows participants to join quizzes using a 6-character alphanumeric code instead of a UUID. This provides a more user-friendly experience for quiz participants.

### How It Works

1. **Code Generation**:
   - When a quiz is created, a unique 6-character alphanumeric code is automatically generated
   - The code uses a carefully selected character set (ABCDEFGHJKLMNPQRSTUVWXYZ23456789) that avoids similar-looking characters

2. **Joining Process**:
   - Participants can join using the new endpoint: `POST /api/v1/quizzes/join`
   - The request body requires only the quiz code and participant name:
     ```json
     {
       "code": "ABC123",
       "name": "Participant Name"
     }
     ```
   - The existing UUID-based joining endpoint remains functional for backward compatibility

3. **Benefits**:
   - Easier to share verbally or write down (e.g., on a whiteboard)
   - More user-friendly than UUIDs for classroom/event settings
   - Improved security by not exposing database IDs directly

4. **Implementation Details**:
   - Added a `code` column to the `quizzes` table with a UNIQUE constraint
   - Enhanced repository layer with methods to find quizzes by code
   - Added new service methods and API endpoints to support code-based joining
   - The Postman collection has been updated with examples of the new endpoint

### API Usage Example

```http
POST /api/v1/quizzes/join
Content-Type: application/json

{
  "code": "ABC123", 
  "name": "John Doe"
}
```

Response:
```json
{
  "success": true,
  "message": "Successfully joined quiz",
  "data": {
    "participant": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "John Doe",
      "quizId": "550e8400-e29b-41d4-a716-446655440001",
      "score": 0,
      "joinedAt": "2025-04-27T09:15:23Z"
    }
  },
  "timestamp": "2025-04-27T09:15:23Z"
}
```

The existing participant API endpoints (`/api/v1/participants/*`) remain unchanged as they operate using UUIDs for internal consistency.

## Dynamic Options and Multiple Choice Questions

The application now supports both dynamic question options and multiple choice questions, providing more flexibility in quiz creation and answering.

### Dynamic Options

Instead of being limited to fixed options (A, B, C, D), quiz creators can now:

1. **Add any number of custom options** to questions
2. **Customize option text** with arbitrary content
3. **Set display order** for options to control presentation
4. **Mark any option as correct** (for single choice questions)
5. **Mark multiple options as correct** (for multiple choice questions)

### Question Types

The system supports two types of questions:

1. **Single Choice Questions** (`SINGLE_CHOICE`):
   - Participants select exactly one answer
   - Only one option can be marked as correct
   - Answers are scored as either correct (full points) or incorrect (no points)

2. **Multiple Choice Questions** (`MULTIPLE_CHOICE`):
   - Participants can select multiple answers
   - Multiple options can be marked as correct
   - For a correct score, participants must select ALL correct options and NO incorrect options

### Implementation Details

1. **Database Changes**:
   - Added `question_type` column to the `questions` table
   - Created a new `question_options` table with columns for:
     - `id`: Unique identifier
     - `question_id`: Reference to the parent question
     - `text`: The option text
     - `is_correct`: Boolean indicating if this is a correct option
     - `display_order`: Integer for controlling option order
   - The original fixed option columns are marked as deprecated

2. **API Changes**:
   - When creating questions, you can now specify:
     - Question type: `SINGLE_CHOICE` or `MULTIPLE_CHOICE`
     - An array of options with text and correct status
   - When submitting answers, participants can provide an array of selected option IDs

3. **Answer Validation**:
   - For single choice: Exactly one option must be selected, and it must be correct
   - For multiple choice: All correct options must be selected, and no incorrect options

### Example API Usage

**Creating a question with dynamic options**:
```json
{
  "text": "Which of the following are planets in our solar system?",
  "questionType": "MULTIPLE_CHOICE",
  "timeLimit": 30,
  "options": [
    {"text": "Earth", "isCorrect": true},
    {"text": "Saturn", "isCorrect": true},
    {"text": "Sun", "isCorrect": false},
    {"text": "Moon", "isCorrect": false},
    {"text": "Neptune", "isCorrect": true}
  ]
}
```

**Submitting an answer to a multiple choice question**:
```json
{
  "questionId": "550e8400-e29b-41d4-a716-446655440000",
  "selectedOptionIds": [
    "550e8400-e29b-41d4-a716-446655440001",
    "550e8400-e29b-41d4-a716-446655440002",
    "550e8400-e29b-41d4-a716-446655440005"
  ]
}
```

These enhancements make the quiz application more versatile, allowing for more complex and educational question types beyond simple single-choice questions.

## Database Migrations

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) to manage database schema changes. We've implemented a structured approach to ensure your database stays in sync with the codebase.

### Migration Files Structure

All migrations are stored in the `migrations/versioned` directory following the naming convention:
```
{version}_{description}.up.sql   # For applying a migration
{version}_{description}.down.sql # For rolling back a migration
```

### Setting Up the Migration Tool

1. Install golang-migrate:
   ```bash
   # macOS (using Homebrew)
   brew install golang-migrate

   # Or using Go
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

2. Run the setup script to convert existing migration files:
   ```bash
   ./scripts/setup_migrations.sh
   ```
   This script will organize your migrations into the versioned format required by golang-migrate.

### Managing Migrations

The Makefile provides several targets to manage database migrations:

1. **Create a new migration**:
   ```bash
   make migrate-create name=add_new_feature
   ```
   This creates two new files in the `migrations/versioned` directory:
   - `{timestamp}_add_new_feature.up.sql` (apply changes)
   - `{timestamp}_add_new_feature.down.sql` (roll back changes)

2. **Apply all pending migrations**:
   ```bash
   DB_URL="postgres://username:password@localhost:5432/quiz_db?sslmode=disable" make migrate-up
   ```

3. **Roll back the last migration**:
   ```bash
   DB_URL="postgres://username:password@localhost:5432/quiz_db?sslmode=disable" make migrate-down-one
   ```

4. **Roll back all migrations**:
   ```bash
   DB_URL="postgres://username:password@localhost:5432/quiz_db?sslmode=disable" make migrate-down
   ```

5. **Check current migration version**:
   ```bash
   DB_URL="postgres://username:password@localhost:5432/quiz_db?sslmode=disable" make migrate-version
   ```

6. **Go to a specific migration version**:
   ```bash
   DB_URL="postgres://username:password@localhost:5432/quiz_db?sslmode=disable" make migrate-goto version=2
   ```

### Best Practices for Database Migrations

1. **Always create "down" migrations**: Ensure each migration can be reversed.
2. **Keep migrations small and focused**: One logical change per migration.
3. **Test migrations before applying to production**: Always verify migrations in a staging environment.
4. **Include migrations in code reviews**: Treat schema changes with the same rigor as code changes.
5. **Use transactions when possible**: Wrap schema changes in transactions for atomicity.
6. **Document complex migrations**: Add comments to explain non-obvious changes.

### Automating Migrations in Deployment

For automatic migrations during deployment, you can add the following to your Dockerfile or deployment script:

```bash
# Example for a Docker entrypoint
migrate -path /app/migrations/versioned -database "${DB_URL}" up
```

Or for docker-compose:
```yaml
command: sh -c "migrate -path /app/migrations/versioned -database \"${DB_URL}\" up && your-app-command"
```

## Recent Changes and Current Implementation Status

- ✅ Refactored the user model to separate Users (creators) and Participants
- ✅ Updated WebSocket implementation to distinguish between creators and participants
- ✅ Improved real-time event broadcasting with specific targeting (to creators or participants)
- ✅ Implemented user registration and login endpoints
- ✅ Enhanced architecture documentation and system design explanations
- ✅ Standardized DTO naming convention - removed redundant "Dto" suffix from DTOs in the dto package
- ✅ Implemented unified API response pattern with standardized error and success handling
- ✅ Created consistent response formatting across all API endpoints
- ✅ Added JWT authentication with token-based authorization
- ✅ Implemented protected routes for quiz and question management
- ✅ Added ownership verification for quiz operations
- ✅ Added quiz joining by code feature (participants can now join using a 6-character code instead of UUID)
- ✅ Added quiz description support for more detailed quiz information
- ✅ Enhanced quiz creation to support creating quizzes with questions in a single API call
- ✅ Implemented support for dynamic options and multiple choice questions
- ✅ Added comprehensive state management system with clear quiz phases and state transitions
- ✅ Standardized WebSocket message format for improved client-side processing

### Current Implementation Status

- ✅ Basic API structure with handlers, services, and repositories
- ✅ User authentication endpoints (registration, login)
- ✅ JWT token-based authentication system
- ✅ WebSocket implementation for real-time communication
- ✅ Quiz creation and management
- ✅ Question management
- ✅ Participant joining
- ✅ Quiz joining by code for improved usability
- ✅ Dedicated participant management API endpoints
- ✅ Real-time leaderboard updates
- ✅ Advanced quiz flow with distinct phases (waiting, between-questions, question-active, showing-results, completed)
- ✅ Role-based WebSocket communication (different events for creators vs participants)
- ✅ Real-time notifications when participants join or leave
- ✅ Bootstrap pattern implementation for improved code organization
- ✅ Modular API structure organized by domain with clear public/private route separation
- ✅ Dynamic options support with ability to add custom options to questions
- ✅ Multiple choice questions that allow selecting multiple correct answers
- ✅ Standardized WebSocket message payloads with timestamps and metadata

### TODO to Fulfill Requirements

1. **Authentication System**:
   - ✅ User registration and login endpoints
   - ✅ JWT authentication middleware
   - ✅ User authorization for quiz management
   - [ ] Secure WebSocket connections

2. **Model Refinements**:
   - ✅ Separate User (creator) and Participant models
   - ✅ Updated WebSocket client model for creator/participant differentiation
   - ✅ Refactored services to work with the new model structure
   - [ ] Add user profile management

3. **Enhanced Quiz Flow**:
   - [ ] Improve timing synchronization for questions
   - [ ] Add question pre-loading for smoother transitions
   - [ ] Implement scoring based on answer speed

4. **UI Differentiation**:
   - ✅ Backend support for different views between creators and participants
   - ✅ Hide question content from participants during answer period
   - ✅ Show correct answers to quiz creators in real-time

5. **Additional Features**:
   - [ ] Quiz templates and reusability
   - [ ] Analytics for quiz creators
   - [ ] Export quiz results

## Getting Started

### Prerequisites
- Go 1.18+
- PostgreSQL
- Redis

### Setup
1. Clone the repository
2. Set up environment variables in `.env` file
3. Run database migrations:
   ```
   make migrate
   ```
4. Start the application:
   ```
   make run
   ```

### Docker
You can also run the application using Docker:
```
docker-compose up
```

## API Documentation
See at `real-time-quiz-postman-collection.json`