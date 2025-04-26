package model

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in a quiz
type UserRole string

const (
	// UserRoleAdmin represents an admin user who creates and manages quizzes
	UserRoleAdmin UserRole = "ADMIN"
	// UserRoleJoiner represents a user who joins and participates in quizzes
	UserRoleJoiner UserRole = "JOINER"
)

// User represents a user in the system
type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	QuizID   uuid.UUID `json:"quizId" db:"quiz_id"`
	Role     UserRole  `json:"role" db:"role"`
	JoinedAt time.Time `json:"joinedAt" db:"joined_at"`
	Score    int       `json:"score" db:"score"`
}

// NewUser creates a new user with the given details
func NewUser(name string, quizID uuid.UUID, role UserRole) *User {
	return &User{
		ID:       uuid.New(),
		Name:     name,
		QuizID:   quizID,
		Role:     role,
		JoinedAt: time.Now(),
		Score:    0,
	}
}

// NewAdmin creates a new admin user
func NewAdmin(name string, quizID uuid.UUID) *User {
	return NewUser(name, quizID, UserRoleAdmin)
}

// NewJoiner creates a new joiner user
func NewJoiner(name string, quizID uuid.UUID) *User {
	return NewUser(name, quizID, UserRoleJoiner)
}
