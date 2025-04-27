-- Add question_type to questions table
ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_type VARCHAR(20) DEFAULT 'SINGLE_CHOICE' NOT NULL;

-- Create question_options table
CREATE TABLE IF NOT EXISTS question_options (
    id UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    display_order INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on question_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_question_options_question_id ON question_options(question_id);

-- Migration of existing data is left intentionally out as per product decision to not maintain backward compatibility
-- We'll drop the old columns after data has been migrated manually if needed

-- Mark these columns as deprecated - we'll keep them for now during the transition
COMMENT ON COLUMN questions.option_a IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_b IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_c IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.option_d IS 'DEPRECATED: use question_options table instead';
COMMENT ON COLUMN questions.correct_answer IS 'DEPRECATED: use question_options table instead';