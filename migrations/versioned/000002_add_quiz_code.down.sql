-- Revert quiz code changes
ALTER TABLE quizzes DROP COLUMN IF EXISTS code;
