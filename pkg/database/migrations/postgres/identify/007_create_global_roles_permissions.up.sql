CREATE TABLE IF NOT EXISTS global_permissions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    key VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS global_roles (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS global_role_permissions (
    role_id UUID NOT NULL REFERENCES global_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES global_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_global_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES global_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_global_roles_role ON user_global_roles(role_id);
