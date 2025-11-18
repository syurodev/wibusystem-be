#!/bin/bash
# Fix PostgreSQL version upgrade issue
# This script will reset PostgreSQL volume and run migrations

echo "==================================================="
echo "FIX POSTGRESQL VERSION UPGRADE"
echo "==================================================="
echo ""
echo "Issue: PostgreSQL 18 incompatible with old data volume"
echo "Solution: Reset volume and re-run migrations"
echo ""
echo "⚠️  WARNING: This will DELETE all PostgreSQL data!"
echo "             Only use in DEVELOPMENT environment."
echo ""
read -p "Continue? Type 'YES' to proceed: " confirm

if [ "$confirm" != "YES" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "Step 1: Stopping all containers..."
docker-compose down

echo ""
echo "Step 2: Removing PostgreSQL volume..."
docker volume rm wibusystem-backend_system_data 2>/dev/null || echo "  (Volume already removed or doesn't exist)"

echo ""
echo "Step 3: Also reset other volumes? (optional)"
read -p "Reset Redis data? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker volume rm wibusystem-backend_redis_data 2>/dev/null || echo "  (Already removed)"
fi

read -p "Reset Loki data? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker volume rm wibusystem-backend_loki_data 2>/dev/null || echo "  (Already removed)"
fi

read -p "Reset Grafana data? (y/N) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker volume rm wibusystem-backend_grafana_data 2>/dev/null || echo "  (Already removed)"
fi

echo ""
echo "Step 4: Starting PostgreSQL..."
docker-compose up -d system_dev

echo ""
echo "Step 5: Waiting for PostgreSQL to initialize (20 seconds)..."
for i in {20..1}; do
    echo -n "$i... "
    sleep 1
done
echo ""

echo ""
echo "Step 6: Verifying PostgreSQL..."
docker-compose exec -T system_dev psql -U system_dev -d system_dev -c "SELECT version();" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "✓ PostgreSQL is ready!"
else
    echo "⚠️  PostgreSQL not ready yet. Waiting 10 more seconds..."
    sleep 10
    docker-compose exec -T system_dev psql -U system_dev -d system_dev -c "SELECT version();" 2>/dev/null || {
        echo "❌ PostgreSQL failed to start. Check logs:"
        echo "   docker logs system_dev"
        exit 1
    }
fi

echo ""
echo "Step 7: Running database migrations..."
make migrate-up

if [ $? -eq 0 ]; then
    echo "✓ Migrations completed successfully!"
else
    echo "⚠️  Migration failed. You may need to:"
    echo "   1. Check if golang-migrate is installed: make migrate-install"
    echo "   2. Manually run: make migrate-up"
fi

echo ""
echo "Step 8: Starting remaining services..."
docker-compose up -d

echo ""
echo "==================================================="
echo "VERIFICATION"
echo "==================================================="
echo ""

echo "✓ All services status:"
docker-compose ps

echo ""
echo "✓ PostgreSQL tables:"
docker-compose exec -T system_dev psql -U system_dev -d system_dev -c "\dt identify.*" 2>/dev/null | head -20

echo ""
echo "==================================================="
echo "✅ SETUP COMPLETE!"
echo "==================================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Verify migrations created tables:"
echo "   make db-shell"
echo "   \\dt identify.*"
echo ""
echo "2. Create OAuth2 keys (if not exists):"
echo "   mkdir -p configs/keys"
echo "   openssl genrsa -out configs/keys/private_key.pem 2048"
echo ""
echo "3. Update .env with required secrets"
echo ""
echo "4. Run application:"
echo "   make run"
echo ""
echo "5. Test logging (if using Docker):"
echo "   ./scripts/restart-logging.sh"
echo ""
