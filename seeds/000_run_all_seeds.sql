-- Master Seed Script
-- Description: Runs all seed scripts in order
-- Usage: psql -U system_dev -d system_dev -f seeds/000_run_all_seeds.sql

\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║                                                                      ║'
\echo '║                    Starting Database Seeding                        ║'
\echo '║                                                                      ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''

-- Start transaction
BEGIN;

-- Run seeds in order
\echo '📦 [1/2] Seeding OAuth2 clients...'
\i seeds/001_oauth2_clients.sql

\echo ''
\echo '📦 [2/2] Seeding users and tenants...'
\i seeds/002_users_and_tenants.sql

-- Commit transaction
COMMIT;

\echo ''
\echo '╔══════════════════════════════════════════════════════════════════════╗'
\echo '║                                                                      ║'
\echo '║                  ✅ Database Seeding Complete!                       ║'
\echo '║                                                                      ║'
\echo '╚══════════════════════════════════════════════════════════════════════╝'
\echo ''
\echo '📊 Summary:'
\echo '━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━'

-- Show summary statistics
SELECT
    (SELECT COUNT(*) FROM identity.oauth2_clients) as oauth2_clients,
    (SELECT COUNT(*) FROM identity.users) as users,
    (SELECT COUNT(*) FROM identity.tenants) as tenants,
    (SELECT COUNT(*) FROM identity.tenant_members) as tenant_members;

\echo ''
\echo '🚀 You can now:'
\echo '   • Login with test credentials'
\echo '   • Use OAuth2 clients for authentication'
\echo '   • Test multi-tenancy features'
\echo ''
\echo '📖 See seeds/README.md for more information'
\echo ''
