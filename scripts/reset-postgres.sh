#!/bin/bash
# Reset PostgreSQL volume for version upgrade

echo "==================================================="
echo "RESET POSTGRESQL VOLUME"
echo "==================================================="
echo ""
echo "⚠️  WARNING: This will DELETE all data in PostgreSQL!"
echo "    Only use this in development environment."
echo ""
read -p "Are you sure? Type 'yes' to continue: " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "1. Stopping all containers..."
docker-compose down

echo ""
echo "2. Removing PostgreSQL volume..."
docker volume rm wibusystem-backend_system_data || echo "Volume doesn't exist or already removed"

echo ""
echo "3. Removing Redis volume (optional, for clean slate)..."
read -p "Also reset Redis? (y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker volume rm wibusystem-backend_redis_data || echo "Volume doesn't exist"
fi

echo ""
echo "4. Starting services..."
docker-compose up -d

echo ""
echo "5. Waiting for PostgreSQL to initialize (30 seconds)..."
sleep 30

echo ""
echo "==================================================="
echo "VERIFICATION"
echo "==================================================="
echo ""

echo "✓ PostgreSQL status:"
docker-compose exec -T system_dev psql -U system_dev -d system_dev -c "SELECT version();" || echo "Not ready yet, wait a bit longer"

echo ""
echo "✓ Docker logs (last 10 lines):"
docker logs system_dev --tail 10

echo ""
echo "==================================================="
echo "NEXT STEPS"
echo "==================================================="
echo ""
echo "1. Run migrations to recreate database schema:"
echo "   make migrate-up"
echo ""
echo "   OR manually:"
echo "   go run main.go migrate"
echo ""
echo "2. Seed test data if needed"
echo ""
echo "3. Continue with logging setup"
echo ""
