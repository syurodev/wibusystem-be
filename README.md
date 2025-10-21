# WibuSystem Backend - Modular Monolith

A modern Go backend application built with modular monolith architecture for scalability, maintainability, and ease of development.

## 🏗️ Architecture

This project follows a **Modular Monolith** pattern, where business domains are organized as independent modules within a single deployable unit.

### Why Modular Monolith?

- **Single Deployment**: Easier to deploy and manage than microservices
- **Shared Infrastructure**: No duplication of database connections, logging, etc.
- **Better Performance**: No network overhead between modules
- **Simpler Development**: Easier to develop and debug
- **Module Independence**: Each module can be extracted to a microservice if needed

---

## 📁 Project Structure

```
wibusystem-be/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── modules/                 # Business domain modules
│   │   ├── identity/            # ✅ Authentication & Authorization (95% complete)
│   │   │   ├── domain/          # Business entities
│   │   │   ├── repository/      # Data access layer
│   │   │   ├── service/         # Business logic
│   │   │   ├── handler/         # HTTP handlers & middleware
│   │   │   └── dto/             # Data transfer objects
│   │   ├── catalog/             # 📋 Content management (planned)
│   │   ├── community/           # 📋 Social features (planned)
│   │   └── payment/             # 📋 Payment processing (planned)
│   ├── infrastructure/          # Shared infrastructure
│   │   ├── database/            # Database connections
│   │   ├── cache/               # Caching layer
│   │   └── messaging/           # Message queue
│   └── platform/                # Cross-cutting concerns
│       └── config/              # Configuration management
├── migrations/                  # Database migrations
├── docs/                        # Documentation
├── postman/                     # API collections
├── docker-compose.yml           # Local development setup
├── go.mod                       # Go dependencies
└── Makefile                     # Common tasks
```

---

## 🎯 Modules Overview

### ✅ Identity Module (95% Complete)

**Purpose:** User authentication, authorization, and tenant management

**Features:**
- ✅ User registration and login
- ✅ Session-based authentication
- ✅ Multi-tenant support
- ✅ Role-based access control (Owner, Admin, Member, Viewer)
- ✅ Password management (reset, change)
- ✅ Email verification
- ✅ Session management

**API Endpoints:** 40+ RESTful endpoints

**Statistics:**
- Domain Layer: 1,447 lines
- Repository Layer: 2,607 lines
- Service Layer: 1,118 lines
- Handler Layer: 3,240 lines
- Total: **9,733 lines** of production code

**Documentation:** `internal/modules/identity/handler/README.md`

---

### 📋 Catalog Module (Planned)

**Purpose:** Content management for novels, chapters, volumes

**Planned Features:**
- Novel CRUD operations
- Chapter and volume management
- Creator profiles
- Character management
- Genre categorization
- Search and filtering

**Status:** Migration from `services/catalog/` pending

**Documentation:** `internal/modules/catalog/README.md`

---

### 📋 Community Module (Planned)

**Purpose:** Social features and user engagement

**Planned Features:**
- Comments and replies
- Reviews and ratings
- Discussions/forums
- Notifications
- Following system
- Bookmarks

**Status:** Design phase

**Documentation:** `internal/modules/community/README.md`

---

### 📋 Payment Module (Planned)

**Purpose:** Payment processing and monetization

**Planned Features:**
- Payment processing (Stripe, PayPal)
- Subscription management
- Wallet system
- Invoice generation
- Refund handling
- Premium content access

**Status:** Design phase

**Documentation:** `internal/modules/payment/README.md`

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14+
- Docker & Docker Compose (for local development)
- Make (optional, for convenience commands)

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/wibusystem-be.git
cd wibusystem-be
```

### 2. Set Up Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 3. Start Dependencies

```bash
# Start PostgreSQL using Docker Compose
docker-compose up -d system-db

# Wait for database to be ready
sleep 5
```

### 4. Install Dependencies

```bash
# Install Go dependencies (including Fiber framework)
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/limiter
go get github.com/gofiber/storage/memory/v2

# Tidy dependencies
go mod tidy
```

### 5. Run Migrations

```bash
# Run database migrations
make migrate-up

# Or manually
go run cmd/migrate/main.go up
```

### 6. Start the Server

```bash
# Development mode
go run cmd/server/main.go

# Or build and run
make build
./bin/server
```

### 7. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "display_name": "John Doe"
  }'
```

---

## 🔧 Development

### Available Commands

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run migrations
make migrate-up
make migrate-down

# Format code
make fmt

# Lint code
make lint

# Clean build artifacts
make clean
```

### Project Setup for New Modules

Each module follows a consistent 4-layer architecture:

1. **Domain Layer** - Business entities and rules
2. **Repository Layer** - Data access interfaces and implementations
3. **Service Layer** - Business logic and use cases
4. **Handler Layer** - HTTP API and middleware

See `internal/modules/identity/` as a reference implementation.

---

## 📚 API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

All protected endpoints require a session token in the Authorization header:

```
Authorization: Bearer <session_token>
```

### Endpoints Overview

**Authentication** (`/api/v1/auth`)
- POST `/register` - Register new user
- POST `/login` - Login user
- POST `/logout` - Logout user
- POST `/refresh` - Refresh session
- GET `/me` - Get current user

**Users** (`/api/v1/users`)
- GET `/profile` - Get user profile
- PUT `/profile` - Update profile
- DELETE `/account` - Delete account
- GET `/sessions` - List sessions

**Tenants** (`/api/v1/tenants`)
- POST `/` - Create tenant
- GET `/` - List tenants
- GET `/:tenantId` - Get tenant details
- PUT `/:tenantId` - Update tenant
- DELETE `/:tenantId` - Delete tenant
- POST `/:tenantId/members` - Add member
- GET `/:tenantId/members` - List members

**Sessions** (`/api/v1/sessions`)
- GET `/` - List user sessions
- DELETE `/:sessionId` - Revoke session
- DELETE `/` - Revoke all sessions

For complete API documentation, see:
- Identity Module: `internal/modules/identity/handler/README.md`
- Postman Collection: `postman/Identity_API.postman_collection.json`

---

## 🧪 Testing

### Run Tests

```bash
# All tests
go test ./...

# Specific module
go test ./internal/modules/identity/...

# With coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

### Test Structure

```
internal/modules/identity/
├── domain/
│   └── user_test.go
├── repository/
│   └── postgres/
│       └── user_repository_test.go
├── service/
│   └── auth_service_test.go
└── handler/
    └── http/
        └── auth_handler_test.go
```

---

## 🔒 Security

### Authentication & Authorization

- Session-based authentication with secure token generation
- Bcrypt password hashing (cost 10)
- Role-based access control (RBAC)
- Tenant-level permissions

### Rate Limiting

| Endpoint Type | Limit | Window |
|--------------|-------|--------|
| Global | 1000 requests | 1 minute |
| Authentication | 5 requests | 15 minutes |
| Registration | 3 requests | 1 hour |
| Password Reset | 3 requests | 1 hour |
| API Calls | 100 requests | 1 minute |

### Input Validation

- Automatic request validation using struct tags
- Custom validators (email, slug, password strength)
- Input sanitization (XSS prevention)
- SQL injection prevention (parameterized queries)

---

## 🚢 Deployment

### Docker

```bash
# Build Docker image
docker build -t wibusystem-be:latest .

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_NAME=system_prod \
  wibusystem-be:latest
```

### Docker Compose

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

### Environment Variables

```env
# Server
SERVER_PORT=8080
ENVIRONMENT=production

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=system_prod
DB_USER=postgres
DB_PASSWORD=your_secure_password

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Security
SESSION_SECRET=your_session_secret
```

---

## 📊 Current Progress

### Overall: 23.75% Complete

| Module | Progress | Status |
|--------|----------|--------|
| Identity | 95% | ✅ Complete (needs testing) |
| Catalog | 0% | 📋 Planned |
| Community | 0% | 📋 Planned |
| Payment | 0% | 📋 Planned |

### Identity Module Breakdown

- ✅ Domain Layer - 100%
- ✅ Repository Layer - 100%
- ✅ Service Layer - 100%
- ✅ Handler Layer - 100%
- ⏳ Testing - In Progress

### Next Steps

1. Complete Identity module testing (2-3 days)
2. Migrate Catalog module (10-12 days)
3. Implement Payment module (12-15 days)
4. Implement Community module (10-12 days)

---

## 📖 Documentation

### Getting Started
- `START_HERE.md` - Quick start guide
- `POC_README.md` - Proof of concept documentation
- `INTEGRATION_CHECKLIST.md` - Integration steps

### Architecture & Design
- `docs/HANDLER_LAYER_GUIDE.md` - Complete handler implementation guide
- `docs/migration/IDENTITY_PROGRESS.md` - Identity module progress
- `MIGRATION_CLEANUP.md` - Migration summary

### Module Documentation
- `internal/modules/identity/handler/README.md` - Identity API reference
- `internal/modules/catalog/README.md` - Catalog migration plan
- `internal/modules/community/README.md` - Community features
- `internal/modules/payment/README.md` - Payment features

---

## 🛠️ Tech Stack

### Core
- **Language:** Go 1.21+
- **Web Framework:** Fiber v2 (high-performance HTTP framework)
- **Database:** PostgreSQL 14+
- **Validation:** go-playground/validator v10

### Libraries
- **UUID:** google/uuid
- **Crypto:** golang.org/x/crypto (bcrypt)
- **Database Driver:** jackc/pgx/v5
- **CORS:** gofiber/cors
- **Rate Limiting:** gofiber/limiter
- **Compression:** gofiber/compress

### Tools
- **Migrations:** golang-migrate
- **Testing:** testify
- **Linting:** golangci-lint
- **Documentation:** Markdown

---

## 🤝 Contributing

### Code Style

- Follow Go best practices and idioms
- Use `gofmt` for formatting
- Write clear, descriptive commit messages
- Add tests for new features
- Update documentation

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Module Development Guidelines

When developing new modules:
1. Follow the 4-layer architecture (Domain, Repository, Service, Handler)
2. Write comprehensive tests (unit, integration, e2e)
3. Document all public APIs
4. Add request/response examples
5. Update module README

---

## 📝 License

This project is proprietary and confidential.

Copyright © 2024-2025 WibuSystem. All rights reserved.

---

## 📞 Support

### Issues & Bugs

Report issues on GitHub: [Issues](https://github.com/yourusername/wibusystem-be/issues)

### Documentation

- API Docs: `internal/modules/*/README.md`
- Architecture: `docs/`
- Examples: `postman/`

### Contact

- Email: support@wibusystem.com
- Website: https://wibusystem.com

---

## 🎯 Roadmap

### Q1 2025
- ✅ Identity Module (Complete)
- ⏳ Catalog Module (In Progress)
- ⏳ Payment Module (Planning)

### Q2 2025
- Community Module
- Advanced search features
- Real-time notifications
- Mobile API optimizations

### Q3 2025
- Analytics dashboard
- Content recommendations
- Performance optimizations
- International support

### Q4 2025
- Advanced monetization features
- Creator tools
- Community gamification
- Scale to 1M+ users

---

**Built with ❤️ by the WibuSystem Team**

**Last Updated:** January 2025  
**Version:** 0.1.0 (Alpha)  
**Status:** Active Development