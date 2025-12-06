-- Drop deprecated columns
ALTER TABLE catalog.genres 
DROP COLUMN IF EXISTS icon,
DROP COLUMN IF EXISTS color,
DROP COLUMN IF EXISTS display_order;

-- Add new count columns
ALTER TABLE catalog.genres 
ADD COLUMN IF NOT EXISTS anime_count INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS manga_count INTEGER NOT NULL DEFAULT 0;
