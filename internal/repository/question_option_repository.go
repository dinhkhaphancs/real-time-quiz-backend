package repository

import (
	"context"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresQuestionOptionRepository implements QuestionOptionRepository for PostgreSQL
type PostgresQuestionOptionRepository struct {
	db *DB
}

// NewPostgresQuestionOptionRepository creates a new PostgreSQL question option repository
func NewPostgresQuestionOptionRepository(db *DB) *PostgresQuestionOptionRepository {
	return &PostgresQuestionOptionRepository{db: db}
}

// CreateQuestionOption creates a new question option
func (r *PostgresQuestionOptionRepository) CreateQuestionOption(ctx context.Context, option *model.QuestionOption) error {
	query := `
		INSERT INTO question_options (id, question_id, text, is_correct, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		option.ID,
		option.QuestionID,
		option.Text,
		option.IsCorrect,
		option.DisplayOrder,
		option.CreatedAt,
		option.UpdatedAt,
	)
	return err
}

// GetQuestionOptionsByQuestionID retrieves all options for a question
func (r *PostgresQuestionOptionRepository) GetQuestionOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*model.QuestionOption, error) {
	query := `
		SELECT id, question_id, text, is_correct, display_order, created_at, updated_at
		FROM question_options
		WHERE question_id = $1
		ORDER BY display_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []*model.QuestionOption
	for rows.Next() {
		var option model.QuestionOption
		if err := rows.Scan(
			&option.ID,
			&option.QuestionID,
			&option.Text,
			&option.IsCorrect,
			&option.DisplayOrder,
			&option.CreatedAt,
			&option.UpdatedAt,
		); err != nil {
			return nil, err
		}
		options = append(options, &option)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, nil
}

// UpdateQuestionOption updates an existing question option
func (r *PostgresQuestionOptionRepository) UpdateQuestionOption(ctx context.Context, option *model.QuestionOption) error {
	query := `
		UPDATE question_options
		SET text = $1, is_correct = $2, display_order = $3, updated_at = $4
		WHERE id = $5
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		option.Text,
		option.IsCorrect,
		option.DisplayOrder,
		time.Now(),
		option.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("question option not found")
	}

	return nil
}

// DeleteQuestionOption deletes a question option
func (r *PostgresQuestionOptionRepository) DeleteQuestionOption(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM question_options
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
		return errors.New("question option not found")
	}

	return nil
}

// DeleteQuestionOptionsByQuestionID deletes all options for a question
func (r *PostgresQuestionOptionRepository) DeleteQuestionOptionsByQuestionID(ctx context.Context, questionID uuid.UUID) error {
	query := `
		DELETE FROM question_options
		WHERE question_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, questionID)
	if err != nil {
		return err
	}

	return nil
}
