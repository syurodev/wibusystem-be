#!/bin/bash
# Check OAuth2 client configuration in database

echo "==================================================="
echo "OAUTH2 CLIENT CONFIGURATION CHECK"
echo "==================================================="
echo ""

CLIENT_ID=${1:-"20000000-0000-0000-0000-000000000001"}

echo "Checking client: $CLIENT_ID"
echo ""

echo "1. Check if client exists in database..."
echo "---------------------------------------------------"
docker compose exec -T system_dev psql -U system_dev -d system_dev -c "
SELECT
    id,
    name,
    is_active,
    created_at
FROM catalog.oauth2_clients
WHERE id = '$CLIENT_ID';
" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "❌ Failed to query database. Is PostgreSQL running?"
    echo "   Run: docker compose ps"
    exit 1
fi

echo ""
echo "2. Check redirect URIs..."
echo "---------------------------------------------------"
docker compose exec -T system_dev psql -U system_dev -d system_dev -c "
SELECT redirect_uri
FROM catalog.oauth2_client_redirect_uris
WHERE client_id = '$CLIENT_ID';
" 2>/dev/null

echo ""
echo "3. Check granted scopes..."
echo "---------------------------------------------------"
docker compose exec -T system_dev psql -U system_dev -d system_dev -c "
SELECT scope
FROM catalog.oauth2_client_scopes
WHERE client_id = '$CLIENT_ID';
" 2>/dev/null

echo ""
echo "4. Check grant types..."
echo "---------------------------------------------------"
docker compose exec -T system_dev psql -U system_dev -d system_dev -c "
SELECT grant_type
FROM catalog.oauth2_client_grant_types
WHERE client_id = '$CLIENT_ID';
" 2>/dev/null

echo ""
echo "==================================================="
echo "COMMON ISSUES"
echo "==================================================="
echo ""
echo "If client not found:"
echo "  → Run SQL to create client (see migrations or admin API)"
echo ""
echo "If redirect_uri mismatch:"
echo "  → Add http://localhost:3000/api/auth/callback to redirect_uris"
echo ""
echo "If client is_active = false:"
echo "  → Update: UPDATE catalog.oauth2_clients SET is_active = true WHERE id = '$CLIENT_ID';"
echo ""
echo "If scopes missing:"
echo "  → Add: openid, profile, email, offline_access"
echo ""
