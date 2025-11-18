-- =====================================================
-- UPDATE NOVELS TABLE
-- Remove author_id (now using novel_authors relationship)
-- Update metadata field to only store additional metadata
-- =====================================================

-- Note: Before running this migration, ensure all existing novels
-- have their authors migrated to the novel_authors table

-- Remove the author_id foreign key constraint and column
ALTER TABLE catalog.novels
DROP COLUMN IF EXISTS author_id;

-- Add original_language field (moved from metadata)
ALTER TABLE catalog.novels
ADD COLUMN IF NOT EXISTS original_language VARCHAR(10),
ADD COLUMN IF NOT EXISTS original_title VARCHAR(500);

-- Update comments
COMMENT ON COLUMN catalog.novels.original_language IS 'Original language of the novel (ISO 639-1)';
COMMENT ON COLUMN catalog.novels.original_title IS 'Original title in source language';
COMMENT ON COLUMN catalog.novels.metadata IS 'Additional metadata (custom fields, external IDs, etc.)';
