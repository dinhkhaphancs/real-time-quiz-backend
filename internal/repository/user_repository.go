package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresUserRepository implements UserRepository interface for PostgreSQL
type PostgresUserRepository struct {
	db *DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser creates a new user
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, name, quiz_id, role, joined_at, score)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.QuizID,
		user.Role,
		user.JoinedAt,
		user.Score,
	)
	return err
}

// GetUserByID retrieves a user by their ID
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, name, quiz_id, role, joined_at, score
		FROM users
		WHERE id = $1
	`
	
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.QuizID,
		&user.Role,
		&user.JoinedAt,
		&user.Score,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	
	return &user, nil
}

// GetUsersByQuizID retrieves all users for a quiz
func (r *PostgresUserRepository) GetUsersByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.User, error) {
	query := `
		SELECT id, name, quiz_id, role, joined_at, score
		FROM users
		WHERE quiz_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.QuizID,
			&user.Role,
			&user.JoinedAt,
			&user.Score,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}

// UpdateUserScore updates a user's score
func (r *PostgresUserRepository) UpdateUserScore(ctx context.Context, userID uuid.UUID, score int) error {
	query := `
		UPDATE users
		SET score = score + $1
		WHERE id = $2
	`
	
	result, err := r.db.ExecContext(ctx, query, score, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	
	return nil
}

// GetLeaderboard retrieves the top users by score for a quiz
func (r *PostgresUserRepository) GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.User, error) {
	query := `
		SELECT id, name, quiz_id, role, joined_at, score
		FROM users
		WHERE quiz_id = $1
		ORDER BY score DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, quizID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.QuizID,
			&user.Role,
			&user.JoinedAt,
			&user.Score,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}