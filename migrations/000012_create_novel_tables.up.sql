-- Create catalog schema if not exists
CREATE SCHEMA IF NOT EXISTS catalog;

-- Create enum types for novel status
CREATE TYPE catalog.novel_status AS ENUM ('draft', 'ongoing', 'completed', 'hiatus', 'dropped');

-- Create enum types for chapter status
CREATE TYPE catalog.chapter_status AS ENUM ('draft', 'published', 'scheduled');

-- =====================================================
-- NOVELS TABLE (Top Level)
-- =====================================================
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE, -- For SEO-friendly URLs
    author_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Synopsis stored as JSONB for rich content structure
    -- Example: {"blocks": [{"type": "paragraph", "content": "..."}], "language": "vi"}
    synopsis JSONB,

    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Statistics
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00, -- 0.00 to 5.00
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Metadata stored as JSONB for flexibility
    -- Example: {"tags": ["fantasy", "action"], "categories": ["xuanhuan"], "language": "vi", "original_language": "zh"}
    metadata JSONB DEFAULT '{}',

    -- Publishing dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft delete

    -- Constraints
    CONSTRAINT novels_rating_check CHECK (rating_average >= 0 AND rating_average <= 5),
    CONSTRAINT novels_counts_check CHECK (total_volumes >= 0 AND total_chapters >= 0 AND total_words >= 0)
);

-- =====================================================
-- VOLUMES TABLE (Middle Level)
-- =====================================================
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    volume_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL, -- For SEO
    description TEXT,

    cover_image_url VARCHAR(1000),

    -- Statistics
    chapter_count INTEGER NOT NULL DEFAULT 0,
    word_count BIGINT NOT NULL DEFAULT 0,

    -- Ordering and status
    display_order INTEGER NOT NULL, -- For custom ordering
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    -- Publishing dates
    published_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT volumes_unique_number UNIQUE (novel_id, volume_number),
    CONSTRAINT volumes_unique_slug UNIQUE (novel_id, slug),
    CONSTRAINT volumes_number_positive CHECK (volume_number > 0),
    CONSTRAINT volumes_counts_check CHECK (chapter_count >= 0 AND word_count >= 0)
);

-- =====================================================
-- CHAPTERS TABLE (Bottom Level)
-- =====================================================
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL, -- Chapters can exist without volumes

    chapter_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL, -- For SEO

    -- Content stored as JSONB for rich formatting
    -- Example: {"blocks": [{"type": "paragraph", "content": "..."}], "version": "1.0"}
    content JSONB NOT NULL,

    -- Content metadata
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Access control
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00, -- For paid chapters
    currency VARCHAR(3) DEFAULT 'VND',

    -- Status and visibility
    status catalog.chapter_status NOT NULL DEFAULT 'draft',

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Ordering
    display_order INTEGER NOT NULL, -- For custom ordering within novel

    -- Author notes (also JSONB for rich content)
    author_notes JSONB,

    -- Publishing dates
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE, -- For scheduled publishing

    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT chapters_unique_number UNIQUE (novel_id, chapter_number),
    CONSTRAINT chapters_unique_slug UNIQUE (novel_id, slug),
    CONSTRAINT chapters_number_positive CHECK (chapter_number > 0),
    CONSTRAINT chapters_counts_check CHECK (word_count >= 0 AND character_count >= 0),
    CONSTRAINT chapters_price_check CHECK (price >= 0)
);

-- =====================================================
-- INDEXES for Performance
-- =====================================================

-- Novels indexes
CREATE INDEX idx_novels_author_id ON catalog.novels(author_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_status ON catalog.novels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_at ON catalog.novels(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_rating ON catalog.novels(rating_average DESC) WHERE deleted_at IS NULL AND status = 'ongoing';
CREATE INDEX idx_novels_views ON catalog.novels(view_count DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata); -- For JSONB queries
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis); -- For full-text search

-- Volumes indexes
CREATE INDEX idx_volumes_novel_id ON catalog.volumes(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_display_order ON catalog.volumes(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_published ON catalog.volumes(novel_id, published_at DESC) WHERE is_published = TRUE AND deleted_at IS NULL;

-- Chapters indexes
CREATE INDEX idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_volume_id ON catalog.chapters(volume_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_display_order ON catalog.chapters(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_published ON catalog.chapters(novel_id, published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapters_status ON catalog.chapters(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_content ON catalog.chapters USING GIN(content); -- For full-text search in content

-- =====================================================
-- TRIGGER FUNCTIONS for Auto-updating Statistics
-- =====================================================

-- Function to update novel statistics when volume changes
CREATE OR REPLACE FUNCTION catalog.update_novel_volume_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE catalog.novels
    SET
        total_volumes = (
            SELECT COUNT(*)
            FROM catalog.volumes
            WHERE novel_id = COALESCE(NEW.novel_id, OLD.novel_id)
            AND deleted_at IS NULL
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.novel_id, OLD.novel_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Function to update novel and volume statistics when chapter changes
CREATE OR REPLACE FUNCTION catalog.update_chapter_stats()
RETURNS TRIGGER AS $$
DECLARE
    v_novel_id UUID;
    v_volume_id UUID;
BEGIN
    v_novel_id := COALESCE(NEW.novel_id, OLD.novel_id);
    v_volume_id := COALESCE(NEW.volume_id, OLD.volume_id);

    -- Update novel statistics
    UPDATE catalog.novels
    SET
        total_chapters = (
            SELECT COUNT(*)
            FROM catalog.chapters
            WHERE novel_id = v_novel_id
            AND deleted_at IS NULL
        ),
        total_words = (
            SELECT COALESCE(SUM(word_count), 0)
            FROM catalog.chapters
            WHERE novel_id = v_novel_id
            AND deleted_at IS NULL
        ),
        last_chapter_at = (
            SELECT MAX(published_at)
            FROM catalog.chapters
            WHERE novel_id = v_novel_id
            AND status = 'published'
            AND deleted_at IS NULL
        ),
        updated_at = NOW()
    WHERE id = v_novel_id;

    -- Update volume statistics if chapter belongs to a volume
    IF v_volume_id IS NOT NULL THEN
        UPDATE catalog.volumes
        SET
            chapter_count = (
                SELECT COUNT(*)
                FROM catalog.chapters
                WHERE volume_id = v_volume_id
                AND deleted_at IS NULL
            ),
            word_count = (
                SELECT COALESCE(SUM(word_count), 0)
                FROM catalog.chapters
                WHERE volume_id = v_volume_id
                AND deleted_at IS NULL
            ),
            updated_at = NOW()
        WHERE id = v_volume_id;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION catalog.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger to update novel statistics when volumes change
CREATE TRIGGER trg_volumes_update_novel_stats
    AFTER INSERT OR UPDATE OR DELETE ON catalog.volumes
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_novel_volume_stats();

-- Trigger to update statistics when chapters change
CREATE TRIGGER trg_chapters_update_stats
    AFTER INSERT OR UPDATE OR DELETE ON catalog.chapters
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_chapter_stats();

-- Triggers to auto-update updated_at
CREATE TRIGGER trg_novels_updated_at
    BEFORE UPDATE ON catalog.novels
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_updated_at_column();

CREATE TRIGGER trg_volumes_updated_at
    BEFORE UPDATE ON catalog.volumes
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_updated_at_column();

CREATE TRIGGER trg_chapters_updated_at
    BEFORE UPDATE ON catalog.chapters
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_updated_at_column();

-- =====================================================
-- COMMENTS for Documentation
-- =====================================================

COMMENT ON TABLE catalog.novels IS 'Top-level table storing novel information';
COMMENT ON TABLE catalog.volumes IS 'Middle-level table organizing chapters into volumes';
COMMENT ON TABLE catalog.chapters IS 'Bottom-level table storing chapter content';

COMMENT ON COLUMN catalog.novels.synopsis IS 'JSONB field for rich synopsis content. Example: {"blocks": [{"type": "paragraph", "content": "text"}]}';
COMMENT ON COLUMN catalog.novels.metadata IS 'JSONB field for tags, categories, language. Example: {"tags": ["fantasy"], "categories": ["xuanhuan"], "language": "vi"}';
COMMENT ON COLUMN catalog.chapters.content IS 'JSONB field for rich chapter content. Example: {"blocks": [{"type": "paragraph", "content": "text"}], "version": "1.0"}';
COMMENT ON COLUMN catalog.chapters.author_notes IS 'JSONB field for author notes with rich formatting';
