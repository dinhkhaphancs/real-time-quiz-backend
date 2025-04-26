package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresAnswerRepository implements AnswerRepository interface for PostgreSQL
type PostgresAnswerRepository struct {
	db *DB
}

// NewPostgresAnswerRepository creates a new PostgreSQL answer repository
func NewPostgresAnswerRepository(db *DB) *PostgresAnswerRepository {
	return &PostgresAnswerRepository{db: db}
}

// CreateAnswer creates a new answer
func (r *PostgresAnswerRepository) CreateAnswer(ctx context.Context, answer *model.Answer) error {
	query := `
		INSERT INTO answers (id, user_id, question_id, selected_option, answered_at, time_taken, is_correct, score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		answer.ID,
		answer.UserID,
		answer.QuestionID,
		answer.SelectedOption,
		answer.AnsweredAt,
		answer.TimeTaken,
		answer.IsCorrect,
		answer.Score,
	)
	return err
}

// GetAnswersByQuestionID retrieves all answers for a question
func (r *PostgresAnswerRepository) GetAnswersByQuestionID(ctx context.Context, questionID uuid.UUID) ([]*model.Answer, error) {
	query := `
		SELECT id, user_id, question_id, selected_option, answered_at, time_taken, is_correct, score
		FROM answers
		WHERE question_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var answers []*model.Answer
	for rows.Next() {
		var answer model.Answer
		if err := rows.Scan(
			&answer.ID,
			&answer.UserID,
			&answer.QuestionID,
			&answer.SelectedOption,
			&answer.AnsweredAt,
			&answer.TimeTaken,
			&answer.IsCorrect,
			&answer.Score,
		); err != nil {
			return nil, err
		}
		answers = append(answers, &answer)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return answers, nil
}

// GetAnswersByUserID retrieves all answers for a user
func (r *PostgresAnswerRepository) GetAnswersByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Answer, error) {
	query := `
		SELECT id, user_id, question_id, selected_option, answered_at, time_taken, is_correct, score
		FROM answers
		WHERE user_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var answers []*model.Answer
	for rows.Next() {
		var answer model.Answer
		if err := rows.Scan(
			&answer.ID,
			&answer.UserID,
			&answer.QuestionID,
			&answer.SelectedOption,
			&answer.AnsweredAt,
			&answer.TimeTaken,
			&answer.IsCorrect,
			&answer.Score,
		); err != nil {
			return nil, err
		}
		answers = append(answers, &answer)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return answers, nil
}

// GetAnswerByUserAndQuestion retrieves a user's answer for a specific question
func (r *PostgresAnswerRepository) GetAnswerByUserAndQuestion(ctx context.Context, userID uuid.UUID, questionID uuid.UUID) (*model.Answer, error) {
	query := `
		SELECT id, user_id, question_id, selected_option, answered_at, time_taken, is_correct, score
		FROM answers
		WHERE user_id = $1 AND question_id = $2
	`
	
	var answer model.Answer
	err := r.db.QueryRowContext(ctx, query, userID, questionID).Scan(
		&answer.ID,
		&answer.UserID,
		&answer.QuestionID,
		&answer.SelectedOption,
		&answer.AnsweredAt,
		&answer.TimeTaken,
		&answer.IsCorrect,
		&answer.Score,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("answer not found")
		}
		return nil, err
	}
	
	return &answer, nil
}