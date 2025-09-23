-- Add profile fields to users table
ALTER TABLE users
ADD COLUMN avatar_url VARCHAR(255),
ADD COLUMN cover_image_url VARCHAR(255),
ADD COLUMN bio JSONB,
ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Create index on is_blocked for performance
CREATE INDEX idx_users_is_blocked ON users(is_blocked);

-- Create trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
