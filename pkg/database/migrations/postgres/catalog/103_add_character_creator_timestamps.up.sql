-- Add timestamp columns to character and creator tables for better tracking
ALTER TABLE character ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE character ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE creator ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE creator ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add unique constraints to prevent duplicates
ALTER TABLE character ADD CONSTRAINT character_name_unique UNIQUE (name);
ALTER TABLE creator ADD CONSTRAINT creator_name_unique UNIQUE (name);

-- Add indexes for better performance
CREATE INDEX idx_character_name ON character(name);
CREATE INDEX idx_creator_name ON creator(name);