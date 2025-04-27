-- Remove description column from quizzes table
ALTER TABLE quizzes DROP COLUMN IF EXISTS description;