-- ============================================================================
-- WIBUSYSTEM DATABASE INITIALIZATION SCRIPT
-- Auto-generated from Ent ORM Schemas
-- Generated: 2025-12-24
-- ============================================================================

-- ============================================================================
-- EXTENSIONS
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================================================
-- SCHEMAS
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE SCHEMA IF NOT EXISTS identify;
CREATE SCHEMA IF NOT EXISTS payment;
CREATE SCHEMA IF NOT EXISTS community;

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

-- identify schema enums
CREATE TYPE identify.user_status AS ENUM ('active', 'suspended', 'deleted');
CREATE TYPE identify.org_member_role AS ENUM ('owner', 'admin', 'moderator', 'member');
CREATE TYPE identify.org_status AS ENUM ('active', 'suspended', 'deleted');
CREATE TYPE identify.org_pending_invite_status AS ENUM ('pending', 'approved', 'rejected', 'expired');
CREATE TYPE identify.org_report_status AS ENUM ('pending', 'org_responded', 'reviewing', 'resolved', 'dismissed');
CREATE TYPE identify.consent_method AS ENUM ('explicit', 'implicit', 'remembered');
CREATE TYPE identify.webauthn_attestation_type AS ENUM ('none', 'indirect', 'direct');
CREATE TYPE identify.webauthn_session_type AS ENUM ('registration', 'authentication');

-- catalog schema enums
CREATE TYPE catalog.novel_status AS ENUM ('draft', 'ongoing', 'completed', 'hiatus', 'dropped');
CREATE TYPE catalog.chapter_status AS ENUM ('draft', 'published', 'scheduled');

-- payment schema enums
CREATE TYPE payment.topup_order_status AS ENUM ('pending', 'success', 'expired', 'cancelled', 'failed');
CREATE TYPE payment.transaction_type AS ENUM ('topup', 'purchase_chapter', 'purchase_series', 'rental', 'subscription', 'refund', 'admin_adjustment');
CREATE TYPE payment.config_value_type AS ENUM ('string', 'number', 'boolean', 'json');

-- community schema enums
CREATE TYPE catalog.unit_progress_status AS ENUM ('in_progress', 'completed');

-- ============================================================================
-- IDENTIFY SCHEMA TABLES
-- ============================================================================

-- Users table
CREATE TABLE identify.users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    phone VARCHAR(50),
    status identify.user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ,
    settings JSONB,
    display_name VARCHAR(255),
    username VARCHAR(255) UNIQUE,
    bio JSONB,
    is_verified BOOLEAN NOT NULL DEFAULT false
);

-- Organizations table
CREATE TABLE identify.organizations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    status identify.org_status NOT NULL DEFAULT 'active',
    description JSONB,
    avatar_url TEXT,
    settings JSONB,
    is_recruiting BOOLEAN NOT NULL DEFAULT false,
    can_translate BOOLEAN NOT NULL DEFAULT true,
    can_proofread BOOLEAN NOT NULL DEFAULT false,
    can_edit BOOLEAN NOT NULL DEFAULT false,
    member_count INTEGER NOT NULL DEFAULT 0,
    active_projects INTEGER NOT NULL DEFAULT 0,
    completed_translations INTEGER NOT NULL DEFAULT 0,
    report_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Organization Members table
CREATE TABLE identify.organization_members (
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    role identify.org_member_role NOT NULL DEFAULT 'member',
    is_active BOOLEAN NOT NULL DEFAULT true,
    contribution_count INTEGER NOT NULL DEFAULT 0,
    quality_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    metadata JSONB,
    invited_by UUID,
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ,
    left_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, organization_id)
);

-- Organization Pending Invites table
CREATE TABLE identify.organization_pending_invites (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    invited_by UUID NOT NULL,
    status identify.org_pending_invite_status NOT NULL DEFAULT 'pending',
    approved_by UUID,
    processed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Organization Reports table
CREATE TABLE identify.organization_reports (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id UUID NOT NULL,
    reporter_id UUID NOT NULL,
    reason VARCHAR(255) NOT NULL,
    description TEXT,
    org_response TEXT,
    org_responded_by UUID,
    org_responded_at TIMESTAMPTZ,
    status identify.org_report_status NOT NULL DEFAULT 'pending',
    resolved_by UUID,
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Permission and Role scope types
CREATE TYPE identify.permission_scope AS ENUM ('global', 'organization');
CREATE TYPE identify.role_scope AS ENUM ('global', 'organization');

-- Permissions table
CREATE TABLE identify.permissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL UNIQUE,
    scope identify.permission_scope NOT NULL,
    description TEXT,
    resource VARCHAR(100),
    action VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT permissions_name_format_check CHECK (name ~ '^[a-z_]+:[a-z_]+$')
);

-- Roles table
CREATE TABLE identify.roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL UNIQUE,
    slug VARCHAR(255) UNIQUE,
    scope identify.role_scope,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT roles_name_format_check CHECK (name ~ '^[A-Z_]+$')
);

-- Role Permissions (junction table)
CREATE TABLE identify.role_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id)
);

-- User Roles table
CREATE TABLE identify.user_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    organization_id UUID,
    metadata JSONB,
    assigned_by UUID,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- User Global Roles table
CREATE TABLE identify.user_global_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, role_id)
);

-- User Organization Roles table
CREATE TABLE identify.user_organization_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    role_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, organization_id, role_id)
);

-- User Statistics table
CREATE TABLE identify.user_statistics (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL UNIQUE,
    follower_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    novel_count INTEGER NOT NULL DEFAULT 0,
    manga_count INTEGER NOT NULL DEFAULT 0,
    anime_count INTEGER NOT NULL DEFAULT 0,
    last_content_updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Sessions table
CREATE TABLE identify.sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    refresh_token TEXT NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address VARCHAR(50),
    expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OAuth2 Clients table
CREATE TABLE identify.oauth2_clients (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    client_name VARCHAR(255) NOT NULL,
    secret_hash TEXT,
    redirect_uris TEXT[],
    grant_types TEXT[],
    response_types TEXT[],
    scopes TEXT[],
    is_public BOOLEAN NOT NULL DEFAULT false,
    is_internal BOOLEAN NOT NULL DEFAULT false,
    token_endpoint_auth_method VARCHAR(100) NOT NULL DEFAULT 'client_secret_basic',
    organization_id UUID,
    owner_user_id UUID,
    client_uri TEXT,
    logo_url TEXT,
    terms_of_service_url TEXT,
    policy_url TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OAuth2 Sessions table
CREATE TABLE identify.oauth2_sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    signature VARCHAR(255) NOT NULL UNIQUE,
    request_id VARCHAR(255) NOT NULL,
    session_type VARCHAR(100) NOT NULL,
    session_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    subject_id VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- OAuth2 Consents table
CREATE TABLE identify.oauth2_consents (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    client_id UUID NOT NULL,
    granted_scopes TEXT[],
    revoked BOOLEAN NOT NULL DEFAULT false,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    consent_method identify.consent_method NOT NULL DEFAULT 'explicit',
    ip_address VARCHAR(50),
    user_agent TEXT
);

-- Email Verification Tokens table
CREATE TABLE identify.email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Password Reset Tokens table
CREATE TABLE identify.password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- WebAuthn Credentials table
CREATE TABLE identify.webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    credential_id VARCHAR(255) NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type identify.webauthn_attestation_type NOT NULL DEFAULT 'none',
    aaguid BYTEA,
    sign_count INTEGER NOT NULL DEFAULT 0,
    transports TEXT[],
    backup_eligible BOOLEAN NOT NULL DEFAULT false,
    backup_state BOOLEAN NOT NULL DEFAULT false,
    credential_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ
);

-- WebAuthn Sessions table
CREATE TABLE identify.webauthn_sessions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID,
    challenge VARCHAR(255) NOT NULL UNIQUE,
    session_type identify.webauthn_session_type NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL
);

-- ============================================================================
-- CATALOG SCHEMA TABLES
-- ============================================================================

-- Genres table
CREATE TABLE catalog.genres (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    parent_id UUID,
    is_active BOOLEAN NOT NULL DEFAULT true,
    novel_count INTEGER NOT NULL DEFAULT 0,
    anime_count INTEGER NOT NULL DEFAULT 0,
    manga_count INTEGER NOT NULL DEFAULT 0,
    active_readers BIGINT NOT NULL DEFAULT 0,
    total_views BIGINT NOT NULL DEFAULT 0,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Authors table
CREATE TABLE catalog.authors (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    biography JSONB,
    avatar_url TEXT,
    social_links JSONB,
    novel_count INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_views BIGINT NOT NULL DEFAULT 0,
    follower_count INTEGER NOT NULL DEFAULT 0,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB,
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Artists table
CREATE TABLE catalog.artists (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    biography JSONB,
    avatar_url TEXT,
    social_links JSONB,
    specialization VARCHAR(255),
    portfolio_url TEXT,
    novel_count INTEGER NOT NULL DEFAULT 0,
    artwork_count INTEGER NOT NULL DEFAULT 0,
    follower_count INTEGER NOT NULL DEFAULT 0,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB,
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Novels table
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    owner_id UUID NOT NULL,
    owner_type VARCHAR(50) NOT NULL DEFAULT 'user',
    synopsis JSONB,
    cover_image_url TEXT,
    thumbnail_url TEXT,
    status catalog.novel_status NOT NULL DEFAULT 'draft',
    is_oneshot BOOLEAN NOT NULL DEFAULT false,
    original_language VARCHAR(10),
    original_title VARCHAR(255),
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DOUBLE PRECISION NOT NULL DEFAULT 0,
    rating_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB,
    first_published_at TIMESTAMPTZ,
    last_chapter_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Novel Volumes table
CREATE TABLE catalog.novel_volumes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL,
    volume_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image_url TEXT,
    chapter_count INTEGER NOT NULL DEFAULT 0,
    word_count BIGINT NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Novel Chapters table
CREATE TABLE catalog.novel_chapters (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL,
    volume_id UUID,
    chapter_number INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content JSONB,
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,
    is_free BOOLEAN NOT NULL DEFAULT true,
    price DOUBLE PRECISION,
    currency VARCHAR(10),
    status catalog.chapter_status NOT NULL DEFAULT 'draft',
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    display_order INTEGER NOT NULL DEFAULT 0,
    author_notes JSONB,
    published_at TIMESTAMPTZ,
    scheduled_at TIMESTAMPTZ,
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Novel Genres (junction table)
CREATE TABLE catalog.novel_genres (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL,
    genre_id UUID NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (novel_id, genre_id)
);

-- Novel Authors (junction table)
CREATE TABLE catalog.novel_authors (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL,
    author_id UUID NOT NULL,
    role VARCHAR(100) NOT NULL DEFAULT 'original_author',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (novel_id, author_id)
);

-- Novel Artists (junction table)
CREATE TABLE catalog.novel_artists (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL,
    artist_id UUID NOT NULL,
    role VARCHAR(100) NOT NULL DEFAULT 'cover_artist',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (novel_id, artist_id)
);

-- Novel Embeddings table (for vector search)
CREATE TABLE catalog.novel_embeddings (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL UNIQUE,
    embedding vector(384) NOT NULL,
    model_version VARCHAR(100) NOT NULL,
    source_hash VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Novel Chapter Histories table (audit log)
CREATE TABLE catalog.novel_chapter_histories (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    chapter_id UUID NOT NULL,
    volume_id UUID,
    novel_id UUID NOT NULL,
    version_number INTEGER NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    slug VARCHAR(255),
    chapter_number INTEGER,
    status VARCHAR(50),
    word_count INTEGER,
    character_count INTEGER,
    changed_fields JSONB,
    change_summary TEXT,
    content_changed BOOLEAN NOT NULL DEFAULT false,
    changed_by UUID NOT NULL,
    request_id VARCHAR(255),
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Novel Volume Histories table (audit log)
CREATE TABLE catalog.novel_volume_histories (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    volume_id UUID NOT NULL,
    novel_id UUID NOT NULL,
    version_number INTEGER NOT NULL DEFAULT 1,
    action VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    slug VARCHAR(255),
    volume_number INTEGER,
    is_published BOOLEAN,
    chapter_count INTEGER,
    word_count INTEGER,
    changed_fields JSONB,
    change_summary TEXT,
    changed_by UUID NOT NULL,
    request_id VARCHAR(255),
    ip_address VARCHAR(50),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================================
-- PAYMENT SCHEMA TABLES
-- ============================================================================

-- User Wallets table
CREATE TABLE payment.user_wallets (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL UNIQUE,
    coin_balance NUMERIC(20,2) NOT NULL DEFAULT 0,
    total_deposited NUMERIC(20,2) NOT NULL DEFAULT 0,
    total_spent NUMERIC(20,2) NOT NULL DEFAULT 0,
    total_subscription_spent NUMERIC(20,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Coin Packages table
CREATE TABLE payment.coin_packages (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    coin_amount NUMERIC(20,2) NOT NULL,
    price_vnd NUMERIC(20,2) NOT NULL,
    bonus_percent INTEGER NOT NULL DEFAULT 0,
    is_popular BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Topup Orders table
CREATE TABLE payment.topup_orders (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    package_id UUID NOT NULL,
    order_code VARCHAR(100) NOT NULL UNIQUE,
    coin_amount NUMERIC(20,2) NOT NULL,
    base_coin_amount NUMERIC(20,2) NOT NULL,
    bonus_coin_amount NUMERIC(20,2) NOT NULL,
    vnd_amount NUMERIC(20,2) NOT NULL,
    status payment.topup_order_status NOT NULL DEFAULT 'pending',
    sepay_transaction_id VARCHAR(255),
    sepay_content TEXT,
    bank_name VARCHAR(255),
    bank_account VARCHAR(100),
    account_name VARCHAR(255),
    completed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Transactions table
CREATE TABLE payment.transactions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    type payment.transaction_type NOT NULL,
    coin_amount NUMERIC(20,2) NOT NULL,
    vnd_amount NUMERIC(20,2),
    balance_after NUMERIC(20,2) NOT NULL,
    reference_type VARCHAR(100),
    reference_id UUID,
    creator_id UUID,
    creator_revenue_vnd NUMERIC(20,2),
    platform_revenue_vnd NUMERIC(20,2),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Payment Configurations table
CREATE TABLE payment.payment_configurations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT NOT NULL DEFAULT '',
    value_type payment.config_value_type NOT NULL DEFAULT 'string',
    description TEXT,
    is_sensitive BOOLEAN NOT NULL DEFAULT false,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================================
-- COMMUNITY SCHEMA TABLES
-- ============================================================================

-- Novel Chapter Translations table
CREATE TABLE community.novel_chapter_translations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    chapter_id UUID NOT NULL,
    language VARCHAR(10) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    translator_notes TEXT,
    organization_id UUID,
    version INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,
    quality_score DOUBLE PRECISION,
    reviewer_rating DOUBLE PRECISION,
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    contribution_count INTEGER NOT NULL DEFAULT 0,
    reviewed_by UUID,
    review_notes TEXT,
    reviewed_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (chapter_id, language)
);

-- Translation Contributions table
CREATE TABLE community.translation_contributions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    chapter_id UUID NOT NULL,
    contributor_id UUID NOT NULL,
    language VARCHAR(10) NOT NULL,
    contribution_type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    contributor_notes TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,
    official_translation_id UUID,
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT false,
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,
    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Media Progress table
CREATE TABLE catalog.media_progress (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    media_type VARCHAR(50) NOT NULL,
    media_id UUID NOT NULL,
    current_unit_id UUID NOT NULL,
    position JSONB,
    total_units INTEGER NOT NULL DEFAULT 0,
    completed_units INTEGER NOT NULL DEFAULT 0,
    progress_percentage DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, media_type, media_id)
);

-- Unit Progress table
CREATE TABLE catalog.unit_progress (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    media_type VARCHAR(50) NOT NULL,
    media_id UUID NOT NULL,
    unit_id UUID NOT NULL,
    status catalog.unit_progress_status NOT NULL DEFAULT 'in_progress',
    position JSONB,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, media_type, unit_id)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- identify.users indexes
CREATE INDEX idx_users_email ON identify.users(email);
CREATE INDEX idx_users_username ON identify.users(username);
CREATE INDEX idx_users_status ON identify.users(status);
CREATE INDEX idx_users_created_at ON identify.users(created_at);

-- identify.organizations indexes
CREATE INDEX idx_organizations_slug ON identify.organizations(slug);
CREATE INDEX idx_organizations_status ON identify.organizations(status);
CREATE INDEX idx_organizations_created_at ON identify.organizations(created_at);

-- identify.organization_members indexes
CREATE INDEX idx_organization_members_user_id ON identify.organization_members(user_id);
CREATE INDEX idx_organization_members_org_id ON identify.organization_members(organization_id);
CREATE INDEX idx_organization_members_role ON identify.organization_members(role);

-- identify.permissions indexes
CREATE INDEX idx_permissions_scope ON identify.permissions(scope);
CREATE INDEX idx_permissions_resource ON identify.permissions(resource);
CREATE INDEX idx_permissions_name ON identify.permissions(name);

-- identify.roles indexes
CREATE INDEX idx_roles_name ON identify.roles(name);
CREATE INDEX idx_roles_scope ON identify.roles(scope);
CREATE INDEX idx_roles_slug ON identify.roles(slug);
CREATE INDEX idx_roles_is_system ON identify.roles(is_system);

-- identify.role_permissions indexes
CREATE INDEX idx_role_permissions_role_id ON identify.role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON identify.role_permissions(permission_id);

-- identify.user_global_roles indexes
CREATE INDEX idx_user_global_roles_user_id ON identify.user_global_roles(user_id);
CREATE INDEX idx_user_global_roles_role_id ON identify.user_global_roles(role_id);

-- identify.user_organization_roles indexes
CREATE INDEX idx_user_organization_roles_user_id ON identify.user_organization_roles(user_id);
CREATE INDEX idx_user_organization_roles_org_id ON identify.user_organization_roles(organization_id);
CREATE INDEX idx_user_organization_roles_role_id ON identify.user_organization_roles(role_id);

-- identify.sessions indexes
CREATE INDEX idx_sessions_user_id ON identify.sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON identify.sessions(expires_at);

-- identify.oauth2_sessions indexes
CREATE INDEX idx_oauth2_sessions_subject_active ON identify.oauth2_sessions(subject_id, active);
CREATE INDEX idx_oauth2_sessions_client_id ON identify.oauth2_sessions(client_id);
CREATE INDEX idx_oauth2_sessions_expires_at ON identify.oauth2_sessions(expires_at);

-- identify.oauth2_consents indexes
CREATE INDEX idx_oauth2_consents_user_id ON identify.oauth2_consents(user_id);
CREATE INDEX idx_oauth2_consents_client_id ON identify.oauth2_consents(client_id);

-- identify.webauthn_credentials indexes
CREATE INDEX idx_webauthn_credentials_user_id ON identify.webauthn_credentials(user_id);

-- catalog.genres indexes
CREATE INDEX idx_genres_slug ON catalog.genres(slug);
CREATE INDEX idx_genres_parent_id ON catalog.genres(parent_id);
CREATE INDEX idx_genres_is_active ON catalog.genres(is_active);

-- catalog.authors indexes
CREATE INDEX idx_authors_slug ON catalog.authors(slug);
CREATE INDEX idx_authors_user_id ON catalog.authors(user_id);

-- catalog.artists indexes
CREATE INDEX idx_artists_slug ON catalog.artists(slug);
CREATE INDEX idx_artists_user_id ON catalog.artists(user_id);

-- catalog.novels indexes
CREATE INDEX idx_novels_slug ON catalog.novels(slug);
CREATE INDEX idx_novels_owner_id ON catalog.novels(owner_id);
CREATE INDEX idx_novels_owner_type ON catalog.novels(owner_type);
CREATE INDEX idx_novels_status ON catalog.novels(status);
CREATE INDEX idx_novels_created_at ON catalog.novels(created_at);
CREATE INDEX idx_novels_view_count ON catalog.novels(view_count);
CREATE INDEX idx_novels_rating_average ON catalog.novels(rating_average);

-- catalog.novel_volumes indexes
CREATE INDEX idx_novel_volumes_novel_id ON catalog.novel_volumes(novel_id);
CREATE INDEX idx_novel_volumes_display_order ON catalog.novel_volumes(display_order);

-- catalog.novel_chapters indexes
CREATE INDEX idx_novel_chapters_novel_id ON catalog.novel_chapters(novel_id);
CREATE INDEX idx_novel_chapters_volume_id ON catalog.novel_chapters(volume_id);
CREATE INDEX idx_novel_chapters_status ON catalog.novel_chapters(status);
CREATE INDEX idx_novel_chapters_display_order ON catalog.novel_chapters(display_order);

-- catalog.novel_genres indexes
CREATE INDEX idx_novel_genres_novel_id ON catalog.novel_genres(novel_id);
CREATE INDEX idx_novel_genres_genre_id ON catalog.novel_genres(genre_id);

-- catalog.novel_authors indexes
CREATE INDEX idx_novel_authors_novel_id ON catalog.novel_authors(novel_id);
CREATE INDEX idx_novel_authors_author_id ON catalog.novel_authors(author_id);

-- catalog.novel_artists indexes
CREATE INDEX idx_novel_artists_novel_id ON catalog.novel_artists(novel_id);
CREATE INDEX idx_novel_artists_artist_id ON catalog.novel_artists(artist_id);

-- catalog.novel_embeddings indexes
CREATE INDEX idx_novel_embeddings_novel_id ON catalog.novel_embeddings(novel_id);
-- HNSW index for vector similarity search
CREATE INDEX idx_novel_embeddings_hnsw ON catalog.novel_embeddings USING hnsw (embedding vector_cosine_ops);

-- catalog.novel_chapter_histories indexes
CREATE INDEX idx_novel_chapter_histories_chapter_id ON catalog.novel_chapter_histories(chapter_id);
CREATE INDEX idx_novel_chapter_histories_novel_id ON catalog.novel_chapter_histories(novel_id);
CREATE INDEX idx_novel_chapter_histories_changed_by ON catalog.novel_chapter_histories(changed_by);

-- catalog.novel_volume_histories indexes
CREATE INDEX idx_novel_volume_histories_volume_id ON catalog.novel_volume_histories(volume_id);
CREATE INDEX idx_novel_volume_histories_novel_id ON catalog.novel_volume_histories(novel_id);
CREATE INDEX idx_novel_volume_histories_changed_by ON catalog.novel_volume_histories(changed_by);

-- payment.user_wallets indexes
CREATE INDEX idx_user_wallets_user_id ON payment.user_wallets(user_id);

-- payment.topup_orders indexes
CREATE INDEX idx_topup_orders_user_id ON payment.topup_orders(user_id);
CREATE INDEX idx_topup_orders_status ON payment.topup_orders(status);
CREATE INDEX idx_topup_orders_order_code ON payment.topup_orders(order_code);
CREATE INDEX idx_topup_orders_created_at ON payment.topup_orders(created_at);

-- payment.transactions indexes
CREATE INDEX idx_transactions_user_id ON payment.transactions(user_id);
CREATE INDEX idx_transactions_type ON payment.transactions(type);
CREATE INDEX idx_transactions_reference_id ON payment.transactions(reference_id);
CREATE INDEX idx_transactions_created_at ON payment.transactions(created_at);

-- payment.payment_configurations indexes
CREATE INDEX idx_payment_configurations_key ON payment.payment_configurations(key);

-- community.novel_chapter_translations indexes
CREATE INDEX idx_novel_chapter_translations_chapter_id ON community.novel_chapter_translations(chapter_id);
CREATE INDEX idx_novel_chapter_translations_org_id ON community.novel_chapter_translations(organization_id);
CREATE INDEX idx_novel_chapter_translations_created_by ON community.novel_chapter_translations(created_by);
CREATE INDEX idx_novel_chapter_translations_status ON community.novel_chapter_translations(status);

-- community.translation_contributions indexes
CREATE INDEX idx_translation_contributions_chapter_id ON community.translation_contributions(chapter_id);
CREATE INDEX idx_translation_contributions_contributor_id ON community.translation_contributions(contributor_id);
CREATE INDEX idx_translation_contributions_status ON community.translation_contributions(status);

-- catalog.media_progress indexes
CREATE INDEX idx_media_progress_user_id ON catalog.media_progress(user_id);
CREATE INDEX idx_media_progress_media ON catalog.media_progress(media_type, media_id);
CREATE INDEX idx_media_progress_last_accessed ON catalog.media_progress(last_accessed_at);

-- catalog.unit_progress indexes
CREATE INDEX idx_unit_progress_user_id ON catalog.unit_progress(user_id);
CREATE INDEX idx_unit_progress_media ON catalog.unit_progress(media_type, media_id);
CREATE INDEX idx_unit_progress_unit_id ON catalog.unit_progress(unit_id);

-- ============================================================================
-- FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- identify schema foreign keys
ALTER TABLE identify.organization_members
    ADD CONSTRAINT fk_org_members_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_org_members_org_id FOREIGN KEY (organization_id) REFERENCES identify.organizations(id) ON DELETE CASCADE;

ALTER TABLE identify.organization_pending_invites
    ADD CONSTRAINT fk_org_pending_invites_org_id FOREIGN KEY (organization_id) REFERENCES identify.organizations(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_org_pending_invites_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_org_pending_invites_invited_by FOREIGN KEY (invited_by) REFERENCES identify.users(id);

ALTER TABLE identify.organization_reports
    ADD CONSTRAINT fk_org_reports_org_id FOREIGN KEY (organization_id) REFERENCES identify.organizations(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_org_reports_reporter_id FOREIGN KEY (reporter_id) REFERENCES identify.users(id);

ALTER TABLE identify.user_roles
    ADD CONSTRAINT fk_user_roles_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_roles_role_id FOREIGN KEY (role_id) REFERENCES identify.roles(id) ON DELETE CASCADE;

ALTER TABLE identify.role_permissions
    ADD CONSTRAINT fk_role_permissions_role_id FOREIGN KEY (role_id) REFERENCES identify.roles(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_role_permissions_permission_id FOREIGN KEY (permission_id) REFERENCES identify.permissions(id) ON DELETE CASCADE;

ALTER TABLE identify.user_global_roles
    ADD CONSTRAINT fk_user_global_roles_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_global_roles_role_id FOREIGN KEY (role_id) REFERENCES identify.roles(id) ON DELETE CASCADE;

ALTER TABLE identify.user_organization_roles
    ADD CONSTRAINT fk_user_org_roles_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_org_roles_org_id FOREIGN KEY (organization_id) REFERENCES identify.organizations(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_user_org_roles_role_id FOREIGN KEY (role_id) REFERENCES identify.roles(id) ON DELETE CASCADE;

ALTER TABLE identify.user_statistics
    ADD CONSTRAINT fk_user_statistics_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE identify.sessions
    ADD CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE identify.oauth2_consents
    ADD CONSTRAINT fk_oauth2_consents_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_oauth2_consents_client_id FOREIGN KEY (client_id) REFERENCES identify.oauth2_clients(id) ON DELETE CASCADE;

ALTER TABLE identify.email_verification_tokens
    ADD CONSTRAINT fk_email_verification_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE identify.password_reset_tokens
    ADD CONSTRAINT fk_password_reset_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE identify.webauthn_credentials
    ADD CONSTRAINT fk_webauthn_credentials_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE identify.webauthn_sessions
    ADD CONSTRAINT fk_webauthn_sessions_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

-- catalog schema foreign keys
ALTER TABLE catalog.genres
    ADD CONSTRAINT fk_genres_parent_id FOREIGN KEY (parent_id) REFERENCES catalog.genres(id) ON DELETE SET NULL;

ALTER TABLE catalog.novel_volumes
    ADD CONSTRAINT fk_novel_volumes_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_chapters
    ADD CONSTRAINT fk_novel_chapters_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_novel_chapters_volume_id FOREIGN KEY (volume_id) REFERENCES catalog.novel_volumes(id) ON DELETE SET NULL;

ALTER TABLE catalog.novel_genres
    ADD CONSTRAINT fk_novel_genres_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_novel_genres_genre_id FOREIGN KEY (genre_id) REFERENCES catalog.genres(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_authors
    ADD CONSTRAINT fk_novel_authors_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_novel_authors_author_id FOREIGN KEY (author_id) REFERENCES catalog.authors(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_artists
    ADD CONSTRAINT fk_novel_artists_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_novel_artists_artist_id FOREIGN KEY (artist_id) REFERENCES catalog.artists(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_embeddings
    ADD CONSTRAINT fk_novel_embeddings_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_chapter_histories
    ADD CONSTRAINT fk_chapter_histories_chapter_id FOREIGN KEY (chapter_id) REFERENCES catalog.novel_chapters(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_chapter_histories_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE;

ALTER TABLE catalog.novel_volume_histories
    ADD CONSTRAINT fk_volume_histories_volume_id FOREIGN KEY (volume_id) REFERENCES catalog.novel_volumes(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_volume_histories_novel_id FOREIGN KEY (novel_id) REFERENCES catalog.novels(id) ON DELETE CASCADE;

-- payment schema foreign keys
ALTER TABLE payment.user_wallets
    ADD CONSTRAINT fk_user_wallets_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE payment.topup_orders
    ADD CONSTRAINT fk_topup_orders_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_topup_orders_package_id FOREIGN KEY (package_id) REFERENCES payment.coin_packages(id);

ALTER TABLE payment.transactions
    ADD CONSTRAINT fk_transactions_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

-- community schema foreign keys
ALTER TABLE community.novel_chapter_translations
    ADD CONSTRAINT fk_chapter_translations_chapter_id FOREIGN KEY (chapter_id) REFERENCES catalog.novel_chapters(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_chapter_translations_org_id FOREIGN KEY (organization_id) REFERENCES identify.organizations(id) ON DELETE SET NULL;

ALTER TABLE community.translation_contributions
    ADD CONSTRAINT fk_translation_contributions_chapter_id FOREIGN KEY (chapter_id) REFERENCES catalog.novel_chapters(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_translation_contributions_contributor_id FOREIGN KEY (contributor_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE catalog.media_progress
    ADD CONSTRAINT fk_media_progress_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

ALTER TABLE catalog.unit_progress
    ADD CONSTRAINT fk_unit_progress_user_id FOREIGN KEY (user_id) REFERENCES identify.users(id) ON DELETE CASCADE;

-- ============================================================================
-- TRIGGERS FOR updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables with updated_at column
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON identify.users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON identify.organizations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organization_members_updated_at BEFORE UPDATE ON identify.organization_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organization_reports_updated_at BEFORE UPDATE ON identify.organization_reports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON identify.roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON identify.user_roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_statistics_updated_at BEFORE UPDATE ON identify.user_statistics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON identify.sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_oauth2_clients_updated_at BEFORE UPDATE ON identify.oauth2_clients FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_webauthn_credentials_updated_at BEFORE UPDATE ON identify.webauthn_credentials FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_genres_updated_at BEFORE UPDATE ON catalog.genres FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_authors_updated_at BEFORE UPDATE ON catalog.authors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_artists_updated_at BEFORE UPDATE ON catalog.artists FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_novels_updated_at BEFORE UPDATE ON catalog.novels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_novel_volumes_updated_at BEFORE UPDATE ON catalog.novel_volumes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_novel_chapters_updated_at BEFORE UPDATE ON catalog.novel_chapters FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_novel_embeddings_updated_at BEFORE UPDATE ON catalog.novel_embeddings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_wallets_updated_at BEFORE UPDATE ON payment.user_wallets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_coin_packages_updated_at BEFORE UPDATE ON payment.coin_packages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_topup_orders_updated_at BEFORE UPDATE ON payment.topup_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_payment_configurations_updated_at BEFORE UPDATE ON payment.payment_configurations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_chapter_translations_updated_at BEFORE UPDATE ON community.novel_chapter_translations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_translation_contributions_updated_at BEFORE UPDATE ON community.translation_contributions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_media_progress_updated_at BEFORE UPDATE ON catalog.media_progress FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON SCHEMA identify IS 'User identity, authentication, and authorization';
COMMENT ON SCHEMA catalog IS 'Content catalog including novels, chapters, authors, artists, genres';
COMMENT ON SCHEMA payment IS 'Payment processing, wallets, transactions';
COMMENT ON SCHEMA community IS 'Community features including translations, progress tracking';

-- identify schema table comments
COMMENT ON TABLE identify.users IS 'User accounts';
COMMENT ON TABLE identify.organizations IS 'Translation groups and organizations';
COMMENT ON TABLE identify.organization_members IS 'Organization membership';
COMMENT ON TABLE identify.permissions IS 'Master list of all permissions in the system';
COMMENT ON TABLE identify.roles IS 'Role definitions for RBAC';
COMMENT ON TABLE identify.role_permissions IS 'Many-to-many junction table linking roles to permissions';
COMMENT ON TABLE identify.sessions IS 'User login sessions';
COMMENT ON TABLE identify.oauth2_clients IS 'OAuth2 client applications';
COMMENT ON TABLE identify.oauth2_sessions IS 'OAuth2 authorization sessions';

-- catalog schema table comments
COMMENT ON TABLE catalog.novels IS 'Novel series';
COMMENT ON TABLE catalog.novel_volumes IS 'Novel volumes/arcs';
COMMENT ON TABLE catalog.novel_chapters IS 'Novel chapters';
COMMENT ON TABLE catalog.genres IS 'Content genres/categories';
COMMENT ON TABLE catalog.authors IS 'Novel authors';
COMMENT ON TABLE catalog.artists IS 'Illustrators and cover artists';
COMMENT ON TABLE catalog.novel_embeddings IS 'Vector embeddings for similarity search';

-- payment schema table comments
COMMENT ON TABLE payment.user_wallets IS 'User coin balances';
COMMENT ON TABLE payment.coin_packages IS 'Available coin packages for purchase';
COMMENT ON TABLE payment.topup_orders IS 'Coin purchase orders';
COMMENT ON TABLE payment.transactions IS 'Transaction history';

-- community schema table comments
COMMENT ON TABLE community.novel_chapter_translations IS 'Translated chapter content';
COMMENT ON TABLE community.translation_contributions IS 'Community translation contributions';
COMMENT ON TABLE catalog.media_progress IS 'User reading/watching progress per series';
COMMENT ON TABLE catalog.unit_progress IS 'User progress per chapter/episode';

-- ============================================================================
-- COLUMN COMMENTS
-- ============================================================================

-- ============================================================================
-- identify.users columns
-- ============================================================================
COMMENT ON COLUMN identify.users.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.users.email IS 'User email address, unique';
COMMENT ON COLUMN identify.users.email_verified IS 'Whether email has been verified';
COMMENT ON COLUMN identify.users.password_hash IS 'Hashed password (bcrypt/argon2)';
COMMENT ON COLUMN identify.users.full_name IS 'User full name';
COMMENT ON COLUMN identify.users.avatar_url IS 'URL to user avatar image';
COMMENT ON COLUMN identify.users.phone IS 'User phone number';
COMMENT ON COLUMN identify.users.status IS 'Account status: active, suspended, deleted';
COMMENT ON COLUMN identify.users.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.users.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN identify.users.last_login_at IS 'Timestamp of last successful login';
COMMENT ON COLUMN identify.users.settings IS 'User preferences and settings (JSONB)';
COMMENT ON COLUMN identify.users.display_name IS 'Public display name';
COMMENT ON COLUMN identify.users.username IS 'Unique username for profile URL';
COMMENT ON COLUMN identify.users.bio IS 'User biography/description (JSONB array)';
COMMENT ON COLUMN identify.users.is_verified IS 'Whether user is verified (blue check)';

-- ============================================================================
-- identify.organizations columns
-- ============================================================================
COMMENT ON COLUMN identify.organizations.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.organizations.name IS 'Organization display name';
COMMENT ON COLUMN identify.organizations.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN identify.organizations.status IS 'Organization status: active, suspended, deleted';
COMMENT ON COLUMN identify.organizations.description IS 'Organization description (JSONB)';
COMMENT ON COLUMN identify.organizations.avatar_url IS 'URL to organization avatar/logo';
COMMENT ON COLUMN identify.organizations.settings IS 'Organization settings (JSONB)';
COMMENT ON COLUMN identify.organizations.is_recruiting IS 'Whether organization is accepting new members';
COMMENT ON COLUMN identify.organizations.can_translate IS 'Whether organization can translate content';
COMMENT ON COLUMN identify.organizations.can_proofread IS 'Whether organization can proofread content';
COMMENT ON COLUMN identify.organizations.can_edit IS 'Whether organization can edit content';
COMMENT ON COLUMN identify.organizations.member_count IS 'Total number of members';
COMMENT ON COLUMN identify.organizations.active_projects IS 'Number of active translation projects';
COMMENT ON COLUMN identify.organizations.completed_translations IS 'Number of completed translations';
COMMENT ON COLUMN identify.organizations.report_count IS 'Number of reports received';
COMMENT ON COLUMN identify.organizations.metadata IS 'Additional metadata (JSONB)';
COMMENT ON COLUMN identify.organizations.created_by IS 'User ID who created this organization';
COMMENT ON COLUMN identify.organizations.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN identify.organizations.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN identify.organizations.version IS 'Optimistic locking version number';
COMMENT ON COLUMN identify.organizations.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.organizations.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN identify.organizations.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- identify.organization_members columns
-- ============================================================================
COMMENT ON COLUMN identify.organization_members.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.organization_members.organization_id IS 'Reference to organization';
COMMENT ON COLUMN identify.organization_members.status IS 'Membership status';
COMMENT ON COLUMN identify.organization_members.role IS 'Member role: owner, admin, moderator, member';
COMMENT ON COLUMN identify.organization_members.is_active IS 'Whether membership is active';
COMMENT ON COLUMN identify.organization_members.contribution_count IS 'Number of contributions made';
COMMENT ON COLUMN identify.organization_members.quality_score IS 'Quality score based on contributions';
COMMENT ON COLUMN identify.organization_members.metadata IS 'Additional member metadata (JSONB)';
COMMENT ON COLUMN identify.organization_members.invited_by IS 'User ID who invited this member';
COMMENT ON COLUMN identify.organization_members.invited_at IS 'Timestamp when invitation was sent';
COMMENT ON COLUMN identify.organization_members.joined_at IS 'Timestamp when member joined';
COMMENT ON COLUMN identify.organization_members.left_at IS 'Timestamp when member left';
COMMENT ON COLUMN identify.organization_members.created_by IS 'User ID who created this record';
COMMENT ON COLUMN identify.organization_members.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN identify.organization_members.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN identify.organization_members.version IS 'Optimistic locking version number';
COMMENT ON COLUMN identify.organization_members.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.organization_members.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN identify.organization_members.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- identify.organization_pending_invites columns
-- ============================================================================
COMMENT ON COLUMN identify.organization_pending_invites.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.organization_pending_invites.organization_id IS 'Reference to organization';
COMMENT ON COLUMN identify.organization_pending_invites.user_id IS 'Reference to invited user';
COMMENT ON COLUMN identify.organization_pending_invites.invited_by IS 'User ID who sent the invitation';
COMMENT ON COLUMN identify.organization_pending_invites.status IS 'Invite status: pending, approved, rejected, expired';
COMMENT ON COLUMN identify.organization_pending_invites.approved_by IS 'User ID who approved the invitation';
COMMENT ON COLUMN identify.organization_pending_invites.processed_at IS 'Timestamp when invitation was processed';
COMMENT ON COLUMN identify.organization_pending_invites.expires_at IS 'Timestamp when invitation expires';
COMMENT ON COLUMN identify.organization_pending_invites.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.organization_reports columns
-- ============================================================================
COMMENT ON COLUMN identify.organization_reports.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.organization_reports.organization_id IS 'Reference to reported organization';
COMMENT ON COLUMN identify.organization_reports.reporter_id IS 'User ID who submitted the report';
COMMENT ON COLUMN identify.organization_reports.reason IS 'Reason for report';
COMMENT ON COLUMN identify.organization_reports.description IS 'Detailed description of the report';
COMMENT ON COLUMN identify.organization_reports.org_response IS 'Organization response to the report';
COMMENT ON COLUMN identify.organization_reports.org_responded_by IS 'User ID who responded on behalf of org';
COMMENT ON COLUMN identify.organization_reports.org_responded_at IS 'Timestamp when org responded';
COMMENT ON COLUMN identify.organization_reports.status IS 'Report status: pending, org_responded, reviewing, resolved, dismissed';
COMMENT ON COLUMN identify.organization_reports.resolved_by IS 'Admin user ID who resolved the report';
COMMENT ON COLUMN identify.organization_reports.resolved_at IS 'Timestamp when report was resolved';
COMMENT ON COLUMN identify.organization_reports.resolution_note IS 'Note about the resolution';
COMMENT ON COLUMN identify.organization_reports.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.organization_reports.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.permissions columns
-- ============================================================================
COMMENT ON COLUMN identify.permissions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.permissions.name IS 'Permission name in format resource:action (e.g., user:create)';
COMMENT ON COLUMN identify.permissions.scope IS 'Scope: global (system-wide) or organization (within org)';
COMMENT ON COLUMN identify.permissions.description IS 'Permission description';
COMMENT ON COLUMN identify.permissions.resource IS 'Resource type this permission applies to';
COMMENT ON COLUMN identify.permissions.action IS 'Action allowed (view, create, update, delete, etc.)';
COMMENT ON COLUMN identify.permissions.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.roles columns
-- ============================================================================
COMMENT ON COLUMN identify.roles.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.roles.name IS 'Role name (UPPER_SNAKE_CASE), unique';
COMMENT ON COLUMN identify.roles.slug IS 'URL-friendly identifier';
COMMENT ON COLUMN identify.roles.scope IS 'Scope: global (system-wide) or organization (within org)';
COMMENT ON COLUMN identify.roles.description IS 'Role description';
COMMENT ON COLUMN identify.roles.is_system IS 'Whether this is a system-defined role that cannot be deleted';
COMMENT ON COLUMN identify.roles.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.roles.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.role_permissions columns
-- ============================================================================
COMMENT ON COLUMN identify.role_permissions.role_id IS 'Reference to role';
COMMENT ON COLUMN identify.role_permissions.permission_id IS 'Reference to permission';
COMMENT ON COLUMN identify.role_permissions.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.user_roles columns
-- ============================================================================
COMMENT ON COLUMN identify.user_roles.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.user_roles.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.user_roles.role_id IS 'Reference to role';
COMMENT ON COLUMN identify.user_roles.organization_id IS 'Organization scope (null for global roles)';
COMMENT ON COLUMN identify.user_roles.metadata IS 'Additional role assignment metadata (JSONB)';
COMMENT ON COLUMN identify.user_roles.assigned_by IS 'User ID who assigned this role';
COMMENT ON COLUMN identify.user_roles.assigned_at IS 'Timestamp when role was assigned';
COMMENT ON COLUMN identify.user_roles.expires_at IS 'Timestamp when role assignment expires';
COMMENT ON COLUMN identify.user_roles.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.user_roles.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.user_global_roles columns
-- ============================================================================
COMMENT ON COLUMN identify.user_global_roles.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.user_global_roles.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.user_global_roles.role_id IS 'Reference to global role';
COMMENT ON COLUMN identify.user_global_roles.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.user_organization_roles columns
-- ============================================================================
COMMENT ON COLUMN identify.user_organization_roles.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.user_organization_roles.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.user_organization_roles.organization_id IS 'Reference to organization';
COMMENT ON COLUMN identify.user_organization_roles.role_id IS 'Reference to organization role';
COMMENT ON COLUMN identify.user_organization_roles.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.user_statistics columns
-- ============================================================================
COMMENT ON COLUMN identify.user_statistics.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.user_statistics.user_id IS 'Reference to user, unique';
COMMENT ON COLUMN identify.user_statistics.follower_count IS 'Number of followers';
COMMENT ON COLUMN identify.user_statistics.following_count IS 'Number of users being followed';
COMMENT ON COLUMN identify.user_statistics.novel_count IS 'Number of novels created';
COMMENT ON COLUMN identify.user_statistics.manga_count IS 'Number of manga created';
COMMENT ON COLUMN identify.user_statistics.anime_count IS 'Number of anime created';
COMMENT ON COLUMN identify.user_statistics.last_content_updated_at IS 'Timestamp of last content update';
COMMENT ON COLUMN identify.user_statistics.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.user_statistics.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.sessions columns
-- ============================================================================
COMMENT ON COLUMN identify.sessions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.sessions.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.sessions.refresh_token IS 'Refresh token for session renewal';
COMMENT ON COLUMN identify.sessions.user_agent IS 'Browser/client user agent';
COMMENT ON COLUMN identify.sessions.ip_address IS 'Client IP address';
COMMENT ON COLUMN identify.sessions.expires_at IS 'Timestamp when session expires';
COMMENT ON COLUMN identify.sessions.is_revoked IS 'Whether session has been revoked';
COMMENT ON COLUMN identify.sessions.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.sessions.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.oauth2_clients columns
-- ============================================================================
COMMENT ON COLUMN identify.oauth2_clients.id IS 'Primary key using UUID v7 (client_id)';
COMMENT ON COLUMN identify.oauth2_clients.client_name IS 'Display name of the OAuth2 client';
COMMENT ON COLUMN identify.oauth2_clients.secret_hash IS 'Hashed client secret';
COMMENT ON COLUMN identify.oauth2_clients.redirect_uris IS 'Allowed redirect URIs';
COMMENT ON COLUMN identify.oauth2_clients.grant_types IS 'Allowed OAuth2 grant types';
COMMENT ON COLUMN identify.oauth2_clients.response_types IS 'Allowed OAuth2 response types';
COMMENT ON COLUMN identify.oauth2_clients.scopes IS 'Allowed OAuth2 scopes';
COMMENT ON COLUMN identify.oauth2_clients.is_public IS 'Whether client is public (no secret required)';
COMMENT ON COLUMN identify.oauth2_clients.is_internal IS 'Whether client is internal/first-party';
COMMENT ON COLUMN identify.oauth2_clients.token_endpoint_auth_method IS 'Token endpoint authentication method';
COMMENT ON COLUMN identify.oauth2_clients.organization_id IS 'Organization that owns this client';
COMMENT ON COLUMN identify.oauth2_clients.owner_user_id IS 'User that owns this client';
COMMENT ON COLUMN identify.oauth2_clients.client_uri IS 'Client application homepage URL';
COMMENT ON COLUMN identify.oauth2_clients.logo_url IS 'Client logo URL';
COMMENT ON COLUMN identify.oauth2_clients.terms_of_service_url IS 'Terms of service URL';
COMMENT ON COLUMN identify.oauth2_clients.policy_url IS 'Privacy policy URL';
COMMENT ON COLUMN identify.oauth2_clients.active IS 'Whether client is active';
COMMENT ON COLUMN identify.oauth2_clients.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.oauth2_clients.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- identify.oauth2_sessions columns
-- ============================================================================
COMMENT ON COLUMN identify.oauth2_sessions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.oauth2_sessions.signature IS 'Token signature for lookup';
COMMENT ON COLUMN identify.oauth2_sessions.request_id IS 'Original authorization request ID';
COMMENT ON COLUMN identify.oauth2_sessions.session_type IS 'Type: access_token, refresh_token, authorization_code';
COMMENT ON COLUMN identify.oauth2_sessions.session_data IS 'Serialized session data (JSONB)';
COMMENT ON COLUMN identify.oauth2_sessions.expires_at IS 'Timestamp when session expires';
COMMENT ON COLUMN identify.oauth2_sessions.client_id IS 'OAuth2 client ID';
COMMENT ON COLUMN identify.oauth2_sessions.subject_id IS 'User ID (subject)';
COMMENT ON COLUMN identify.oauth2_sessions.active IS 'Whether session is active';
COMMENT ON COLUMN identify.oauth2_sessions.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.oauth2_consents columns
-- ============================================================================
COMMENT ON COLUMN identify.oauth2_consents.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.oauth2_consents.user_id IS 'Reference to user who granted consent';
COMMENT ON COLUMN identify.oauth2_consents.client_id IS 'Reference to OAuth2 client';
COMMENT ON COLUMN identify.oauth2_consents.granted_scopes IS 'Scopes that user granted';
COMMENT ON COLUMN identify.oauth2_consents.revoked IS 'Whether consent has been revoked';
COMMENT ON COLUMN identify.oauth2_consents.granted_at IS 'Timestamp when consent was granted';
COMMENT ON COLUMN identify.oauth2_consents.revoked_at IS 'Timestamp when consent was revoked';
COMMENT ON COLUMN identify.oauth2_consents.last_used_at IS 'Timestamp when consent was last used';
COMMENT ON COLUMN identify.oauth2_consents.expires_at IS 'Timestamp when consent expires';
COMMENT ON COLUMN identify.oauth2_consents.consent_method IS 'How consent was given: explicit, implicit, remembered';
COMMENT ON COLUMN identify.oauth2_consents.ip_address IS 'Client IP when consent was given';
COMMENT ON COLUMN identify.oauth2_consents.user_agent IS 'Browser user agent when consent was given';

-- ============================================================================
-- identify.email_verification_tokens columns
-- ============================================================================
COMMENT ON COLUMN identify.email_verification_tokens.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.email_verification_tokens.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.email_verification_tokens.token IS 'Verification token';
COMMENT ON COLUMN identify.email_verification_tokens.expires_at IS 'Timestamp when token expires';
COMMENT ON COLUMN identify.email_verification_tokens.used_at IS 'Timestamp when token was used';
COMMENT ON COLUMN identify.email_verification_tokens.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.password_reset_tokens columns
-- ============================================================================
COMMENT ON COLUMN identify.password_reset_tokens.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.password_reset_tokens.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.password_reset_tokens.token IS 'Password reset token';
COMMENT ON COLUMN identify.password_reset_tokens.expires_at IS 'Timestamp when token expires';
COMMENT ON COLUMN identify.password_reset_tokens.used_at IS 'Timestamp when token was used';
COMMENT ON COLUMN identify.password_reset_tokens.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- identify.webauthn_credentials columns
-- ============================================================================
COMMENT ON COLUMN identify.webauthn_credentials.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.webauthn_credentials.user_id IS 'Reference to user';
COMMENT ON COLUMN identify.webauthn_credentials.credential_id IS 'WebAuthn credential ID';
COMMENT ON COLUMN identify.webauthn_credentials.public_key IS 'Public key bytes';
COMMENT ON COLUMN identify.webauthn_credentials.attestation_type IS 'Attestation type: none, indirect, direct';
COMMENT ON COLUMN identify.webauthn_credentials.aaguid IS 'Authenticator AAGUID';
COMMENT ON COLUMN identify.webauthn_credentials.sign_count IS 'Signature counter for replay protection';
COMMENT ON COLUMN identify.webauthn_credentials.transports IS 'Supported transports: usb, nfc, ble, internal';
COMMENT ON COLUMN identify.webauthn_credentials.backup_eligible IS 'Whether credential can be backed up';
COMMENT ON COLUMN identify.webauthn_credentials.backup_state IS 'Whether credential is currently backed up';
COMMENT ON COLUMN identify.webauthn_credentials.credential_name IS 'User-friendly credential name';
COMMENT ON COLUMN identify.webauthn_credentials.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.webauthn_credentials.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN identify.webauthn_credentials.last_used_at IS 'Timestamp when credential was last used';

-- ============================================================================
-- identify.webauthn_sessions columns
-- ============================================================================
COMMENT ON COLUMN identify.webauthn_sessions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN identify.webauthn_sessions.user_id IS 'Reference to user (null for registration)';
COMMENT ON COLUMN identify.webauthn_sessions.challenge IS 'WebAuthn challenge string';
COMMENT ON COLUMN identify.webauthn_sessions.session_type IS 'Type: registration, authentication';
COMMENT ON COLUMN identify.webauthn_sessions.user_agent IS 'Browser user agent';
COMMENT ON COLUMN identify.webauthn_sessions.ip_address IS 'Client IP address';
COMMENT ON COLUMN identify.webauthn_sessions.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN identify.webauthn_sessions.expires_at IS 'Timestamp when session expires';

-- ============================================================================
-- catalog.genres columns
-- ============================================================================
COMMENT ON COLUMN catalog.genres.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.genres.name IS 'Genre display name';
COMMENT ON COLUMN catalog.genres.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN catalog.genres.description IS 'Genre description';
COMMENT ON COLUMN catalog.genres.parent_id IS 'Parent genre for hierarchical structure';
COMMENT ON COLUMN catalog.genres.is_active IS 'Whether genre is active';
COMMENT ON COLUMN catalog.genres.novel_count IS 'Number of novels in this genre';
COMMENT ON COLUMN catalog.genres.anime_count IS 'Number of anime in this genre';
COMMENT ON COLUMN catalog.genres.manga_count IS 'Number of manga in this genre';
COMMENT ON COLUMN catalog.genres.active_readers IS 'Number of active readers';
COMMENT ON COLUMN catalog.genres.total_views IS 'Total view count';
COMMENT ON COLUMN catalog.genres.created_by IS 'User ID who created this genre';
COMMENT ON COLUMN catalog.genres.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.genres.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.genres.version IS 'Optimistic locking version number';
COMMENT ON COLUMN catalog.genres.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.genres.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.genres.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.authors columns
-- ============================================================================
COMMENT ON COLUMN catalog.authors.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.authors.user_id IS 'Linked user account (if author has account)';
COMMENT ON COLUMN catalog.authors.name IS 'Author display name';
COMMENT ON COLUMN catalog.authors.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN catalog.authors.biography IS 'Author biography (JSONB)';
COMMENT ON COLUMN catalog.authors.avatar_url IS 'URL to author avatar';
COMMENT ON COLUMN catalog.authors.social_links IS 'Social media links (JSONB)';
COMMENT ON COLUMN catalog.authors.novel_count IS 'Number of novels authored';
COMMENT ON COLUMN catalog.authors.total_chapters IS 'Total chapters across all novels';
COMMENT ON COLUMN catalog.authors.total_views IS 'Total view count across all works';
COMMENT ON COLUMN catalog.authors.follower_count IS 'Number of followers';
COMMENT ON COLUMN catalog.authors.is_verified IS 'Whether author is verified';
COMMENT ON COLUMN catalog.authors.metadata IS 'Additional metadata (JSONB)';
COMMENT ON COLUMN catalog.authors.created_by IS 'User ID who created this record';
COMMENT ON COLUMN catalog.authors.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.authors.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.authors.version IS 'Optimistic locking version number';
COMMENT ON COLUMN catalog.authors.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.authors.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.authors.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.artists columns
-- ============================================================================
COMMENT ON COLUMN catalog.artists.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.artists.user_id IS 'Linked user account (if artist has account)';
COMMENT ON COLUMN catalog.artists.name IS 'Artist display name';
COMMENT ON COLUMN catalog.artists.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN catalog.artists.biography IS 'Artist biography (JSONB)';
COMMENT ON COLUMN catalog.artists.avatar_url IS 'URL to artist avatar';
COMMENT ON COLUMN catalog.artists.social_links IS 'Social media links (JSONB)';
COMMENT ON COLUMN catalog.artists.specialization IS 'Artist specialization/style';
COMMENT ON COLUMN catalog.artists.portfolio_url IS 'Link to artist portfolio';
COMMENT ON COLUMN catalog.artists.novel_count IS 'Number of novels with artwork';
COMMENT ON COLUMN catalog.artists.artwork_count IS 'Total artworks created';
COMMENT ON COLUMN catalog.artists.follower_count IS 'Number of followers';
COMMENT ON COLUMN catalog.artists.is_verified IS 'Whether artist is verified';
COMMENT ON COLUMN catalog.artists.metadata IS 'Additional metadata (JSONB)';
COMMENT ON COLUMN catalog.artists.created_by IS 'User ID who created this record';
COMMENT ON COLUMN catalog.artists.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.artists.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.artists.version IS 'Optimistic locking version number';
COMMENT ON COLUMN catalog.artists.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.artists.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.artists.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.novels columns
-- ============================================================================
COMMENT ON COLUMN catalog.novels.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novels.title IS 'Novel title';
COMMENT ON COLUMN catalog.novels.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN catalog.novels.owner_id IS 'Owner ID (user or organization)';
COMMENT ON COLUMN catalog.novels.owner_type IS 'Owner type: user, organization';
COMMENT ON COLUMN catalog.novels.synopsis IS 'Novel synopsis/description (JSONB)';
COMMENT ON COLUMN catalog.novels.cover_image_url IS 'URL to cover image';
COMMENT ON COLUMN catalog.novels.thumbnail_url IS 'URL to thumbnail image';
COMMENT ON COLUMN catalog.novels.status IS 'Status: draft, ongoing, completed, hiatus, dropped';
COMMENT ON COLUMN catalog.novels.is_oneshot IS 'Whether novel is a one-shot';
COMMENT ON COLUMN catalog.novels.original_language IS 'Original language code';
COMMENT ON COLUMN catalog.novels.original_title IS 'Title in original language';
COMMENT ON COLUMN catalog.novels.total_volumes IS 'Total number of volumes';
COMMENT ON COLUMN catalog.novels.total_chapters IS 'Total number of chapters';
COMMENT ON COLUMN catalog.novels.total_words IS 'Total word count';
COMMENT ON COLUMN catalog.novels.view_count IS 'Total view count';
COMMENT ON COLUMN catalog.novels.favorite_count IS 'Number of users who favorited';
COMMENT ON COLUMN catalog.novels.rating_average IS 'Average rating (0-5)';
COMMENT ON COLUMN catalog.novels.rating_count IS 'Number of ratings';
COMMENT ON COLUMN catalog.novels.metadata IS 'Additional metadata (JSONB)';
COMMENT ON COLUMN catalog.novels.first_published_at IS 'Timestamp of first publication';
COMMENT ON COLUMN catalog.novels.last_chapter_at IS 'Timestamp of last chapter release';
COMMENT ON COLUMN catalog.novels.completed_at IS 'Timestamp when novel was completed';
COMMENT ON COLUMN catalog.novels.created_by IS 'User ID who created this novel';
COMMENT ON COLUMN catalog.novels.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.novels.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.novels.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.novels.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.novels.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.novel_volumes columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_volumes.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_volumes.novel_id IS 'Reference to parent novel';
COMMENT ON COLUMN catalog.novel_volumes.volume_number IS 'Volume number in sequence';
COMMENT ON COLUMN catalog.novel_volumes.title IS 'Volume title';
COMMENT ON COLUMN catalog.novel_volumes.slug IS 'URL-friendly identifier';
COMMENT ON COLUMN catalog.novel_volumes.description IS 'Volume description';
COMMENT ON COLUMN catalog.novel_volumes.cover_image_url IS 'URL to volume cover image';
COMMENT ON COLUMN catalog.novel_volumes.chapter_count IS 'Number of chapters in volume';
COMMENT ON COLUMN catalog.novel_volumes.word_count IS 'Total word count in volume';
COMMENT ON COLUMN catalog.novel_volumes.display_order IS 'Display order for sorting';
COMMENT ON COLUMN catalog.novel_volumes.is_published IS 'Whether volume is published';
COMMENT ON COLUMN catalog.novel_volumes.published_at IS 'Timestamp when published';
COMMENT ON COLUMN catalog.novel_volumes.created_by IS 'User ID who created this volume';
COMMENT ON COLUMN catalog.novel_volumes.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.novel_volumes.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.novel_volumes.version IS 'Optimistic locking version number';
COMMENT ON COLUMN catalog.novel_volumes.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.novel_volumes.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.novel_volumes.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.novel_chapters columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_chapters.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_chapters.novel_id IS 'Reference to parent novel';
COMMENT ON COLUMN catalog.novel_chapters.volume_id IS 'Reference to parent volume (optional)';
COMMENT ON COLUMN catalog.novel_chapters.chapter_number IS 'Chapter number in sequence';
COMMENT ON COLUMN catalog.novel_chapters.title IS 'Chapter title';
COMMENT ON COLUMN catalog.novel_chapters.slug IS 'URL-friendly identifier';
COMMENT ON COLUMN catalog.novel_chapters.content IS 'Chapter content (JSONB)';
COMMENT ON COLUMN catalog.novel_chapters.word_count IS 'Word count of chapter';
COMMENT ON COLUMN catalog.novel_chapters.character_count IS 'Character count of chapter';
COMMENT ON COLUMN catalog.novel_chapters.is_free IS 'Whether chapter is free to read';
COMMENT ON COLUMN catalog.novel_chapters.price IS 'Price to unlock (if not free)';
COMMENT ON COLUMN catalog.novel_chapters.currency IS 'Currency for price';
COMMENT ON COLUMN catalog.novel_chapters.status IS 'Status: draft, published, scheduled';
COMMENT ON COLUMN catalog.novel_chapters.view_count IS 'Total view count';
COMMENT ON COLUMN catalog.novel_chapters.like_count IS 'Number of likes';
COMMENT ON COLUMN catalog.novel_chapters.comment_count IS 'Number of comments';
COMMENT ON COLUMN catalog.novel_chapters.display_order IS 'Display order for sorting';
COMMENT ON COLUMN catalog.novel_chapters.author_notes IS 'Author notes (JSONB)';
COMMENT ON COLUMN catalog.novel_chapters.published_at IS 'Timestamp when published';
COMMENT ON COLUMN catalog.novel_chapters.scheduled_at IS 'Timestamp for scheduled publishing';
COMMENT ON COLUMN catalog.novel_chapters.created_by IS 'User ID who created this chapter';
COMMENT ON COLUMN catalog.novel_chapters.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN catalog.novel_chapters.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN catalog.novel_chapters.version IS 'Optimistic locking version number';
COMMENT ON COLUMN catalog.novel_chapters.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.novel_chapters.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN catalog.novel_chapters.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.novel_genres columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_genres.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_genres.novel_id IS 'Reference to novel';
COMMENT ON COLUMN catalog.novel_genres.genre_id IS 'Reference to genre';
COMMENT ON COLUMN catalog.novel_genres.display_order IS 'Display order for sorting';
COMMENT ON COLUMN catalog.novel_genres.created_by IS 'User ID who created this association';
COMMENT ON COLUMN catalog.novel_genres.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- catalog.novel_authors columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_authors.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_authors.novel_id IS 'Reference to novel';
COMMENT ON COLUMN catalog.novel_authors.author_id IS 'Reference to author';
COMMENT ON COLUMN catalog.novel_authors.role IS 'Author role: original_author, co_author';
COMMENT ON COLUMN catalog.novel_authors.display_order IS 'Display order for sorting';
COMMENT ON COLUMN catalog.novel_authors.created_by IS 'User ID who created this association';
COMMENT ON COLUMN catalog.novel_authors.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- catalog.novel_artists columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_artists.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_artists.novel_id IS 'Reference to novel';
COMMENT ON COLUMN catalog.novel_artists.artist_id IS 'Reference to artist';
COMMENT ON COLUMN catalog.novel_artists.role IS 'Artist role: cover_artist, illustrator, character_designer';
COMMENT ON COLUMN catalog.novel_artists.display_order IS 'Display order for sorting';
COMMENT ON COLUMN catalog.novel_artists.created_by IS 'User ID who created this association';
COMMENT ON COLUMN catalog.novel_artists.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- catalog.novel_embeddings columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_embeddings.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_embeddings.novel_id IS 'Reference to novel, unique';
COMMENT ON COLUMN catalog.novel_embeddings.embedding IS 'Vector embedding for similarity search';
COMMENT ON COLUMN catalog.novel_embeddings.model_version IS 'Version of embedding model used';
COMMENT ON COLUMN catalog.novel_embeddings.source_hash IS 'Hash of source content for change detection';
COMMENT ON COLUMN catalog.novel_embeddings.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.novel_embeddings.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- catalog.novel_chapter_histories columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_chapter_histories.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_chapter_histories.chapter_id IS 'Reference to chapter';
COMMENT ON COLUMN catalog.novel_chapter_histories.volume_id IS 'Reference to volume at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.novel_id IS 'Reference to novel';
COMMENT ON COLUMN catalog.novel_chapter_histories.version_number IS 'Version number of this history entry';
COMMENT ON COLUMN catalog.novel_chapter_histories.action IS 'Action: created, updated, published, deleted';
COMMENT ON COLUMN catalog.novel_chapter_histories.title IS 'Chapter title at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.slug IS 'Chapter slug at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.chapter_number IS 'Chapter number at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.status IS 'Status at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.word_count IS 'Word count at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.character_count IS 'Character count at time of change';
COMMENT ON COLUMN catalog.novel_chapter_histories.changed_fields IS 'List of changed fields (JSONB)';
COMMENT ON COLUMN catalog.novel_chapter_histories.change_summary IS 'Summary of changes';
COMMENT ON COLUMN catalog.novel_chapter_histories.content_changed IS 'Whether content was changed';
COMMENT ON COLUMN catalog.novel_chapter_histories.changed_by IS 'User ID who made the change';
COMMENT ON COLUMN catalog.novel_chapter_histories.request_id IS 'Request ID for tracing';
COMMENT ON COLUMN catalog.novel_chapter_histories.ip_address IS 'IP address of requester';
COMMENT ON COLUMN catalog.novel_chapter_histories.user_agent IS 'User agent of requester';
COMMENT ON COLUMN catalog.novel_chapter_histories.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- catalog.novel_volume_histories columns
-- ============================================================================
COMMENT ON COLUMN catalog.novel_volume_histories.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.novel_volume_histories.volume_id IS 'Reference to volume';
COMMENT ON COLUMN catalog.novel_volume_histories.novel_id IS 'Reference to novel';
COMMENT ON COLUMN catalog.novel_volume_histories.version_number IS 'Version number of this history entry';
COMMENT ON COLUMN catalog.novel_volume_histories.action IS 'Action: created, updated, published, unpublished, deleted';
COMMENT ON COLUMN catalog.novel_volume_histories.title IS 'Volume title at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.slug IS 'Volume slug at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.volume_number IS 'Volume number at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.is_published IS 'Published state at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.chapter_count IS 'Chapter count at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.word_count IS 'Word count at time of change';
COMMENT ON COLUMN catalog.novel_volume_histories.changed_fields IS 'List of changed fields (JSONB)';
COMMENT ON COLUMN catalog.novel_volume_histories.change_summary IS 'Summary of changes';
COMMENT ON COLUMN catalog.novel_volume_histories.changed_by IS 'User ID who made the change';
COMMENT ON COLUMN catalog.novel_volume_histories.request_id IS 'Request ID for tracing';
COMMENT ON COLUMN catalog.novel_volume_histories.ip_address IS 'IP address of requester';
COMMENT ON COLUMN catalog.novel_volume_histories.user_agent IS 'User agent of requester';
COMMENT ON COLUMN catalog.novel_volume_histories.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- payment.user_wallets columns
-- ============================================================================
COMMENT ON COLUMN payment.user_wallets.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN payment.user_wallets.user_id IS 'Reference to user, unique';
COMMENT ON COLUMN payment.user_wallets.coin_balance IS 'Current coin balance';
COMMENT ON COLUMN payment.user_wallets.total_deposited IS 'Total coins ever deposited';
COMMENT ON COLUMN payment.user_wallets.total_spent IS 'Total coins ever spent';
COMMENT ON COLUMN payment.user_wallets.total_subscription_spent IS 'Total coins spent on subscriptions';
COMMENT ON COLUMN payment.user_wallets.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN payment.user_wallets.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- payment.coin_packages columns
-- ============================================================================
COMMENT ON COLUMN payment.coin_packages.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN payment.coin_packages.name IS 'Package display name';
COMMENT ON COLUMN payment.coin_packages.slug IS 'URL-friendly identifier, unique';
COMMENT ON COLUMN payment.coin_packages.coin_amount IS 'Base coin amount in package';
COMMENT ON COLUMN payment.coin_packages.price_vnd IS 'Price in VND';
COMMENT ON COLUMN payment.coin_packages.bonus_percent IS 'Bonus percentage on top of base';
COMMENT ON COLUMN payment.coin_packages.is_popular IS 'Whether to highlight as popular';
COMMENT ON COLUMN payment.coin_packages.is_active IS 'Whether package is available for purchase';
COMMENT ON COLUMN payment.coin_packages.display_order IS 'Display order for sorting';
COMMENT ON COLUMN payment.coin_packages.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN payment.coin_packages.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- payment.topup_orders columns
-- ============================================================================
COMMENT ON COLUMN payment.topup_orders.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN payment.topup_orders.user_id IS 'Reference to user making purchase';
COMMENT ON COLUMN payment.topup_orders.package_id IS 'Reference to coin package';
COMMENT ON COLUMN payment.topup_orders.order_code IS 'Unique order code for payment reference';
COMMENT ON COLUMN payment.topup_orders.coin_amount IS 'Total coins to receive (base + bonus)';
COMMENT ON COLUMN payment.topup_orders.base_coin_amount IS 'Base coin amount before bonus';
COMMENT ON COLUMN payment.topup_orders.bonus_coin_amount IS 'Bonus coin amount';
COMMENT ON COLUMN payment.topup_orders.vnd_amount IS 'Amount in VND to pay';
COMMENT ON COLUMN payment.topup_orders.status IS 'Order status: pending, success, expired, cancelled, failed';
COMMENT ON COLUMN payment.topup_orders.sepay_transaction_id IS 'SePay transaction ID';
COMMENT ON COLUMN payment.topup_orders.sepay_content IS 'SePay transfer content';
COMMENT ON COLUMN payment.topup_orders.bank_name IS 'Destination bank name';
COMMENT ON COLUMN payment.topup_orders.bank_account IS 'Destination bank account number';
COMMENT ON COLUMN payment.topup_orders.account_name IS 'Destination account holder name';
COMMENT ON COLUMN payment.topup_orders.completed_at IS 'Timestamp when order was completed';
COMMENT ON COLUMN payment.topup_orders.expired_at IS 'Timestamp when order expires';
COMMENT ON COLUMN payment.topup_orders.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN payment.topup_orders.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- payment.transactions columns
-- ============================================================================
COMMENT ON COLUMN payment.transactions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN payment.transactions.user_id IS 'Reference to user';
COMMENT ON COLUMN payment.transactions.type IS 'Transaction type: topup, purchase_chapter, etc.';
COMMENT ON COLUMN payment.transactions.coin_amount IS 'Coin amount (positive for credit, negative for debit)';
COMMENT ON COLUMN payment.transactions.vnd_amount IS 'Equivalent VND amount';
COMMENT ON COLUMN payment.transactions.balance_after IS 'Wallet balance after transaction';
COMMENT ON COLUMN payment.transactions.reference_type IS 'Type of referenced entity: chapter, series, etc.';
COMMENT ON COLUMN payment.transactions.reference_id IS 'ID of referenced entity';
COMMENT ON COLUMN payment.transactions.creator_id IS 'Content creator who receives revenue';
COMMENT ON COLUMN payment.transactions.creator_revenue_vnd IS 'Revenue share for creator in VND';
COMMENT ON COLUMN payment.transactions.platform_revenue_vnd IS 'Platform fee in VND';
COMMENT ON COLUMN payment.transactions.description IS 'Transaction description';
COMMENT ON COLUMN payment.transactions.metadata IS 'Additional metadata (JSONB)';
COMMENT ON COLUMN payment.transactions.created_at IS 'Timestamp when record was created';

-- ============================================================================
-- payment.payment_configurations columns
-- ============================================================================
COMMENT ON COLUMN payment.payment_configurations.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN payment.payment_configurations.key IS 'Configuration key, unique';
COMMENT ON COLUMN payment.payment_configurations.value IS 'Configuration value';
COMMENT ON COLUMN payment.payment_configurations.value_type IS 'Value type: string, number, boolean, json';
COMMENT ON COLUMN payment.payment_configurations.description IS 'Description of this configuration';
COMMENT ON COLUMN payment.payment_configurations.is_sensitive IS 'Whether value should be masked';
COMMENT ON COLUMN payment.payment_configurations.updated_by IS 'User who last updated this config';
COMMENT ON COLUMN payment.payment_configurations.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN payment.payment_configurations.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- community.novel_chapter_translations columns
-- ============================================================================
COMMENT ON COLUMN community.novel_chapter_translations.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN community.novel_chapter_translations.chapter_id IS 'Reference to original chapter';
COMMENT ON COLUMN community.novel_chapter_translations.language IS 'Target language code';
COMMENT ON COLUMN community.novel_chapter_translations.title IS 'Translated chapter title';
COMMENT ON COLUMN community.novel_chapter_translations.content IS 'Translated chapter content';
COMMENT ON COLUMN community.novel_chapter_translations.translator_notes IS 'Notes from translator';
COMMENT ON COLUMN community.novel_chapter_translations.organization_id IS 'Organization that made this translation';
COMMENT ON COLUMN community.novel_chapter_translations.version IS 'Translation version number';
COMMENT ON COLUMN community.novel_chapter_translations.status IS 'Status: draft, pending_review, published';
COMMENT ON COLUMN community.novel_chapter_translations.word_count IS 'Word count of translation';
COMMENT ON COLUMN community.novel_chapter_translations.character_count IS 'Character count of translation';
COMMENT ON COLUMN community.novel_chapter_translations.quality_score IS 'AI-calculated quality score';
COMMENT ON COLUMN community.novel_chapter_translations.reviewer_rating IS 'Rating given by human reviewer';
COMMENT ON COLUMN community.novel_chapter_translations.view_count IS 'Total view count';
COMMENT ON COLUMN community.novel_chapter_translations.like_count IS 'Number of likes';
COMMENT ON COLUMN community.novel_chapter_translations.comment_count IS 'Number of comments';
COMMENT ON COLUMN community.novel_chapter_translations.contribution_count IS 'Number of community contributions';
COMMENT ON COLUMN community.novel_chapter_translations.reviewed_by IS 'User ID who reviewed this translation';
COMMENT ON COLUMN community.novel_chapter_translations.review_notes IS 'Notes from reviewer';
COMMENT ON COLUMN community.novel_chapter_translations.reviewed_at IS 'Timestamp when reviewed';
COMMENT ON COLUMN community.novel_chapter_translations.published_at IS 'Timestamp when published';
COMMENT ON COLUMN community.novel_chapter_translations.created_by IS 'User ID who created this translation';
COMMENT ON COLUMN community.novel_chapter_translations.updated_by IS 'User ID who last updated this record';
COMMENT ON COLUMN community.novel_chapter_translations.deleted_by IS 'User ID who deleted this record';
COMMENT ON COLUMN community.novel_chapter_translations.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN community.novel_chapter_translations.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN community.novel_chapter_translations.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- community.translation_contributions columns
-- ============================================================================
COMMENT ON COLUMN community.translation_contributions.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN community.translation_contributions.chapter_id IS 'Reference to original chapter';
COMMENT ON COLUMN community.translation_contributions.contributor_id IS 'User ID who made this contribution';
COMMENT ON COLUMN community.translation_contributions.language IS 'Target language code';
COMMENT ON COLUMN community.translation_contributions.contribution_type IS 'Type: full, partial, correction';
COMMENT ON COLUMN community.translation_contributions.title IS 'Contributed title translation';
COMMENT ON COLUMN community.translation_contributions.content IS 'Contributed content translation';
COMMENT ON COLUMN community.translation_contributions.contributor_notes IS 'Notes from contributor';
COMMENT ON COLUMN community.translation_contributions.status IS 'Status: draft, pending_review, approved, rejected';
COMMENT ON COLUMN community.translation_contributions.reviewed_by IS 'User ID who reviewed this contribution';
COMMENT ON COLUMN community.translation_contributions.reviewed_at IS 'Timestamp when reviewed';
COMMENT ON COLUMN community.translation_contributions.review_notes IS 'Notes from reviewer';
COMMENT ON COLUMN community.translation_contributions.official_translation_id IS 'Official translation this was merged into';
COMMENT ON COLUMN community.translation_contributions.credit_points IS 'Points earned for this contribution';
COMMENT ON COLUMN community.translation_contributions.is_credited IS 'Whether contributor has been credited';
COMMENT ON COLUMN community.translation_contributions.word_count IS 'Word count of contribution';
COMMENT ON COLUMN community.translation_contributions.character_count IS 'Character count of contribution';
COMMENT ON COLUMN community.translation_contributions.upvote_count IS 'Number of upvotes';
COMMENT ON COLUMN community.translation_contributions.downvote_count IS 'Number of downvotes';
COMMENT ON COLUMN community.translation_contributions.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN community.translation_contributions.updated_at IS 'Timestamp when record was last updated';
COMMENT ON COLUMN community.translation_contributions.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- catalog.media_progress columns
-- ============================================================================
COMMENT ON COLUMN catalog.media_progress.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.media_progress.user_id IS 'Reference to user';
COMMENT ON COLUMN catalog.media_progress.media_type IS 'Type: novel, manga, anime';
COMMENT ON COLUMN catalog.media_progress.media_id IS 'Reference to media (novel, manga, anime)';
COMMENT ON COLUMN catalog.media_progress.current_unit_id IS 'Current chapter/episode ID';
COMMENT ON COLUMN catalog.media_progress.position IS 'Position within current unit (JSONB)';
COMMENT ON COLUMN catalog.media_progress.total_units IS 'Total units in media';
COMMENT ON COLUMN catalog.media_progress.completed_units IS 'Number of completed units';
COMMENT ON COLUMN catalog.media_progress.progress_percentage IS 'Overall progress percentage (0-100)';
COMMENT ON COLUMN catalog.media_progress.last_accessed_at IS 'Timestamp of last access';
COMMENT ON COLUMN catalog.media_progress.created_at IS 'Timestamp when record was created';
COMMENT ON COLUMN catalog.media_progress.updated_at IS 'Timestamp when record was last updated';

-- ============================================================================
-- catalog.unit_progress columns
-- ============================================================================
COMMENT ON COLUMN catalog.unit_progress.id IS 'Primary key using UUID v7';
COMMENT ON COLUMN catalog.unit_progress.user_id IS 'Reference to user';
COMMENT ON COLUMN catalog.unit_progress.media_type IS 'Type: novel, manga, anime';
COMMENT ON COLUMN catalog.unit_progress.media_id IS 'Reference to media';
COMMENT ON COLUMN catalog.unit_progress.unit_id IS 'Reference to unit (chapter/episode)';
COMMENT ON COLUMN catalog.unit_progress.status IS 'Status: in_progress, completed';
COMMENT ON COLUMN catalog.unit_progress.position IS 'Position within unit (JSONB)';
COMMENT ON COLUMN catalog.unit_progress.started_at IS 'Timestamp when started reading/watching';
COMMENT ON COLUMN catalog.unit_progress.completed_at IS 'Timestamp when completed';
COMMENT ON COLUMN catalog.unit_progress.last_accessed_at IS 'Timestamp of last access';
