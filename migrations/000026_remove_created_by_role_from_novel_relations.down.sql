-- =====================================================
-- Migration 000026: Remove created_by and role from novel_authors and novel_artists (ROLLBACK)
-- Description: Restore role and created_by columns
-- =====================================================

-- Restore columns to novel_authors
ALTER TABLE catalog.novel_authors
    ADD COLUMN IF NOT EXISTS role catalog.author_role NOT NULL DEFAULT 'original_author',
    ADD COLUMN IF NOT EXISTS created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

-- Restore columns to novel_artists
ALTER TABLE catalog.novel_artists
    ADD COLUMN IF NOT EXISTS role catalog.artist_role NOT NULL DEFAULT 'cover_artist',
    ADD COLUMN IF NOT EXISTS created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

-- Restore unique constraint for novel_artists (include role)
ALTER TABLE catalog.novel_artists
    DROP CONSTRAINT IF EXISTS novel_artists_novel_id_artist_id_key;

ALTER TABLE catalog.novel_artists
    ADD CONSTRAINT novel_artists_novel_id_artist_id_role_key UNIQUE (novel_id, artist_id, role);

-- Restore role-related indexes
CREATE INDEX IF NOT EXISTS idx_novel_authors_role ON catalog.novel_authors(novel_id, role);
CREATE INDEX IF NOT EXISTS idx_novel_artists_role ON catalog.novel_artists(novel_id, role);
