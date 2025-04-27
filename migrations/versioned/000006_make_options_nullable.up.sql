-- Make the deprecated option columns nullable
ALTER TABLE questions ALTER COLUMN option_a DROP NOT NULL;
ALTER TABLE questions ALTER COLUMN option_b DROP NOT NULL;
ALTER TABLE questions ALTER COLUMN option_c DROP NOT NULL;
ALTER TABLE questions ALTER COLUMN option_d DROP NOT NULL;
ALTER TABLE questions ALTER COLUMN correct_answer DROP NOT NULL;

-- Update the comment to clarify
COMMENT ON COLUMN questions.option_a IS 'DEPRECATED: use question_options table instead. Column made nullable for backward compatibility.';
COMMENT ON COLUMN questions.option_b IS 'DEPRECATED: use question_options table instead. Column made nullable for backward compatibility.';
COMMENT ON COLUMN questions.option_c IS 'DEPRECATED: use question_options table instead. Column made nullable for backward compatibility.';
COMMENT ON COLUMN questions.option_d IS 'DEPRECATED: use question_options table instead. Column made nullable for backward compatibility.';
COMMENT ON COLUMN questions.correct_answer IS 'DEPRECATED: use question_options table instead. Column made nullable for backward compatibility.';