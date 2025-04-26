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
        uuid creator_id FK
        enum status
        timestamp created_at
        timestamp updated_at
    }
    
    QUESTION {
        uuid id PK
        uuid quiz_id FK
        string content
        string option_a
        string option_b
        string option_c
        string option_d
        string correct_option
        int order
        int time_limit
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
        string selected_option
        int score
        timestamp submitted_at
    }
    
    QUIZ_SESSION {
        uuid quiz_id PK
        uuid current_question_id FK
        enum status
        timestamp started_at
        timestamp ended_at
        timestamp current_question_started_at
    }
    
    USER ||--o{ QUIZ : creates
    QUIZ ||--o{ QUESTION : contains
    QUIZ ||--o{ PARTICIPANT : joins
    PARTICIPANT ||--o{ ANSWER : submits
    QUESTION ||--o{ ANSWER : has
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

## Recent Changes and Current Implementation Status

- ✅ Refactored the user model to separate Users (creators) and Participants
- ✅ Updated WebSocket implementation to distinguish between creators and participants
- ✅ Improved real-time event broadcasting with specific targeting (to creators or participants)
- ✅ Implemented user registration and login endpoints
- ✅ Enhanced architecture documentation and system design explanations

### Current Implementation Status

- ✅ Basic API structure with handlers, services, and repositories
- ✅ User authentication endpoints (registration, login)
- ✅ WebSocket implementation for real-time communication
- ✅ Quiz creation and management
- ✅ Question management
- ✅ Participant joining
- ✅ Real-time leaderboard updates
- ✅ Basic quiz flow (waiting, active, completed states)
- ✅ Role-based WebSocket communication (different events for creators vs participants)

### TODO to Fulfill Requirements

1. **Authentication System**:
   - ✅ User registration and login endpoints
   - [ ] JWT authentication middleware
   - [ ] User authorization for quiz management
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
   - [ ] Hide question content from participants during answer period
   - [ ] Show correct answers to quiz creators in real-time

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

API documentation is available at `/swagger/index.html` when running the application.