#!/bin/bash
# Start full application stack with logging
# This includes PostgreSQL, Redis, Loki, Promtail, Grafana, and the Application

set -e  # Exit on error

echo "==================================================="
echo "STARTING FULL APPLICATION STACK"
echo "==================================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Check if PostgreSQL volume exists and is compatible
echo -e "${BLUE}Step 1: Checking PostgreSQL volume...${NC}"
if docker volume inspect wibusystem-backend_system_data &>/dev/null; then
    echo -e "${YELLOW}⚠️  PostgreSQL volume exists.${NC}"
    echo -e "${YELLOW}   If you're upgrading from PostgreSQL <18, you may need to run:${NC}"
    echo -e "${YELLOW}   ./scripts/fix-postgres-upgrade.sh${NC}"
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
else
    echo -e "${GREEN}✓ No existing PostgreSQL volume found (fresh start)${NC}"
fi

# Step 2: Check if OAuth2 keys exist
echo ""
echo -e "${BLUE}Step 2: Checking OAuth2 keys...${NC}"
if [ ! -f "configs/keys/private_key.pem" ]; then
    echo -e "${YELLOW}⚠️  OAuth2 keys not found. Generating...${NC}"
    mkdir -p configs/keys
    openssl genrsa -out configs/keys/private_key.pem 2048 2>/dev/null
    openssl rsa -in configs/keys/private_key.pem -pubout -out configs/keys/public_key.pem 2>/dev/null
    echo -e "${GREEN}✓ OAuth2 keys generated${NC}"
else
    echo -e "${GREEN}✓ OAuth2 keys already exist${NC}"
fi

# Step 3: Stop any existing containers
echo ""
echo -e "${BLUE}Step 3: Stopping existing containers...${NC}"
docker compose down 2>/dev/null || true
echo -e "${GREEN}✓ Containers stopped${NC}"

# Step 4: Start infrastructure (PostgreSQL, Redis)
echo ""
echo -e "${BLUE}Step 4: Starting infrastructure services...${NC}"
docker compose up -d system_dev redis
echo -e "${GREEN}✓ Infrastructure services starting${NC}"

# Step 5: Wait for PostgreSQL to be ready
echo ""
echo -e "${BLUE}Step 5: Waiting for PostgreSQL to be ready...${NC}"
for i in {30..1}; do
    if docker compose exec -T system_dev pg_isready -U system_dev &>/dev/null; then
        echo -e "${GREEN}✓ PostgreSQL is ready!${NC}"
        break
    fi
    echo -n "$i... "
    sleep 1
done
echo ""

# Step 6: Check if migrations are needed
echo ""
echo -e "${BLUE}Step 6: Checking database migrations...${NC}"
MIGRATION_VERSION=$(make migrate-version 2>&1 | grep -oP '\d+' | tail -1 || echo "0")
if [ "$MIGRATION_VERSION" == "0" ] || [ -z "$MIGRATION_VERSION" ]; then
    echo -e "${YELLOW}⚠️  No migrations applied. Running migrations...${NC}"
    make migrate-up
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Migrations completed${NC}"
    else
        echo -e "${RED}❌ Migration failed!${NC}"
        echo -e "${YELLOW}   Check if golang-migrate is installed: make migrate-install${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ Database is at migration version: $MIGRATION_VERSION${NC}"
fi

# Step 7: Build application image
echo ""
echo -e "${BLUE}Step 7: Building application Docker image...${NC}"
docker compose build app
echo -e "${GREEN}✓ Application image built${NC}"

# Step 8: Start all services
echo ""
echo -e "${BLUE}Step 8: Starting all services (app, logging stack)...${NC}"
docker compose up -d
echo -e "${GREEN}✓ All services started${NC}"

# Step 9: Wait for application to be ready
echo ""
echo -e "${BLUE}Step 9: Waiting for application to be ready...${NC}"
for i in {20..1}; do
    if curl -s http://localhost:8080/health &>/dev/null; then
        echo -e "${GREEN}✓ Application is ready!${NC}"
        break
    fi
    echo -n "$i... "
    sleep 1
done
echo ""

# Step 10: Display status
echo ""
echo "==================================================="
echo -e "${GREEN}✅ FULL STACK STARTED SUCCESSFULLY!${NC}"
echo "==================================================="
echo ""
echo -e "${BLUE}Services Status:${NC}"
docker compose ps
echo ""

echo -e "${BLUE}📊 Access Points:${NC}"
echo "  • Application:  http://localhost:8080"
echo "  • Health Check: http://localhost:8080/health"
echo "  • Grafana:      http://localhost:5555 (admin/admin)"
echo "  • Loki API:     http://localhost:3100"
echo "  • Promtail:     http://localhost:9080"
echo ""

echo -e "${BLUE}📝 View Logs:${NC}"
echo "  • Application:  docker logs -f wibusystem-be"
echo "  • PostgreSQL:   docker logs -f system_dev"
echo "  • All services: docker compose logs -f"
echo ""

echo -e "${BLUE}🔍 Test Logging:${NC}"
echo "  1. Generate some logs:"
echo "     curl http://localhost:8080/health"
echo ""
echo "  2. Check logs in Loki (wait ~10 seconds for ingestion):"
echo "     curl -G http://localhost:3100/loki/api/v1/query \\"
echo "       --data-urlencode 'query={container_name=\"wibusystem-be\"}' | jq"
echo ""
echo "  3. View in Grafana:"
echo "     - Open http://localhost:5555"
echo "     - Go to Explore"
echo "     - Select Loki datasource"
echo "     - Query: {container_name=\"wibusystem-be\"}"
echo ""

echo -e "${BLUE}🛠️  Useful Commands:${NC}"
echo "  • Stop all:     docker compose down"
echo "  • View logs:    docker compose logs -f app"
echo "  • Rebuild app:  docker compose up -d --build app"
echo "  • Database:     make db-shell"
echo ""
