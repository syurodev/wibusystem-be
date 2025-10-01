-- Add summary field to novel_translation table
-- This allows storing translated summaries in JSONB format from Plate editor

ALTER TABLE novel_translation
ADD COLUMN summary JSONB; -- Summary content from Plate editor

-- Add comment to clarify the column purpose
COMMENT ON COLUMN novel_translation.summary IS 'Summary content in JSONB format from Plate editor';