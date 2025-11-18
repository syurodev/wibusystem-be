-- =====================================================
-- GENRES SYSTEM
-- =====================================================

-- Bảng genres (thể loại)
CREATE TABLE catalog.genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Parent genre for hierarchical structure
    parent_id UUID REFERENCES catalog.genres(id) ON DELETE SET NULL,

    -- Display
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Many-to-many: novels <-> genres
CREATE TABLE catalog.novel_genres (
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES catalog.genres(id) ON DELETE CASCADE,
    display_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (novel_id, genre_id)
);

-- =====================================================
-- CONTRIBUTORS SYSTEM (Authors, Artists, Translators)
-- =====================================================

-- Bảng authors (tác giả)
CREATE TABLE catalog.authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL, -- Link to user account if registered

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    -- Biography stored as JSONB
    biography JSONB,

    avatar_url VARCHAR(1000),

    -- Social links stored as JSONB
    -- Example: {"facebook": "...", "twitter": "...", "website": "..."}
    social_links JSONB DEFAULT '{}',

    -- Statistics
    novel_count INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_views BIGINT NOT NULL DEFAULT 0,
    follower_count INTEGER NOT NULL DEFAULT 0,

    -- Status
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Bảng artists (hoạ sĩ/minh hoạ)
CREATE TABLE catalog.artists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    biography JSONB,
    avatar_url VARCHAR(1000),
    social_links JSONB DEFAULT '{}',

    -- Specialization: cover_artist, illustrator, manga_artist, etc.
    specialization VARCHAR(50),

    -- Statistics
    novel_count INTEGER NOT NULL DEFAULT 0,
    artwork_count INTEGER NOT NULL DEFAULT 0,
    follower_count INTEGER NOT NULL DEFAULT 0,

    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Bảng translators (người dịch)
CREATE TABLE catalog.translators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    biography JSONB,
    avatar_url VARCHAR(1000),

    -- Languages: ["vi", "en", "zh", "ja", "ko"]
    languages JSONB NOT NULL DEFAULT '[]',

    -- Statistics
    novel_count INTEGER NOT NULL DEFAULT 0,
    chapter_count INTEGER NOT NULL DEFAULT 0,
    word_count BIGINT NOT NULL DEFAULT 0,
    contribution_count INTEGER NOT NULL DEFAULT 0,
    follower_count INTEGER NOT NULL DEFAULT 0,

    -- Quality rating (0-5)
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    is_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =====================================================
-- NOVEL CONTRIBUTORS RELATIONSHIPS
-- =====================================================

-- Many-to-many: novels <-> authors
CREATE TABLE catalog.novel_authors (
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES catalog.authors(id) ON DELETE CASCADE,

    -- Role: original_author, co_author, etc.
    role VARCHAR(50) NOT NULL DEFAULT 'author',
    display_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (novel_id, author_id)
);

-- Many-to-many: novels <-> artists
CREATE TABLE catalog.novel_artists (
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    artist_id UUID NOT NULL REFERENCES catalog.artists(id) ON DELETE CASCADE,

    -- Role: cover_artist, illustrator, character_designer, etc.
    role VARCHAR(50) NOT NULL DEFAULT 'illustrator',
    display_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (novel_id, artist_id, role)
);

-- Many-to-many: novels <-> translators (primary translators)
CREATE TABLE catalog.novel_translators (
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    translator_id UUID NOT NULL REFERENCES catalog.translators(id) ON DELETE CASCADE,

    -- Language being translated to
    target_language VARCHAR(10) NOT NULL,

    -- Role: lead_translator, translator, proofreader, etc.
    role VARCHAR(50) NOT NULL DEFAULT 'translator',
    display_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (novel_id, translator_id, target_language)
);

-- =====================================================
-- TRANSLATION SYSTEM
-- =====================================================

-- Create enum for translation status
CREATE TYPE catalog.translation_status AS ENUM ('draft', 'pending_review', 'approved', 'rejected', 'published');

-- Create enum for contribution type
CREATE TYPE catalog.contribution_type AS ENUM ('new_translation', 'improvement', 'proofreading', 'correction');

-- Bảng chapter_translations (bản dịch chính thức của chapter)
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,

    -- Language code (ISO 639-1)
    language VARCHAR(10) NOT NULL,

    -- Title and content in target language
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,

    -- Translator notes
    translator_notes JSONB,

    -- Primary translator
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,

    -- Version tracking
    version INTEGER NOT NULL DEFAULT 1,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Publishing
    published_at TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Unique: one published translation per chapter per language
    CONSTRAINT chapter_translations_unique UNIQUE (chapter_id, language)
);

-- Bảng translation_contributions (đóng góp bản dịch từ community)
CREATE TABLE catalog.translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,

    -- Contributor
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Target language
    language VARCHAR(10) NOT NULL,

    -- Contribution type
    contribution_type catalog.contribution_type NOT NULL,

    -- Content
    title VARCHAR(500),
    content JSONB NOT NULL,
    contributor_notes TEXT,

    -- Status and review
    status catalog.translation_status NOT NULL DEFAULT 'pending_review',

    -- Reviewer information
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- If approved, link to the official translation
    official_translation_id UUID REFERENCES catalog.chapter_translations(id) ON DELETE SET NULL,

    -- Credit and rewards
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT FALSE,

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Community feedback
    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =====================================================
-- TRANSLATION HISTORY (Version Control)
-- =====================================================

-- Track all changes to translations
CREATE TABLE catalog.translation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,

    version INTEGER NOT NULL,

    -- Snapshot of content at this version
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,

    -- Who made the change
    changed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    change_description TEXT,

    -- Metrics at this version
    word_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- =====================================================
-- INDEXES
-- =====================================================

-- Genres indexes
CREATE INDEX idx_genres_parent_id ON catalog.genres(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_genres_slug ON catalog.genres(slug);
CREATE INDEX idx_genres_active ON catalog.genres(is_active) WHERE is_active = TRUE;

-- Novel genres indexes
CREATE INDEX idx_novel_genres_genre_id ON catalog.novel_genres(genre_id);
CREATE INDEX idx_novel_genres_novel_id ON catalog.novel_genres(novel_id);

-- Authors indexes
CREATE INDEX idx_authors_user_id ON catalog.authors(user_id) WHERE user_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_authors_slug ON catalog.authors(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_authors_verified ON catalog.authors(is_verified) WHERE is_verified = TRUE AND deleted_at IS NULL;

-- Artists indexes
CREATE INDEX idx_artists_user_id ON catalog.artists(user_id) WHERE user_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_artists_slug ON catalog.artists(slug) WHERE deleted_at IS NULL;

-- Translators indexes
CREATE INDEX idx_translators_user_id ON catalog.translators(user_id) WHERE user_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_translators_slug ON catalog.translators(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_translators_languages ON catalog.translators USING GIN(languages);

-- Novel contributors indexes
CREATE INDEX idx_novel_authors_author_id ON catalog.novel_authors(author_id);
CREATE INDEX idx_novel_artists_artist_id ON catalog.novel_artists(artist_id);
CREATE INDEX idx_novel_translators_translator_id ON catalog.novel_translators(translator_id);
CREATE INDEX idx_novel_translators_language ON catalog.novel_translators(target_language);

-- Chapter translations indexes
CREATE INDEX idx_chapter_translations_chapter_id ON catalog.chapter_translations(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_language ON catalog.chapter_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_translator_id ON catalog.chapter_translations(translator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_status ON catalog.chapter_translations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_content ON catalog.chapter_translations USING GIN(content);

-- Translation contributions indexes
CREATE INDEX idx_translation_contributions_chapter_id ON catalog.translation_contributions(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributor_id ON catalog.translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_status ON catalog.translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_language ON catalog.translation_contributions(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_reviewed_by ON catalog.translation_contributions(reviewed_by) WHERE reviewed_by IS NOT NULL;

-- Translation history indexes
CREATE INDEX idx_translation_history_translation_id ON catalog.translation_history(translation_id);
CREATE INDEX idx_translation_history_version ON catalog.translation_history(translation_id, version DESC);

-- =====================================================
-- TRIGGER FUNCTIONS
-- =====================================================

-- Update author statistics when novel_authors changes
CREATE OR REPLACE FUNCTION catalog.update_author_novel_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE catalog.authors
    SET novel_count = (
        SELECT COUNT(DISTINCT novel_id)
        FROM catalog.novel_authors
        WHERE author_id = COALESCE(NEW.author_id, OLD.author_id)
    )
    WHERE id = COALESCE(NEW.author_id, OLD.author_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Update translator statistics when translations are created/updated
CREATE OR REPLACE FUNCTION catalog.update_translator_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.translator_id IS NOT NULL THEN
        UPDATE catalog.translators
        SET
            chapter_count = (
                SELECT COUNT(*)
                FROM catalog.chapter_translations
                WHERE translator_id = NEW.translator_id
                AND status = 'published'
                AND deleted_at IS NULL
            ),
            word_count = (
                SELECT COALESCE(SUM(word_count), 0)
                FROM catalog.chapter_translations
                WHERE translator_id = NEW.translator_id
                AND status = 'published'
                AND deleted_at IS NULL
            )
        WHERE id = NEW.translator_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create version history when translation is updated
CREATE OR REPLACE FUNCTION catalog.create_translation_history()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE' AND (OLD.content != NEW.content OR OLD.title != NEW.title)) THEN
        INSERT INTO catalog.translation_history (
            translation_id, version, title, content,
            changed_by, change_description, word_count
        ) VALUES (
            NEW.id, NEW.version, NEW.title, NEW.content,
            NULL, -- Will be set by application
            'Content updated', NEW.word_count
        );

        -- Increment version
        NEW.version = OLD.version + 1;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Trigger for author statistics
CREATE TRIGGER trg_novel_authors_update_stats
    AFTER INSERT OR DELETE ON catalog.novel_authors
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_author_novel_count();

-- Trigger for translator statistics
CREATE TRIGGER trg_chapter_translations_update_stats
    AFTER INSERT OR UPDATE ON catalog.chapter_translations
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_translator_stats();

-- Trigger for translation version history
CREATE TRIGGER trg_chapter_translations_history
    BEFORE UPDATE ON catalog.chapter_translations
    FOR EACH ROW
    EXECUTE FUNCTION catalog.create_translation_history();

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE catalog.genres IS 'Thể loại novel (có thể phân cấp)';
COMMENT ON TABLE catalog.authors IS 'Tác giả novel';
COMMENT ON TABLE catalog.artists IS 'Hoạ sĩ/minh họa';
COMMENT ON TABLE catalog.translators IS 'Người dịch';
COMMENT ON TABLE catalog.chapter_translations IS 'Bản dịch chính thức của chapter';
COMMENT ON TABLE catalog.translation_contributions IS 'Đóng góp bản dịch từ community';
COMMENT ON TABLE catalog.translation_history IS 'Lịch sử thay đổi bản dịch';

COMMENT ON COLUMN catalog.authors.biography IS 'JSONB field for rich biography content';
COMMENT ON COLUMN catalog.chapter_translations.content IS 'JSONB field for translated chapter content';
COMMENT ON COLUMN catalog.translation_contributions.content IS 'JSONB field for contribution content';
