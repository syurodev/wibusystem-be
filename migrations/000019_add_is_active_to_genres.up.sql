-- Add is_active column to genres table
ALTER TABLE catalog.genres
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Create index for active genres filtering
CREATE INDEX idx_genres_is_active ON catalog.genres(is_active) WHERE deleted_at IS NULL;

-- Comment
COMMENT ON COLUMN catalog.genres.is_active IS 'Whether the genre is active and visible to users';
