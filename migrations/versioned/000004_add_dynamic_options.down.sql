-- Drop the question_options table
DROP TABLE IF EXISTS question_options;

-- Remove question_type column from questions table
ALTER TABLE questions DROP COLUMN IF EXISTS question_type;