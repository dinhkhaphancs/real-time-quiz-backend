package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/model"
	"github.com/google/uuid"
)

// stateRepositoryImpl implements the StateRepository interface
type stateRepositoryImpl struct {
	db *sql.DB
}

// NewStateRepository creates a new state repository
func NewStateRepository(db *sql.DB) StateRepository {
	return &stateRepositoryImpl{db: db}
}

// StoreEvent stores a quiz event in the database
func (r *stateRepositoryImpl) StoreEvent(ctx context.Context, event *model.QuizEvent) error {
	query := `
		INSERT INTO quiz_events (quiz_id, event_type, payload, sequence_number, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		event.QuizID,
		event.EventType,
		event.Payload,
		event.SequenceNumber,
		event.CreatedAt,
	).Scan(&id)

	if err != nil {
		return err
	}

	event.ID = id
	return nil
}

// GetMissedEvents retrieves events a client missed since their last connection
func (r *stateRepositoryImpl) GetMissedEvents(
	ctx context.Context,
	quizID uuid.UUID,
	lastSequence int64,
	limit int,
) ([]*model.QuizEvent, error) {
	query := `
		SELECT id, quiz_id, event_type, payload, sequence_number, created_at
		FROM quiz_events
		WHERE quiz_id = $1 AND sequence_number > $2
		ORDER BY sequence_number ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, quizID, lastSequence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.QuizEvent
	for rows.Next() {
		event := &model.QuizEvent{}
		err := rows.Scan(
			&event.ID,
			&event.QuizID,
			&event.EventType,
			&event.Payload,
			&event.SequenceNumber,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// UpdateParticipantConnection updates or creates a participant connection
func (r *stateRepositoryImpl) UpdateParticipantConnection(ctx context.Context, conn *model.ParticipantConnection) error {
	query := `
		INSERT INTO participant_connections (participant_id, quiz_id, is_connected, last_seen, instance_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (participant_id, quiz_id) DO UPDATE
		SET is_connected = $3, last_seen = $4, instance_id = $5
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		conn.ParticipantID,
		conn.QuizID,
		conn.IsConnected,
		conn.LastSeen,
		conn.InstanceID,
	)

	return err
}

// GetActiveParticipantConnections retrieves all active participant connections for a quiz
func (r *stateRepositoryImpl) GetActiveParticipantConnections(
	ctx context.Context,
	quizID uuid.UUID,
	cutoffTime time.Time,
) ([]*model.ParticipantConnection, error) {
	query := `
		SELECT participant_id, quiz_id, is_connected, last_seen, instance_id
		FROM participant_connections
		WHERE quiz_id = $1 AND is_connected = true AND last_seen > $2
	`

	rows, err := r.db.QueryContext(ctx, query, quizID, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*model.ParticipantConnection
	for rows.Next() {
		conn := &model.ParticipantConnection{}
		err := rows.Scan(
			&conn.ParticipantID,
			&conn.QuizID,
			&conn.IsConnected,
			&conn.LastSeen,
			&conn.InstanceID,
		)
		if err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}

// RegisterInstance registers a server instance
func (r *stateRepositoryImpl) RegisterInstance(ctx context.Context, instance *model.ServerInstance) error {
	query := `
		INSERT INTO server_instances (instance_id, last_heartbeat)
		VALUES ($1, $2)
		ON CONFLICT (instance_id) DO UPDATE
		SET last_heartbeat = $2
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		instance.InstanceID,
		instance.LastHeartbeat,
	)

	return err
}

// UpdateInstanceHeartbeat updates the heartbeat for a server instance
func (r *stateRepositoryImpl) UpdateInstanceHeartbeat(ctx context.Context, instanceID string) error {
	query := `
		UPDATE server_instances
		SET last_heartbeat = $1
		WHERE instance_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), instanceID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("instance not found")
	}

	return nil
}

// GetActiveInstances retrieves all active server instances
func (r *stateRepositoryImpl) GetActiveInstances(
	ctx context.Context,
	cutoffTime time.Time,
) ([]*model.ServerInstance, error) {
	query := `
		SELECT instance_id, last_heartbeat
		FROM server_instances
		WHERE last_heartbeat > $1
	`

	rows, err := r.db.QueryContext(ctx, query, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*model.ServerInstance
	for rows.Next() {
		instance := &model.ServerInstance{}
		err := rows.Scan(
			&instance.InstanceID,
			&instance.LastHeartbeat,
		)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

// IncrementSequenceNumber increments and returns the sequence number for a quiz
func (r *stateRepositoryImpl) IncrementSequenceNumber(ctx context.Context, quizID uuid.UUID) (int64, error) {
	// Using a transaction to ensure atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Get the current max sequence number for this quiz
	var currentSeq int64 = 0
	query := `
		SELECT COALESCE(MAX(sequence_number), 0)
		FROM quiz_events
		WHERE quiz_id = $1
	`
	err = tx.QueryRowContext(ctx, query, quizID).Scan(&currentSeq)
	if err != nil {
		return 0, err
	}

	// Increment the sequence number
	newSeq := currentSeq + 1

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return newSeq, nil
}
