-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS server_instances;
DROP TABLE IF EXISTS participant_connections;
DROP TABLE IF EXISTS quiz_events;

-- Remove added columns from quiz_sessions
ALTER TABLE quiz_sessions
DROP COLUMN IF EXISTS next_question_id,
DROP COLUMN IF EXISTS current_question_ended_at,
DROP COLUMN IF EXISTS current_phase;