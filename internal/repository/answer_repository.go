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
		INSERT INTO answers (id, participant_id, question_id, selected_option, selected_options_json, answered_at, time_taken, is_correct, score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		answer.ID,
		answer.ParticipantID,
		answer.QuestionID,
		"", // Deprecated field, keeping for backward compatibility
		answer.SelectedJSON,
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
		SELECT id, participant_id, question_id, selected_option, selected_options_json, answered_at, time_taken, is_correct, score
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
		var selectedJSON sql.NullString
		if err := rows.Scan(
			&answer.ID,
			&answer.ParticipantID,
			&answer.QuestionID,
			nil, // Skip the old selected_option field
			&selectedJSON,
			&answer.AnsweredAt,
			&answer.TimeTaken,
			&answer.IsCorrect,
			&answer.Score,
		); err != nil {
			return nil, err
		}

		// Set the selected JSON if it's not null
		if selectedJSON.Valid {
			answer.SelectedJSON = selectedJSON.String
			// Parse the JSON into the SelectedOptions array
			if _, err := answer.GetSelectedOptions(); err != nil {
				return nil, err
			}
		}

		answers = append(answers, &answer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return answers, nil
}

// GetAnswersByParticipantID retrieves all answers for a participant
func (r *PostgresAnswerRepository) GetAnswersByParticipantID(ctx context.Context, participantID uuid.UUID) ([]*model.Answer, error) {
	query := `
		SELECT id, participant_id, question_id, selected_option, selected_options_json, answered_at, time_taken, is_correct, score
		FROM answers
		WHERE participant_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, participantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []*model.Answer
	for rows.Next() {
		var answer model.Answer
		var selectedOption sql.NullString
		var selectedJSON sql.NullString
		if err := rows.Scan(
			&answer.ID,
			&answer.ParticipantID,
			&answer.QuestionID,
			&selectedOption,
			&selectedJSON,
			&answer.AnsweredAt,
			&answer.TimeTaken,
			&answer.IsCorrect,
			&answer.Score,
		); err != nil {
			return nil, err
		}

		// Set the selected JSON if it's not null
		if selectedJSON.Valid {
			answer.SelectedJSON = selectedJSON.String
			// Parse the JSON into the SelectedOptions array
			if _, err := answer.GetSelectedOptions(); err != nil {
				return nil, err
			}
		} else if selectedOption.Valid {
			// For backward compatibility, convert old single option to array
			if err := answer.SetSelectedOptions([]string{selectedOption.String}); err != nil {
				return nil, err
			}
		}

		answers = append(answers, &answer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return answers, nil
}

// GetAnswerByParticipantAndQuestion retrieves a participant's answer for a specific question
func (r *PostgresAnswerRepository) GetAnswerByParticipantAndQuestion(ctx context.Context, participantID uuid.UUID, questionID uuid.UUID) (*model.Answer, error) {
	query := `
		SELECT id, participant_id, question_id, selected_option, selected_options_json, answered_at, time_taken, is_correct, score
		FROM answers
		WHERE participant_id = $1 AND question_id = $2
	`

	var answer model.Answer
	var selectedOption sql.NullString
	var selectedJSON sql.NullString
	err := r.db.QueryRowContext(ctx, query, participantID, questionID).Scan(
		&answer.ID,
		&answer.ParticipantID,
		&answer.QuestionID,
		&selectedOption,
		&selectedJSON,
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

	// Set the selected JSON if it's not null
	if selectedJSON.Valid {
		answer.SelectedJSON = selectedJSON.String
		// Parse the JSON into the SelectedOptions array
		if _, err := answer.GetSelectedOptions(); err != nil {
			return nil, err
		}
	} else if selectedOption.Valid {
		// TODO remove this when all answers are migrated
		// For backward compatibility, convert old single option to array
		if err := answer.SetSelectedOptions([]string{selectedOption.String}); err != nil {
			return nil, err
		}
	}

	return &answer, nil
}
