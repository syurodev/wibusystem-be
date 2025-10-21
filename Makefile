# ============================================================================
# WibuSystem Modular Monolith - Makefile
# ============================================================================

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Development Commands
# ============================================================================

.PHONY: setup
setup: ## Initial setup for development
	@echo "🔧 Setting up development environment..."
	cp -n .env.example .env || true
	@echo "✅ .env file created (edit if needed)"
	@echo "📦 Downloading Go dependencies..."
	go mod download
	@echo "✅ Setup complete!"

.PHONY: run
run: ## Run the application
	@echo "🚀 Starting WibuSystem..."
	go run ./cmd/server

.PHONY: build
build: ## Build the application
	@echo "🔨 Building application..."
	go build -o bin/wibusystem ./cmd/server
	@echo "✅ Binary created at: bin/wibusystem"

.PHONY: build-linux
build-linux: ## Build for Linux (useful for Docker)
	@echo "🔨 Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o bin/wibusystem-linux ./cmd/server
	@echo "✅ Linux binary created at: bin/wibusystem-linux"

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@echo "🔥 Starting with hot reload..."
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/cosmtrek/air@latest; }
	air

# ============================================================================
# Database Commands
# ============================================================================

.PHONY: db-up
db-up: ## Start database and Redis
	@echo "🗄️  Starting database services..."
	docker compose up -d system_dev redis
	@echo "✅ Database services started"
	@echo "   PostgreSQL: localhost:5432"
	@echo "   Redis:      localhost:6379"

.PHONY: db-down
db-down: ## Stop database and Redis
	@echo "🛑 Stopping database services..."
	docker compose down
	@echo "✅ Database services stopped"

.PHONY: db-logs
db-logs: ## Show database logs
	docker compose logs -f system_dev

.PHONY: db-connect
db-connect: ## Connect to database with psql
	@echo "🔗 Connecting to database..."
	docker exec -it system_dev psql -U system_dev -d system_dev

.PHONY: db-reset
db-reset: ## Reset database (WARNING: destroys all data)
	@echo "⚠️  WARNING: This will destroy all data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose down -v; \
		docker compose up -d system_dev redis; \
		sleep 3; \
		make run; \
	fi

.PHONY: db-migrate-up
db-migrate-up: ## Run migrations (handled automatically by app)
	@echo "Migrations are run automatically when the application starts"
	@echo "Or just run: make run"

.PHONY: db-shell
db-shell: ## Open database shell
	@echo "📊 Database shell commands:"
	@echo "  \\dn                    - List schemas"
	@echo "  \\dt identity.*         - List identity tables"
	@echo "  \\d identity.users      - Describe users table"
	@echo "  \\q                     - Quit"
	@echo ""
	docker exec -it system_dev psql -U system_dev -d system_dev

.PHONY: db-status
db-status: ## Show database status
	@echo "🔍 Database Status:"
	@docker compose ps system_dev redis
	@echo ""
	@echo "📊 Database Info:"
	@docker exec -it system_dev psql -U system_dev -d system_dev -c "\dn" 2>/dev/null || echo "Database not running"

# ============================================================================
# Testing Commands
# ============================================================================

.PHONY: test
test: ## Run all tests
	@echo "🧪 Running tests..."
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	go test ./... -tags=integration -v

.PHONY: bench
bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	go test ./... -bench=. -benchmem

# ============================================================================
# Code Quality Commands
# ============================================================================

.PHONY: fmt
fmt: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "🧹 Tidying go.mod..."
	go mod tidy
	@echo "✅ go.mod tidied"

.PHONY: check
check: fmt vet lint ## Run all code quality checks
	@echo "✅ All checks passed"

# ============================================================================
# Load Testing
# ============================================================================

.PHONY: load-test
load-test: ## Run load test on health endpoint
	@echo "📊 Running load test..."
	@command -v hey >/dev/null 2>&1 || { echo "Installing hey..."; go install github.com/rakyll/hey@latest; }
	hey -n 10000 -c 100 http://localhost:8080/health

.PHONY: load-test-api
load-test-api: ## Run load test on API endpoint
	@echo "📊 Running load test on API..."
	@command -v hey >/dev/null 2>&1 || { echo "Installing hey..."; go install github.com/rakyll/hey@latest; }
	hey -n 10000 -c 100 http://localhost:8080/api/v1/

# ============================================================================
# Docker Commands
# ============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t wibusystem:latest -f deployments/docker/Dockerfile .

.PHONY: docker-run
docker-run: ## Run application in Docker
	@echo "🐳 Running Docker container..."
	docker run --rm -p 8080:8080 --env-file .env wibusystem:latest

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	docker compose logs -f

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "🧹 Cleaning Docker resources..."
	docker compose down -v
	docker system prune -f

# ============================================================================
# Migration Management
# ============================================================================

.PHONY: migration-create
migration-create: ## Create new migration (usage: make migration-create NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Error: NAME is required"; \
		echo "Usage: make migration-create NAME=add_users_table"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(NAME)"
	@TIMESTAMP=$$(date +%s); \
	touch migrations/$${TIMESTAMP}_$(NAME).up.sql; \
	touch migrations/$${TIMESTAMP}_$(NAME).down.sql; \
	echo "✅ Created migrations/$${TIMESTAMP}_$(NAME).up.sql"; \
	echo "✅ Created migrations/$${TIMESTAMP}_$(NAME).down.sql"

# ============================================================================
# Monitoring & Debugging
# ============================================================================

.PHONY: logs
logs: ## Show application logs (when running with docker)
	docker compose logs -f

.PHONY: health
health: ## Check application health
	@echo "🏥 Checking application health..."
	@curl -s http://localhost:8080/health | jq . || echo "Application not running or jq not installed"

.PHONY: stats
stats: ## Show database pool statistics
	@echo "📊 Database Pool Statistics:"
	@echo "Check application logs for: 'DB Pool Stats'"

# ============================================================================
# Cleanup Commands
# ============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	rm -f coverage.out coverage.html
	@echo "✅ Cleaned"

.PHONY: clean-all
clean-all: clean docker-clean ## Clean everything (build artifacts + docker)
	@echo "✅ Everything cleaned"

# ============================================================================
# Installation Commands
# ============================================================================

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/rakyll/hey@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Tools installed"

# ============================================================================
# PoC Specific Commands
# ============================================================================

.PHONY: poc-start
poc-start: ## Start PoC from scratch
	@echo "🚀 Starting PoC..."
	make setup
	make db-up
	sleep 3
	make run

.PHONY: poc-verify
poc-verify: ## Verify PoC is working
	@echo "✅ Verifying PoC..."
	@echo ""
	@echo "1. Checking database connection..."
	@docker compose ps system_dev | grep -q "Up" && echo "   ✅ Database is running" || echo "   ❌ Database is not running"
	@echo ""
	@echo "2. Checking schemas..."
	@docker exec system_dev psql -U system_dev -d system_dev -c "\dn" 2>/dev/null | grep -q "identity" && echo "   ✅ Identity schema exists" || echo "   ❌ Identity schema missing"
	@echo ""
	@echo "3. Checking application health..."
	@curl -s http://localhost:8080/health >/dev/null 2>&1 && echo "   ✅ Application is healthy" || echo "   ❌ Application is not responding"
	@echo ""
	@echo "4. API endpoint..."
	@curl -s http://localhost:8080/api/v1/ >/dev/null 2>&1 && echo "   ✅ API is accessible" || echo "   ❌ API is not accessible"

.PHONY: poc-info
poc-info: ## Show PoC information
	@echo "📚 WibuSystem Modular Monolith PoC"
	@echo ""
	@echo "🌐 Endpoints:"
	@echo "   Health:      http://localhost:8080/health"
	@echo "   Readiness:   http://localhost:8080/health/ready"
	@echo "   Liveness:    http://localhost:8080/health/live"
	@echo "   API:         http://localhost:8080/api/v1/"
	@echo ""
	@echo "🗄️  Database:"
	@echo "   Host:        localhost:5432"
	@echo "   Database:    system_dev"
	@echo "   User:        system_dev"
	@echo "   Password:    system_dev"
	@echo ""
	@echo "📖 Documentation:"
	@echo "   PoC Guide:   POC_README.md"
	@echo "   Migration:   docs/migration/to-modular-monolith.md"
	@echo "   Comparison:  docs/migration/comparison.md"

# ============================================================================
# Default Target
# ============================================================================

.DEFAULT_GOAL := help
