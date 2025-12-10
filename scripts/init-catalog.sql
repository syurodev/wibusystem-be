CREATE TYPE novel_status AS ENUM ('draft', 'ongoing', 'completed', 'hiatus', 'dropped');

ALTER TYPE novel_status OWNER TO system_dev;

CREATE TYPE chapter_status AS ENUM ('draft', 'published', 'scheduled');

ALTER TYPE chapter_status OWNER TO system_dev;

CREATE TYPE transfer_status AS ENUM ('pending', 'approved', 'rejected', 'cancelled');

ALTER TYPE transfer_status OWNER TO system_dev;

CREATE TYPE report_status AS ENUM ('pending', 'under_review', 'resolved', 'rejected');

ALTER TYPE report_status OWNER TO system_dev;

CREATE TYPE translation_status AS ENUM ('draft', 'pending_review', 'published', 'rejected');

ALTER TYPE translation_status OWNER TO system_dev;

CREATE TYPE contribution_status AS ENUM ('pending', 'accepted', 'rejected');

ALTER TYPE contribution_status OWNER TO system_dev;

CREATE TYPE audit_action AS ENUM ('created', 'updated', 'deleted', 'restored', 'published', 'unpublished', 'transferred');

ALTER TYPE audit_action OWNER TO system_dev;

CREATE TYPE author_role AS ENUM ('original_author', 'co_author', 'ghostwriter');

ALTER TYPE author_role OWNER TO system_dev;

CREATE TYPE artist_role AS ENUM ('cover_artist', 'illustrator', 'character_designer');

ALTER TYPE artist_role OWNER TO system_dev;

CREATE TYPE translator_role AS ENUM ('translator', 'localizer', 'adapter');

ALTER TYPE translator_role OWNER TO system_dev;

CREATE TYPE assignment_status AS ENUM ('active', 'inactive', 'suspended');

ALTER TYPE assignment_status OWNER TO system_dev;

CREATE TABLE IF NOT EXISTS novels
(
    id                 uuid                     DEFAULT uuidv7( )                     NOT NULL PRIMARY KEY,
    title              VARCHAR(500)                                                   NOT NULL,
    slug               VARCHAR(500)                                                   NOT NULL UNIQUE,
    owner_type         VARCHAR(20)                                                    NOT NULL
        CONSTRAINT novels_owner_type_check CHECK ((owner_type)::TEXT = ANY
                                                  ((ARRAY ['user'::CHARACTER VARYING, 'tenant'::CHARACTER VARYING])::TEXT[])),
    owner_id           uuid                                                           NOT NULL,
    created_by         uuid                                                           NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by         uuid                                                           REFERENCES identify.users ON DELETE SET NULL,
    version            INTEGER                  DEFAULT 1                             NOT NULL,
    synopsis           jsonb,
    cover_image_url    VARCHAR(1000),
    thumbnail_url      VARCHAR(1000),
    original_language  VARCHAR(10)                                                    NOT NULL,
    original_title     VARCHAR(500),
    status             catalog.novel_status     DEFAULT 'draft'::catalog.novel_status NOT NULL,
    total_volumes      INTEGER                  DEFAULT 0                             NOT NULL,
    total_chapters     INTEGER                  DEFAULT 0                             NOT NULL,
    total_words        BIGINT                   DEFAULT 0                             NOT NULL,
    view_count         BIGINT                   DEFAULT 0                             NOT NULL,
    favorite_count     INTEGER                  DEFAULT 0                             NOT NULL,
    rating_average     NUMERIC(3, 2)            DEFAULT 0.00
        CONSTRAINT novels_rating_average_check CHECK ((rating_average >= (0)::NUMERIC) AND (rating_average <= (5)::NUMERIC)),
    rating_count       INTEGER                  DEFAULT 0                             NOT NULL,
    metadata           jsonb                    DEFAULT '{}'::jsonb,
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at    TIMESTAMP WITH TIME ZONE,
    completed_at       TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                        NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                        NOT NULL,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    deleted_by         uuid                                                           REFERENCES identify.users ON DELETE SET NULL,
    is_oneshot         BOOLEAN                  DEFAULT FALSE                         NOT NULL
);

COMMENT ON TABLE novels IS 'Top-level table storing novel information with polymorphic ownership';

COMMENT ON COLUMN novels.owner_type IS 'Polymorphic owner type: user or tenant';

COMMENT ON COLUMN novels.owner_id IS 'Reference to users.id OR tenants.id based on owner_type (validated in application)';

COMMENT ON COLUMN novels.created_by IS 'User who created the novel (never changes)';

COMMENT ON COLUMN novels.updated_by IS 'User who last updated the novel';

COMMENT ON COLUMN novels.version IS 'Version number, auto-incremented on each update';

COMMENT ON COLUMN novels.synopsis IS 'Synopsis in ORIGINAL language only. Translations go to novel_synopsis_translations';

ALTER TABLE novels
    OWNER TO system_dev;

CREATE INDEX idx_novels_owner ON novels ( owner_type, owner_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_created_by ON novels ( created_by ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_status ON novels ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_original_language ON novels ( original_language ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_slug ON novels ( slug );

CREATE INDEX idx_novels_synopsis ON novels USING gin ( synopsis ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_metadata ON novels USING gin ( metadata ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_created_at ON novels ( created_at DESC ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_rating ON novels ( rating_average DESC ) WHERE ((status = 'ongoing'::catalog.novel_status) AND (deleted_at IS NULL));

CREATE INDEX idx_novels_views ON novels ( view_count DESC ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novels_title_trgm ON novels USING gin ( title public.gin_trgm_ops );

CREATE INDEX idx_novels_original_title_trgm ON novels USING gin ( original_title public.gin_trgm_ops );

CREATE TABLE IF NOT EXISTS novel_volumes
(
    id              uuid                     DEFAULT uuidv7( ) NOT NULL
        CONSTRAINT volumes_pkey PRIMARY KEY,
    novel_id        uuid                                       NOT NULL
        CONSTRAINT volumes_novel_id_fkey REFERENCES novels ON DELETE CASCADE,
    volume_number   INTEGER                                    NOT NULL
        CONSTRAINT volumes_volume_number_check CHECK (volume_number > 0),
    title           VARCHAR(500)                               NOT NULL,
    slug            VARCHAR(500)                               NOT NULL,
    description     TEXT,
    cover_image_url VARCHAR(1000),
    created_by      uuid                                       NOT NULL
        CONSTRAINT volumes_created_by_fkey REFERENCES identify.users ON DELETE RESTRICT,
    updated_by      uuid
        CONSTRAINT volumes_updated_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    version         INTEGER                  DEFAULT 1         NOT NULL,
    chapter_count   INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT volumes_chapter_count_check CHECK (chapter_count >= 0),
    word_count      BIGINT                   DEFAULT 0         NOT NULL
        CONSTRAINT volumes_word_count_check CHECK (word_count >= 0),
    display_order   INTEGER                                    NOT NULL,
    is_published    BOOLEAN                  DEFAULT FALSE     NOT NULL,
    published_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    deleted_at      TIMESTAMP WITH TIME ZONE,
    deleted_by      uuid
        CONSTRAINT volumes_deleted_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT volumes_novel_id_volume_number_key UNIQUE ( novel_id, volume_number ),
    CONSTRAINT volumes_novel_id_slug_key UNIQUE ( novel_id, slug )
);

COMMENT ON TABLE novel_volumes IS 'Middle-level table organizing chapters into volumes';

COMMENT ON COLUMN novel_volumes.chapter_count IS 'Auto-updated by application when chapters change';

ALTER TABLE novel_volumes
    OWNER TO system_dev;

CREATE INDEX idx_volumes_novel_id ON novel_volumes ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_volumes_display_order ON novel_volumes ( novel_id, display_order ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_volumes_published ON novel_volumes ( novel_id ASC, published_at DESC ) WHERE ((is_published = TRUE) AND (deleted_at IS NULL));

CREATE TABLE IF NOT EXISTS novel_chapters
(
    id              uuid                     DEFAULT uuidv7( )                       NOT NULL
        CONSTRAINT chapters_pkey PRIMARY KEY,
    novel_id        uuid                                                             NOT NULL
        CONSTRAINT chapters_novel_id_fkey REFERENCES novels ON DELETE CASCADE,
    volume_id       uuid
        CONSTRAINT chapters_volume_id_fkey REFERENCES novel_volumes ON DELETE SET NULL,
    chapter_number  INTEGER                                                          NOT NULL
        CONSTRAINT chapters_chapter_number_check CHECK (chapter_number > 0),
    title           VARCHAR(500)                                                     NOT NULL,
    slug            VARCHAR(500)                                                     NOT NULL,
    content         jsonb                                                            NOT NULL,
    created_by      uuid                                                             NOT NULL
        CONSTRAINT chapters_created_by_fkey REFERENCES identify.users ON DELETE RESTRICT,
    updated_by      uuid
        CONSTRAINT chapters_updated_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    version         INTEGER                  DEFAULT 1                               NOT NULL,
    word_count      INTEGER                  DEFAULT 0                               NOT NULL
        CONSTRAINT chapters_word_count_check CHECK (word_count >= 0),
    character_count INTEGER                  DEFAULT 0                               NOT NULL
        CONSTRAINT chapters_character_count_check CHECK (character_count >= 0),
    is_free         BOOLEAN                  DEFAULT TRUE                            NOT NULL,
    price           NUMERIC(10, 2)           DEFAULT 0.00
        CONSTRAINT chapters_price_check CHECK (price >= (0)::NUMERIC),
    currency        VARCHAR(3)               DEFAULT 'VND'::CHARACTER VARYING,
    status          catalog.chapter_status   DEFAULT 'draft'::catalog.chapter_status NOT NULL,
    view_count      BIGINT                   DEFAULT 0                               NOT NULL,
    like_count      INTEGER                  DEFAULT 0                               NOT NULL,
    comment_count   INTEGER                  DEFAULT 0                               NOT NULL,
    display_order   INTEGER                                                          NOT NULL,
    author_notes    jsonb,
    published_at    TIMESTAMP WITH TIME ZONE,
    scheduled_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                          NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                          NOT NULL,
    deleted_at      TIMESTAMP WITH TIME ZONE,
    deleted_by      uuid
        CONSTRAINT chapters_deleted_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT chapters_novel_id_chapter_number_key UNIQUE ( novel_id, chapter_number ),
    CONSTRAINT chapters_novel_id_slug_key UNIQUE ( novel_id, slug )
);

COMMENT ON TABLE novel_chapters IS 'Bottom-level table storing chapter content';

COMMENT ON COLUMN novel_chapters.volume_id IS 'Nullable - chapters can exist without belonging to a volume';

COMMENT ON COLUMN novel_chapters.content IS 'Chapter content in ORIGINAL language only. Translations go to chapter_translations';

ALTER TABLE novel_chapters
    OWNER TO system_dev;

CREATE INDEX idx_chapters_novel_id ON novel_chapters ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapters_volume_id ON novel_chapters ( volume_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapters_status ON novel_chapters ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapters_display_order ON novel_chapters ( novel_id, display_order ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapters_published ON novel_chapters ( published_at DESC ) WHERE (
    (status = 'published'::catalog.chapter_status) AND (deleted_at IS NULL));

CREATE INDEX idx_chapters_scheduled ON novel_chapters ( scheduled_at ) WHERE (
    (status = 'scheduled'::catalog.chapter_status) AND (deleted_at IS NULL));

CREATE TABLE IF NOT EXISTS ownership_transfers
(
    id                uuid                     DEFAULT uuidv7( )                          NOT NULL PRIMARY KEY,
    novel_id          uuid                                                                NOT NULL REFERENCES novels ON DELETE CASCADE,
    from_owner_type   VARCHAR(20)                                                         NOT NULL
        CONSTRAINT ownership_transfers_from_owner_type_check CHECK ((from_owner_type)::TEXT = ANY
                                                                    ((ARRAY ['user'::CHARACTER VARYING, 'organization'::CHARACTER VARYING])::TEXT[])),
    from_owner_id     uuid                                                                NOT NULL,
    to_owner_type     VARCHAR(20)                                                         NOT NULL
        CONSTRAINT ownership_transfers_to_owner_type_check CHECK ((to_owner_type)::TEXT = ANY
                                                                  ((ARRAY ['user'::CHARACTER VARYING, 'organization'::CHARACTER VARYING])::TEXT[])),
    to_owner_id       uuid                                                                NOT NULL,
    status            catalog.transfer_status  DEFAULT 'pending'::catalog.transfer_status NOT NULL,
    reason            TEXT,
    requires_approval BOOLEAN                  DEFAULT FALSE                              NOT NULL,
    reviewed_by       uuid                                                                REFERENCES identify.users ON DELETE SET NULL,
    review_notes      TEXT,
    reviewed_at       TIMESTAMP WITH TIME ZONE,
    created_by        uuid                                                                NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by        uuid                                                                REFERENCES identify.users ON DELETE SET NULL,
    version           INTEGER                  DEFAULT 1                                  NOT NULL,
    requested_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                             NOT NULL,
    completed_at      TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                             NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                             NOT NULL,
    deleted_at        TIMESTAMP WITH TIME ZONE,
    deleted_by        uuid                                                                REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT ownership_transfers_check CHECK (((from_owner_type)::TEXT <> (to_owner_type)::TEXT) OR
                                                (from_owner_id <> to_owner_id))
);

COMMENT ON TABLE ownership_transfers IS 'Tracks ownership transfer requests between users and organizations';

COMMENT ON COLUMN ownership_transfers.from_owner_type IS 'Source owner type: user or organization';

COMMENT ON COLUMN ownership_transfers.from_owner_id IS 'Source owner UUID (validated in application)';

COMMENT ON COLUMN ownership_transfers.to_owner_type IS 'Target owner type: user or organization';

COMMENT ON COLUMN ownership_transfers.to_owner_id IS 'Target owner UUID (validated in application)';

COMMENT ON COLUMN ownership_transfers.requires_approval IS 'TRUE for 2-way transfers (user<->organization or organization<->organization), FALSE for 1-way (user->organization)';

ALTER TABLE ownership_transfers
    OWNER TO system_dev;

CREATE INDEX idx_ownership_transfers_novel_id ON ownership_transfers ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_ownership_transfers_from ON ownership_transfers ( from_owner_type, from_owner_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_ownership_transfers_to ON ownership_transfers ( to_owner_type, to_owner_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_ownership_transfers_status ON ownership_transfers ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_ownership_transfers_pending_approval ON ownership_transfers ( status, requires_approval ) WHERE (
    (requires_approval = TRUE) AND (status = 'pending'::catalog.transfer_status) AND (deleted_at IS NULL));

CREATE TABLE IF NOT EXISTS exclusive_translation_reports
(
    id                        uuid                     DEFAULT uuidv7( )                        NOT NULL PRIMARY KEY,
    novel_id                  uuid                                                              NOT NULL REFERENCES novels ON DELETE CASCADE,
    language                  VARCHAR(10)                                                       NOT NULL,
    reporting_organization_id uuid                                                              NOT NULL,
    reported_organization_id  uuid                                                              NOT NULL,
    reason                    TEXT                                                              NOT NULL,
    evidence_urls             TEXT[],
    status                    catalog.report_status    DEFAULT 'pending'::catalog.report_status NOT NULL,
    reviewed_by               uuid                                                              REFERENCES identify.users ON DELETE SET NULL,
    review_notes              TEXT,
    reviewed_at               TIMESTAMP WITH TIME ZONE,
    resolution                TEXT,
    resolved_at               TIMESTAMP WITH TIME ZONE,
    created_by                uuid                                                              NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by                uuid                                                              REFERENCES identify.users ON DELETE SET NULL,
    version                   INTEGER                  DEFAULT 1                                NOT NULL,
    created_at                TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                           NOT NULL,
    updated_at                TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                           NOT NULL,
    deleted_at                TIMESTAMP WITH TIME ZONE,
    deleted_by                uuid                                                              REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT exclusive_translation_reports_check CHECK (reporting_organization_id <> reported_organization_id)
);

COMMENT ON TABLE exclusive_translation_reports IS 'Reports for organizations claiming exclusive translation rights';

COMMENT ON COLUMN exclusive_translation_reports.reporting_organization_id IS 'Organization filing the report';

COMMENT ON COLUMN exclusive_translation_reports.reported_organization_id IS 'Organization being reported for claiming exclusive rights';

COMMENT ON COLUMN exclusive_translation_reports.evidence_urls IS 'URLs to evidence supporting the report';

ALTER TABLE exclusive_translation_reports
    OWNER TO system_dev;

CREATE INDEX idx_exclusive_reports_novel_id ON exclusive_translation_reports ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_exclusive_reports_language ON exclusive_translation_reports ( language ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_exclusive_reports_reporting_organization ON exclusive_translation_reports ( reporting_organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_exclusive_reports_reported_organization ON exclusive_translation_reports ( reported_organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_exclusive_reports_status ON exclusive_translation_reports ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_exclusive_reports_pending ON exclusive_translation_reports ( status ) WHERE ((status = 'pending'::catalog.report_status) AND (deleted_at IS NULL));

CREATE TABLE IF NOT EXISTS novel_synopsis_translations
(
    id                 uuid                       DEFAULT uuidv7( )                           NOT NULL PRIMARY KEY,
    novel_id           uuid                                                                   NOT NULL REFERENCES novels ON DELETE CASCADE,
    language           VARCHAR(10)                                                            NOT NULL,
    synopsis           jsonb                                                                  NOT NULL,
    organization_id    uuid                                                                   REFERENCES identify.organizations ON DELETE SET NULL,
    status             catalog.translation_status DEFAULT 'draft'::catalog.translation_status NOT NULL,
    quality_score      NUMERIC(3, 2)              DEFAULT 0.00
        CONSTRAINT novel_synopsis_translations_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    reviewer_rating    NUMERIC(3, 2)              DEFAULT 0.00
        CONSTRAINT novel_synopsis_translations_reviewer_rating_check CHECK ((reviewer_rating >= (0)::NUMERIC) AND
                                                                            (reviewer_rating <= (5)::NUMERIC)),
    contribution_count INTEGER                    DEFAULT 0                                   NOT NULL
        CONSTRAINT novel_synopsis_translations_contribution_count_check CHECK (contribution_count >= 0),
    view_count         BIGINT                     DEFAULT 0                                   NOT NULL,
    like_count         INTEGER                    DEFAULT 0                                   NOT NULL,
    reviewed_by        uuid                                                                   REFERENCES identify.users ON DELETE SET NULL,
    review_notes       TEXT,
    reviewed_at        TIMESTAMP WITH TIME ZONE,
    created_by         uuid                                                                   NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by         uuid                                                                   REFERENCES identify.users ON DELETE SET NULL,
    version            INTEGER                    DEFAULT 1                                   NOT NULL,
    published_at       TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE   DEFAULT NOW( )                              NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE   DEFAULT NOW( )                              NOT NULL,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    deleted_by         uuid                                                                   REFERENCES identify.users ON DELETE SET NULL,
    UNIQUE ( novel_id, language )
);

COMMENT ON TABLE novel_synopsis_translations IS 'Translated synopsis for novels';

COMMENT ON COLUMN novel_synopsis_translations.synopsis IS 'Translated synopsis content in JSONB format';

COMMENT ON COLUMN novel_synopsis_translations.organization_id IS 'Optional organization responsible for this translation';

COMMENT ON COLUMN novel_synopsis_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';

COMMENT ON COLUMN novel_synopsis_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

ALTER TABLE novel_synopsis_translations
    OWNER TO system_dev;

CREATE INDEX idx_synopsis_translations_novel_id ON novel_synopsis_translations ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_translations_language ON novel_synopsis_translations ( language ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_translations_organization_id ON novel_synopsis_translations ( organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_translations_status ON novel_synopsis_translations ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_translations_published ON novel_synopsis_translations ( novel_id ASC, language ASC, published_at DESC ) WHERE (
    (status = 'published'::catalog.translation_status) AND (deleted_at IS NULL));

CREATE INDEX idx_synopsis_translations_content ON novel_synopsis_translations USING gin ( synopsis ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS synopsis_translation_contributions
(
    id                      uuid                        DEFAULT uuidv7( )                              NOT NULL PRIMARY KEY,
    synopsis_translation_id uuid                                                                       NOT NULL REFERENCES novel_synopsis_translations ON DELETE CASCADE,
    contributor_id          uuid                                                                       NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    contribution_type       VARCHAR(50)                                                                NOT NULL,
    contribution_notes      TEXT,
    status                  catalog.contribution_status DEFAULT 'pending'::catalog.contribution_status NOT NULL,
    changes                 jsonb,
    quality_score           NUMERIC(3, 2)               DEFAULT 0.00
        CONSTRAINT synopsis_translation_contributions_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    reviewed_by             uuid                                                                       REFERENCES identify.users ON DELETE SET NULL,
    review_notes            TEXT,
    reviewed_at             TIMESTAMP WITH TIME ZONE,
    created_by              uuid                                                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by              uuid                                                                       REFERENCES identify.users ON DELETE SET NULL,
    version                 INTEGER                     DEFAULT 1                                      NOT NULL,
    contributed_at          TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    created_at              TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    updated_at              TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    deleted_at              TIMESTAMP WITH TIME ZONE,
    deleted_by              uuid                                                                       REFERENCES identify.users ON DELETE SET NULL
);

COMMENT ON TABLE synopsis_translation_contributions IS 'Tracks individual contributions to synopsis translations';

COMMENT ON COLUMN synopsis_translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review';

COMMENT ON COLUMN synopsis_translation_contributions.changes IS 'JSONB documenting what was changed in this contribution';

ALTER TABLE synopsis_translation_contributions
    OWNER TO system_dev;

CREATE INDEX idx_synopsis_contributions_translation ON synopsis_translation_contributions ( synopsis_translation_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_contributions_contributor ON synopsis_translation_contributions ( contributor_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_contributions_type ON synopsis_translation_contributions ( contribution_type ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_contributions_status ON synopsis_translation_contributions ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_synopsis_contributions_contributed ON synopsis_translation_contributions ( contributed_at DESC ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS novel_chapter_translations
(
    id                 uuid                       DEFAULT uuidv7( )                           NOT NULL
        CONSTRAINT chapter_translations_pkey PRIMARY KEY,
    chapter_id         uuid                                                                   NOT NULL
        CONSTRAINT chapter_translations_chapter_id_fkey REFERENCES novel_chapters ON DELETE CASCADE,
    language           VARCHAR(10)                                                            NOT NULL,
    content            jsonb                                                                  NOT NULL,
    title              VARCHAR(500)                                                           NOT NULL,
    organization_id    uuid
        CONSTRAINT chapter_translations_organization_id_fkey REFERENCES identify.organizations ON DELETE SET NULL,
    status             catalog.translation_status DEFAULT 'draft'::catalog.translation_status NOT NULL,
    quality_score      NUMERIC(3, 2)              DEFAULT 0.00
        CONSTRAINT chapter_translations_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    reviewer_rating    NUMERIC(3, 2)              DEFAULT 0.00
        CONSTRAINT chapter_translations_reviewer_rating_check CHECK ((reviewer_rating >= (0)::NUMERIC) AND
                                                                     (reviewer_rating >= (0)::NUMERIC)),
    word_count         INTEGER                    DEFAULT 0                                   NOT NULL
        CONSTRAINT chapter_translations_word_count_check CHECK (word_count >= 0),
    character_count    INTEGER                    DEFAULT 0                                   NOT NULL
        CONSTRAINT chapter_translations_character_count_check CHECK (character_count >= 0),
    contribution_count INTEGER                    DEFAULT 0                                   NOT NULL
        CONSTRAINT chapter_translations_contribution_count_check CHECK (contribution_count >= 0),
    view_count         BIGINT                     DEFAULT 0                                   NOT NULL,
    like_count         INTEGER                    DEFAULT 0                                   NOT NULL,
    comment_count      INTEGER                    DEFAULT 0                                   NOT NULL,
    reviewed_by        uuid
        CONSTRAINT chapter_translations_reviewed_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    review_notes       TEXT,
    reviewed_at        TIMESTAMP WITH TIME ZONE,
    translator_notes   jsonb,
    created_by         uuid                                                                   NOT NULL
        CONSTRAINT chapter_translations_created_by_fkey REFERENCES identify.users ON DELETE RESTRICT,
    updated_by         uuid
        CONSTRAINT chapter_translations_updated_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    version            INTEGER                    DEFAULT 1                                   NOT NULL,
    published_at       TIMESTAMP WITH TIME ZONE,
    created_at         TIMESTAMP WITH TIME ZONE   DEFAULT NOW( )                              NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE   DEFAULT NOW( )                              NOT NULL,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    deleted_by         uuid
        CONSTRAINT chapter_translations_deleted_by_fkey REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT chapter_translations_chapter_id_language_key UNIQUE ( chapter_id, language )
);

COMMENT ON TABLE novel_chapter_translations IS 'Translated chapter content';

COMMENT ON COLUMN novel_chapter_translations.content IS 'Translated chapter content in JSONB format';

COMMENT ON COLUMN novel_chapter_translations.organization_id IS 'Optional organization responsible for this translation';

COMMENT ON COLUMN novel_chapter_translations.quality_score IS 'Aggregate quality score from community (0-5 scale)';

COMMENT ON COLUMN novel_chapter_translations.reviewer_rating IS 'Quality rating from official reviewers (0-5 scale)';

ALTER TABLE novel_chapter_translations
    OWNER TO system_dev;

CREATE INDEX idx_chapter_translations_chapter_id ON novel_chapter_translations ( chapter_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapter_translations_language ON novel_chapter_translations ( language ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapter_translations_organization_id ON novel_chapter_translations ( organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapter_translations_status ON novel_chapter_translations ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapter_translations_published ON novel_chapter_translations ( chapter_id ASC, language ASC, published_at DESC ) WHERE (
    (status = 'published'::catalog.translation_status) AND (deleted_at IS NULL));

CREATE INDEX idx_chapter_translations_content ON novel_chapter_translations USING gin ( content ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_chapter_translations_views ON novel_chapter_translations ( view_count DESC ) WHERE (
    (status = 'published'::catalog.translation_status) AND (deleted_at IS NULL));

CREATE TABLE IF NOT EXISTS translation_contributions
(
    id                     uuid                        DEFAULT uuidv7( )                              NOT NULL PRIMARY KEY,
    chapter_translation_id uuid                                                                       NOT NULL REFERENCES novel_chapter_translations ON DELETE CASCADE,
    contributor_id         uuid                                                                       NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    contribution_type      VARCHAR(50)                                                                NOT NULL,
    contribution_notes     TEXT,
    status                 catalog.contribution_status DEFAULT 'pending'::catalog.contribution_status NOT NULL,
    changes                jsonb,
    quality_score          NUMERIC(3, 2)               DEFAULT 0.00
        CONSTRAINT translation_contributions_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    reviewed_by            uuid                                                                       REFERENCES identify.users ON DELETE SET NULL,
    review_notes           TEXT,
    reviewed_at            TIMESTAMP WITH TIME ZONE,
    created_by             uuid                                                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by             uuid                                                                       REFERENCES identify.users ON DELETE SET NULL,
    version                INTEGER                     DEFAULT 1                                      NOT NULL,
    contributed_at         TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    created_at             TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    updated_at             TIMESTAMP WITH TIME ZONE    DEFAULT NOW( )                                 NOT NULL,
    deleted_at             TIMESTAMP WITH TIME ZONE,
    deleted_by             uuid                                                                       REFERENCES identify.users ON DELETE SET NULL
);

COMMENT ON TABLE translation_contributions IS 'Tracks individual contributions to chapter translations';

COMMENT ON COLUMN translation_contributions.contribution_type IS 'Type of contribution: translation, proofread, edit, review, typeset';

COMMENT ON COLUMN translation_contributions.changes IS 'JSONB documenting metadata about changes (NOT full content to save space)';

ALTER TABLE translation_contributions
    OWNER TO system_dev;

CREATE INDEX idx_translation_contributions_chapter ON translation_contributions ( chapter_translation_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translation_contributions_contributor ON translation_contributions ( contributor_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translation_contributions_type ON translation_contributions ( contribution_type ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translation_contributions_status ON translation_contributions ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translation_contributions_contributed ON translation_contributions ( contributed_at DESC ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS translation_history
(
    id                     uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    chapter_translation_id uuid                                       NOT NULL REFERENCES novel_chapter_translations ON DELETE CASCADE,
    version_number         INTEGER                                    NOT NULL
        CONSTRAINT translation_history_version_number_check CHECK (version_number > 0),
    title                  VARCHAR(500),
    word_count             INTEGER,
    character_count        INTEGER,
    status                 catalog.translation_status,
    change_summary         TEXT,
    changed_fields         jsonb,
    changed_by             uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    UNIQUE ( chapter_translation_id, version_number )
);

COMMENT ON TABLE translation_history IS 'Version control for chapter translations (metadata only, NOT full content)';

COMMENT ON COLUMN translation_history.change_summary IS 'Human-readable summary of what changed';

COMMENT ON COLUMN translation_history.changed_fields IS 'JSONB array of field names that changed in this version';

ALTER TABLE translation_history
    OWNER TO system_dev;

CREATE INDEX idx_translation_history_chapter ON translation_history ( chapter_translation_id );

CREATE INDEX idx_translation_history_version ON translation_history ( chapter_translation_id ASC, version_number DESC );

CREATE INDEX idx_translation_history_changed_by ON translation_history ( changed_by );

CREATE INDEX idx_translation_history_created ON translation_history ( created_at DESC );

CREATE TABLE IF NOT EXISTS novel_history
(
    id             uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    novel_id       uuid                                       NOT NULL REFERENCES novels ON DELETE CASCADE,
    version_number INTEGER                                    NOT NULL
        CONSTRAINT novel_history_version_number_check CHECK (version_number > 0),
    action         catalog.audit_action                       NOT NULL,
    title          VARCHAR(500),
    slug           VARCHAR(500),
    status         catalog.novel_status,
    owner_type     VARCHAR(20),
    owner_id       uuid,
    total_volumes  INTEGER,
    total_chapters INTEGER,
    total_words    BIGINT,
    changed_fields jsonb,
    change_summary TEXT,
    changed_by     uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    request_id     VARCHAR(100),
    ip_address     inet,
    user_agent     TEXT,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    UNIQUE ( novel_id, version_number )
);

COMMENT ON TABLE novel_history IS 'Audit log for novel changes (metadata only, logged at application layer)';

COMMENT ON COLUMN novel_history.changed_fields IS 'JSONB array of field names that changed';

COMMENT ON COLUMN novel_history.change_summary IS 'Human-readable description of changes';

COMMENT ON COLUMN novel_history.request_id IS 'Request ID for tracing';

ALTER TABLE novel_history
    OWNER TO system_dev;

CREATE INDEX idx_novel_history_novel ON novel_history ( novel_id );

CREATE INDEX idx_novel_history_version ON novel_history ( novel_id ASC, version_number DESC );

CREATE INDEX idx_novel_history_action ON novel_history ( action );

CREATE INDEX idx_novel_history_changed_by ON novel_history ( changed_by );

CREATE INDEX idx_novel_history_created ON novel_history ( created_at DESC );

CREATE INDEX idx_novel_history_request ON novel_history ( request_id ) WHERE (request_id IS NOT NULL);

CREATE TABLE IF NOT EXISTS novel_volume_histories
(
    id             uuid                     DEFAULT uuidv7( ) NOT NULL
        CONSTRAINT volume_history_pkey PRIMARY KEY,
    volume_id      uuid                                       NOT NULL
        CONSTRAINT volume_history_volume_id_fkey REFERENCES novel_volumes ON DELETE CASCADE,
    novel_id       uuid                                       NOT NULL
        CONSTRAINT volume_history_novel_id_fkey REFERENCES novels ON DELETE CASCADE,
    version_number INTEGER                                    NOT NULL
        CONSTRAINT volume_history_version_number_check CHECK (version_number > 0),
    action         catalog.audit_action                       NOT NULL,
    title          VARCHAR(500),
    slug           VARCHAR(500),
    volume_number  INTEGER,
    is_published   BOOLEAN,
    chapter_count  INTEGER,
    word_count     BIGINT,
    changed_fields jsonb,
    change_summary TEXT,
    changed_by     uuid                                       NOT NULL
        CONSTRAINT volume_history_changed_by_fkey REFERENCES identify.users ON DELETE RESTRICT,
    request_id     VARCHAR(100),
    ip_address     inet,
    user_agent     TEXT,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    CONSTRAINT volume_history_volume_id_version_number_key UNIQUE ( volume_id, version_number )
);

COMMENT ON TABLE novel_volume_histories IS 'Audit log for volume changes (metadata only, logged at application layer)';

COMMENT ON COLUMN novel_volume_histories.changed_fields IS 'JSONB array of field names that changed';

ALTER TABLE novel_volume_histories
    OWNER TO system_dev;

CREATE INDEX idx_volume_history_volume ON novel_volume_histories ( volume_id );

CREATE INDEX idx_volume_history_novel ON novel_volume_histories ( novel_id );

CREATE INDEX idx_volume_history_version ON novel_volume_histories ( volume_id ASC, version_number DESC );

CREATE INDEX idx_volume_history_action ON novel_volume_histories ( action );

CREATE INDEX idx_volume_history_changed_by ON novel_volume_histories ( changed_by );

CREATE INDEX idx_volume_history_created ON novel_volume_histories ( created_at DESC );

CREATE TABLE IF NOT EXISTS novel_chapter_histories
(
    id              uuid                     DEFAULT uuidv7( ) NOT NULL
        CONSTRAINT chapter_history_pkey PRIMARY KEY,
    chapter_id      uuid                                       NOT NULL
        CONSTRAINT chapter_history_chapter_id_fkey REFERENCES novel_chapters ON DELETE CASCADE,
    volume_id       uuid
        CONSTRAINT chapter_history_volume_id_fkey REFERENCES novel_volumes ON DELETE SET NULL,
    novel_id        uuid                                       NOT NULL
        CONSTRAINT chapter_history_novel_id_fkey REFERENCES novels ON DELETE CASCADE,
    version_number  INTEGER                                    NOT NULL
        CONSTRAINT chapter_history_version_number_check CHECK (version_number > 0),
    action          catalog.audit_action                       NOT NULL,
    title           VARCHAR(500),
    slug            VARCHAR(500),
    chapter_number  INTEGER,
    status          catalog.chapter_status,
    word_count      INTEGER,
    character_count INTEGER,
    changed_fields  jsonb,
    change_summary  TEXT,
    content_changed BOOLEAN                  DEFAULT FALSE,
    changed_by      uuid                                       NOT NULL
        CONSTRAINT chapter_history_changed_by_fkey REFERENCES identify.users ON DELETE RESTRICT,
    request_id      VARCHAR(100),
    ip_address      inet,
    user_agent      TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    CONSTRAINT chapter_history_chapter_id_version_number_key UNIQUE ( chapter_id, version_number )
);

COMMENT ON TABLE novel_chapter_histories IS 'Audit log for chapter changes (metadata only, logged at application layer)';

COMMENT ON COLUMN novel_chapter_histories.changed_fields IS 'JSONB array of field names that changed';

COMMENT ON COLUMN novel_chapter_histories.change_summary IS 'Human-readable description like "Updated chapter content and title"';

COMMENT ON COLUMN novel_chapter_histories.content_changed IS 'Flag indicating if content JSONB changed (actual content NOT stored here to save space)';

ALTER TABLE novel_chapter_histories
    OWNER TO system_dev;

CREATE INDEX idx_chapter_history_chapter ON novel_chapter_histories ( chapter_id );

CREATE INDEX idx_chapter_history_volume ON novel_chapter_histories ( volume_id );

CREATE INDEX idx_chapter_history_novel ON novel_chapter_histories ( novel_id );

CREATE INDEX idx_chapter_history_version ON novel_chapter_histories ( chapter_id ASC, version_number DESC );

CREATE INDEX idx_chapter_history_action ON novel_chapter_histories ( action );

CREATE INDEX idx_chapter_history_changed_by ON novel_chapter_histories ( changed_by );

CREATE INDEX idx_chapter_history_created ON novel_chapter_histories ( created_at DESC );

CREATE INDEX idx_chapter_history_content_changed ON novel_chapter_histories ( chapter_id, content_changed ) WHERE (content_changed = TRUE);

CREATE TABLE IF NOT EXISTS genres
(
    id             uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name           VARCHAR(100)                               NOT NULL,
    slug           VARCHAR(100)                               NOT NULL UNIQUE,
    description    TEXT,
    parent_id      uuid                                       REFERENCES genres ON DELETE SET NULL,
    novel_count    INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT genres_novel_count_check CHECK (novel_count >= 0),
    active_readers BIGINT                   DEFAULT 0         NOT NULL
        CONSTRAINT genres_active_readers_check CHECK (active_readers >= 0),
    total_views    BIGINT                   DEFAULT 0         NOT NULL
        CONSTRAINT genres_total_views_check CHECK (total_views >= 0),
    created_by     uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    version        INTEGER                  DEFAULT 1         NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    deleted_at     TIMESTAMP WITH TIME ZONE,
    deleted_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    is_active      BOOLEAN                  DEFAULT TRUE      NOT NULL,
    anime_count    INTEGER                  DEFAULT 0         NOT NULL,
    manga_count    INTEGER                  DEFAULT 0         NOT NULL
);

COMMENT ON TABLE genres IS 'Genre definitions for novels';

COMMENT ON COLUMN genres.parent_id IS 'Optional parent genre for hierarchical organization';

COMMENT ON COLUMN genres.is_active IS 'Whether the genre is active and visible to users';

ALTER TABLE genres
    OWNER TO system_dev;

CREATE INDEX idx_genres_slug ON genres ( slug ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_genres_parent ON genres ( parent_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_genres_is_active ON genres ( is_active ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS novel_genres
(
    id            uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    novel_id      uuid                                       NOT NULL REFERENCES novels ON DELETE CASCADE,
    genre_id      uuid                                       NOT NULL REFERENCES genres ON DELETE CASCADE,
    display_order INTEGER                  DEFAULT 0         NOT NULL,
    created_by    uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    UNIQUE ( novel_id, genre_id )
);

COMMENT ON TABLE novel_genres IS 'Junction table linking novels to genres';

ALTER TABLE novel_genres
    OWNER TO system_dev;

CREATE INDEX idx_novel_genres_novel ON novel_genres ( novel_id );

CREATE INDEX idx_novel_genres_genre ON novel_genres ( genre_id );

CREATE TABLE IF NOT EXISTS authors
(
    id             uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name           VARCHAR(200)                               NOT NULL,
    slug           VARCHAR(200)                               NOT NULL UNIQUE,
    biography      TEXT,
    avatar_url     VARCHAR(1000),
    user_id        uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    novel_count    INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT authors_novel_count_check CHECK (novel_count >= 0),
    total_views    BIGINT                   DEFAULT 0         NOT NULL,
    metadata       jsonb                    DEFAULT '{}'::jsonb,
    created_by     uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    version        INTEGER                  DEFAULT 1         NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    deleted_at     TIMESTAMP WITH TIME ZONE,
    deleted_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    total_chapters INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT authors_total_chapters_check CHECK (total_chapters >= 0),
    follower_count INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT authors_follower_count_check CHECK (follower_count >= 0),
    is_verified    BOOLEAN                  DEFAULT FALSE     NOT NULL,
    social_links   jsonb                    DEFAULT '{}'::jsonb
);

COMMENT ON TABLE authors IS 'Author information';

COMMENT ON COLUMN authors.biography IS 'Author biography in JSONB format';

COMMENT ON COLUMN authors.user_id IS 'Optional link to user account if author is registered';

COMMENT ON COLUMN authors.total_chapters IS 'Total number of chapters written by this author across all novels';

COMMENT ON COLUMN authors.follower_count IS 'Number of followers for this author';

COMMENT ON COLUMN authors.is_verified IS 'Whether the author has been verified by the platform';

COMMENT ON COLUMN authors.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';

ALTER TABLE authors
    OWNER TO system_dev;

CREATE INDEX idx_authors_slug ON authors ( slug ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_authors_user_id ON authors ( user_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_authors_metadata ON authors USING gin ( metadata ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_authors_is_verified ON authors ( is_verified ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_authors_follower_count ON authors ( follower_count DESC ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS novel_authors
(
    id            uuid    DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    novel_id      uuid                      NOT NULL REFERENCES novels ON DELETE CASCADE,
    author_id     uuid                      NOT NULL REFERENCES authors ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0         NOT NULL,
    UNIQUE ( novel_id, author_id )
);

COMMENT ON TABLE novel_authors IS 'Junction table linking novels to authors';

ALTER TABLE novel_authors
    OWNER TO system_dev;

CREATE INDEX idx_novel_authors_novel ON novel_authors ( novel_id );

CREATE INDEX idx_novel_authors_author ON novel_authors ( author_id );

CREATE TABLE IF NOT EXISTS artists
(
    id             uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name           VARCHAR(200)                               NOT NULL,
    slug           VARCHAR(200)                               NOT NULL UNIQUE,
    biography      TEXT,
    avatar_url     VARCHAR(1000),
    portfolio_url  VARCHAR(1000),
    user_id        uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    novel_count    INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT artists_novel_count_check CHECK (novel_count >= 0),
    metadata       jsonb                    DEFAULT '{}'::jsonb,
    created_by     uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    version        INTEGER                  DEFAULT 1         NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    deleted_at     TIMESTAMP WITH TIME ZONE,
    deleted_by     uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    specialization VARCHAR(100),
    artwork_count  INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT artists_artwork_count_check CHECK (artwork_count >= 0),
    follower_count INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT artists_follower_count_check CHECK (follower_count >= 0),
    is_verified    BOOLEAN                  DEFAULT FALSE     NOT NULL,
    social_links   jsonb                    DEFAULT '{}'::jsonb
);

COMMENT ON TABLE artists IS 'Artist information (cover artists, illustrators, etc.)';

COMMENT ON COLUMN artists.biography IS 'Artist biography in JSONB format';

COMMENT ON COLUMN artists.specialization IS 'Artist specialization (e.g., cover_artist, illustrator, manga_artist)';

COMMENT ON COLUMN artists.artwork_count IS 'Total number of artworks created by this artist';

COMMENT ON COLUMN artists.follower_count IS 'Number of followers for this artist';

COMMENT ON COLUMN artists.is_verified IS 'Whether the artist has been verified by the platform';

COMMENT ON COLUMN artists.social_links IS 'Social media links in JSONB format (e.g., {"facebook": "...", "twitter": "..."})';

ALTER TABLE artists
    OWNER TO system_dev;

CREATE INDEX idx_artists_slug ON artists ( slug ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_artists_user_id ON artists ( user_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_artists_metadata ON artists USING gin ( metadata ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_artists_specialization ON artists ( specialization ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_artists_is_verified ON artists ( is_verified ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_artists_follower_count ON artists ( follower_count DESC ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS novel_artists
(
    id            uuid    DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    novel_id      uuid                      NOT NULL REFERENCES novels ON DELETE CASCADE,
    artist_id     uuid                      NOT NULL REFERENCES artists ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0         NOT NULL,
    UNIQUE ( novel_id, artist_id )
);

COMMENT ON TABLE novel_artists IS 'Junction table linking novels to artists';

ALTER TABLE novel_artists
    OWNER TO system_dev;

CREATE INDEX idx_novel_artists_novel ON novel_artists ( novel_id );

CREATE INDEX idx_novel_artists_artist ON novel_artists ( artist_id );

CREATE TABLE IF NOT EXISTS translators
(
    id                     uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name                   VARCHAR(200)                               NOT NULL,
    slug                   VARCHAR(200)                               NOT NULL UNIQUE,
    bio                    TEXT,
    avatar_url             VARCHAR(1000),
    user_id                uuid                                       NOT NULL REFERENCES identify.users ON DELETE CASCADE,
    source_languages       VARCHAR(10)[]                              NOT NULL,
    target_languages       VARCHAR(10)[]                              NOT NULL,
    translation_count      INTEGER                  DEFAULT 0         NOT NULL
        CONSTRAINT translators_translation_count_check CHECK (translation_count >= 0),
    total_words_translated BIGINT                   DEFAULT 0         NOT NULL,
    quality_score          NUMERIC(3, 2)            DEFAULT 0.00
        CONSTRAINT translators_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    metadata               jsonb                    DEFAULT '{}'::jsonb,
    created_by             uuid                                       NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by             uuid                                       REFERENCES identify.users ON DELETE SET NULL,
    version                INTEGER                  DEFAULT 1         NOT NULL,
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    deleted_at             TIMESTAMP WITH TIME ZONE,
    deleted_by             uuid                                       REFERENCES identify.users ON DELETE SET NULL
);

COMMENT ON TABLE translators IS 'Translator information';

COMMENT ON COLUMN translators.source_languages IS 'Array of languages the translator can translate from';

COMMENT ON COLUMN translators.target_languages IS 'Array of languages the translator can translate to';

ALTER TABLE translators
    OWNER TO system_dev;

CREATE INDEX idx_translators_slug ON translators ( slug ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translators_user_id ON translators ( user_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translators_source_langs ON translators USING gin ( source_languages ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translators_target_langs ON translators USING gin ( target_languages ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_translators_quality ON translators ( quality_score DESC ) WHERE (deleted_at IS NULL);

CREATE TABLE IF NOT EXISTS novel_translators
(
    id                  uuid                     DEFAULT uuidv7( )                             NOT NULL PRIMARY KEY,
    novel_id            uuid                                                                   NOT NULL REFERENCES novels ON DELETE CASCADE,
    translator_id       uuid                                                                   NOT NULL REFERENCES translators ON DELETE CASCADE,
    language            VARCHAR(10)                                                            NOT NULL,
    role                catalog.translator_role  DEFAULT 'translator'::catalog.translator_role NOT NULL,
    chapters_translated INTEGER                  DEFAULT 0                                     NOT NULL
        CONSTRAINT novel_translators_chapters_translated_check CHECK (chapters_translated >= 0),
    created_by          uuid                                                                   NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                                NOT NULL,
    UNIQUE ( novel_id, translator_id, language )
);

COMMENT ON TABLE novel_translators IS 'Junction table linking novels to translators';

COMMENT ON COLUMN novel_translators.language IS 'Target language for this translator on this novel';

ALTER TABLE novel_translators
    OWNER TO system_dev;

CREATE INDEX idx_novel_translators_novel ON novel_translators ( novel_id );

CREATE INDEX idx_novel_translators_translator ON novel_translators ( translator_id );

CREATE INDEX idx_novel_translators_language ON novel_translators ( novel_id, language );

CREATE TABLE IF NOT EXISTS novel_organization_assignments
(
    id                   uuid                      DEFAULT uuidv7( )                           NOT NULL PRIMARY KEY,
    novel_id             uuid                                                                  NOT NULL REFERENCES novels ON DELETE CASCADE,
    organization_id      uuid                                                                  NOT NULL REFERENCES identify.organizations ON DELETE CASCADE,
    language             VARCHAR(10)                                                           NOT NULL,
    status               catalog.assignment_status DEFAULT 'active'::catalog.assignment_status NOT NULL,
    has_exclusive_rights BOOLEAN                   DEFAULT FALSE                               NOT NULL,
    chapters_translated  INTEGER                   DEFAULT 0                                   NOT NULL
        CONSTRAINT novel_organization_assignments_chapters_translated_check CHECK (chapters_translated >= 0),
    chapters_proofread   INTEGER                   DEFAULT 0                                   NOT NULL
        CONSTRAINT novel_organization_assignments_chapters_proofread_check CHECK (chapters_proofread >= 0),
    metadata             jsonb                     DEFAULT '{}'::jsonb,
    created_by           uuid                                                                  NOT NULL REFERENCES identify.users ON DELETE RESTRICT,
    updated_by           uuid                                                                  REFERENCES identify.users ON DELETE SET NULL,
    version              INTEGER                   DEFAULT 1                                   NOT NULL,
    assigned_at          TIMESTAMP WITH TIME ZONE  DEFAULT NOW( )                              NOT NULL,
    last_activity_at     TIMESTAMP WITH TIME ZONE,
    created_at           TIMESTAMP WITH TIME ZONE  DEFAULT NOW( )                              NOT NULL,
    updated_at           TIMESTAMP WITH TIME ZONE  DEFAULT NOW( )                              NOT NULL,
    deleted_at           TIMESTAMP WITH TIME ZONE,
    deleted_by           uuid                                                                  REFERENCES identify.users ON DELETE SET NULL,
    CONSTRAINT novel_organization_assignment_novel_id_language_organizatio_key UNIQUE ( novel_id, language, organization_id )
);

COMMENT ON TABLE novel_organization_assignments IS 'Assigns organizations to translate specific novels in specific languages';

COMMENT ON COLUMN novel_organization_assignments.status IS 'Assignment status: active, inactive, or suspended';

COMMENT ON COLUMN novel_organization_assignments.has_exclusive_rights IS 'If TRUE, this organization claims exclusive translation rights (can be challenged via reports)';

ALTER TABLE novel_organization_assignments
    OWNER TO system_dev;

CREATE INDEX idx_novel_organization_assignments_novel ON novel_organization_assignments ( novel_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novel_organization_assignments_organization ON novel_organization_assignments ( organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novel_organization_assignments_language ON novel_organization_assignments ( novel_id, language ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_novel_organization_assignments_exclusive ON novel_organization_assignments ( novel_id, language, has_exclusive_rights ) WHERE (
    (has_exclusive_rights = TRUE) AND (status = 'active'::catalog.assignment_status) AND (deleted_at IS NULL));

CREATE INDEX idx_novel_organization_assignments_status ON novel_organization_assignments ( status ) WHERE (deleted_at IS NULL);

CREATE FUNCTION increment_version() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    NEW.version := OLD.version + 1;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$;

ALTER FUNCTION increment_version() OWNER TO system_dev;

CREATE TRIGGER trg_novels_version
    BEFORE UPDATE
    ON novels
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_volumes_version
    BEFORE UPDATE
    ON novel_volumes
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_chapters_version
    BEFORE UPDATE
    ON novel_chapters
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_ownership_transfers_version
    BEFORE UPDATE
    ON ownership_transfers
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_exclusive_reports_version
    BEFORE UPDATE
    ON exclusive_translation_reports
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_synopsis_translations_version
    BEFORE UPDATE
    ON novel_synopsis_translations
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_synopsis_contributions_version
    BEFORE UPDATE
    ON synopsis_translation_contributions
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_chapter_translations_version
    BEFORE UPDATE
    ON novel_chapter_translations
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_translation_contributions_version
    BEFORE UPDATE
    ON translation_contributions
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_genres_version
    BEFORE UPDATE
    ON genres
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_authors_version
    BEFORE UPDATE
    ON authors
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_artists_version
    BEFORE UPDATE
    ON artists
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_translators_version
    BEFORE UPDATE
    ON translators
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE TRIGGER trg_novel_organization_assignments_version
    BEFORE UPDATE
    ON novel_organization_assignments
    FOR EACH ROW
    WHEN (old.* IS DISTINCT FROM new.*)
EXECUTE PROCEDURE increment_version( );

CREATE FUNCTION update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$;

ALTER FUNCTION update_timestamp() OWNER TO system_dev;

