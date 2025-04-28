-- Add fields for better quiz state tracking
ALTER TABLE quiz_sessions
ADD COLUMN current_phase VARCHAR(30) NOT NULL DEFAULT 'BETWEEN_QUESTIONS',
ADD COLUMN current_question_ended_at TIMESTAMP NULL,
ADD COLUMN next_question_id UUID NULL REFERENCES questions(id);

-- Create tables for event tracking and connection management
CREATE TABLE IF NOT EXISTS quiz_events (
    id SERIAL PRIMARY KEY,
    quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    sequence_number BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE (quiz_id, sequence_number)
);
CREATE INDEX idx_quiz_events_quiz_seq ON quiz_events(quiz_id, sequence_number);

CREATE TABLE IF NOT EXISTS participant_connections (
    participant_id UUID REFERENCES participants(id) ON DELETE CASCADE,
    quiz_id UUID REFERENCES quizzes(id) ON DELETE CASCADE,
    is_connected BOOLEAN NOT NULL DEFAULT false,
    last_seen TIMESTAMP NOT NULL,
    instance_id VARCHAR(50),
    PRIMARY KEY (participant_id, quiz_id)
);

CREATE TABLE IF NOT EXISTS server_instances (
    instance_id VARCHAR(50) PRIMARY KEY,
    last_heartbeat TIMESTAMP NOT NULL
);