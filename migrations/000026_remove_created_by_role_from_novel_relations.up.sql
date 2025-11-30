-- =====================================================
-- Migration 000026: Remove created_by and role from novel_authors and novel_artists
-- Description: Simplify novel_authors and novel_artists tables
-- =====================================================

-- Remove role and created_by from novel_authors
ALTER TABLE catalog.novel_authors
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_at;

-- Remove role and created_by from novel_artists
ALTER TABLE catalog.novel_artists
    DROP COLUMN IF EXISTS role,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_at;

-- Update unique constraint for novel_artists (remove role from constraint)
ALTER TABLE catalog.novel_artists
    DROP CONSTRAINT IF EXISTS novel_artists_novel_id_artist_id_role_key;

ALTER TABLE catalog.novel_artists
    ADD CONSTRAINT novel_artists_novel_id_artist_id_key UNIQUE (novel_id, artist_id);

-- Drop role-related indexes
DROP INDEX IF EXISTS catalog.idx_novel_authors_role;
DROP INDEX IF EXISTS catalog.idx_novel_artists_role;
