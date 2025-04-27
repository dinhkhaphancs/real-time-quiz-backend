-- Restore the NOT NULL constraints on option columns
ALTER TABLE questions ALTER COLUMN option_a SET NOT NULL;
ALTER TABLE questions ALTER COLUMN option_b SET NOT NULL;
ALTER TABLE questions ALTER COLUMN option_c SET NOT NULL;
ALTER TABLE questions ALTER COLUMN option_d SET NOT NULL;
ALTER TABLE questions ALTER COLUMN correct_answer SET NOT NULL;

-- Restore original comments
COMMENT ON COLUMN questions.option_a IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_b IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_c IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_d IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.correct_answer IS 'DEPRECATED: use question_options table instead';