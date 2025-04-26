package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// PostgresParticipantRepository implements ParticipantRepository interface for PostgreSQL
type PostgresParticipantRepository struct {
	db *DB
}

// NewPostgresParticipantRepository creates a new PostgreSQL participant repository
func NewPostgresParticipantRepository(db *DB) *PostgresParticipantRepository {
	return &PostgresParticipantRepository{db: db}
}

// CreateParticipant creates a new participant
func (r *PostgresParticipantRepository) CreateParticipant(ctx context.Context, participant *model.Participant) error {
	query := `
		INSERT INTO participants (id, name, quiz_id, score, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		participant.ID,
		participant.Name,
		participant.QuizID,
		participant.Score,
		participant.JoinedAt,
	)
	return err
}

// GetParticipantByID retrieves a participant by their ID
func (r *PostgresParticipantRepository) GetParticipantByID(ctx context.Context, id uuid.UUID) (*model.Participant, error) {
	query := `
		SELECT id, name, quiz_id, score, joined_at
		FROM participants
		WHERE id = $1
	`
	
	var participant model.Participant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&participant.ID,
		&participant.Name,
		&participant.QuizID,
		&participant.Score,
		&participant.JoinedAt,
	)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("participant not found")
		}
		return nil, err
	}
	
	return &participant, nil
}

// GetParticipantsByQuizID retrieves all participants for a quiz
func (r *PostgresParticipantRepository) GetParticipantsByQuizID(ctx context.Context, quizID uuid.UUID) ([]*model.Participant, error) {
	query := `
		SELECT id, name, quiz_id, score, joined_at
		FROM participants
		WHERE quiz_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var participants []*model.Participant
	for rows.Next() {
		var participant model.Participant
		if err := rows.Scan(
			&participant.ID,
			&participant.Name,
			&participant.QuizID,
			&participant.Score,
			&participant.JoinedAt,
		); err != nil {
			return nil, err
		}
		participants = append(participants, &participant)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return participants, nil
}

// UpdateParticipantScore updates a participant's score
func (r *PostgresParticipantRepository) UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, score int) error {
	query := `
		UPDATE participants
		SET score = score + $1
		WHERE id = $2
	`
	
	result, err := r.db.ExecContext(ctx, query, score, participantID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("participant not found")
	}
	
	return nil
}

// GetLeaderboard retrieves the top participants by score for a quiz
func (r *PostgresParticipantRepository) GetLeaderboard(ctx context.Context, quizID uuid.UUID, limit int) ([]*model.Participant, error) {
	query := `
		SELECT id, name, quiz_id, score, joined_at
		FROM participants
		WHERE quiz_id = $1
		ORDER BY score DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, quizID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var participants []*model.Participant
	for rows.Next() {
		var participant model.Participant
		if err := rows.Scan(
			&participant.ID,
			&participant.Name,
			&participant.QuizID,
			&participant.Score,
			&participant.JoinedAt,
		); err != nil {
			return nil, err
		}
		participants = append(participants, &participant)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return participants, nil
}