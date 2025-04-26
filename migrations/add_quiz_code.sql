-- Migration to add code column to quizzes table

-- Add code column with unique constraint
ALTER TABLE quizzes ADD COLUMN code VARCHAR(10) UNIQUE;

-- Update existing quizzes with random codes
-- This is a placeholder - in production, you would want to handle existing records more carefully
UPDATE quizzes SET code = CONCAT(
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1),
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1),
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1),
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1),
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1),
  SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ23456789' FROM (random() * 33)::int + 1 FOR 1)
);

-- Make the column NOT NULL after populating it
ALTER TABLE quizzes ALTER COLUMN code SET NOT NULL;