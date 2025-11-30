-- Enable pg_trgm extension for fuzzy search and ILIKE support
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create GIN indexes using gin_trgm_ops for fast ILIKE searches
-- This supports multi-language and partial matching better than simple FTS
CREATE INDEX idx_novels_title_trgm ON catalog.novels USING GIN (title gin_trgm_ops);
CREATE INDEX idx_novels_original_title_trgm ON catalog.novels USING GIN (original_title gin_trgm_ops);
