package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresQuestionRepository implements QuestionRepository interface for PostgreSQL
type PostgresQuestionRepository struct {
	db *DB
}

// NewPostgresQuestionRepository creates a new PostgreSQL question repository
func NewPostgresQuestionRepository(db *DB) *PostgresQuestionRepository {
	return &PostgresQuestionRepository{db: db}
}

// CreateQuestion creates a new question
func (r *PostgresQuestionRepository) CreateQuestion(ctx context.Context, question *model.Question) error {
	query := `
		INSERT INTO questions (id, quiz_id, text, time_limit, "order", question_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		question.ID,
		question.QuizID,
		question.Text,
		question.TimeLimit,
		question.Order,
		question.QuestionType,
		question.CreatedAt,
		question.UpdatedAt,
	)
	return err
}

// GetQuestionsByQuizID retrieves all questions for a quiz
func (r *PostgresQuestionRepository) GetQuestionsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Question, error) {
	query := `
		SELECT id, quiz_id, text, time_limit, "order", question_type, created_at, updated_at
		FROM questions
		WHERE quiz_id = $1
		ORDER BY "order" ASC
	`

	rows, err := r.db.QueryContext(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*model.Question
	for rows.Next() {
		var q model.Question
		if err := rows.Scan(
			&q.ID,
			&q.QuizID,
			&q.Text,
			&q.TimeLimit,
			&q.Order,
			&q.QuestionType,
			&q.CreatedAt,
			&q.UpdatedAt,
		); err != nil {
			return nil, err
		}
		questions = append(questions, &q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

// GetQuestionByID retrieves a question by its ID
func (r *PostgresQuestionRepository) GetQuestionByID(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	query := `
		SELECT id, quiz_id, text, time_limit, "order", question_type, created_at, updated_at
		FROM questions
		WHERE id = $1
	`

	var q model.Question
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&q.ID,
		&q.QuizID,
		&q.Text,
		&q.TimeLimit,
		&q.Order,
		&q.QuestionType,
		&q.CreatedAt,
		&q.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("question not found")
		}
		return nil, err
	}

	return &q, nil
}

// GetNextQuestion retrieves the next question after the current one
func (r *PostgresQuestionRepository) GetNextQuestion(ctx context.Context, quizID uuid.UUID, currentOrder int) (*model.Question, error) {
	query := `
		SELECT id, quiz_id, text, time_limit, "order", question_type, created_at, updated_at
		FROM questions
		WHERE quiz_id = $1 AND "order" > $2
		ORDER BY "order" ASC
		LIMIT 1
	`

	var q model.Question
	err := r.db.QueryRowContext(ctx, query, quizID, currentOrder).Scan(
		&q.ID,
		&q.QuizID,
		&q.Text,
		&q.TimeLimit,
		&q.Order,
		&q.QuestionType,
		&q.CreatedAt,
		&q.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no more questions")
		}
		return nil, err
	}

	return &q, nil
}

// UpdateQuestion updates an existing question
func (r *PostgresQuestionRepository) UpdateQuestion(ctx context.Context, question *model.Question) error {
	query := `
		UPDATE questions
		SET text = $1, time_limit = $2, "order" = $3, question_type = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		question.Text,
		question.TimeLimit,
		question.Order,
		question.QuestionType,
		time.Now(),
		question.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("question not found")
	}

	return nil
}

// DeleteQuestion deletes a question
func (r *PostgresQuestionRepository) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM questions
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
		return errors.New("question not found")
	}

	return nil
}
