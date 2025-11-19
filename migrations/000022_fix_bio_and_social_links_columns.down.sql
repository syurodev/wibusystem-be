-- Revert artists table
ALTER TABLE catalog.artists
DROP COLUMN IF EXISTS social_links;

ALTER TABLE catalog.artists
RENAME COLUMN biography TO bio;

-- Revert authors table
ALTER TABLE catalog.authors
DROP COLUMN IF EXISTS social_links;

ALTER TABLE catalog.authors
RENAME COLUMN biography TO bio;
