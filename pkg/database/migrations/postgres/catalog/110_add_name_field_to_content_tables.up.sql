-- Add name field to main content tables for direct title access
-- This provides a primary/default title while keeping translation support

-- Add name field to anime table
ALTER TABLE anime ADD COLUMN name TEXT;
COMMENT ON COLUMN anime.name IS 'Primary/default title of the anime series';

-- Add name field to manga table
ALTER TABLE manga ADD COLUMN name TEXT;
COMMENT ON COLUMN manga.name IS 'Primary/default title of the manga series';

-- Add name field to novel table
ALTER TABLE novel ADD COLUMN name TEXT;
COMMENT ON COLUMN novel.name IS 'Primary/default title of the novel series';

-- Add indexes for name-based searches
CREATE INDEX idx_anime_name ON anime(name);
CREATE INDEX idx_manga_name ON manga(name);
CREATE INDEX idx_novel_name ON novel(name);

-- Add text search indexes for better search performance
CREATE INDEX idx_anime_name_text_search ON anime USING GIN(to_tsvector('simple', name));
CREATE INDEX idx_manga_name_text_search ON manga USING GIN(to_tsvector('simple', name));
CREATE INDEX idx_novel_name_text_search ON novel USING GIN(to_tsvector('simple', name));