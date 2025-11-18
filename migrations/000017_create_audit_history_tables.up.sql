-- =====================================================
-- Migration 000017: Audit History Tables
-- Description: History tables for novels, volumes, and chapters
-- Note: These tables store METADATA only, NOT full content snapshots
-- Logging is done at APPLICATION layer, NOT via database triggers
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Action type
CREATE TYPE catalog.audit_action AS ENUM (
    'created',
    'updated',
    'deleted',
    'restored',
    'published',
    'unpublished',
    'transferred'
);

-- =====================================================
-- NOVEL HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.novel_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content)
    title VARCHAR(500),
    slug VARCHAR(500),
    status catalog.novel_status,
    owner_type VARCHAR(20),
    owner_id UUID,

    -- Statistics snapshot
    total_volumes INTEGER,
    total_chapters INTEGER,
    total_words BIGINT,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed: ["title", "status", "cover_image_url"]
    change_summary TEXT, -- Human-readable summary

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context (rich application context)
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, version_number)
);

-- Indexes for novel_history
CREATE INDEX idx_novel_history_novel ON catalog.novel_history(novel_id);
CREATE INDEX idx_novel_history_version ON catalog.novel_history(novel_id, version_number DESC);
CREATE INDEX idx_novel_history_action ON catalog.novel_history(action);
CREATE INDEX idx_novel_history_changed_by ON catalog.novel_history(changed_by);
CREATE INDEX idx_novel_history_created ON catalog.novel_history(created_at DESC);
CREATE INDEX idx_novel_history_request ON catalog.novel_history(request_id) WHERE request_id IS NOT NULL;

-- Comments
COMMENT ON TABLE catalog.novel_history IS 'Audit log for novel changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.novel_history.changed_fields IS 'JSONB array of field names that changed';
COMMENT ON COLUMN catalog.novel_history.change_summary IS 'Human-readable description of changes';
COMMENT ON COLUMN catalog.novel_history.request_id IS 'Request ID for tracing';

-- =====================================================
-- VOLUME HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.volume_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volume_id UUID NOT NULL REFERENCES catalog.volumes(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content)
    title VARCHAR(500),
    slug VARCHAR(500),
    volume_number INTEGER,
    is_published BOOLEAN,

    -- Statistics snapshot
    chapter_count INTEGER,
    word_count BIGINT,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed
    change_summary TEXT,

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(volume_id, version_number)
);

-- Indexes for volume_history
CREATE INDEX idx_volume_history_volume ON catalog.volume_history(volume_id);
CREATE INDEX idx_volume_history_novel ON catalog.volume_history(novel_id);
CREATE INDEX idx_volume_history_version ON catalog.volume_history(volume_id, version_number DESC);
CREATE INDEX idx_volume_history_action ON catalog.volume_history(action);
CREATE INDEX idx_volume_history_changed_by ON catalog.volume_history(changed_by);
CREATE INDEX idx_volume_history_created ON catalog.volume_history(created_at DESC);

-- Comments
COMMENT ON TABLE catalog.volume_history IS 'Audit log for volume changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.volume_history.changed_fields IS 'JSONB array of field names that changed';

-- =====================================================
-- CHAPTER HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.chapter_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Action
    action catalog.audit_action NOT NULL,

    -- Metadata (NOT full content - content stored separately if needed)
    title VARCHAR(500),
    slug VARCHAR(500),
    chapter_number INTEGER,
    status catalog.chapter_status,

    -- Metrics snapshot
    word_count INTEGER,
    character_count INTEGER,

    -- Change tracking
    changed_fields JSONB, -- Array of field names that changed
    change_summary TEXT,
    content_changed BOOLEAN DEFAULT FALSE, -- Flag indicating if content JSONB changed

    -- Who made this change
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Request context
    request_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(chapter_id, version_number)
);

-- Indexes for chapter_history
CREATE INDEX idx_chapter_history_chapter ON catalog.chapter_history(chapter_id);
CREATE INDEX idx_chapter_history_volume ON catalog.chapter_history(volume_id);
CREATE INDEX idx_chapter_history_novel ON catalog.chapter_history(novel_id);
CREATE INDEX idx_chapter_history_version ON catalog.chapter_history(chapter_id, version_number DESC);
CREATE INDEX idx_chapter_history_action ON catalog.chapter_history(action);
CREATE INDEX idx_chapter_history_changed_by ON catalog.chapter_history(changed_by);
CREATE INDEX idx_chapter_history_created ON catalog.chapter_history(created_at DESC);
CREATE INDEX idx_chapter_history_content_changed ON catalog.chapter_history(chapter_id, content_changed) WHERE content_changed = TRUE;

-- Comments
COMMENT ON TABLE catalog.chapter_history IS 'Audit log for chapter changes (metadata only, logged at application layer)';
COMMENT ON COLUMN catalog.chapter_history.changed_fields IS 'JSONB array of field names that changed';
COMMENT ON COLUMN catalog.chapter_history.content_changed IS 'Flag indicating if content JSONB changed (actual content NOT stored here to save space)';
COMMENT ON COLUMN catalog.chapter_history.change_summary IS 'Human-readable description like "Updated chapter content and title"';
