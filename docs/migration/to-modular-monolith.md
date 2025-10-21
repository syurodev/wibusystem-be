# Migration Plan: Microservices to Modular Monolith

## 📋 Tổng quan

**Mục tiêu:** Chuyển đổi kiến trúc từ Microservices sang Modular Monolith để đơn giản hóa development, deployment và maintenance trong khi vẫn giữ được tính module hóa tốt.

**Timeline dự kiến:** 2-3 tuần  
**Ngày bắt đầu:** TBD  
**Status:** Planning

---

## 🎯 Lý do Chuyển đổi

### Vấn đề hiện tại với Microservices:

1. **Complexity không cần thiết:**
   - Chỉ có 2 services đang chạy (identify, catalog)
   - Overhead của distributed systems (network latency, error handling)
   - gRPC inter-service communication phức tạp

2. **Development Experience:**
   - Debugging khó khăn qua nhiều services
   - Testing cần mock nhiều services
   - Distributed transactions phức tạp

3. **Operational Overhead:**
   - Deploy và monitor nhiều services
   - Multiple databases cần sync
   - Infrastructure cost cao

### Lợi ích của Modular Monolith:

1. **Simplified Development:**
   - Single codebase, dễ navigate
   - Function calls thay vì network calls (nhanh hơn 10-100x)
   - Debugging đơn giản trong 1 process
   - Refactoring dễ dàng

2. **Operational Simplicity:**
   - Single deployment artifact
   - Đơn giản monitoring & logging
   - Chi phí infrastructure thấp hơn

3. **Maintainability:**
   - Vẫn giữ module boundaries rõ ràng
   - Dễ onboard developers mới
   - Có thể tách ra microservices sau nếu cần scale

---

## 🏗️ Kiến trúc Mục tiêu: Modular Monolith

### Cấu trúc Tổng thể:

```
wibusystem-be/
├── cmd/
│   └── server/
│       └── main.go                    # Single entry point
│
├── internal/
│   ├── modules/                       # Business modules (bounded contexts)
│   │   ├── identity/                  # Identity & Auth module
│   │   │   ├── domain/               # Domain models, entities, value objects
│   │   │   │   ├── user.go
│   │   │   │   ├── tenant.go
│   │   │   │   └── session.go
│   │   │   ├── repository/           # Data access layer
│   │   │   │   ├── user_repository.go
│   │   │   │   └── tenant_repository.go
│   │   │   ├── service/              # Business logic
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── user_service.go
│   │   │   │   └── tenant_service.go
│   │   │   ├── handler/              # HTTP/gRPC handlers
│   │   │   │   ├── http/
│   │   │   │   └── grpc/
│   │   │   ├── dto/                  # DTOs for this module
│   │   │   └── module.go             # Module registration
│   │   │
│   │   ├── catalog/                   # Catalog module
│   │   │   ├── domain/
│   │   │   │   ├── anime.go
│   │   │   │   ├── manga.go
│   │   │   │   ├── novel.go
│   │   │   │   └── shared.go         # character, creator, genre
│   │   │   ├── repository/
│   │   │   ├── service/
│   │   │   ├── handler/
│   │   │   └── module.go
│   │   │
│   │   ├── community/                 # Community module (future)
│   │   │   └── module.go
│   │   │
│   │   └── payment/                   # Payment module (future)
│   │       └── module.go
│   │
│   ├── shared/                        # Shared utilities (không có business logic)
│   │   ├── auth/                     # Auth helpers
│   │   ├── crypto/                   # Encryption, hashing
│   │   ├── email/                    # Email sender
│   │   ├── errors/                   # Error types
│   │   ├── logger/                   # Structured logging
│   │   ├── validator/                # Input validation
│   │   └── utils/                    # Common utilities
│   │
│   ├── infrastructure/                # Infrastructure layer
│   │   ├── database/                 # Database factory, connections
│   │   │   ├── postgres/
│   │   │   ├── redis/
│   │   │   └── factory.go
│   │   ├── cache/                    # Caching layer
│   │   ├── messaging/                # Event bus (in-memory hoặc external)
│   │   └── storage/                  # File storage (S3, local)
│   │
│   └── platform/                      # Platform services
│       ├── config/                   # Configuration management
│       ├── middleware/               # HTTP middleware
│       ├── router/                   # HTTP router setup
│       ├── grpc/                     # gRPC server setup
│       └── server/                   # HTTP server
│
├── migrations/                        # Database migrations (all schemas)
│   ├── 001_create_identity_schema.up.sql
│   ├── 001_create_identity_schema.down.sql
│   ├── 002_create_catalog_schema.up.sql
│   ├── 002_create_catalog_schema.down.sql
│   └── ...
│
├── api/                              # API definitions
│   ├── openapi/                      # OpenAPI specs
│   └── proto/                        # Protocol buffers (nếu cần gRPC external)
│
├── scripts/                          # Build & deployment scripts
│   ├── build.sh
│   ├── migrate.sh
│   └── seed.sh
│
├── deployments/                      # Deployment configs
│   ├── docker/
│   │   └── Dockerfile
│   └── k8s/                         # Nếu cần deploy lên k8s
│
├── tests/                           # Integration tests
│   ├── integration/
│   └── e2e/
│
├── go.mod                           # Single go.mod
├── go.sum
├── .env.example
├── docker-compose.yml               # Chỉ databases & redis
└── README.md
```

### Nguyên tắc Thiết kế:

#### 1. **Module Independence (Bounded Context)**
- Mỗi module đại diện cho một domain/subdomain riêng biệt
- Modules KHÔNG import trực tiếp từ nhau
- Communication giữa modules qua:
  - **Interfaces/Contracts** (dependency injection)
  - **Events** (event-driven architecture)

#### 2. **Dependency Direction**
```
┌─────────────────────────────────────────┐
│            HTTP/gRPC Layer              │ ← Presentation
├─────────────────────────────────────────┤
│      Modules (Business Logic)           │ ← Domain
│  identity │ catalog │ community │ ...   │
├─────────────────────────────────────────┤
│        Infrastructure Layer             │ ← Infrastructure
│   database │ cache │ messaging │ ...    │
└─────────────────────────────────────────┘

Dependencies flow: Presentation → Domain → Infrastructure
```

#### 3. **Single Database với Multiple Schemas**
```sql
-- PostgreSQL với schemas riêng cho mỗi module
CREATE SCHEMA identity;
CREATE SCHEMA catalog;
CREATE SCHEMA community;
CREATE SCHEMA payment;

-- Isolation ở schema level, không phải database level
-- Vẫn có thể JOIN cross-schema nếu cần (nhưng nên tránh)
```

#### 4. **Event-Driven Communication**
```go
// Module A publish event
eventBus.Publish(UserRegisteredEvent{
    UserID: user.ID,
    Email: user.Email,
})

// Module B subscribe
eventBus.Subscribe("user.registered", func(event Event) {
    // Handle event
})
```

---

## 📅 Migration Roadmap

### **Phase 1: Setup & Preparation (3-4 ngày)**

#### Day 1: Project Structure Setup

**Tasks:**
1. ✅ Tạo cấu trúc folder mới theo modular monolith
2. ✅ Setup single `go.mod` ở root
3. ✅ Tạo `cmd/server/main.go` entry point
4. ✅ Setup `internal/platform` với config, router, server base

**Deliverables:**
- [ ] Folder structure hoàn chỉnh
- [ ] Build script chạy được (empty app)
- [ ] Docker setup mới (single service)

#### Day 2: Shared Infrastructure

**Tasks:**
1. ✅ Di chuyển `pkg/database` → `internal/infrastructure/database`
2. ✅ Di chuyển `pkg/common` → `internal/shared`
3. ✅ Setup event bus (in-memory) trong `internal/infrastructure/messaging`
4. ✅ Setup logging, config management

**Deliverables:**
- [ ] Infrastructure layer hoàn chỉnh
- [ ] Config từ env variables
- [ ] Database connection working

#### Day 3-4: Database Migration Strategy

**Tasks:**
1. ✅ Consolidate migrations từ tất cả services
2. ✅ Create database schemas (identity, catalog, community, payment)
3. ✅ Test migration scripts
4. ✅ Seed data scripts

**Deliverables:**
- [ ] Single migrations folder với tất cả schemas
- [ ] Migration tool working
- [ ] Seed scripts for dev environment

---

### **Phase 2: Identity Module Migration (4-5 ngày)**

#### Day 5-6: Identity Domain Layer

**Tasks:**
1. ✅ Di chuyển domain models từ `services/identify`
2. ✅ Refactor repositories để dùng new DB factory
3. ✅ Implement module interfaces (để catalog có thể dùng)

**Deliverables:**
- [ ] `internal/modules/identity/domain` complete
- [ ] `internal/modules/identity/repository` complete
- [ ] Interfaces exported cho other modules

#### Day 7-8: Identity Service & Handlers

**Tasks:**
1. ✅ Di chuyển business logic vào services
2. ✅ Convert HTTP handlers
3. ✅ Convert gRPC handlers (if needed for external clients)
4. ✅ OAuth2/OIDC provider integration

**Deliverables:**
- [ ] Auth endpoints working
- [ ] User management working
- [ ] Tenant management working
- [ ] Tests passing

#### Day 9: Identity Integration Testing

**Tasks:**
1. ✅ Integration tests cho auth flows
2. ✅ Test OAuth2 flows
3. ✅ Test multi-tenant features

**Deliverables:**
- [ ] All identity features tested
- [ ] Performance baseline established

---

### **Phase 3: Catalog Module Migration (5-6 ngày)**

#### Day 10-11: Catalog Domain Layer

**Tasks:**
1. ✅ Di chuyển domain models (anime, manga, novel, shared)
2. ✅ Refactor repositories
3. ✅ Handle complex relationships (characters, creators, genres)

**Deliverables:**
- [ ] Domain models complete
- [ ] Repositories complete
- [ ] Multi-language support working

#### Day 12-13: Catalog Service & Handlers

**Tasks:**
1. ✅ Business logic services
2. ✅ HTTP handlers cho CRUD operations
3. ✅ Content ownership logic
4. ✅ Access control integration với Identity module

**Deliverables:**
- [ ] CRUD endpoints working
- [ ] Content ownership working
- [ ] Permission checks working

#### Day 14-15: Catalog Advanced Features

**Tasks:**
1. ✅ Monetization logic (purchases, rentals, subscriptions)
2. ✅ Content relationships
3. ✅ Search & filtering

**Deliverables:**
- [ ] All catalog features working
- [ ] Integration tests passing

---

### **Phase 4: Module Integration (3-4 ngày)**

#### Day 16-17: Cross-Module Communication

**Tasks:**
1. ✅ Implement event bus for inter-module events
2. ✅ Identity module events (user.created, user.updated, etc.)
3. ✅ Catalog module events (content.published, purchase.completed, etc.)
4. ✅ Event handlers in subscribing modules

**Example Events:**
```go
// Identity module publishes
UserRegisteredEvent
UserEmailVerifiedEvent
TenantCreatedEvent

// Catalog module subscribes to identity events
// Catalog module publishes
ContentPublishedEvent
ContentPurchasedEvent
ContentAccessGrantedEvent
```

**Deliverables:**
- [ ] Event bus working
- [ ] Key events published/subscribed
- [ ] Async operations working

#### Day 18: Access Control Integration

**Tasks:**
1. ✅ Catalog uses Identity service for auth checks
2. ✅ Tenant-scoped content access
3. ✅ Content ownership verification

**Deliverables:**
- [ ] Catalog respects identity permissions
- [ ] Multi-tenant isolation working

---

### **Phase 5: Testing & Optimization (3-4 ngày)**

#### Day 19-20: Integration Testing

**Tasks:**
1. ✅ End-to-end test scenarios
2. ✅ Load testing
3. ✅ Memory profiling
4. ✅ Query optimization

**Test Scenarios:**
- User registration → content creation → purchase flow
- Multi-tenant isolation
- Content access with various permission levels
- Concurrent operations

**Deliverables:**
- [ ] Integration test suite
- [ ] Performance benchmarks
- [ ] No major bottlenecks

#### Day 21: Documentation

**Tasks:**
1. ✅ Update README with new architecture
2. ✅ API documentation (OpenAPI)
3. ✅ Developer guide
4. ✅ Deployment guide

**Deliverables:**
- [ ] Comprehensive documentation
- [ ] Setup guide for new developers

---

### **Phase 6: Deployment & Cutover (2-3 ngày)**

#### Day 22: Deployment Preparation

**Tasks:**
1. ✅ Build Docker image
2. ✅ Update docker-compose for monolith
3. ✅ CI/CD pipeline
4. ✅ Rollback plan

**Deliverables:**
- [ ] Production-ready Docker image
- [ ] Deployment scripts
- [ ] Monitoring setup

#### Day 23: Production Deployment

**Tasks:**
1. ✅ Deploy to staging
2. ✅ Smoke tests
3. ✅ Deploy to production
4. ✅ Monitor for issues

**Deliverables:**
- [ ] Monolith running in production
- [ ] Old services deprecated

---

## 🔧 Technical Details

### Database Migration Strategy

**Option 1: Keep Multiple Databases → Single Database with Schemas**
```sql
-- Step 1: Create schemas in single database
CREATE SCHEMA identity;
CREATE SCHEMA catalog;
CREATE SCHEMA community;
CREATE SCHEMA payment;

-- Step 2: Migrate tables with schema prefix
ALTER TABLE users SET SCHEMA identity;
ALTER TABLE anime SET SCHEMA catalog;

-- Step 3: Update application code to use schema-qualified tables
-- FROM: SELECT * FROM users
-- TO:   SELECT * FROM identity.users
```

**Option 2: Keep Separate Databases (if needed)**
- Monolith có thể vẫn connect đến multiple databases
- Sử dụng transaction coordinator nếu cần distributed transactions
- Migrate về single DB sau

### Module Communication Patterns

#### Pattern 1: Direct Interface Call (Synchronous)
```go
// identity module exposes interface
type UserService interface {
    GetUserByID(ctx context.Context, userID string) (*User, error)
    ValidateToken(ctx context.Context, token string) (*Claims, error)
}

// catalog module uses it via dependency injection
type CatalogHandler struct {
    userService identity.UserService
}

func (h *CatalogHandler) CreateContent(c *gin.Context) {
    user, err := h.userService.GetUserByID(ctx, userID)
    // ...
}
```

#### Pattern 2: Event-Driven (Asynchronous)
```go
// identity module publishes event
type UserRegisteredEvent struct {
    UserID    string
    Email     string
    Timestamp time.Time
}

eventBus.Publish("user.registered", event)

// catalog module subscribes
eventBus.Subscribe("user.registered", func(event Event) {
    // Initialize user's content library
})
```

#### Pattern 3: Shared Data Access (AVOID if possible)
```go
// Only for read-only, rarely-changing data
// Example: catalog reads identity.users for display names
// But prefer calling identity service instead
```

### Configuration Management

**Single config with sections:**
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Identity IdentityModuleConfig
    Catalog  CatalogModuleConfig
    Redis    RedisConfig
    // ...
}
```

**Environment variables:**
```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=wibusystem
DB_USER=wibusystem
DB_PASSWORD=secret

# Identity Module
IDENTITY_JWT_SECRET=xxx
IDENTITY_OAUTH2_ISSUER=http://localhost:8080

# Catalog Module
CATALOG_CDN_URL=https://cdn.example.com
```

---

## ✅ Migration Checklist

### Pre-Migration
- [ ] Backup current databases
- [ ] Document current API contracts
- [ ] List all inter-service dependencies
- [ ] Plan downtime window (if needed)

### During Migration
- [ ] Create new project structure
- [ ] Migrate identity module
- [ ] Migrate catalog module
- [ ] Setup event bus
- [ ] Migrate shared code
- [ ] Update tests
- [ ] Performance testing
- [ ] Documentation

### Post-Migration
- [ ] Deploy to staging
- [ ] Run integration tests
- [ ] Performance comparison (vs microservices)
- [ ] Deploy to production
- [ ] Monitor logs & metrics
- [ ] Cleanup old services
- [ ] Update team documentation

---

## 📊 Success Metrics

### Performance
- [ ] API response time < 100ms (p95)
- [ ] No memory leaks under load
- [ ] Database query time < 50ms (p95)

### Development
- [ ] Build time < 1 minute
- [ ] Test suite runs < 5 minutes
- [ ] Developer can run full app locally

### Operations
- [ ] Single deployment artifact
- [ ] Deployment time < 5 minutes
- [ ] Zero-downtime deployments
- [ ] Rollback < 1 minute

---

## 🚨 Risks & Mitigation

### Risk 1: Loss of Service Isolation
**Mitigation:**
- Strong module boundaries
- No circular dependencies
- Code review process
- Architecture decision records

### Risk 2: Database Bottleneck
**Mitigation:**
- Connection pooling
- Read replicas if needed
- Caching layer (Redis)
- Query optimization

### Risk 3: Testing Complexity
**Mitigation:**
- Comprehensive integration tests
- Test fixtures & factories
- Database seeding for tests
- Mocking infrastructure

### Risk 4: Breaking Changes During Migration
**Mitigation:**
- API compatibility layer
- Feature flags
- Gradual rollout
- Rollback plan

---

## 🔄 Future: Path to Microservices (If Needed)

Modular Monolith làm nền tảng tốt để tách ra microservices sau:

### When to Extract a Service:
1. **High load on specific module** → Extract to scale independently
2. **Different technology needs** → Extract to use different stack
3. **Team scaling** → Extract to enable team autonomy

### Easy Migration Path:
```
Modular Monolith
    ↓
Extract high-traffic module (e.g., Content Delivery)
    ↓
Monolith + 1 Microservice
    ↓
Continue extracting as needed
```

**Key advantage:** Module boundaries đã rõ ràng, chỉ cần:
1. Add HTTP/gRPC API layer cho module
2. Deploy module riêng
3. Update monolith to call external API instead of local interface

---

## 📚 References

- [Modular Monolith: A Primer](https://www.kamilgrzybek.com/design/modular-monolith-primer/)
- [Monolith to Microservices (Sam Newman)](https://samnewman.io/books/monolith-to-microservices/)
- [The Majestic Monolith (DHH)](https://m.signalvnoise.com/the-majestic-monolith/)
- [Shopify's Modular Monolith](https://shopify.engineering/deconstructing-monolith-designing-software-maximizes-developer-productivity)

---

## 👥 Team & Roles

**Migration Lead:** TBD  
**Backend Engineers:** TBD  
**DevOps:** TBD  
**QA:** TBD

**Communication:**
- Daily standups during migration
- Weekly progress reports
- Dedicated Slack channel
- Document decisions in ADRs

---

## 📝 Notes

- Tài liệu này sẽ được update liên tục trong quá trình migration
- Mọi architecture decision quan trọng cần ghi vào ADR (Architecture Decision Record)
- Keep old microservices code in separate branch for reference