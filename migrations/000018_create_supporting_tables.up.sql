-- =====================================================
-- Migration 000018: Supporting Tables
-- Description: Genres, authors, artists, translators, and their associations
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Contributor role types
CREATE TYPE catalog.author_role AS ENUM (
    'original_author',
    'co_author',
    'ghostwriter'
);

CREATE TYPE catalog.artist_role AS ENUM (
    'cover_artist',
    'illustrator',
    'character_designer'
);

CREATE TYPE catalog.translator_role AS ENUM (
    'translator',
    'localizer',
    'adapter'
);

-- =====================================================
-- GENRES TABLE
-- =====================================================
CREATE TABLE catalog.genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Hierarchy (optional parent for sub-genres)
    parent_id UUID REFERENCES catalog.genres(id) ON DELETE SET NULL,

    -- Display
    icon VARCHAR(100), -- Icon name or emoji
    color VARCHAR(7), -- Hex color code
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),
    active_readers BIGINT NOT NULL DEFAULT 0 CHECK (active_readers >= 0),
    total_views BIGINT NOT NULL DEFAULT 0 CHECK (total_views >= 0),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for genres
CREATE INDEX idx_genres_slug ON catalog.genres(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_genres_parent ON catalog.genres(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_genres_display_order ON catalog.genres(display_order) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.genres IS 'Genre definitions for novels';
COMMENT ON COLUMN catalog.genres.parent_id IS 'Optional parent genre for hierarchical organization';

-- =====================================================
-- NOVEL_GENRES JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES catalog.genres(id) ON DELETE CASCADE,

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, genre_id)
);

-- Indexes for novel_genres
CREATE INDEX idx_novel_genres_novel ON catalog.novel_genres(novel_id);
CREATE INDEX idx_novel_genres_genre ON catalog.novel_genres(genre_id);

-- Comments
COMMENT ON TABLE catalog.novel_genres IS 'Junction table linking novels to genres';

-- =====================================================
-- AUTHORS TABLE
-- =====================================================
CREATE TABLE catalog.authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),

    -- Optional link to user account
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),
    total_views BIGINT NOT NULL DEFAULT 0,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for authors
CREATE INDEX idx_authors_slug ON catalog.authors(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_user_id ON catalog.authors(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_metadata ON catalog.authors USING GIN(metadata) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.authors IS 'Author information';
COMMENT ON COLUMN catalog.authors.user_id IS 'Optional link to user account if author is registered';

-- =====================================================
-- NOVEL_AUTHORS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES catalog.authors(id) ON DELETE CASCADE,

    -- Role
    role catalog.author_role NOT NULL DEFAULT 'original_author',

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, author_id)
);

-- Indexes for novel_authors
CREATE INDEX idx_novel_authors_novel ON catalog.novel_authors(novel_id);
CREATE INDEX idx_novel_authors_author ON catalog.novel_authors(author_id);
CREATE INDEX idx_novel_authors_role ON catalog.novel_authors(novel_id, role);

-- Comments
COMMENT ON TABLE catalog.novel_authors IS 'Junction table linking novels to authors';

-- =====================================================
-- ARTISTS TABLE
-- =====================================================
CREATE TABLE catalog.artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),
    portfolio_url VARCHAR(1000),

    -- Optional link to user account
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- Statistics (auto-updated by application)
    novel_count INTEGER NOT NULL DEFAULT 0 CHECK (novel_count >= 0),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for artists
CREATE INDEX idx_artists_slug ON catalog.artists(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_user_id ON catalog.artists(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_artists_metadata ON catalog.artists USING GIN(metadata) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.artists IS 'Artist information (cover artists, illustrators, etc.)';

-- =====================================================
-- NOVEL_ARTISTS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES catalog.artists(id) ON DELETE CASCADE,

    -- Role
    role catalog.artist_role NOT NULL DEFAULT 'cover_artist',

    -- Ordering
    display_order INTEGER NOT NULL DEFAULT 0,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, artist_id, role)
);

-- Indexes for novel_artists
CREATE INDEX idx_novel_artists_novel ON catalog.novel_artists(novel_id);
CREATE INDEX idx_novel_artists_artist ON catalog.novel_artists(artist_id);
CREATE INDEX idx_novel_artists_role ON catalog.novel_artists(novel_id, role);

-- Comments
COMMENT ON TABLE catalog.novel_artists IS 'Junction table linking novels to artists';

-- =====================================================
-- TRANSLATORS TABLE
-- =====================================================
CREATE TABLE catalog.translators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(1000),

    -- Link to user account (usually required for translators)
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Languages
    source_languages VARCHAR(10)[] NOT NULL, -- Array of ISO 639-1 codes
    target_languages VARCHAR(10)[] NOT NULL, -- Array of ISO 639-1 codes

    -- Statistics (auto-updated by application)
    translation_count INTEGER NOT NULL DEFAULT 0 CHECK (translation_count >= 0),
    total_words_translated BIGINT NOT NULL DEFAULT 0,
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes for translators
CREATE INDEX idx_translators_slug ON catalog.translators(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_user_id ON catalog.translators(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_source_langs ON catalog.translators USING GIN(source_languages) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_target_langs ON catalog.translators USING GIN(target_languages) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_quality ON catalog.translators(quality_score DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.translators IS 'Translator information';
COMMENT ON COLUMN catalog.translators.source_languages IS 'Array of languages the translator can translate from';
COMMENT ON COLUMN catalog.translators.target_languages IS 'Array of languages the translator can translate to';

-- =====================================================
-- NOVEL_TRANSLATORS JUNCTION TABLE
-- =====================================================
CREATE TABLE catalog.novel_translators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    translator_id UUID NOT NULL REFERENCES catalog.translators(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- Target language

    -- Role
    role catalog.translator_role NOT NULL DEFAULT 'translator',

    -- Statistics (auto-updated by application)
    chapters_translated INTEGER NOT NULL DEFAULT 0 CHECK (chapters_translated >= 0),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, translator_id, language)
);

-- Indexes for novel_translators
CREATE INDEX idx_novel_translators_novel ON catalog.novel_translators(novel_id);
CREATE INDEX idx_novel_translators_translator ON catalog.novel_translators(translator_id);
CREATE INDEX idx_novel_translators_language ON catalog.novel_translators(novel_id, language);

-- Comments
COMMENT ON TABLE catalog.novel_translators IS 'Junction table linking novels to translators';
COMMENT ON COLUMN catalog.novel_translators.language IS 'Target language for this translator on this novel';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Genres triggers
CREATE TRIGGER trg_genres_version
    BEFORE UPDATE ON catalog.genres
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Authors triggers
CREATE TRIGGER trg_authors_version
    BEFORE UPDATE ON catalog.authors
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Artists triggers
CREATE TRIGGER trg_artists_version
    BEFORE UPDATE ON catalog.artists
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Translators triggers
CREATE TRIGGER trg_translators_version
    BEFORE UPDATE ON catalog.translators
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
