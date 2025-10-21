# Quick Reference Card - WibuSystem Backend

**Last Updated:** January 2025  
**For:** Developers working on WibuSystem Backend

---

## 🚀 Quick Start (5 minutes)

```bash
# 1. Start database
docker-compose up -d system-db

# 2. Install dependencies
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/limiter
go get github.com/gofiber/storage/memory/v2
go mod tidy

# 3. Run migrations
make migrate-up

# 4. Start server
go run cmd/server/main.go
```

Server runs at: `http://localhost:8080`

---

## 📁 Project Structure

```
internal/modules/identity/        # ✅ Complete (95%)
├── domain/                       # Business entities
├── repository/                   # Data access
├── service/                      # Business logic
├── handler/                      # HTTP API
└── dto/                          # Request/Response

internal/modules/catalog/         # 📋 Planned
internal/modules/community/       # 📋 Planned
internal/modules/payment/         # 📋 Planned
```

---

## 🔧 Common Commands

```bash
# Development
make run                          # Start server
make build                        # Build binary
make test                         # Run tests
make fmt                          # Format code
make lint                         # Lint code

# Database
make migrate-up                   # Run migrations
make migrate-down                 # Rollback migrations
make db-reset                     # Reset database

# Docker
docker-compose up -d              # Start all services
docker-compose down               # Stop all services
docker-compose logs -f server     # View logs
```

---

## 🌐 API Endpoints (Identity Module)

### Base URL
```
http://localhost:8080/api/v1
```

### Authentication
```bash
# Register
POST /auth/register
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "display_name": "John Doe"
}

# Login
POST /auth/login
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
# Returns: session_token

# Logout
POST /auth/logout
Authorization: Bearer <token>

# Get current user
GET /auth/me
Authorization: Bearer <token>
```

### Users
```bash
# Get profile
GET /users/profile
Authorization: Bearer <token>

# Update profile
PUT /users/profile
Authorization: Bearer <token>
{
  "display_name": "New Name"
}

# List sessions
GET /users/sessions
Authorization: Bearer <token>
```

### Tenants
```bash
# Create tenant
POST /tenants
Authorization: Bearer <token>
{
  "name": "My Organization",
  "slug": "my-org",
  "description": "Description"
}

# List my tenants
GET /tenants/my-tenants
Authorization: Bearer <token>

# Add member
POST /tenants/:tenantId/members
Authorization: Bearer <token>
{
  "user_id": "uuid",
  "role": "admin"
}
```

### System
```bash
# Health check
GET /health

# Metrics (dev only)
GET /metrics
```

---

## 🔒 Authentication

All protected endpoints require:
```
Authorization: Bearer <session_token>
```

Get token from login response:
```json
{
  "session_token": "your-token-here",
  "expires_at": "2025-01-22T10:30:00Z"
}
```

---

## 🧪 Testing

```bash
# Quick API test
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}'

# Use token (replace YOUR_TOKEN)
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## ⚙️ Configuration

### Environment Variables (.env)
```env
# Server
SERVER_PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=system_dev
DB_USER=postgres
DB_PASSWORD=postgres

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

---

## 📊 Database

### Connection
```
postgresql://postgres:postgres@localhost:5432/system_dev
```

### Schemas
- `identity` - Users, tenants, sessions
- `catalog` - Novels, chapters (future)
- `community` - Comments, reviews (future)
- `payment` - Transactions (future)

### Migrations Location
```
migrations/
├── 000001_create_identity_schema.up.sql
└── 000001_create_identity_schema.down.sql
```

---

## 🐛 Troubleshooting

### Server won't start
```bash
# Check port 8080
lsof -i :8080

# Check database
docker-compose ps
psql -h localhost -U postgres -d system_dev
```

### Database connection failed
```bash
# Restart database
docker-compose restart system-db

# Check logs
docker-compose logs system-db
```

### Compilation errors
```bash
# Clean and reinstall
go clean -modcache
go mod download
go mod tidy
```

### Rate limit errors
- Wait 15 minutes (auth endpoints)
- Or restart server (dev only)

---

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| `README.md` | Main documentation |
| `START_HERE.md` | Getting started |
| `INTEGRATION_CHECKLIST.md` | Step-by-step setup |
| `HANDLER_LAYER_COMPLETE.md` | Implementation summary |
| `internal/modules/identity/handler/README.md` | API reference |
| `docs/HANDLER_LAYER_GUIDE.md` | Complete guide |

---

## 🔑 Key Files

```
cmd/server/main.go                     # Entry point
internal/platform/config/config.go     # Configuration
internal/infrastructure/database/      # Database setup
internal/modules/identity/             # Identity module
docker-compose.yml                     # Services setup
.env                                   # Environment config
Makefile                              # Common tasks
```

---

## 🎯 Module Status

| Module | Status | Progress | Endpoints |
|--------|--------|----------|-----------|
| Identity | ✅ Complete | 95% | 42 |
| Catalog | 📋 Planned | 0% | ~30-40 |
| Community | 📋 Planned | 0% | ~35-40 |
| Payment | 📋 Planned | 0% | ~30-35 |

---

## 🛡️ Security

### Rate Limits
- Global: 1000 req/min
- Auth: 5 req/15min
- Registration: 3 req/hour
- Password Reset: 3 req/hour
- API: 100 req/min

### Password Requirements
- Minimum 8 characters
- Maximum 72 characters
- Mix of upper, lower, numbers recommended

### Roles (Tenant)
- `owner` - Full control
- `admin` - Manage members and content
- `member` - Create and edit own content
- `viewer` - Read-only access

---

## 🚨 Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `unauthorized` | Missing/invalid token | Login again |
| `rate_limit_exceeded` | Too many requests | Wait and retry |
| `validation_error` | Invalid input | Check request format |
| `not_found` | Resource doesn't exist | Verify ID/slug |
| `forbidden` | Insufficient permissions | Check role |

---

## 💡 Tips

### Development
- Use `make run` for hot reload (with air)
- Check `make help` for all commands
- Use Postman collection for API testing
- Enable debug logs with `DEBUG=true`

### Database
- Always use migrations, never manual SQL
- Test rollback: `make migrate-down && make migrate-up`
- Backup before schema changes

### Testing
- Write tests alongside code
- Use `testify` for assertions
- Mock repositories for service tests
- Use real DB for integration tests

### Code Style
- Follow Go conventions
- Run `gofmt` before commit
- Add comments for exported functions
- Keep functions small (<50 lines)

---

## 🔗 Useful Links

**Internal:**
- Health: http://localhost:8080/health
- Metrics: http://localhost:8080/metrics
- API: http://localhost:8080/api/v1

**Database:**
- pgAdmin: http://localhost:5050
- Direct: postgresql://localhost:5432/system_dev

**Documentation:**
- API Docs: `internal/modules/identity/handler/README.md`
- Architecture: `docs/`

---

## 📞 Getting Help

**Documentation:**
1. Check this Quick Reference first
2. Read `START_HERE.md`
3. Check module README
4. Review `INTEGRATION_CHECKLIST.md`

**Debugging:**
1. Check server logs
2. Check database logs: `docker-compose logs system-db`
3. Test with curl
4. Review error response

**Code Examples:**
- See `internal/modules/identity/` for patterns
- Check handler implementations
- Review service layer
- Study domain models

---

## ✅ Daily Checklist

**Before Starting:**
- [ ] Database running: `docker-compose ps`
- [ ] Dependencies updated: `go mod tidy`
- [ ] Environment configured: `.env` exists

**During Development:**
- [ ] Code formatted: `make fmt`
- [ ] Tests passing: `make test`
- [ ] No lint errors: `make lint`
- [ ] Documentation updated

**Before Committing:**
- [ ] All tests pass
- [ ] Code formatted
- [ ] No compilation errors
- [ ] README updated if needed

---

## 🎓 Learning Path

**New to Project:**
1. Read `README.md` (10 min)
2. Read `START_HERE.md` (5 min)
3. Follow `INTEGRATION_CHECKLIST.md` (60 min)
4. Review Identity module code (30 min)

**Understanding Architecture:**
1. Read Domain layer: `internal/modules/identity/domain/`
2. Read Repository pattern: `internal/modules/identity/repository/`
3. Read Service layer: `internal/modules/identity/service/`
4. Read Handler layer: `internal/modules/identity/handler/`

**Adding Features:**
1. Study existing patterns
2. Follow 4-layer architecture
3. Write tests first (TDD)
4. Update documentation

---

**Keep this file handy for quick reference!**

**Version:** 1.0.0  
**Maintained by:** WibuSystem Team