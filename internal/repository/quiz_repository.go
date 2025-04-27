package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresQuizRepository implements QuizRepository interface for PostgreSQL
type PostgresQuizRepository struct {
	db *DB
}

// NewPostgresQuizRepository creates a new PostgreSQL quiz repository
func NewPostgresQuizRepository(db *DB) *PostgresQuizRepository {
	return &PostgresQuizRepository{db: db}
}

// CreateQuiz creates a new quiz
func (r *PostgresQuizRepository) CreateQuiz(ctx context.Context, quiz *model.Quiz) error {
	query := `
		INSERT INTO quizzes (id, title, description, creator_id, status, code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		quiz.ID,
		quiz.Title,
		quiz.Description,
		quiz.CreatorID,
		quiz.Status,
		quiz.Code,
		quiz.CreatedAt,
		quiz.UpdatedAt,
	)
	return err
}

// GetQuizByID retrieves a quiz by its ID
func (r *PostgresQuizRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*model.Quiz, error) {
	query := `
		SELECT id, title, description, creator_id, status, code, created_at, updated_at
		FROM quizzes
		WHERE id = $1
	`

	var quiz model.Quiz
	var description sql.NullString // Use sql.NullString to handle NULL values
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&quiz.ID,
		&quiz.Title,
		&description,
		&quiz.CreatorID,
		&quiz.Status,
		&quiz.Code,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("quiz not found")
		}
		return nil, err
	}

	// Convert NullString to string
	if description.Valid {
		quiz.Description = description.String
	} else {
		quiz.Description = ""
	}

	return &quiz, nil
}

// GetQuizByCode retrieves a quiz by its code
func (r *PostgresQuizRepository) GetQuizByCode(ctx context.Context, code string) (*model.Quiz, error) {
	query := `
		SELECT id, title, description, creator_id, status, code, created_at, updated_at
		FROM quizzes
		WHERE code = $1
	`

	var quiz model.Quiz
	var description sql.NullString // Use sql.NullString to handle NULL values
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&quiz.ID,
		&quiz.Title,
		&description,
		&quiz.CreatorID,
		&quiz.Status,
		&quiz.Code,
		&quiz.CreatedAt,
		&quiz.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("quiz not found")
		}
		return nil, err
	}

	// Convert NullString to string
	if description.Valid {
		quiz.Description = description.String
	} else {
		quiz.Description = ""
	}

	return &quiz, nil
}

// GetQuizzesByCreatorID retrieves all quizzes created by a user
func (r *PostgresQuizRepository) GetQuizzesByCreatorID(ctx context.Context, creatorID uuid.UUID) ([]*model.Quiz, error) {
	query := `
		SELECT id, title, description, creator_id, status, code, created_at, updated_at
		FROM quizzes
		WHERE creator_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quizzes []*model.Quiz
	for rows.Next() {
		var quiz model.Quiz
		var description sql.NullString // Use sql.NullString to handle NULL values
		if err := rows.Scan(
			&quiz.ID,
			&quiz.Title,
			&description,
			&quiz.CreatorID,
			&quiz.Status,
			&quiz.Code,
			&quiz.CreatedAt,
			&quiz.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Convert NullString to string
		if description.Valid {
			quiz.Description = description.String
		} else {
			quiz.Description = ""
		}

		quizzes = append(quizzes, &quiz)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return quizzes, nil
}

// UpdateQuizStatus updates the status of a quiz
func (r *PostgresQuizRepository) UpdateQuizStatus(ctx context.Context, id uuid.UUID, status model.QuizStatus) error {
	query := `
		UPDATE quizzes
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quiz not found")
	}

	return nil
}

// UpdateQuiz updates a quiz's title and description
func (r *PostgresQuizRepository) UpdateQuiz(ctx context.Context, quiz *model.Quiz) error {
	query := `
		UPDATE quizzes
		SET title = $1, description = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		quiz.Title,
		quiz.Description,
		time.Now(),
		quiz.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quiz not found")
	}

	return nil
}

// DeleteQuiz deletes a quiz and all its related data
func (r *PostgresQuizRepository) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	// Due to the ON DELETE CASCADE constraints set up in the database,
	// deleting a quiz will automatically delete all related questions, participants, answers,
	// and the quiz session.
	query := `
		DELETE FROM quizzes
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quiz not found")
	}

	return nil
}

// CreateQuizSession creates a new quiz session
func (r *PostgresQuizRepository) CreateQuizSession(ctx context.Context, session *model.QuizSession) error {
	query := `
		INSERT INTO quiz_sessions (quiz_id, status)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.QuizID,
		session.Status,
	)
	return err
}

// GetQuizSession retrieves a quiz session
func (r *PostgresQuizRepository) GetQuizSession(ctx context.Context, quizID uuid.UUID) (*model.QuizSession, error) {
	query := `
		SELECT quiz_id, current_question_id, status, started_at, ended_at, current_question_started_at
		FROM quiz_sessions
		WHERE quiz_id = $1
	`

	var session model.QuizSession
	err := r.db.QueryRowContext(ctx, query, quizID).Scan(
		&session.QuizID,
		&session.CurrentQuestionID,
		&session.Status,
		&session.StartedAt,
		&session.EndedAt,
		&session.CurrentQuestionStartedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("quiz session not found")
		}
		return nil, err
	}

	return &session, nil
}

// UpdateQuizSession updates a quiz session
func (r *PostgresQuizRepository) UpdateQuizSession(ctx context.Context, session *model.QuizSession) error {
	query := `
		UPDATE quiz_sessions
		SET current_question_id = $1,
			status = $2,
			started_at = $3,
			ended_at = $4,
			current_question_started_at = $5
		WHERE quiz_id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		session.CurrentQuestionID,
		session.Status,
		session.StartedAt,
		session.EndedAt,
		session.CurrentQuestionStartedAt,
		session.QuizID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("quiz session not found")
	}

	return nil
}
