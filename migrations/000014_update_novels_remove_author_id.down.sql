-- Restore original_title and original_language columns
ALTER TABLE catalog.novels
DROP COLUMN IF EXISTS original_title,
DROP COLUMN IF EXISTS original_language;

-- Restore author_id column
ALTER TABLE catalog.novels
ADD COLUMN IF NOT EXISTS author_id UUID REFERENCES identify.users(id) ON DELETE RESTRICT;

-- Restore index
CREATE INDEX IF NOT EXISTS idx_novels_author_id ON catalog.novels(author_id) WHERE deleted_at IS NULL;
