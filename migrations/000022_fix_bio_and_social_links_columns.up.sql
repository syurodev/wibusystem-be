-- Fix authors table: rename bio to biography and add social_links
ALTER TABLE catalog.authors
RENAME COLUMN bio TO biography;

ALTER TABLE catalog.authors
ADD COLUMN social_links JSONB DEFAULT '{}';

COMMENT ON COLUMN catalog.authors.biography IS 'Author biography in JSONB format';
COMMENT ON COLUMN catalog.authors.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';

-- Fix artists table: rename bio to biography and add social_links
ALTER TABLE catalog.artists
RENAME COLUMN bio TO biography;

ALTER TABLE catalog.artists
ADD COLUMN social_links JSONB DEFAULT '{}';

COMMENT ON COLUMN catalog.artists.biography IS 'Artist biography in JSONB format';
COMMENT ON COLUMN catalog.artists.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';
