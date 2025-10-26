  # Database Migrations

Thư mục này chứa tất cả database migrations cho project sử dụng [golang-migrate](https://github.com/golang-migrate/migrate).

## Cấu trúc Migrations

```
migrations/
├── 000001_create_core_tables.up.sql                  # Tạo tenants, users, permissions, roles
├── 000001_create_core_tables.down.sql                # Rollback migration 1
├── 000002_create_rbac_relations.up.sql               # Tạo RBAC junction tables
├── 000002_create_rbac_relations.down.sql             # Rollback migration 2
├── 000003_create_oauth2_tables.up.sql                # Tạo OAuth2 tables
├── 000003_create_oauth2_tables.down.sql              # Rollback migration 3
├── 000004_create_schemas_and_move_tables.up.sql      # Tạo schemas và di chuyển tables
├── 000004_create_schemas_and_move_tables.down.sql    # Rollback migration 4
├── 000005_move_functions_to_identify_schema.up.sql   # Di chuyển functions vào identify schema
├── 000005_move_functions_to_identify_schema.down.sql # Rollback migration 5
├── 000006_fix_function_search_paths.up.sql           # Fix search_path cho functions
├── 000006_fix_function_search_paths.down.sql         # Rollback migration 6
└── README.md                                          # File này
```

## Migration Overview

### Migration 000001: Core Tables
Tạo các bảng cốt lõi cho hệ thống multi-tenant với RBAC:
- **tenants** - Thông tin tổ chức/khách hàng
- **users** - Tài khoản người dùng (global)
- **permissions** - Danh sách permissions (global/tenant scope)
- **roles** - Danh sách roles (global/tenant scope)

**Seed Data:**
- 5 global permissions + 11 tenant permissions
- 6 default roles (SUPER_ADMIN, PLATFORM_ADMIN, TENANT_ADMIN, TENANT_MANAGER, TENANT_USER, TENANT_VIEWER)

### Migration 000002: RBAC Relations
Tạo các bảng quan hệ cho RBAC và multi-tenant:
- **role_permissions** - Many-to-many: roles ↔ permissions
- **user_tenant_memberships** - User membership trong tenant
- **user_tenant_roles** - User roles trong tenant context
- **user_global_roles** - Global roles của user

**Helper Functions:**
- `user_has_tenant_permission()` - Check permission trong tenant
- `user_has_global_permission()` - Check global permission
- `get_user_tenant_permissions()` - Lấy tất cả permissions của user trong tenant
- `get_user_global_permissions()` - Lấy tất cả global permissions

**Seed Data:**
- Permission mappings cho tất cả default roles

### Migration 000003: OAuth2 Tables
Tạo các bảng cho OAuth 2.0 Authorization Server (Fosite-compatible):
- **oauth2_clients** - OAuth clients (support multi-tenant)
- **oauth2_sessions** - Sessions, codes, tokens (JSONB storage)
- **oauth2_jti_blacklist** - Revoked tokens blacklist

**Helper Functions:**
- `cleanup_expired_oauth2_data()` - Cleanup expired sessions/blacklist
- `is_token_revoked()` - Check token revocation
- `revoke_token()` - Revoke single token
- `revoke_user_client_tokens()` - Revoke user tokens cho một client
- `revoke_all_user_tokens()` - Global logout

**Seed Data:**
- Demo "System Admin Dashboard" client

### Migration 000004: Create Schemas and Move Tables
Tạo multi-schema architecture và di chuyển tables vào đúng schemas:
- Tạo 4 schemas: **identify**, **catalog**, **community**, **payment**
- Di chuyển tất cả 11 tables vào **identify** schema:
  - Core: tenants, users, permissions, roles
  - RBAC: role_permissions, user_tenant_memberships, user_tenant_roles, user_global_roles
  - OAuth2: oauth2_clients, oauth2_sessions, oauth2_jti_blacklist

**Lý do:**
- Tách biệt domain logic theo schemas
- Dễ quản lý permissions và security
- Chuẩn bị cho việc scale (có thể tách schema ra database riêng sau này)

### Migration 000005: Move Functions to Identify Schema
Di chuyển tất cả helper functions vào **identify** schema:
- 4 RBAC functions: permission checking và retrieval
- 5 OAuth2 functions: token management và cleanup

**Lý do:**
- Functions và tables nên cùng schema
- Nhất quán về organization
- Dễ quản lý access control

### Migration 000006: Fix Function Search Paths
Set `search_path = identify, public` cho tất cả functions trong identify schema:

**Lý do:**
- Functions cần tìm tables trong identify schema
- Mặc định PostgreSQL search path chỉ có public
- Sau khi move functions sang schema khác, cần update search_path

**Kết quả:**
- Tất cả functions hoạt động bình thường
- Có thể gọi: `SELECT identify.user_has_tenant_permission(...)`

## Prerequisites

### 1. Cài đặt golang-migrate

```bash
# Sử dụng Makefile
make migrate-install

# Hoặc cài thủ công
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 2. Start Database

```bash
# Start PostgreSQL và Redis
make docker-up

# Hoặc
docker-compose up -d
```

## Commands

### Sử dụng Makefile (Recommended)

```bash
# Xem tất cả commands
make help

# Setup development environment (install tool + docker + migrations)
make setup

# Chạy migrations
make migrate-up

# Rollback last migration
make migrate-down

# Rollback all migrations
make migrate-down-all

# Check migration version
make migrate-version

# Force migration version
make migrate-force VERSION=2

# Tạo migration mới
make migrate-create NAME=add_users_phone
```

### Sử dụng golang-migrate trực tiếp

```bash
# Database URL
DATABASE_URL="postgres://system_dev:system_dev@localhost:5432/system_dev?sslmode=disable"

# Migrate up
migrate -path ./migrations -database "$DATABASE_URL" up

# Migrate down
migrate -path ./migrations -database "$DATABASE_URL" down 1

# Check version
migrate -path ./migrations -database "$DATABASE_URL" version

# Force version
migrate -path ./migrations -database "$DATABASE_URL" force 2
```

## Creating New Migrations

### Sử dụng Makefile

```bash
make migrate-create NAME=add_users_avatar
```

Sẽ tạo 2 files:
```
migrations/000004_add_users_avatar.up.sql
migrations/000004_add_users_avatar.down.sql
```

### Best Practices

1. **Luôn tạo cả `.up.sql` và `.down.sql`**
   - Up: Thực hiện migration
   - Down: Rollback migration

2. **Atomic migrations**
   - Mỗi migration nên làm một việc cụ thể
   - Tránh mix multiple concerns trong một migration

3. **Idempotent migrations**
   - Sử dụng `IF EXISTS`, `IF NOT EXISTS`
   - Migration có thể chạy nhiều lần mà không lỗi

4. **Comment rõ ràng**
   - Giải thích mục đích của migration
   - Document breaking changes

5. **Test migrations**
   ```bash
   make migrate-up    # Apply
   make migrate-down  # Rollback
   make migrate-up    # Re-apply
   ```

## Troubleshooting

### Migration bị "dirty"

Nếu migration fail ở giữa, database sẽ ở trạng thái "dirty":

```bash
# Check version
make migrate-version
# Output: 2/d (dirty)

# Fix bằng cách force version
make migrate-force VERSION=2

# Hoặc rollback và thử lại
make migrate-down
make migrate-up
```

### Reset database hoàn toàn

```bash
# Cách 1: Drop + Create database
make db-reset
make migrate-up

# Cách 2: Rollback all migrations
make migrate-down-all
make migrate-up
```

### Connection refused

```bash
# Check PostgreSQL đang chạy
make docker-ps

# Start nếu chưa chạy
make docker-up

# Check logs
make docker-logs
```

## Database Schema Documentation

### Core Tables Relationships

```
tenants (1) ──────< (N) user_tenant_memberships (N) >────── (1) users
                            │
                            │ (1)
                            │
                            ↓
                          (N) user_tenant_roles (N) >────── (1) roles
                                                                  │
                                                                  │ (1)
                                                                  │
                                                                  ↓
                                                                (N) role_permissions (N) >── (1) permissions
```

### OAuth2 Tables Relationships

```
tenants (1) ─────< (0..N) oauth2_clients (1) ──────< (N) oauth2_sessions (N) >────── (1) users
                                                            │
                                                            │
                                                            ↓
                                                      oauth2_jti_blacklist
```

### Scopes

- **Global Scope**: Permissions/Roles áp dụng toàn hệ thống
  - Example: `SUPER_ADMIN`, `system:manage_all`

- **Tenant Scope**: Permissions/Roles áp dụng trong tenant
  - Example: `TENANT_ADMIN`, `user:create`

## Seeded Data

### Default Roles

| Role | Scope | Description |
|------|-------|-------------|
| SUPER_ADMIN | global | Full system access |
| PLATFORM_ADMIN | global | Platform management |
| TENANT_ADMIN | tenant | Full tenant access |
| TENANT_MANAGER | tenant | Manage users & content |
| TENANT_USER | tenant | Create & edit content |
| TENANT_VIEWER | tenant | Read-only access |

### Permission Examples

| Permission | Scope | Description |
|------------|-------|-------------|
| system:manage_all | global | Manage entire system |
| tenant:create | global | Create new tenants |
| user:create | tenant | Create users in tenant |
| content:view | tenant | View content |

## Maintenance

### Cleanup expired OAuth2 data

```sql
-- Manual cleanup
SELECT cleanup_expired_oauth2_data();

-- Setup cron job (recommended)
-- Run daily at 2 AM
0 2 * * * psql -U system_dev -d system_dev -c "SELECT cleanup_expired_oauth2_data();"
```

### Monitoring

```sql
-- Check migration version
SELECT * FROM schema_migrations;

-- Count records in tables
SELECT
    'tenants' as table_name,
    COUNT(*) as count
FROM tenants
UNION ALL
SELECT 'users', COUNT(*) FROM users
UNION ALL
SELECT 'oauth2_clients', COUNT(*) FROM oauth2_clients
UNION ALL
SELECT 'oauth2_sessions', COUNT(*) FROM oauth2_sessions;

-- Check active sessions
SELECT
    session_type,
    COUNT(*) as count,
    COUNT(*) FILTER (WHERE active = TRUE) as active_count
FROM oauth2_sessions
GROUP BY session_type;
```

## References

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [OAuth 2.0 Specification](https://tools.ietf.org/html/rfc6749)
- [Fosite Library](https://github.com/ory/fosite)
