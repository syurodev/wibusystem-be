CREATE TABLE IF NOT EXISTS users
(
    id                      uuid                     DEFAULT uuidv7( )                   NOT NULL PRIMARY KEY,
    email                   VARCHAR(255)                                                 NOT NULL UNIQUE,
    email_verified          BOOLEAN                  DEFAULT FALSE                       NOT NULL,
    password_hash           VARCHAR(255),
    full_name               VARCHAR(255),
    avatar_url              TEXT,
    phone                   VARCHAR(50),
    status                  VARCHAR(50)              DEFAULT 'active'::CHARACTER VARYING NOT NULL
        CONSTRAINT users_status_check CHECK ((status)::TEXT = ANY
                                             ((ARRAY ['active'::CHARACTER VARYING, 'suspended'::CHARACTER VARYING, 'deleted'::CHARACTER VARYING])::TEXT[])),
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                      NOT NULL,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                      NOT NULL,
    last_login_at           TIMESTAMP WITH TIME ZONE,
    settings                jsonb                    DEFAULT '{}'::jsonb,
    display_name            VARCHAR(255),
    username                VARCHAR(100),
    bio                     jsonb,
    is_verified             BOOLEAN                  DEFAULT FALSE,
    follower_count          INTEGER                  DEFAULT 0,
    works_count             INTEGER                  DEFAULT 0,
    last_content_updated_at TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE users IS 'Bảng toàn cục chứa tất cả người dùng trong hệ thống';

COMMENT ON COLUMN users.email_verified IS 'Trạng thái xác thực email';

COMMENT ON COLUMN users.password_hash IS 'Password hash (NULL for passwordless accounts using only WebAuthn)';

COMMENT ON COLUMN users.settings IS 'User preferences (theme, language, notifications, etc.)';

ALTER TABLE users
    OWNER TO system_dev;

CREATE INDEX idx_users_email ON users ( email );

CREATE INDEX idx_users_status ON users ( status );

CREATE INDEX idx_users_created_at ON users ( created_at );

CREATE INDEX idx_users_last_login_at ON users ( last_login_at );

CREATE UNIQUE INDEX idx_users_username ON users ( username ) WHERE (username IS NOT NULL);

CREATE INDEX idx_users_is_verified ON users ( is_verified );

CREATE INDEX idx_users_last_content_updated_at ON users ( last_content_updated_at );

CREATE INDEX idx_users_follower_count ON users ( follower_count );

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column( );

CREATE TABLE IF NOT EXISTS organizations
(
    id                     uuid                     DEFAULT uuidv7( )                   NOT NULL PRIMARY KEY,
    name                   VARCHAR(255)                                                 NOT NULL,
    slug                   VARCHAR(255)                                                 NOT NULL UNIQUE,
    status                 VARCHAR(50)              DEFAULT 'active'::CHARACTER VARYING NOT NULL
        CONSTRAINT organizations_status_check CHECK ((status)::TEXT = ANY
                                                     ((ARRAY ['active'::CHARACTER VARYING, 'suspended'::CHARACTER VARYING, 'archived'::CHARACTER VARYING])::TEXT[])),
    settings               jsonb,
    description            jsonb,
    avatar_url             VARCHAR(1000),
    is_recruiting          BOOLEAN                  DEFAULT FALSE                       NOT NULL,
    can_translate          BOOLEAN                  DEFAULT TRUE                        NOT NULL,
    can_proofread          BOOLEAN                  DEFAULT TRUE                        NOT NULL,
    can_edit               BOOLEAN                  DEFAULT TRUE                        NOT NULL,
    member_count           INTEGER                  DEFAULT 0                           NOT NULL
        CONSTRAINT organizations_member_count_check CHECK (member_count >= 0),
    active_projects        INTEGER                  DEFAULT 0                           NOT NULL
        CONSTRAINT organizations_active_projects_check CHECK (active_projects >= 0),
    completed_translations INTEGER                  DEFAULT 0                           NOT NULL
        CONSTRAINT organizations_completed_translations_check CHECK (completed_translations >= 0),
    metadata               jsonb                    DEFAULT '{}'::jsonb,
    created_by             uuid REFERENCES users ON DELETE RESTRICT,
    updated_by             uuid                                                         REFERENCES users ON DELETE SET NULL,
    deleted_by             uuid                                                         REFERENCES users ON DELETE SET NULL,
    version                INTEGER                  DEFAULT 1                           NOT NULL,
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                      NOT NULL,
    updated_at             TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                      NOT NULL,
    deleted_at             TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE organizations IS 'Bảng lưu trữ thông tin về các tổ chức (organizations/translation teams) trong hệ thống';

COMMENT ON COLUMN organizations.slug IS 'URL-friendly identifier cho organization';

COMMENT ON COLUMN organizations.settings IS 'Cấu hình tùy chỉnh cho organization (JSONB format)';

COMMENT ON COLUMN organizations.description IS 'Mô tả organization (Plate editor JSON output)';

COMMENT ON COLUMN organizations.can_translate IS 'Quyền dịch nội dung';

COMMENT ON COLUMN organizations.can_proofread IS 'Quyền hiệu đính bản dịch';

COMMENT ON COLUMN organizations.can_edit IS 'Quyền chỉnh sửa bản dịch';

ALTER TABLE organizations
    OWNER TO system_dev;

CREATE INDEX idx_organizations_slug ON organizations ( slug ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_organizations_status ON organizations ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_organizations_created_at ON organizations ( created_at ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_organizations_recruiting ON organizations ( is_recruiting ) WHERE ((is_recruiting = TRUE) AND (deleted_at IS NULL));

CREATE INDEX idx_organizations_metadata ON organizations USING gin ( metadata ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_organizations_description ON organizations USING gin ( description ) WHERE (deleted_at IS NULL);

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE
    ON organizations
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column( );

CREATE TABLE IF NOT EXISTS permissions
(
    id          uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name        VARCHAR(255)                               NOT NULL UNIQUE
        CONSTRAINT permissions_name_format_check CHECK ((name)::TEXT ~ '^[a-z_]+:[a-z_]+$'::TEXT),
    scope       permission_scope                           NOT NULL,
    description TEXT,
    resource    VARCHAR(100),
    action      VARCHAR(100),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL
);

COMMENT ON TABLE permissions IS 'Seeded với permissions cơ bản cho system và organization operations';

COMMENT ON COLUMN permissions.name IS 'Tên unique của permission, format: resource:action (e.g., user:create)';

COMMENT ON COLUMN permissions.scope IS 'Phạm vi áp dụng: global hoặc tenant';

COMMENT ON COLUMN permissions.resource IS 'Resource type mà permission này áp dụng';

COMMENT ON COLUMN permissions.action IS 'Hành động được phép (view, create, update, delete, etc.)';

ALTER TABLE permissions
    OWNER TO system_dev;

CREATE INDEX idx_permissions_scope ON permissions ( scope );

CREATE INDEX idx_permissions_resource ON permissions ( resource );

CREATE INDEX idx_permissions_name ON permissions ( name );

CREATE TABLE IF NOT EXISTS roles
(
    id          uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    name        VARCHAR(255)                               NOT NULL UNIQUE
        CONSTRAINT roles_name_format_check CHECK ((name)::TEXT ~ '^[A-Z_]+$'::TEXT),
    slug        VARCHAR(255)                               NOT NULL UNIQUE,
    scope       role_scope                                 NOT NULL,
    description TEXT,
    is_system   BOOLEAN                  DEFAULT FALSE     NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL
);

COMMENT ON TABLE roles IS 'Seeded với roles cơ bản: SUPER_ADMIN, PLATFORM_ADMIN, ORGANIZATION_ADMIN, ORGANIZATION_MANAGER, ORGANIZATION_MEMBER, ORGANIZATION_VIEWER';

COMMENT ON COLUMN roles.name IS 'Tên unique của role (UPPER_SNAKE_CASE)';

COMMENT ON COLUMN roles.slug IS 'URL-friendly identifier cho role';

COMMENT ON COLUMN roles.scope IS 'Phạm vi áp dụng: global hoặc tenant';

COMMENT ON COLUMN roles.is_system IS 'System role không thể xóa hoặc sửa đổi';

ALTER TABLE roles
    OWNER TO system_dev;

CREATE INDEX idx_roles_scope ON roles ( scope );

CREATE INDEX idx_roles_slug ON roles ( slug );

CREATE INDEX idx_roles_is_system ON roles ( is_system );

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE
    ON roles
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column( );

CREATE TABLE IF NOT EXISTS role_permissions
(
    role_id       uuid                                    NOT NULL REFERENCES roles ON DELETE CASCADE,
    permission_id uuid                                    NOT NULL REFERENCES permissions ON DELETE CASCADE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW( ) NOT NULL,
    PRIMARY KEY ( role_id, permission_id )
);

COMMENT ON TABLE role_permissions IS 'Seeded với permission mappings cho tất cả default roles';

COMMENT ON COLUMN role_permissions.role_id IS 'ID của role';

COMMENT ON COLUMN role_permissions.permission_id IS 'ID của permission được gán cho role';

ALTER TABLE role_permissions
    OWNER TO system_dev;

CREATE INDEX idx_role_permissions_role_id ON role_permissions ( role_id );

CREATE INDEX idx_role_permissions_permission_id ON role_permissions ( permission_id );

CREATE TABLE IF NOT EXISTS user_organization_memberships
(
    user_id            uuid                                                                NOT NULL REFERENCES users ON DELETE CASCADE,
    organization_id    uuid                                                                NOT NULL REFERENCES organizations ON DELETE CASCADE,
    status             VARCHAR(50)              DEFAULT 'active'::CHARACTER VARYING        NOT NULL
        CONSTRAINT user_organization_memberships_status_check CHECK ((status)::TEXT = ANY
                                                                     ((ARRAY ['active'::CHARACTER VARYING, 'pending_invite'::CHARACTER VARYING, 'suspended'::CHARACTER VARYING])::TEXT[])),
    role               organization_member_role DEFAULT 'member'::organization_member_role NOT NULL,
    is_active          BOOLEAN                  DEFAULT TRUE                               NOT NULL,
    contribution_count INTEGER                  DEFAULT 0                                  NOT NULL
        CONSTRAINT user_organization_memberships_contribution_count_check CHECK (contribution_count >= 0),
    quality_score      NUMERIC(3, 2)            DEFAULT 0.00
        CONSTRAINT user_organization_memberships_quality_score_check CHECK ((quality_score >= (0)::NUMERIC) AND (quality_score <= (5)::NUMERIC)),
    metadata           jsonb                    DEFAULT '{}'::jsonb,
    invited_by         uuid                                                                REFERENCES users ON DELETE SET NULL,
    invited_at         TIMESTAMP WITH TIME ZONE,
    joined_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW( ),
    left_at            TIMESTAMP WITH TIME ZONE,
    created_by         uuid                                                                REFERENCES users ON DELETE SET NULL,
    updated_by         uuid                                                                REFERENCES users ON DELETE SET NULL,
    deleted_by         uuid                                                                REFERENCES users ON DELETE SET NULL,
    version            INTEGER                  DEFAULT 1                                  NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                             NOT NULL,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                             NOT NULL,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY ( user_id, organization_id )
);

COMMENT ON TABLE user_organization_memberships IS 'Bảng liên kết users với organizations - một user có thể thuộc nhiều organizations. Merged with team_members.';

COMMENT ON COLUMN user_organization_memberships.status IS 'Trạng thái membership: active (đang hoạt động), pending_invite (chờ chấp nhận), suspended (bị đình chỉ)';

COMMENT ON COLUMN user_organization_memberships.role IS 'Vai trò trong organization: leader, translator, proofreader, editor, member';

COMMENT ON COLUMN user_organization_memberships.quality_score IS 'Điểm chất lượng trung bình từ reviewers (thang điểm 0-5)';

COMMENT ON COLUMN user_organization_memberships.invited_by IS 'User đã mời thành viên này vào organization';

ALTER TABLE user_organization_memberships
    OWNER TO system_dev;

CREATE INDEX idx_user_organization_memberships_user_id ON user_organization_memberships ( user_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_user_organization_memberships_organization_id ON user_organization_memberships ( organization_id ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_user_organization_memberships_status ON user_organization_memberships ( status ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_user_organization_memberships_invited_by ON user_organization_memberships ( invited_by ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_user_organization_memberships_role ON user_organization_memberships ( organization_id, role ) WHERE (deleted_at IS NULL);

CREATE INDEX idx_user_organization_memberships_active ON user_organization_memberships ( organization_id, is_active ) WHERE ((is_active = TRUE) AND (deleted_at IS NULL));

CREATE TRIGGER update_user_organization_memberships_updated_at
    BEFORE UPDATE
    ON user_organization_memberships
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column( );

CREATE TABLE IF NOT EXISTS user_organization_roles
(
    user_id         uuid                                    NOT NULL,
    organization_id uuid                                    NOT NULL,
    role_id         uuid                                    NOT NULL REFERENCES roles ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW( ) NOT NULL,
    PRIMARY KEY ( user_id, organization_id, role_id ),
    FOREIGN KEY ( user_id, organization_id ) REFERENCES user_organization_memberships ON DELETE CASCADE
);

COMMENT ON TABLE user_organization_roles IS 'Gán RBAC roles cho user trong context của một organization cụ thể. QUAN TRỌNG: Application phải đảm bảo chỉ gán organization-scoped roles';

ALTER TABLE user_organization_roles
    OWNER TO system_dev;

CREATE INDEX idx_user_organization_roles_user_id ON user_organization_roles ( user_id );

CREATE INDEX idx_user_organization_roles_organization_id ON user_organization_roles ( organization_id );

CREATE INDEX idx_user_organization_roles_role_id ON user_organization_roles ( role_id );

CREATE INDEX idx_user_organization_roles_user_organization ON user_organization_roles ( user_id, organization_id );

CREATE TABLE IF NOT EXISTS user_global_roles
(
    user_id    uuid                                    NOT NULL REFERENCES users ON DELETE CASCADE,
    role_id    uuid                                    NOT NULL REFERENCES roles ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW( ) NOT NULL,
    PRIMARY KEY ( user_id, role_id )
);

COMMENT ON TABLE user_global_roles IS 'Gán global roles cho user, độc lập với organization. QUAN TRỌNG: Application phải đảm bảo chỉ gán global-scoped roles';

ALTER TABLE user_global_roles
    OWNER TO system_dev;

CREATE INDEX idx_user_global_roles_user_id ON user_global_roles ( user_id );

CREATE INDEX idx_user_global_roles_role_id ON user_global_roles ( role_id );

CREATE TABLE IF NOT EXISTS oauth2_clients
(
    id                         uuid                     DEFAULT uuidv7( )                                NOT NULL PRIMARY KEY,
    client_name                VARCHAR(255)                                                              NOT NULL,
    secret_hash                VARCHAR(255)                                                              NOT NULL,
    redirect_uris              TEXT[]                   DEFAULT '{}'::TEXT[]                             NOT NULL,
    grant_types                TEXT[]                   DEFAULT '{}'::TEXT[]                             NOT NULL,
    response_types             TEXT[]                   DEFAULT '{}'::TEXT[]                             NOT NULL,
    scopes                     TEXT[]                   DEFAULT '{}'::TEXT[]                             NOT NULL,
    is_public                  BOOLEAN                  DEFAULT FALSE                                    NOT NULL,
    organization_id            uuid REFERENCES organizations ON DELETE CASCADE,
    owner_user_id              uuid                                                                      REFERENCES users ON DELETE SET NULL,
    logo_url                   TEXT,
    terms_of_service_url       TEXT,
    policy_url                 TEXT,
    client_uri                 TEXT,
    token_endpoint_auth_method VARCHAR(50)              DEFAULT 'client_secret_basic'::CHARACTER VARYING NOT NULL
        CONSTRAINT oauth2_clients_auth_method_check CHECK ((token_endpoint_auth_method)::TEXT = ANY
                                                           ((ARRAY ['client_secret_basic'::CHARACTER VARYING, 'client_secret_post'::CHARACTER VARYING, 'client_secret_jwt'::CHARACTER VARYING, 'private_key_jwt'::CHARACTER VARYING, 'none'::CHARACTER VARYING])::TEXT[])),
    created_at                 TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                                   NOT NULL,
    updated_at                 TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                                   NOT NULL,
    is_internal                BOOLEAN                  DEFAULT FALSE                                    NOT NULL,
    active                     BOOLEAN                  DEFAULT TRUE                                     NOT NULL,
    CONSTRAINT oauth2_clients_public_check CHECK (((is_public = TRUE) AND
                                                   ((token_endpoint_auth_method)::TEXT = 'none'::TEXT)) OR
                                                  (is_public = FALSE))
);

COMMENT ON TABLE oauth2_clients IS 'Seeded với demo admin dashboard client';

COMMENT ON COLUMN oauth2_clients.secret_hash IS 'Hashed client secret (bcrypt hoặc argon2)';

COMMENT ON COLUMN oauth2_clients.redirect_uris IS 'Danh sách redirect URIs được phép (OAuth 2.0 callback)';

COMMENT ON COLUMN oauth2_clients.grant_types IS 'Grant types được phép: authorization_code, refresh_token, client_credentials, etc.';

COMMENT ON COLUMN oauth2_clients.is_public IS 'TRUE cho public clients (mobile apps, SPAs) không có secret';

COMMENT ON COLUMN oauth2_clients.organization_id IS 'NULL = global client (first-party), NOT NULL = organization-specific client (third-party)';

COMMENT ON COLUMN oauth2_clients.is_internal IS 'TRUE = Internal/first-party client (full access), FALSE = External/third-party client (limited features). Internal clients typically have organization_id = NULL.';

COMMENT ON COLUMN oauth2_clients.active IS 'TRUE = Active client, FALSE = Inactive/Disabled client';

ALTER TABLE oauth2_clients
    OWNER TO system_dev;

CREATE INDEX idx_oauth2_clients_organization_id ON oauth2_clients ( organization_id );

CREATE INDEX idx_oauth2_clients_owner_user_id ON oauth2_clients ( owner_user_id );

CREATE INDEX idx_oauth2_clients_is_public ON oauth2_clients ( is_public );

CREATE INDEX idx_oauth2_clients_is_internal ON oauth2_clients ( is_internal );

CREATE INDEX idx_oauth2_clients_active ON oauth2_clients ( active );

CREATE TRIGGER update_oauth2_clients_updated_at
    BEFORE UPDATE
    ON oauth2_clients
    FOR EACH ROW
EXECUTE PROCEDURE public.update_updated_at_column( );

CREATE TABLE IF NOT EXISTS oauth2_sessions
(
    signature    VARCHAR(255)                            NOT NULL PRIMARY KEY,
    request_id   VARCHAR(255)                            NOT NULL,
    session_type session_type                            NOT NULL,
    active       BOOLEAN                  DEFAULT TRUE   NOT NULL,
    session_data jsonb                                   NOT NULL,
    expires_at   TIMESTAMP WITH TIME ZONE                NOT NULL,
    client_id    uuid                                    NOT NULL REFERENCES oauth2_clients ON DELETE CASCADE,
    subject_id   uuid REFERENCES users ON DELETE CASCADE,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW( ) NOT NULL
);

COMMENT ON TABLE oauth2_sessions IS 'Bảng chung lưu trữ tất cả OAuth 2.0 sessions, codes, và tokens';

COMMENT ON COLUMN oauth2_sessions.signature IS 'Unique signature của token/code (thường là hash)';

COMMENT ON COLUMN oauth2_sessions.request_id IS 'ID của authorization request';

COMMENT ON COLUMN oauth2_sessions.session_data IS 'Fosite Requester object serialized thành JSON';

COMMENT ON COLUMN oauth2_sessions.subject_id IS 'User ID (NULL cho client_credentials grant)';

ALTER TABLE oauth2_sessions
    OWNER TO system_dev;

CREATE INDEX idx_oauth2_sessions_request_id ON oauth2_sessions ( request_id );

CREATE INDEX idx_oauth2_sessions_session_type ON oauth2_sessions ( session_type );

CREATE INDEX idx_oauth2_sessions_expires_at ON oauth2_sessions ( expires_at );

CREATE INDEX idx_oauth2_sessions_client_id ON oauth2_sessions ( client_id );

CREATE INDEX idx_oauth2_sessions_subject_id ON oauth2_sessions ( subject_id );

CREATE INDEX idx_oauth2_sessions_active ON oauth2_sessions ( active );

CREATE INDEX idx_oauth2_sessions_cleanup ON oauth2_sessions ( expires_at, active );

CREATE INDEX idx_oauth2_sessions_session_data ON oauth2_sessions USING gin ( session_data );

CREATE TABLE IF NOT EXISTS oauth2_jti_blacklist
(
    signature  VARCHAR(255)                            NOT NULL PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE                NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW( ) NOT NULL
);

COMMENT ON TABLE oauth2_jti_blacklist IS 'Blacklist cho revoked tokens - ngăn chặn replay attacks';

COMMENT ON COLUMN oauth2_jti_blacklist.signature IS 'Token signature đã bị revoke';

COMMENT ON COLUMN oauth2_jti_blacklist.expires_at IS 'Thời điểm token hết hạn (có thể xóa khỏi blacklist sau đó)';

ALTER TABLE oauth2_jti_blacklist
    OWNER TO system_dev;

CREATE INDEX idx_oauth2_jti_blacklist_expires_at ON oauth2_jti_blacklist ( expires_at );

CREATE TABLE IF NOT EXISTS oauth2_consents
(
    id             uuid                     DEFAULT uuidv7( )                     NOT NULL PRIMARY KEY,
    user_id        uuid                                                           NOT NULL REFERENCES users ON DELETE CASCADE,
    client_id      uuid                                                           NOT NULL REFERENCES oauth2_clients ON DELETE CASCADE,
    granted_scopes TEXT[]                   DEFAULT '{}'::TEXT[]                  NOT NULL,
    revoked        BOOLEAN                  DEFAULT FALSE                         NOT NULL,
    granted_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                        NOT NULL,
    revoked_at     TIMESTAMP WITH TIME ZONE,
    last_used_at   TIMESTAMP WITH TIME ZONE,
    expires_at     TIMESTAMP WITH TIME ZONE,
    consent_method VARCHAR(50)              DEFAULT 'explicit'::CHARACTER VARYING NOT NULL
        CONSTRAINT oauth2_consents_method_check CHECK ((consent_method)::TEXT = ANY
                                                       ((ARRAY ['explicit'::CHARACTER VARYING, 'implicit'::CHARACTER VARYING, 'remembered'::CHARACTER VARYING])::TEXT[])),
    ip_address     inet,
    user_agent     TEXT,
    CONSTRAINT oauth2_consents_user_client_unique UNIQUE ( user_id, client_id )
);

COMMENT ON TABLE oauth2_consents IS 'Lưu trữ user consents cho OAuth2 clients - quản lý quyền truy cập';

COMMENT ON COLUMN oauth2_consents.user_id IS 'User đã cấp quyền';

COMMENT ON COLUMN oauth2_consents.client_id IS 'Client được cấp quyền';

COMMENT ON COLUMN oauth2_consents.granted_scopes IS 'Danh sách scopes đã được user chấp thuận';

COMMENT ON COLUMN oauth2_consents.revoked IS 'TRUE nếu consent đã bị thu hồi';

COMMENT ON COLUMN oauth2_consents.expires_at IS 'Thời điểm consent hết hạn (NULL = persistent)';

COMMENT ON COLUMN oauth2_consents.consent_method IS 'Phương thức consent: explicit (user clicked allow), implicit (trusted first-party), remembered (previous consent)';

ALTER TABLE oauth2_consents
    OWNER TO system_dev;

CREATE INDEX idx_oauth2_consents_user_id ON oauth2_consents ( user_id );

CREATE INDEX idx_oauth2_consents_client_id ON oauth2_consents ( client_id );

CREATE INDEX idx_oauth2_consents_revoked ON oauth2_consents ( revoked );

CREATE INDEX idx_oauth2_consents_granted_at ON oauth2_consents ( granted_at );

CREATE INDEX idx_oauth2_consents_expires_at ON oauth2_consents ( expires_at );

CREATE INDEX idx_oauth2_consents_cleanup ON oauth2_consents ( expires_at, revoked ) WHERE (expires_at IS NOT NULL);

CREATE TABLE IF NOT EXISTS email_verification_tokens
(
    id         uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    user_id    uuid                                       NOT NULL REFERENCES users ON DELETE CASCADE,
    token      VARCHAR(255)                               NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE                   NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL
);

COMMENT ON TABLE email_verification_tokens IS 'Lưu trữ tokens để xác thực email address khi đăng ký hoặc thay đổi email';

COMMENT ON COLUMN email_verification_tokens.token IS 'Random token gửi qua email (hashed hoặc plain - tùy implementation)';

COMMENT ON COLUMN email_verification_tokens.expires_at IS 'Token hết hạn sau 24 giờ';

COMMENT ON COLUMN email_verification_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã verify thành công';

ALTER TABLE email_verification_tokens
    OWNER TO system_dev;

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens ( user_id );

CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens ( token );

CREATE INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens ( expires_at );

CREATE TABLE IF NOT EXISTS password_reset_tokens
(
    id         uuid                     DEFAULT uuidv7( ) NOT NULL PRIMARY KEY,
    user_id    uuid                                       NOT NULL REFERENCES users ON DELETE CASCADE,
    token      VARCHAR(255)                               NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE                   NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW( )    NOT NULL
);

COMMENT ON TABLE password_reset_tokens IS 'Lưu trữ tokens để reset password';

COMMENT ON COLUMN password_reset_tokens.token IS 'Random token gửi qua email';

COMMENT ON COLUMN password_reset_tokens.expires_at IS 'Token hết hạn sau 1 giờ';

COMMENT ON COLUMN password_reset_tokens.used_at IS 'NULL = chưa sử dụng, NOT NULL = đã reset thành công';

ALTER TABLE password_reset_tokens
    OWNER TO system_dev;

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens ( user_id );

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens ( token );

CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens ( expires_at );

CREATE TABLE IF NOT EXISTS webauthn_credentials
(
    id               uuid                     DEFAULT uuidv7( )                 NOT NULL PRIMARY KEY,
    user_id          uuid                                                       NOT NULL
        CONSTRAINT fk_webauthn_credentials_user REFERENCES users ON DELETE CASCADE,
    credential_id    TEXT                                                       NOT NULL UNIQUE,
    public_key       bytea                                                      NOT NULL,
    attestation_type VARCHAR(50)              DEFAULT 'none'::CHARACTER VARYING NOT NULL
        CONSTRAINT webauthn_credentials_attestation_type_check CHECK ((attestation_type)::TEXT = ANY
                                                                      ((ARRAY ['none'::CHARACTER VARYING, 'indirect'::CHARACTER VARYING, 'direct'::CHARACTER VARYING])::TEXT[])),
    aaguid           bytea,
    sign_count       INTEGER                  DEFAULT 0                         NOT NULL,
    transports       TEXT[],
    backup_eligible  BOOLEAN                  DEFAULT FALSE                     NOT NULL,
    backup_state     BOOLEAN                  DEFAULT FALSE                     NOT NULL,
    credential_name  VARCHAR(255),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                    NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                    NOT NULL,
    last_used_at     TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE webauthn_credentials IS 'Bảng lưu trữ WebAuthn/FIDO2 credentials cho passwordless authentication';

COMMENT ON COLUMN webauthn_credentials.credential_id IS 'Unique identifier của credential từ authenticator (Base64URL encoded)';

COMMENT ON COLUMN webauthn_credentials.public_key IS 'Public key của credential trong binary format';

COMMENT ON COLUMN webauthn_credentials.attestation_type IS 'Loại attestation: none (no attestation), indirect (anonymized), direct (full)';

COMMENT ON COLUMN webauthn_credentials.aaguid IS 'Authenticator AAGUID (16 bytes) - identifies authenticator model';

COMMENT ON COLUMN webauthn_credentials.sign_count IS 'Counter tăng dần mỗi lần authentication - dùng để phát hiện cloned authenticators';

COMMENT ON COLUMN webauthn_credentials.transports IS 'Supported transports (usb, nfc, ble, internal, hybrid)';

COMMENT ON COLUMN webauthn_credentials.backup_eligible IS 'Credential có thể được backup (multi-device credentials)';

COMMENT ON COLUMN webauthn_credentials.backup_state IS 'Credential hiện đang được backup hay không';

COMMENT ON COLUMN webauthn_credentials.credential_name IS 'User-defined name để nhận diện credential';

COMMENT ON COLUMN webauthn_credentials.last_used_at IS 'Timestamp lần cuối credential được sử dụng để authentication';

ALTER TABLE webauthn_credentials
    OWNER TO system_dev;

CREATE INDEX idx_webauthn_credentials_user_id ON webauthn_credentials ( user_id );

CREATE INDEX idx_webauthn_credentials_credential_id ON webauthn_credentials ( credential_id );

CREATE INDEX idx_webauthn_credentials_created_at ON webauthn_credentials ( created_at );

CREATE INDEX idx_webauthn_credentials_last_used_at ON webauthn_credentials ( last_used_at );

CREATE TABLE IF NOT EXISTS webauthn_sessions
(
    id           uuid                     DEFAULT uuidv7( )                       NOT NULL PRIMARY KEY,
    user_id      uuid
        CONSTRAINT fk_webauthn_sessions_user REFERENCES users ON DELETE CASCADE,
    challenge    TEXT                                                             NOT NULL,
    session_type VARCHAR(50)                                                      NOT NULL
        CONSTRAINT webauthn_sessions_session_type_check CHECK ((session_type)::TEXT = ANY
                                                               ((ARRAY ['registration'::CHARACTER VARYING, 'authentication'::CHARACTER VARYING])::TEXT[])),
    user_agent   TEXT,
    ip_address   VARCHAR(45),
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW( )                          NOT NULL,
    expires_at   TIMESTAMP WITH TIME ZONE DEFAULT (NOW( ) + '00:05:00'::INTERVAL) NOT NULL
);

COMMENT ON TABLE webauthn_sessions IS 'Bảng lưu trữ temporary sessions cho WebAuthn registration/authentication flow';

COMMENT ON COLUMN webauthn_sessions.challenge IS 'Random challenge string (Base64URL encoded) dùng cho ceremony';

COMMENT ON COLUMN webauthn_sessions.session_type IS 'Loại session: registration (đăng ký credential mới) hoặc authentication (xác thực)';

COMMENT ON COLUMN webauthn_sessions.expires_at IS 'Session timeout (default 5 minutes)';

ALTER TABLE webauthn_sessions
    OWNER TO system_dev;

CREATE INDEX idx_webauthn_sessions_user_id ON webauthn_sessions ( user_id );

CREATE INDEX idx_webauthn_sessions_challenge ON webauthn_sessions ( challenge );

CREATE INDEX idx_webauthn_sessions_expires_at ON webauthn_sessions ( expires_at );

CREATE INDEX idx_webauthn_sessions_created_at ON webauthn_sessions ( created_at );

CREATE FUNCTION user_has_organization_permission(p_user_id uuid, p_organization_id uuid, p_permission_name character varying) RETURNS boolean
    STABLE SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_organization_roles uor
        JOIN role_permissions rp ON uor.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE uor.user_id = p_user_id
            AND uor.organization_id = p_organization_id
            AND p.name = p_permission_name
    );
END;
$$;

COMMENT ON FUNCTION user_has_organization_permission(uuid, uuid, VARCHAR) IS 'Kiểm tra xem user có permission cụ thể trong organization không';

ALTER FUNCTION user_has_organization_permission(uuid, uuid, VARCHAR) OWNER TO system_dev;

CREATE FUNCTION user_has_global_permission(p_user_id uuid, p_permission_name character varying) RETURNS boolean
    STABLE SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM user_global_roles ugr
        JOIN role_permissions rp ON ugr.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE ugr.user_id = p_user_id
            AND p.name = p_permission_name
    );
END;
$$;

COMMENT ON FUNCTION user_has_global_permission(uuid, VARCHAR) IS 'Kiểm tra xem user có global permission cụ thể không';

ALTER FUNCTION user_has_global_permission(uuid, VARCHAR) OWNER TO system_dev;

CREATE FUNCTION get_user_organization_permissions(p_user_id uuid, p_organization_id uuid)
    RETURNS TABLE(permission_name character varying, permission_description text)
    STABLE SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT p.name, p.description
    FROM user_organization_roles uor
    JOIN role_permissions rp ON uor.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE uor.user_id = p_user_id
        AND uor.organization_id = p_organization_id
    ORDER BY p.name;
END;
$$;

COMMENT ON FUNCTION get_user_organization_permissions(uuid, uuid) IS 'Lấy danh sách tất cả permissions của user trong organization';

ALTER FUNCTION get_user_organization_permissions(uuid, uuid) OWNER TO system_dev;

CREATE FUNCTION get_user_global_permissions(p_user_id uuid)
    RETURNS TABLE(permission_name character varying, permission_description text)
    STABLE SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT p.name, p.description
    FROM user_global_roles ugr
    JOIN role_permissions rp ON ugr.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ugr.user_id = p_user_id
    ORDER BY p.name;
END;
$$;

COMMENT ON FUNCTION get_user_global_permissions(uuid) IS 'Lấy danh sách tất cả global permissions của user';

ALTER FUNCTION get_user_global_permissions(uuid) OWNER TO system_dev;

CREATE FUNCTION cleanup_expired_oauth2_data() RETURNS integer
    SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete expired sessions
    WITH deleted AS (
        DELETE FROM oauth2_sessions
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    -- Delete expired blacklist entries
    DELETE FROM oauth2_jti_blacklist
    WHERE expires_at < NOW();

    RETURN deleted_count;
END;
$$;

COMMENT ON FUNCTION cleanup_expired_oauth2_data() IS 'Cleanup expired OAuth2 sessions và blacklist entries. Nên chạy định kỳ (cron job).';

ALTER FUNCTION cleanup_expired_oauth2_data() OWNER TO system_dev;

CREATE FUNCTION is_token_revoked(p_signature character varying) RETURNS boolean
    STABLE SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM oauth2_jti_blacklist
        WHERE signature = p_signature
            AND expires_at > NOW()
    );
END;
$$;

COMMENT ON FUNCTION is_token_revoked(VARCHAR) IS 'Kiểm tra xem token có trong blacklist không';

ALTER FUNCTION is_token_revoked(VARCHAR) OWNER TO system_dev;

CREATE FUNCTION revoke_token(p_signature character varying, p_expires_at timestamp with time zone) RETURNS void
    SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
BEGIN
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    VALUES (p_signature, p_expires_at)
    ON CONFLICT (signature) DO NOTHING;

    -- Mark session as inactive
    UPDATE oauth2_sessions
    SET active = FALSE
    WHERE signature = p_signature;
END;
$$;

COMMENT ON FUNCTION revoke_token(VARCHAR, TIMESTAMP WITH TIME ZONE) IS 'Revoke một token bằng cách thêm vào blacklist và mark session inactive';

ALTER FUNCTION revoke_token(VARCHAR, TIMESTAMP WITH TIME ZONE) OWNER TO system_dev;

CREATE FUNCTION revoke_user_client_tokens(p_user_id uuid, p_client_id uuid) RETURNS integer
    SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    -- Mark sessions as inactive
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND client_id = p_client_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    -- Add to blacklist
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$;

COMMENT ON FUNCTION revoke_user_client_tokens(uuid, uuid) IS 'Revoke tất cả tokens của một user cho một client cụ thể';

ALTER FUNCTION revoke_user_client_tokens(uuid, uuid) OWNER TO system_dev;

CREATE FUNCTION revoke_all_user_tokens(p_user_id uuid) RETURNS integer
    SET search_path = identify, public
    LANGUAGE plpgsql
AS $$
DECLARE
    revoked_count INTEGER;
BEGIN
    WITH updated AS (
        UPDATE oauth2_sessions
        SET active = FALSE
        WHERE subject_id = p_user_id
            AND active = TRUE
        RETURNING signature, expires_at
    )
    INSERT INTO oauth2_jti_blacklist (signature, expires_at)
    SELECT signature, expires_at FROM updated
    ON CONFLICT (signature) DO NOTHING;

    GET DIAGNOSTICS revoked_count = ROW_COUNT;
    RETURN revoked_count;
END;
$$;

COMMENT ON FUNCTION revoke_all_user_tokens(uuid) IS 'Revoke tất cả tokens của một user (global logout)';

ALTER FUNCTION revoke_all_user_tokens(uuid) OWNER TO system_dev;

CREATE FUNCTION cleanup_expired_consents() RETURNS integer
    LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.oauth2_consents
        WHERE expires_at IS NOT NULL
            AND expires_at < NOW()
            AND revoked = FALSE
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$;

COMMENT ON FUNCTION cleanup_expired_consents() IS 'Xóa các consents đã hết hạn - nên chạy định kỳ';

ALTER FUNCTION cleanup_expired_consents() OWNER TO system_dev;

CREATE FUNCTION revoke_consent(p_user_id uuid, p_client_id uuid) RETURNS boolean
    LANGUAGE plpgsql
AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND client_id = p_client_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected > 0;
END;
$$;

COMMENT ON FUNCTION revoke_consent(uuid, uuid) IS 'Thu hồi consent của user cho một client';

ALTER FUNCTION revoke_consent(uuid, uuid) OWNER TO system_dev;

CREATE FUNCTION revoke_all_user_consents(p_user_id uuid) RETURNS integer
    LANGUAGE plpgsql
AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    UPDATE identify.oauth2_consents
    SET revoked = TRUE,
        revoked_at = NOW()
    WHERE user_id = p_user_id
        AND revoked = FALSE;

    GET DIAGNOSTICS rows_affected = ROW_COUNT;
    RETURN rows_affected;
END;
$$;

COMMENT ON FUNCTION revoke_all_user_consents(uuid) IS 'Thu hồi tất cả consents của một user';

ALTER FUNCTION revoke_all_user_consents(uuid) OWNER TO system_dev;

CREATE FUNCTION get_active_consent(p_user_id uuid, p_client_id uuid)
    RETURNS TABLE(id uuid, granted_scopes text[], granted_at timestamp with time zone, expires_at timestamp with time zone)
    STABLE
    LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.granted_scopes,
        c.granted_at,
        c.expires_at
    FROM identify.oauth2_consents c
    WHERE c.user_id = p_user_id
        AND c.client_id = p_client_id
        AND c.revoked = FALSE
        AND (c.expires_at IS NULL OR c.expires_at > NOW());
END;
$$;

COMMENT ON FUNCTION get_active_consent(uuid, uuid) IS 'Lấy active consent của user cho một client';

ALTER FUNCTION get_active_consent(uuid, uuid) OWNER TO system_dev;

CREATE FUNCTION cleanup_expired_verification_tokens() RETURNS integer
    LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.email_verification_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$;

COMMENT ON FUNCTION cleanup_expired_verification_tokens() IS 'Cleanup expired verification tokens. Nên chạy định kỳ (cron job).';

ALTER FUNCTION cleanup_expired_verification_tokens() OWNER TO system_dev;

CREATE FUNCTION cleanup_expired_password_reset_tokens() RETURNS integer
    LANGUAGE plpgsql
AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    WITH deleted AS (
        DELETE FROM identify.password_reset_tokens
        WHERE expires_at < NOW()
        RETURNING 1
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;

    RETURN deleted_count;
END;
$$;

COMMENT ON FUNCTION cleanup_expired_password_reset_tokens() IS 'Cleanup expired password reset tokens. Nên chạy định kỳ (cron job).';

ALTER FUNCTION cleanup_expired_password_reset_tokens() OWNER TO system_dev;

