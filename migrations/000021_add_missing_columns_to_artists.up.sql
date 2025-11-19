-- Add missing columns to artists table
ALTER TABLE catalog.artists
ADD COLUMN specialization VARCHAR(100),
ADD COLUMN artwork_count INTEGER NOT NULL DEFAULT 0 CHECK (artwork_count >= 0),
ADD COLUMN follower_count INTEGER NOT NULL DEFAULT 0 CHECK (follower_count >= 0),
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;

-- Create indexes for filtering
CREATE INDEX idx_artists_specialization ON catalog.artists(specialization) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_is_verified ON catalog.artists(is_verified) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_follower_count ON catalog.artists(follower_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN catalog.artists.specialization IS 'Artist specialization (e.g., cover_artist, illustrator, manga_artist)';
COMMENT ON COLUMN catalog.artists.artwork_count IS 'Total number of artworks created by this artist';
COMMENT ON COLUMN catalog.artists.follower_count IS 'Number of followers for this artist';
COMMENT ON COLUMN catalog.artists.is_verified IS 'Whether the artist has been verified by the platform';
