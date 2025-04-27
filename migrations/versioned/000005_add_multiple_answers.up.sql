-- Add selected_options_json column to answers table
ALTER TABLE answers ADD COLUMN selected_options_json TEXT;

-- Migrate existing data - convert single selected_option to JSON array
UPDATE answers 
SET selected_options_json = CONCAT('["', selected_option, '"]')
WHERE selected_option IS NOT NULL;

-- Mark the old column as deprecated
COMMENT ON COLUMN answers.selected_option IS 'DEPRECATED: use selected_options_json instead';