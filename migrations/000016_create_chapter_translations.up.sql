-- =====================================================
-- Migration 000016: Chapter Translations
-- Description: Translations, contributions, and version history for chapters
-- =====================================================

-- =====================================================
-- CHAPTER TRANSLATIONS TABLE
-- =====================================================
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Translation content
    content JSONB NOT NULL,
    title VARCHAR(500) NOT NULL,

    -- Team assignment (optional - can be contributed by individuals)
    team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),
    reviewer_rating DECIMAL(3,2) DEFAULT 0.00 CHECK (reviewer_rating >= 0 AND reviewer_rating >= 0),

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    character_count INTEGER NOT NULL DEFAULT 0 CHECK (character_count >= 0),

    -- Statistics (auto-updated by application)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Author/Translator notes
    translator_notes JSONB,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One translation per chapter per language
    UNIQUE(chapter_id, language)
);

-- Indexes for chapter_translations
CREATE INDEX idx_chapter_translations_chapter_id ON catalog.chapter_translations(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_language ON catalog.chapter_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_team_id ON catalog.chapter_translations(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_status ON catalog.chapter_translations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_published ON catalog.chapter_translations(chapter_id, language, published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_content ON catalog.chapter_translations USING GIN(content) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_views ON catalog.chapter_translations(view_count DESC) WHERE status = 'published' AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.chapter_translations IS 'Translated chapter content';
COMMENT ON COLUMN catalog.chapter_translations.content IS 'Translated chapter content in JSONB format';
COMMENT ON COLUMN catalog.chapter_translations.team_id IS 'Optional team responsible for this translation';
COMMENT ON COLUMN catalog.chapter_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';
COMMENT ON COLUMN catalog.chapter_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

-- =====================================================
-- TRANSLATION CONTRIBUTIONS TABLE
-- =====================================================
CREATE TABLE catalog.translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Contribution details
    contribution_type VARCHAR(50) NOT NULL, -- 'translation', 'proofread', 'edit', 'review', 'typeset'
    contribution_notes TEXT,
    status catalog.contribution_status NOT NULL DEFAULT 'pending',

    -- Changes (optional - stores metadata about changes, NOT full content)
    changes JSONB,

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    contributed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for translation_contributions
CREATE INDEX idx_translation_contributions_chapter ON catalog.translation_contributions(chapter_translation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributor ON catalog.translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_type ON catalog.translation_contributions(contribution_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_status ON catalog.translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributed ON catalog.translation_contributions(contributed_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.translation_contributions IS 'Tracks individual contributions to chapter translations';
COMMENT ON COLUMN catalog.translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review, typeset';
COMMENT ON COLUMN catalog.translation_contributions.changes IS 'JSONB documenting metadata about changes (NOT full content to save space)';

-- =====================================================
-- TRANSLATION HISTORY TABLE
-- =====================================================
CREATE TABLE catalog.translation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,

    -- Version tracking
    version_number INTEGER NOT NULL CHECK (version_number > 0),

    -- Metadata (NOT full content - to save space)
    title VARCHAR(500),
    word_count INTEGER,
    character_count INTEGER,
    status catalog.translation_status,

    -- Change summary
    change_summary TEXT,
    changed_fields JSONB, -- Array of field names that changed

    -- Who made this version
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(chapter_translation_id, version_number)
);

-- Indexes for translation_history
CREATE INDEX idx_translation_history_chapter ON catalog.translation_history(chapter_translation_id);
CREATE INDEX idx_translation_history_version ON catalog.translation_history(chapter_translation_id, version_number DESC);
CREATE INDEX idx_translation_history_changed_by ON catalog.translation_history(changed_by);
CREATE INDEX idx_translation_history_created ON catalog.translation_history(created_at DESC);

-- Comments
COMMENT ON TABLE catalog.translation_history IS 'Version control for chapter translations (metadata only, NOT full content)';
COMMENT ON COLUMN catalog.translation_history.changed_fields IS 'JSONB array of field names that changed in this version';
COMMENT ON COLUMN catalog.translation_history.change_summary IS 'Human-readable summary of what changed';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Chapter translations triggers
CREATE TRIGGER trg_chapter_translations_version
    BEFORE UPDATE ON catalog.chapter_translations
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Translation contributions triggers
CREATE TRIGGER trg_translation_contributions_version
    BEFORE UPDATE ON catalog.translation_contributions
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
