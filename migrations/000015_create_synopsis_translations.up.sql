-- =====================================================
-- Migration 000015: Synopsis Translations
-- Description: Translations and contributions for novel synopsis
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Translation status
CREATE TYPE catalog.translation_status AS ENUM (
    'draft',
    'pending_review',
    'published',
    'rejected'
);

-- Contribution status
CREATE TYPE catalog.contribution_status AS ENUM (
    'pending',
    'accepted',
    'rejected'
);

-- =====================================================
-- NOVEL SYNOPSIS TRANSLATIONS TABLE
-- =====================================================
CREATE TABLE catalog.novel_synopsis_translations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Translation content
    synopsis JSONB NOT NULL,

    -- Organization assignment (optional - can be contributed by individuals)
    organization_id UUID REFERENCES identify.organizations(id) ON DELETE SET NULL,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),
    reviewer_rating DECIMAL(3,2) DEFAULT 0.00 CHECK (reviewer_rating >= 0 AND reviewer_rating <= 5),

    -- Statistics (auto-updated by application)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    review_notes TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,

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

    -- One translation per novel per language
    UNIQUE(novel_id, language)
);

-- Indexes for novel_synopsis_translations
CREATE INDEX idx_synopsis_translations_novel_id ON catalog.novel_synopsis_translations(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_language ON catalog.novel_synopsis_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_organization_id ON catalog.novel_synopsis_translations(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_status ON catalog.novel_synopsis_translations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_published ON catalog.novel_synopsis_translations(novel_id, language, published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_content ON catalog.novel_synopsis_translations USING GIN(synopsis) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novel_synopsis_translations IS 'Translated synopsis for novels';
COMMENT ON COLUMN catalog.novel_synopsis_translations.synopsis IS 'Translated synopsis content in JSONB format';
COMMENT ON COLUMN catalog.novel_synopsis_translations.organization_id IS 'Optional organization responsible for this translation';
COMMENT ON COLUMN catalog.novel_synopsis_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';
COMMENT ON COLUMN catalog.novel_synopsis_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

-- =====================================================
-- SYNOPSIS TRANSLATION CONTRIBUTIONS TABLE
-- =====================================================
CREATE TABLE catalog.synopsis_translation_contributions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    synopsis_translation_id UUID NOT NULL REFERENCES catalog.novel_synopsis_translations(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Contribution details
    contribution_type VARCHAR(50) NOT NULL, -- 'translation', 'proofread', 'edit', 'review'
    contribution_notes TEXT,
    status catalog.contribution_status NOT NULL DEFAULT 'pending',

    -- Changes (optional - for tracking what was changed)
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

-- Indexes for synopsis_translation_contributions
CREATE INDEX idx_synopsis_contributions_translation ON catalog.synopsis_translation_contributions(synopsis_translation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_contributor ON catalog.synopsis_translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_type ON catalog.synopsis_translation_contributions(contribution_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_status ON catalog.synopsis_translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_contributed ON catalog.synopsis_translation_contributions(contributed_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.synopsis_translation_contributions IS 'Tracks individual contributions to synopsis translations';
COMMENT ON COLUMN catalog.synopsis_translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review';
COMMENT ON COLUMN catalog.synopsis_translation_contributions.changes IS 'JSONB documenting what was changed in this contribution';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Synopsis translations triggers
CREATE TRIGGER trg_synopsis_translations_version
    BEFORE UPDATE ON catalog.novel_synopsis_translations
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Synopsis contributions triggers
CREATE TRIGGER trg_synopsis_contributions_version
    BEFORE UPDATE ON catalog.synopsis_translation_contributions
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
