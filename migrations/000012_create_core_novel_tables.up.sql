-- =====================================================
-- Migration 000012: Core Novel System Tables
-- Description: novels, volumes, chapters with audit fields
-- =====================================================

-- Create catalog schema if not exists
CREATE SCHEMA IF NOT EXISTS catalog;

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Novel status
CREATE TYPE catalog.novel_status AS ENUM (
    'draft',
    'ongoing',
    'completed',
    'hiatus',
    'dropped'
);

-- Chapter status
CREATE TYPE catalog.chapter_status AS ENUM (
    'draft',
    'published',
    'scheduled'
);

-- =====================================================
-- NOVELS TABLE (Top Level)
-- =====================================================
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE,

    -- Ownership (Polymorphic - validated in application)
    owner_type VARCHAR(20) NOT NULL CHECK (owner_type IN ('user', 'tenant')),
    owner_id UUID NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Content
    synopsis JSONB, -- Synopsis in ORIGINAL language only

    -- Images
    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    -- Original info
    original_language VARCHAR(10) NOT NULL, -- ISO 639-1: vi, en, zh, ja, ko
    original_title VARCHAR(500),

    -- Status
    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Statistics (auto-updated by application)
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00 CHECK (rating_average >= 0 AND rating_average <= 5),
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Additional metadata
    metadata JSONB DEFAULT '{}',

    -- Dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for novels
CREATE INDEX idx_novels_owner ON catalog.novels(owner_type, owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_by ON catalog.novels(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_status ON catalog.novels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_original_language ON catalog.novels(original_language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_slug ON catalog.novels(slug);
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_at ON catalog.novels(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_rating ON catalog.novels(rating_average DESC) WHERE status = 'ongoing' AND deleted_at IS NULL;
CREATE INDEX idx_novels_views ON catalog.novels(view_count DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novels IS 'Top-level table storing novel information with polymorphic ownership';
COMMENT ON COLUMN catalog.novels.owner_type IS 'Polymorphic owner type: user or tenant';
COMMENT ON COLUMN catalog.novels.owner_id IS 'Reference to users.id OR tenants.id based on owner_type (validated in application)';
COMMENT ON COLUMN catalog.novels.created_by IS 'User who created the novel (never changes)';
COMMENT ON COLUMN catalog.novels.updated_by IS 'User who last updated the novel';
COMMENT ON COLUMN catalog.novels.version IS 'Version number, auto-incremented on each update';
COMMENT ON COLUMN catalog.novels.synopsis IS 'Synopsis in ORIGINAL language only. Translations go to novel_synopsis_translations';

-- =====================================================
-- VOLUMES TABLE (Middle Level)
-- =====================================================
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Basic info
    volume_number INTEGER NOT NULL CHECK (volume_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    description TEXT,

    -- Images
    cover_image_url VARCHAR(1000),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Statistics (auto-updated by application)
    chapter_count INTEGER NOT NULL DEFAULT 0 CHECK (chapter_count >= 0),
    word_count BIGINT NOT NULL DEFAULT 0 CHECK (word_count >= 0),

    -- Ordering
    display_order INTEGER NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, volume_number),
    UNIQUE(novel_id, slug)
);

-- Indexes for volumes
CREATE INDEX idx_volumes_novel_id ON catalog.volumes(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_display_order ON catalog.volumes(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_published ON catalog.volumes(novel_id, published_at DESC) WHERE is_published = TRUE AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.volumes IS 'Middle-level table organizing chapters into volumes';
COMMENT ON COLUMN catalog.volumes.chapter_count IS 'Auto-updated by application when chapters change';

-- =====================================================
-- CHAPTERS TABLE (Bottom Level)
-- =====================================================
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,

    -- Basic info
    chapter_number INTEGER NOT NULL CHECK (chapter_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,

    -- Content in ORIGINAL language only (JSONB)
    content JSONB NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    character_count INTEGER NOT NULL DEFAULT 0 CHECK (character_count >= 0),

    -- Access control
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00 CHECK (price >= 0),
    currency VARCHAR(3) DEFAULT 'VND',

    -- Status
    status catalog.chapter_status NOT NULL DEFAULT 'draft',

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Ordering
    display_order INTEGER NOT NULL,

    -- Author notes (JSONB)
    author_notes JSONB,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, chapter_number),
    UNIQUE(novel_id, slug)
);

-- Indexes for chapters
CREATE INDEX idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_volume_id ON catalog.chapters(volume_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_status ON catalog.chapters(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_display_order ON catalog.chapters(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_published ON catalog.chapters(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapters_scheduled ON catalog.chapters(scheduled_at ASC) WHERE status = 'scheduled' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.chapters IS 'Bottom-level table storing chapter content';
COMMENT ON COLUMN catalog.chapters.content IS 'Chapter content in ORIGINAL language only. Translations go to chapter_translations';
COMMENT ON COLUMN catalog.chapters.volume_id IS 'Nullable - chapters can exist without belonging to a volume';

-- =====================================================
-- TRIGGER FUNCTIONS (Minimal - only version increment)
-- =====================================================

-- Function to increment version and update timestamp
CREATE OR REPLACE FUNCTION catalog.increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version := OLD.version + 1;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION catalog.update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Novels triggers
CREATE TRIGGER trg_novels_version
    BEFORE UPDATE ON catalog.novels
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Volumes triggers
CREATE TRIGGER trg_volumes_version
    BEFORE UPDATE ON catalog.volumes
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Chapters triggers
CREATE TRIGGER trg_chapters_version
    BEFORE UPDATE ON catalog.chapters
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
