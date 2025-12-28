-- =====================================================
-- Migration 000045: Media Progress Tracking System
-- Description: Create tables for tracking user reading/watching progress
-- 
-- ARCHITECTURE:
-- ┌────────────────────────────────────────────────────────────────────────────┐
-- │                        MEDIA PROGRESS SYSTEM                                │
-- ├────────────────────────────────────────────────────────────────────────────┤
-- │                                                                            │
-- │  ┌──────────────────────┐         ┌──────────────────────┐                │
-- │  │   media_progress     │         │    unit_progress     │                │
-- │  │   (1 per media)      │         │  (1 per chapter)     │                │
-- │  ├──────────────────────┤         ├──────────────────────┤                │
-- │  │ - current_unit_id    │◄────────│ - unit_id            │                │
-- │  │ - position (JSONB)   │         │ - status (completed/ │                │
-- │  │ - completed_units    │         │   in_progress)       │                │
-- │  │ - progress_percentage│         │ - position (JSONB)   │                │
-- │  │ - last_accessed_at   │         │ - started_at         │                │
-- │  └──────────────────────┘         │ - completed_at       │                │
-- │                                   └──────────────────────┘                │
-- │                                                                            │
-- │  USE CASES:                                                                │
-- │  ━━━━━━━━━━                                                                │
-- │  1. "Continue Reading" Section                                             │
-- │     → Query media_progress, JOIN với novels/chapters                       │
-- │     → ORDER BY last_accessed_at DESC, LIMIT N                             │
-- │                                                                            │
-- │  2. "Chapter List" với trạng thái đã đọc                                  │
-- │     → Query unit_progress WHERE media_id = X                              │
-- │     → JOIN với chapters để hiển thị icon "đã đọc"                         │
-- │                                                                            │
-- │  3. Khi user đọc chapter                                                   │
-- │     → UPSERT unit_progress (in_progress)                                  │
-- │     → UPDATE media_progress (current_unit, position)                      │
-- │                                                                            │
-- │  4. Khi user đọc xong chapter                                             │
-- │     → UPDATE unit_progress (completed)                                    │
-- │     → UPDATE media_progress (completed_units++)                           │
-- │                                                                            │
-- └────────────────────────────────────────────────────────────────────────────┘
-- =====================================================

-- =============================================================================
-- TABLE 1: media_progress
-- Tracks overall progress for each media (novel/manga/anime) per user
-- =============================================================================
CREATE TABLE catalog.media_progress (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    
    -- User reference
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    
    -- Media reference (polymorphic - validated in application)
    -- Supports: novel, manga, anime (future)
    media_type VARCHAR(20) NOT NULL CHECK (media_type IN ('novel', 'manga', 'anime')),
    media_id UUID NOT NULL,
    
    -- Current position: which unit (chapter/episode) user is currently reading/watching
    current_unit_id UUID NOT NULL,
    
    -- Position details WITHIN the current unit
    -- Stored as JSONB for flexibility across different media types:
    --   Novel: { "node_id": "paragraph-uuid", "preview": "First 100 chars..." }
    --   Manga: { "page": 15 }
    --   Anime: { "time": "12:34", "seconds": 754 }
    position JSONB NOT NULL DEFAULT '{}',
    
    -- Progress statistics (denormalized for fast reads, updated by triggers/app)
    total_units INTEGER NOT NULL DEFAULT 0,       -- Total chapters/episodes in media
    completed_units INTEGER NOT NULL DEFAULT 0,   -- Number of completed units
    progress_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00 
        CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    
    -- Timestamps
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Unique constraint: one progress entry per user per media
    UNIQUE(user_id, media_type, media_id)
);

-- Indexes for media_progress
-- Primary query: Get recent items for "Continue Reading"
CREATE INDEX idx_media_progress_user_recent 
    ON catalog.media_progress(user_id, last_accessed_at DESC);

-- For media-specific queries (e.g., get all users reading a novel)
CREATE INDEX idx_media_progress_media 
    ON catalog.media_progress(media_type, media_id);

-- Comments
COMMENT ON TABLE catalog.media_progress IS 
    'Tracks overall reading/watching progress per user per media (novel/manga/anime)';
COMMENT ON COLUMN catalog.media_progress.current_unit_id IS 
    'The chapter/episode user is currently reading/watching';
COMMENT ON COLUMN catalog.media_progress.position IS 
    'Position within current unit: Novel{node_id,preview}, Manga{page}, Anime{time,seconds}';
COMMENT ON COLUMN catalog.media_progress.completed_units IS 
    'Denormalized count of completed chapters/episodes (for fast progress display)';


-- =============================================================================
-- TABLE 2: unit_progress
-- Tracks individual chapter/episode progress for each user
-- Allows answering: "Which chapters has user X read?"
-- =============================================================================
CREATE TABLE catalog.unit_progress (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    
    -- User reference
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,
    
    -- Media reference (for efficient querying by media)
    media_type VARCHAR(20) NOT NULL CHECK (media_type IN ('novel', 'manga', 'anime')),
    media_id UUID NOT NULL,
    
    -- Unit reference (chapter/episode)
    unit_id UUID NOT NULL,
    
    -- Read/Watch status
    -- in_progress: User has started but not finished
    -- completed: User has finished reading/watching
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' 
        CHECK (status IN ('in_progress', 'completed')),
    
    -- Position within unit (for resuming)
    -- Same format as media_progress.position
    position JSONB NOT NULL DEFAULT '{}',
    
    -- Time tracking
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),    -- When user first opened
    completed_at TIMESTAMP WITH TIME ZONE,                          -- When user finished (NULL if in_progress)
    last_accessed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(), -- Last time user accessed
    
    -- Unique constraint: one entry per user per unit
    UNIQUE(user_id, unit_id)
);

-- Indexes for unit_progress
-- Query: Get all chapters a user has read for a specific novel
CREATE INDEX idx_unit_progress_user_media 
    ON catalog.unit_progress(user_id, media_type, media_id);

-- Query: Get recent reading activity for a user
CREATE INDEX idx_unit_progress_user_recent 
    ON catalog.unit_progress(user_id, last_accessed_at DESC);

-- Query: Get all users who have read a specific chapter
CREATE INDEX idx_unit_progress_unit 
    ON catalog.unit_progress(unit_id);

-- Comments
COMMENT ON TABLE catalog.unit_progress IS 
    'Tracks individual chapter/episode read status per user';
COMMENT ON COLUMN catalog.unit_progress.status IS 
    'Read status: in_progress (started but not finished), completed (finished)';
COMMENT ON COLUMN catalog.unit_progress.started_at IS 
    'Timestamp when user first opened this chapter/episode';
COMMENT ON COLUMN catalog.unit_progress.completed_at IS 
    'Timestamp when user finished (NULL if still in_progress)';


-- =============================================================================
-- TRIGGER: Auto-update updated_at on media_progress
-- =============================================================================
CREATE OR REPLACE FUNCTION catalog.update_media_progress_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_media_progress_updated_at
    BEFORE UPDATE ON catalog.media_progress
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_media_progress_timestamp();
