# WibuSystem Backend

Backend API cho WibuSystem - nền tảng đọc light novel.

## Tech Stack

- **Language**: Go 1.22+
- **Framework**: Gin
- **Database**: PostgreSQL + ClickHouse (analytics)
- **Cache**: Redis
- **Auth**: OAuth2 + WebAuthn

---

## Cấu Trúc Thư Mục

```
wibusystem-be/
├── cmd/                    # Entry points
│   └── server/             # Main server application
│
├── configs/                # Configuration files & loaders
│
├── internal/               # Private application code
│   ├── app/                # Application bootstrap
│   │   ├── middleware/     # HTTP middlewares (auth, CORS, rate-limit)
│   │   ├── router/         # Route definitions & dependencies
│   │   └── worker/         # Background workers
│   │
│   ├── domain/             # Domain entities & repository interfaces
│   │   ├── novel.go        # Novel entity
│   │   ├── novel_volume.go # Volume entity
│   │   ├── novel_chapter.go# Chapter entity
│   │   ├── user.go         # User entity
│   │   ├── organization.go # Organization entity
│   │   ├── rbac.go         # Roles & Permissions
│   │   └── ...             # Other entities
│   │
│   ├── dto/                # Data Transfer Objects (API request/response)
│   │   ├── novel/          # Novel DTOs
│   │   ├── novel_volume/   # Volume DTOs
│   │   ├── novel_chapter/  # Chapter DTOs
│   │   ├── genre/          # Genre DTOs
│   │   ├── author/         # Author DTOs
│   │   ├── artist/         # Artist DTOs
│   │   ├── media/          # Media DTOs (shared)
│   │   ├── user/           # User DTOs
│   │   └── auth/           # Auth DTOs
│   │
│   ├── modules/            # Feature modules (handler, service, repository)
│   │   ├── novel/          # Novel CRUD
│   │   ├── novel_volume/   # Volume management
│   │   ├── novel_chapter/  # Chapter management
│   │   ├── genre/          # Genre management
│   │   ├── author/         # Author management
│   │   ├── artist/         # Artist management
│   │   ├── auth/           # Authentication (login, register)
│   │   ├── oauth2/         # OAuth2 server
│   │   ├── user/           # User profile & settings
│   │   ├── creator/        # Creator dashboard
│   │   ├── media/          # Media aggregation
│   │   ├── analytics/      # View tracking & analytics
│   │   └── email/          # Email service
│   │
│   └── platform/           # Infrastructure adapters
│       ├── database/       # PostgreSQL & ClickHouse connections
│       ├── cache/          # Redis cache service
│       ├── oauth2/         # OAuth2 provider (Fosite)
│       ├── logger/         # Structured logging (Zap)
│       ├── i18n/           # Internationalization
│       └── resend/         # Email provider (Resend)
│
├── pkg/                    # Public reusable packages
│   ├── errors/             # Custom error types
│   └── util/               # Utility functions
│       ├── response/       # HTTP response helpers
│       ├── timeutil/       # Time formatting
│       └── ...
│
├── migrations/             # Database migrations (goose)
│
├── scripts/                # SQL scripts for initialization
│   ├── init-identify.sql   # User/Auth schema
│   ├── init-catalog.sql    # Content schema
│   └── init-clickhouse.sql # Analytics schema
│
├── deploy/                 # Deployment configs (Docker, K8s)
│
├── docs/                   # Documentation
│
└── web/                    # Static web assets (if any)
```

---

## Module Structure

Mỗi module trong `internal/modules/` có cấu trúc:

```
modules/{module_name}/
├── handler.go              # HTTP handlers (Gin)
├── routes.go               # Route registration
├── i18nkeys.go             # i18n keys
│
├── service.go              # Domain service (business logic)
├── service_interfaces.go   # Service interface
│
├── usecase_interfaces.go   # UseCase interfaces + Input types
├── usecase_external.go     # External service interfaces (UC prefix)
├── usecase_create.go       # CreateXxxUseCase
├── usecase_update.go       # UpdateXxxUseCase (future)
├── usecase_delete.go       # DeleteXxxUseCase (future)
│
├── repository.go           # Database queries
├── locale/                 # i18n locale files
└── queries/                # Embedded SQL files
```

### Architecture Layers

```
┌──────────────────────────────────────────────────────────────┐
│                      HTTP Handler                            │
│  - Gin binding, validation                                   │
│  - Parse DTO → UseCase Input                                 │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                        UseCase                               │
│  - Transaction management (owns RunInTx)                     │
│  - Orchestrates multiple services                            │
│  - Cross-domain operations                                   │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                     Domain Service                           │
│  - Pure business logic                                       │
│  - Only interacts with its own repository                    │
│  - Transaction-agnostic                                      │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                       Repository                             │
│  - Database access (uses db.GetDB for tx propagation)        │
│  - Embedded SQL queries                                      │
└──────────────────────────────────────────────────────────────┘
```

### File Naming Convention

| Prefix     | Purpose                         |
| ---------- | ------------------------------- |
| `usecase_` | Use case files grouped together |
| `service_` | Service-related files           |
| `i18n`     | Internationalization keys       |

---

## Domain Layer

`internal/domain/` chứa:

- **Entities**: Structs map với database tables
- **Repository Interfaces**: Định nghĩa data access contracts
- **Enums/Constants**: Status types, roles, etc.

---

## DTO Layer

`internal/dto/` chứa các Data Transfer Objects:

- **Response structs**: JSON response cho API
- **Request structs**: (đang ở modules, sẽ di chuyển)

---

## Getting Started

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Start server
make run
```

---

## Environment Variables

Xem `.env.production.example` để biết các biến môi trường cần thiết.
