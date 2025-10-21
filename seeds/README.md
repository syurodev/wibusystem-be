# Database Seeds

This directory contains SQL seed scripts to populate the database with initial development data.

## 📁 Seed Files

| File | Description | Dependencies |
|------|-------------|--------------|
| `000_run_all_seeds.sql` | Master script - runs all seeds in order | All seeds |
| `001_oauth2_clients.sql` | OAuth2 clients for authentication | pgcrypto |
| `002_users_and_tenants.sql` | Demo users and tenants | 001_oauth2_clients.sql |

## 🚀 Quick Start

### Run All Seeds

The easiest way to seed the database:

```bash
make db-seed
```

Or manually:

```bash
docker exec -i system_dev psql -U system_dev -d system_dev < seeds/000_run_all_seeds.sql
```

### Run Individual Seeds

```bash
# OAuth2 clients only
docker exec -i system_dev psql -U system_dev -d system_dev < seeds/001_oauth2_clients.sql

# Users and tenants only
docker exec -i system_dev psql -U system_dev -d system_dev < seeds/002_users_and_tenants.sql
```

### Reset and Re-seed

```bash
# Clean database and re-seed
make db-reset
# Then wait for app to run migrations
# Then seed
make db-seed
```

## 📊 Seeded Data

### OAuth2 Clients (4 clients)

| Client ID | Type | Grant Types | Use Case |
|-----------|------|-------------|----------|
| `wibusystem-web` | Confidential | authorization_code, refresh_token | Web apps |
| `wibusystem-mobile` | Confidential | authorization_code, refresh_token, password | Mobile apps |
| `wibusystem-spa` | Public | authorization_code, refresh_token | SPAs |
| `wibusystem-service` | Confidential | client_credentials | M2M services |

**Client Secrets (Development Only):**
- `wibusystem-web-secret-dev`
- `wibusystem-mobile-secret-dev`
- `wibusystem-service-secret-dev`
- SPA has no secret (public client)

### Users (4 users)

| Email | Display Name | Status | Password |
|-------|--------------|--------|----------|
| admin@wibusystem.dev | System Administrator | active | password123 |
| user1@wibusystem.dev | John Doe | active | password123 |
| user2@wibusystem.dev | Jane Smith | active | password123 |
| pending@wibusystem.dev | Pending User | pending_verification | password123 |

### Tenants (3 tenants)

| Name | Slug | Owner | Status | Members |
|------|------|-------|--------|---------|
| WibuSystem HQ | wibusystem-hq | admin@wibusystem.dev | active | 1 |
| John's Workspace | john-workspace | user1@wibusystem.dev | active | 1 |
| Team Collaboration | team-collab | user1@wibusystem.dev | trial | 2 |

## 🔐 Password Hashing

Seeds use PostgreSQL's **pgcrypto extension** with **bcrypt** for password hashing:

```sql
-- Enable extension
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Hash password
crypt('password123', gen_salt('bf', 10))  -- bcrypt with cost factor 10
```

### Why pgcrypto?

- ✅ Native PostgreSQL extension
- ✅ Bcrypt algorithm (same as Go's bcrypt)
- ✅ Compatible with application code
- ✅ No external dependencies
- ✅ Secure salt generation

### Bcrypt Cost Factor

Seeds use cost factor **10** for development (faster):
- Development: cost 10 (~100ms)
- Production: cost 12-14 recommended (~500ms-2s)

## 📝 Seed Script Features

### Idempotent

All seeds use `ON CONFLICT ... DO UPDATE` or `ON CONFLICT ... DO NOTHING`:

```sql
INSERT INTO identity.oauth2_clients (...)
VALUES (...)
ON CONFLICT (id) DO UPDATE SET
    client_name = EXCLUDED.client_name,
    updated_at = CURRENT_TIMESTAMP;
```

Safe to run multiple times without errors.

### Transactional

Master seed script runs in a transaction:

```sql
BEGIN;
-- All seeds here
COMMIT;
```

All seeds succeed or all fail together.

### Informative Output

Seeds display:
- Progress messages
- Seeded data tables
- Credentials for testing
- Summary statistics

## 🛠️ Creating New Seeds

### Naming Convention

```
NNN_description.sql
```

Where `NNN` is a 3-digit sequence number (001, 002, 003, etc.)

### Template

```sql
-- Seed: Your Seed Name
-- Description: What this seed does
-- Dependencies: Previous seeds required

-- Enable extensions if needed
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert data with conflict handling
INSERT INTO schema.table (...)
VALUES (...)
ON CONFLICT (unique_column) DO UPDATE SET
    column = EXCLUDED.column,
    updated_at = CURRENT_TIMESTAMP;

-- Display results
\echo '✅ Your data seeded successfully!'
SELECT * FROM schema.table;
```

### Best Practices

1. **Use pgcrypto for passwords**: `crypt(password, gen_salt('bf', 10))`
2. **Handle conflicts**: Always use `ON CONFLICT`
3. **Use UUIDs**: `gen_random_uuid()` for IDs
4. **Add comments**: Explain what data is being seeded
5. **Show output**: Display what was seeded
6. **Test idempotency**: Run twice to verify

## 🔍 Verifying Seeds

### Check OAuth2 Clients

```sql
SELECT id, client_name, public FROM identity.oauth2_clients;
```

Or:
```bash
make oauth2-list-clients
```

### Check Users

```sql
SELECT email, display_name, status FROM identity.users;
```

### Check Tenants

```sql
SELECT t.name, t.slug, u.email as owner 
FROM identity.tenants t 
JOIN identity.users u ON t.owner_id = u.id;
```

### Check Tenant Members

```sql
SELECT 
    t.name as tenant,
    u.email as user,
    tm.role
FROM identity.tenant_members tm
JOIN identity.tenants t ON tm.tenant_id = t.id
JOIN identity.users u ON tm.user_id = u.id;
```

## ⚠️ Important Notes

### Development Only

**NEVER use these seeds in production!**

- Passwords are weak (`password123`)
- Secrets are publicly known
- Data is for testing only

### Production Seeding

For production:

1. Create separate production seed scripts
2. Use strong, randomly generated passwords
3. Store secrets in environment variables
4. Use higher bcrypt cost factor (12-14)
5. Remove demo/test data
6. Keep only essential initial data

### Password Verification

To verify a password works with bcrypt:

```sql
SELECT 
    email,
    (password_hash = crypt('password123', password_hash)) as password_matches
FROM identity.users
WHERE email = 'admin@wibusystem.dev';
```

Should return `password_matches: true`

## 📖 Related Documentation

- [OAuth2 Client Management](../docs/OAUTH2_CLIENT_MANAGEMENT.md)
- [Database Schema](../migrations/000001_create_identity_schema.up.sql)
- [START_HERE.md](../START_HERE.md)

## 🆘 Troubleshooting

### "extension pgcrypto does not exist"

```bash
# Enable extension manually
docker exec system_dev psql -U system_dev -d system_dev -c "CREATE EXTENSION pgcrypto;"
```

### Seeds fail with constraint violations

```bash
# Clean database and start fresh
make db-clean-migrations
# Restart app to run migrations
go run ./cmd/server/main.go
# Then seed
make db-seed
```

### Need to re-seed

```bash
# Delete seeded data (careful!)
docker exec system_dev psql -U system_dev -d system_dev << 'EOF'
DELETE FROM identity.tenant_members;
DELETE FROM identity.tenants;
DELETE FROM identity.users;
DELETE FROM identity.oauth2_clients;
EOF

# Re-run seeds
make db-seed
```

## 🎯 Next Steps

1. Run seeds: `make db-seed`
2. Verify data: Check tables in database
3. Test login: Use seeded credentials
4. Test OAuth2: Use seeded clients
5. Develop features: Use seeded tenants

---

**Happy seeding!** 🌱
