# Makefile for System Project
# Author: System Team
# Description: Helper commands cho development, testing và deployment

# =====================================================
# Variables
# =====================================================
BINARY_NAME=server
BINARY_DIR=bin
MAIN_PATH=./cmd/server
MIGRATIONS_DIR=./migrations

# Database connection (from .env)
DB_HOST?=localhost
DB_PORT?=5432
DB_NAME?=system_dev
DB_USER?=system_dev
DB_PASSWORD?=system_dev
DB_SSL_MODE?=disable

# Build migration database URL
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# =====================================================
# Help Command
# =====================================================
.PHONY: help
help: ## Hiển thị help message
	@echo "$(BLUE)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# =====================================================
# Development Commands
# =====================================================
.PHONY: run
run: ## Chạy application
	@echo "$(BLUE)Starting application...$(NC)"
	go run $(MAIN_PATH)/main.go

.PHONY: build
build: ## Build application binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build completed: $(BINARY_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: clean
clean: ## Xóa build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BINARY_DIR)
	@echo "$(GREEN)✓ Clean completed$(NC)"

.PHONY: test
test: ## Chạy tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests completed$(NC)"

.PHONY: test-coverage
test-coverage: test ## Xem test coverage
	@echo "$(BLUE)Opening coverage report...$(NC)"
	go tool cover -html=coverage.out

.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	gofmt -s -w .
	@echo "$(GREEN)✓ Format completed$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run ./...
	@echo "$(GREEN)✓ Lint completed$(NC)"

.PHONY: tidy
tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	go mod tidy
	@echo "$(GREEN)✓ Tidy completed$(NC)"

# =====================================================
# Database Commands
# =====================================================
.PHONY: db-create
db-create: ## Tạo database
	@echo "$(BLUE)Creating database $(DB_NAME)...$(NC)"
	docker exec -it system_dev psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);" 2>/dev/null || true
	@echo "$(GREEN)✓ Database created (or already exists)$(NC)"

.PHONY: db-drop
db-drop: ## Xóa database (⚠️ CẢNH BÁO: Mất toàn bộ dữ liệu!)
	@echo "$(RED)⚠️  WARNING: This will delete the database $(DB_NAME)!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker exec -it system_dev psql -U postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"; \
		echo "$(GREEN)✓ Database dropped$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

.PHONY: db-reset
db-reset: db-drop db-create ## Reset database (drop + create)
	@echo "$(GREEN)✓ Database reset completed$(NC)"

.PHONY: db-shell
db-shell: ## Kết nối vào PostgreSQL shell
	@echo "$(BLUE)Connecting to PostgreSQL...$(NC)"
	docker exec -it system_dev psql -U $(DB_USER) -d $(DB_NAME)

# =====================================================
# Migration Commands
# =====================================================
.PHONY: migrate-install
migrate-install: ## Cài đặt golang-migrate tool
	@echo "$(BLUE)Installing golang-migrate...$(NC)"
	@which migrate > /dev/null || ( \
		echo "$(YELLOW)Installing migrate...$(NC)" && \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
	)
	@echo "$(GREEN)✓ golang-migrate installed$(NC)"
	@migrate -version

.PHONY: migrate-up
migrate-up: ## Chạy all pending migrations
	@echo "$(BLUE)Running migrations up...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations up completed$(NC)"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Migration rolled back$(NC)"

.PHONY: migrate-down-all
migrate-down-all: ## Rollback all migrations (⚠️ CẢNH BÁO!)
	@echo "$(RED)⚠️  WARNING: This will rollback ALL migrations!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down; \
		echo "$(GREEN)✓ All migrations rolled back$(NC)"; \
	else \
		echo "$(YELLOW)Cancelled$(NC)"; \
	fi

.PHONY: migrate-force
migrate-force: ## Force migration version (sử dụng: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Usage: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration to version $(VERSION)...$(NC)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)
	@echo "$(GREEN)✓ Migration forced to version $(VERSION)$(NC)"

.PHONY: migrate-version
migrate-version: ## Hiển thị current migration version
	@echo "$(BLUE)Current migration version:$(NC)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-create
migrate-create: ## Tạo migration file mới (sử dụng: make migrate-create NAME=create_users)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=create_users$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)...$(NC)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "$(GREEN)✓ Migration files created$(NC)"

.PHONY: migrate-status
migrate-status: ## Hiển thị migration status
	@echo "$(BLUE)Migration status:$(NC)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version || echo "No migrations applied yet"

.PHONY: migrate-clickhouse
migrate-clickhouse: ## Chạy ClickHouse migrations
	@echo "$(BLUE)Running ClickHouse migrations...$(NC)"
	go run scripts/init_clickhouse.go
	@echo "$(GREEN)✓ ClickHouse migrations completed$(NC)"

# =====================================================
# Docker Commands
# =====================================================
.PHONY: docker-up
docker-up: ## Start docker services
	@echo "$(BLUE)Starting Docker services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Docker services started$(NC)"

.PHONY: docker-down
docker-down: ## Stop docker services
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Docker services stopped$(NC)"

.PHONY: docker-logs
docker-logs: ## Xem docker logs
	docker-compose logs -f

.PHONY: docker-ps
docker-ps: ## Hiển thị docker containers
	docker-compose ps

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart docker services

# =====================================================
# Seed Commands
# =====================================================
.PHONY: seed
seed: ## Seed roles and permissions
	@echo "$(BLUE)Seeding roles and permissions...$(NC)"
	docker exec -i system_dev psql -U $(DB_USER) -d $(DB_NAME) < scripts/seed_roles_permissions.sql
	@echo "$(GREEN)✓ Seed completed$(NC)"

.PHONY: seed-admin
seed-admin: ## Seed admin user and internal OAuth2 clients
	@echo "$(BLUE)Seeding admin user and OAuth2 clients...$(NC)"
	docker exec -i system_dev psql -U $(DB_USER) -d $(DB_NAME) < scripts/seed_admin_and_clients.sql
	@echo "$(GREEN)✓ Admin and clients seeded$(NC)"

.PHONY: seed-test
seed-test: ## Seed test data (users and OAuth2 clients)
	@echo "$(BLUE)Seeding test data...$(NC)"
	docker exec -i system_dev psql -U $(DB_USER) -d $(DB_NAME) < scripts/seed_test_data.sql
	@echo "$(GREEN)✓ Test data seeded$(NC)"

.PHONY: seed-all
seed-all: seed seed-admin seed-test ## Seed all data (roles, permissions, admin, and test data)
	@echo "$(GREEN)✓ All seed data completed$(NC)"

# =====================================================
# Setup Commands
# =====================================================
.PHONY: setup
setup: migrate-install docker-up ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@sleep 3 # Wait for database to be ready
	@$(MAKE) migrate-up
	@$(MAKE) migrate-clickhouse
	@$(MAKE) seed
	@$(MAKE) seed-admin
	@echo "$(GREEN)✓ Setup completed!$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Run '$(GREEN)make run$(NC)' to start the application"
	@echo "  2. Run '$(GREEN)make test$(NC)' to run tests"
	@echo "  3. Run '$(GREEN)make seed-test$(NC)' to seed test data (optional)"

.PHONY: dev
dev: ## Start development environment (docker + migrations + run)
	@$(MAKE) docker-up
	@sleep 3
	@$(MAKE) migrate-up
	@$(MAKE) run

# =====================================================
# Default Target
# =====================================================
.DEFAULT_GOAL := help
