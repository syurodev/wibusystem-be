-- =====================================================
-- Migration 000014: Translation Teams
-- Description: Teams, members, and novel assignments
-- =====================================================

-- =====================================================
-- ENUM TYPES
-- =====================================================

-- Team member role
CREATE TYPE catalog.team_member_role AS ENUM (
    'leader',
    'translator',
    'proofreader',
    'editor'
);

-- Team assignment status
CREATE TYPE catalog.assignment_status AS ENUM (
    'active',
    'inactive',
    'suspended'
);

-- =====================================================
-- TRANSLATION TEAMS TABLE
-- =====================================================
CREATE TABLE catalog.translation_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL, -- References identify.tenants (validated in application)

    -- Basic info
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL,
    description TEXT,
    avatar_url VARCHAR(1000),

    -- Settings
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_recruiting BOOLEAN NOT NULL DEFAULT FALSE,

    -- Permissions (structured boolean fields for clarity)
    can_translate BOOLEAN NOT NULL DEFAULT TRUE,
    can_proofread BOOLEAN NOT NULL DEFAULT TRUE,
    can_edit BOOLEAN NOT NULL DEFAULT TRUE,

    -- Statistics (auto-updated by application)
    member_count INTEGER NOT NULL DEFAULT 0 CHECK (member_count >= 0),
    active_projects INTEGER NOT NULL DEFAULT 0 CHECK (active_projects >= 0),
    completed_translations INTEGER NOT NULL DEFAULT 0 CHECK (completed_translations >= 0),

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
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(tenant_id, slug)
);

-- Indexes for translation_teams
CREATE INDEX idx_translation_teams_tenant_id ON catalog.translation_teams(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_teams_slug ON catalog.translation_teams(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_teams_active ON catalog.translation_teams(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_translation_teams_recruiting ON catalog.translation_teams(is_recruiting) WHERE is_recruiting = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_translation_teams_metadata ON catalog.translation_teams USING GIN(metadata) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.translation_teams IS 'Translation teams belonging to tenants';
COMMENT ON COLUMN catalog.translation_teams.tenant_id IS 'Reference to identify.tenants (validated in application)';
COMMENT ON COLUMN catalog.translation_teams.can_translate IS 'Permission to translate content';
COMMENT ON COLUMN catalog.translation_teams.can_proofread IS 'Permission to proofread translations';
COMMENT ON COLUMN catalog.translation_teams.can_edit IS 'Permission to edit translations';

-- =====================================================
-- TEAM MEMBERS TABLE
-- =====================================================
CREATE TABLE catalog.team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    -- Role and permissions
    role catalog.team_member_role NOT NULL DEFAULT 'translator',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Statistics (auto-updated by application)
    contribution_count INTEGER NOT NULL DEFAULT 0 CHECK (contribution_count >= 0),
    quality_score DECIMAL(3,2) DEFAULT 0.00 CHECK (quality_score >= 0 AND quality_score <= 5),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(team_id, user_id)
);

-- Indexes for team_members
CREATE INDEX idx_team_members_team_id ON catalog.team_members(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_members_user_id ON catalog.team_members(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_members_role ON catalog.team_members(team_id, role) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_members_active ON catalog.team_members(team_id, is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.team_members IS 'Members of translation teams';
COMMENT ON COLUMN catalog.team_members.quality_score IS 'Average quality score from reviewers (0-5 scale)';

-- =====================================================
-- NOVEL TEAM ASSIGNMENTS TABLE
-- =====================================================
CREATE TABLE catalog.novel_team_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- ISO 639-1: en, vi, zh, ja, ko, etc.

    -- Assignment details
    status catalog.assignment_status NOT NULL DEFAULT 'active',
    has_exclusive_rights BOOLEAN NOT NULL DEFAULT FALSE,

    -- Statistics (auto-updated by application)
    chapters_translated INTEGER NOT NULL DEFAULT 0 CHECK (chapters_translated >= 0),
    chapters_proofread INTEGER NOT NULL DEFAULT 0 CHECK (chapters_proofread >= 0),

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Dates
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    -- One team per novel per language with exclusive rights
    -- Multiple teams can work on same novel+language if no exclusive rights
    UNIQUE(novel_id, language, team_id)
);

-- Indexes for novel_team_assignments
CREATE INDEX idx_novel_team_assignments_novel ON catalog.novel_team_assignments(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_team_assignments_team ON catalog.novel_team_assignments(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_team_assignments_language ON catalog.novel_team_assignments(novel_id, language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novel_team_assignments_exclusive ON catalog.novel_team_assignments(novel_id, language, has_exclusive_rights) WHERE has_exclusive_rights = TRUE AND status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_novel_team_assignments_status ON catalog.novel_team_assignments(status) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE catalog.novel_team_assignments IS 'Assigns teams to translate specific novels in specific languages';
COMMENT ON COLUMN catalog.novel_team_assignments.has_exclusive_rights IS 'If TRUE, this team claims exclusive translation rights (can be challenged via reports)';
COMMENT ON COLUMN catalog.novel_team_assignments.status IS 'Assignment status: active, inactive, or suspended';

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Translation teams triggers
CREATE TRIGGER trg_translation_teams_version
    BEFORE UPDATE ON catalog.translation_teams
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Team members triggers
CREATE TRIGGER trg_team_members_version
    BEFORE UPDATE ON catalog.team_members
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();

-- Novel team assignments triggers
CREATE TRIGGER trg_novel_team_assignments_version
    BEFORE UPDATE ON catalog.novel_team_assignments
    FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*)
    EXECUTE FUNCTION catalog.increment_version();
