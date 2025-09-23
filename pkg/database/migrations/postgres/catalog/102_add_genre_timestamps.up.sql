-- Add timestamp columns to genre table for better tracking
ALTER TABLE genre ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE genre ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add unique constraint on genre name to prevent duplicates
ALTER TABLE genre ADD CONSTRAINT genre_name_unique UNIQUE (name);

-- Add index on name for faster lookups
CREATE INDEX idx_genre_name ON genre(name);