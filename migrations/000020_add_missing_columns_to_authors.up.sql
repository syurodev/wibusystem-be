-- Add missing columns to authors table
ALTER TABLE catalog.authors
ADD COLUMN total_chapters INTEGER NOT NULL DEFAULT 0 CHECK (total_chapters >= 0),
ADD COLUMN follower_count INTEGER NOT NULL DEFAULT 0 CHECK (follower_count >= 0),
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;

-- Create indexes for filtering
CREATE INDEX idx_authors_is_verified ON catalog.authors(is_verified) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_follower_count ON catalog.authors(follower_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN catalog.authors.total_chapters IS 'Total number of chapters written by this author across all novels';
COMMENT ON COLUMN catalog.authors.follower_count IS 'Number of followers for this author';
COMMENT ON COLUMN catalog.authors.is_verified IS 'Whether the author has been verified by the platform';
