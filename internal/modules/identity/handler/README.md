# Identity Module - Handler Layer

## Overview

The Handler Layer provides HTTP API endpoints for authentication, user management, tenant management, and session management in the Identity module.

## Structure

```
handler/
├── http/
│   ├── auth_handler.go       # Authentication endpoints (register, login, logout, etc.)
│   ├── user_handler.go        # User profile and account management
│   ├── tenant_handler.go      # Multi-tenant operations
│   ├── session_handler.go     # Session management
│   └── router.go              # Route configuration and setup
└── middleware/
    ├── auth_middleware.go     # JWT/session authentication
    ├── error_handler.go       # Standardized error handling
    ├── validation.go          # Request validation
    ├── cors.go               # CORS configuration
    └── rate_limiter.go       # Rate limiting protection
```

## Quick Start

### 1. Install Dependencies

```bash
# Add Fiber web framework
go get github.com/gofiber/fiber/v2

# Add Fiber middleware
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/limiter
go get github.com/gofiber/fiber/v2/middleware/compress
go get github.com/gofiber/fiber/v2/middleware/logger
go get github.com/gofiber/fiber/v2/middleware/requestid
go get github.com/gofiber/fiber/v2/middleware/monitor

# Add storage for rate limiting
go get github.com/gofiber/storage/memory/v2

# Tidy dependencies
go mod tidy
```

### 2. Integration Example

```go
package main

import (
    "wibusystem/internal/modules/identity/handler/http"
    identityService "wibusystem/internal/modules/identity/service"

    "github.com/gofiber/fiber/v2"
)

func main() {
    // Initialize services (from service layer)
    authService := identityService.NewAuthService(userRepo, sessionRepo)
    userService := identityService.NewUserService(userRepo)
    tenantService := identityService.NewTenantService(tenantRepo, memberRepo, userRepo)

    // Create Fiber app
    app := fiber.New()

    // Setup routes
    http.SetupRouter(app, http.RouterConfig{
        AuthService:    authService,
        UserService:    userService,
        TenantService:  tenantService,
        Environment:    "development",
        AllowedOrigins: []string{"http://localhost:3000"},
    })

    // Start server
    app.Listen(":8080")
}
```

## API Endpoints

### Authentication (`/api/v1/auth`)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/register` | Register new user | No |
| POST | `/login` | Login user | No |
| POST | `/logout` | Logout user | Yes |
| POST | `/refresh` | Refresh session token | Yes |
| POST | `/verify-email` | Verify email address | No |
| POST | `/forgot-password` | Request password reset | No |
| POST | `/reset-password` | Reset password | No |
| POST | `/change-password` | Change password | Yes |
| GET | `/validate` | Validate current session | Yes |
| GET | `/me` | Get current user | Yes |

### Users (`/api/v1/users`)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/profile` | Get user profile | Yes |
| PUT | `/profile` | Update user profile | Yes |
| DELETE | `/account` | Delete user account | Yes |
| GET | `` | List all users (admin) | Yes |
| GET | `/search` | Search users (admin) | Yes |
| GET | `/:userId` | Get user by ID (admin) | Yes |
| GET | `/:userId/stats` | Get user statistics (admin) | Yes |
| GET | `/sessions` | List user sessions | Yes |
| DELETE | `/sessions/:sessionId` | Revoke specific session | Yes |
| DELETE | `/sessions` | Revoke all sessions | Yes |

### Tenants (`/api/v1/tenants`)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| POST | `` | Create tenant | Yes | - |
| GET | `` | List tenants | Yes | - |
| GET | `/my-tenants` | Get user's tenants | Yes | - |
| GET | `/my-owned-tenants` | Get owned tenants | Yes | - |
| GET | `/:tenantId` | Get tenant by ID | Yes | Member |
| PUT | `/:tenantId` | Update tenant | Yes | Owner/Admin |
| DELETE | `/:tenantId` | Delete tenant | Yes | Owner |
| GET | `/:tenantId/stats` | Get tenant statistics | Yes | Member |
| GET | `/:tenantId/members` | List tenant members | Yes | Owner/Admin |
| POST | `/:tenantId/members` | Add member | Yes | Owner/Admin |
| PUT | `/:tenantId/members/:userId` | Update member role | Yes | Owner/Admin |
| DELETE | `/:tenantId/members/:userId` | Remove member | Yes | Owner/Admin |
| POST | `/:tenantId/transfer-ownership` | Transfer ownership | Yes | Owner |

### Sessions (`/api/v1/sessions`)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `` | List user sessions | Yes |
| GET | `/count` | Get active sessions count | Yes |
| GET | `/:sessionId` | Get session details | Yes |
| DELETE | `/:sessionId` | Revoke specific session | Yes |
| DELETE | `` | Revoke all sessions | Yes |
| DELETE | `/current` | Revoke current session | Yes |

## Authentication

All protected endpoints require a session token in the `Authorization` header:

```
Authorization: Bearer <session_token>
```

## Request/Response Examples

### Register User

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "display_name": "John Doe"
  }'
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "email_verified": false,
    "display_name": "John Doe",
    "status": "active",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "message": "Registration successful. Please check your email to verify your account."
}
```

### Login User

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "remember_me": true
  }'
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "email_verified": true,
    "display_name": "John Doe",
    "status": "active",
    "last_login_at": "2024-01-15T10:35:00Z"
  },
  "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-02-14T10:35:00Z",
  "message": "Login successful"
}
```

### Create Tenant

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <session_token>" \
  -d '{
    "name": "My Organization",
    "slug": "my-org",
    "description": "Our awesome organization"
  }'
```

**Response:**
```json
{
  "tenant": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "name": "My Organization",
    "slug": "my-org",
    "description": "Our awesome organization",
    "owner_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "created_at": "2024-01-15T10:40:00Z",
    "updated_at": "2024-01-15T10:40:00Z"
  },
  "message": "Tenant created successfully"
}
```

## Error Responses

All errors follow a standardized format:

```json
{
  "error": "validation_error",
  "message": "Request validation failed",
  "errors": [
    {
      "field": "email",
      "message": "email must be a valid email address",
      "tag": "email"
    }
  ]
}
```

## Middleware

### Authentication Middleware

```go
// Require authentication
authProtected := router.Group("", middleware.AuthMiddleware(authService))

// Optional authentication
optionalAuth := router.Group("", middleware.OptionalAuthMiddleware(authService))
```

### Authorization Middleware

```go
// Require tenant membership
router.Use(middleware.RequireTenantMembership(tenantService))

// Require specific role
router.Use(middleware.RequireTenantRole(tenantService, "owner", "admin"))
```

### Rate Limiting

Different rate limits are applied based on endpoint sensitivity:

- **Global**: 1000 requests per minute
- **Authentication**: 5 requests per 15 minutes
- **Registration**: 3 requests per hour
- **Password Reset**: 3 requests per hour
- **API Calls**: 100 requests per minute

## Validation

Request validation is automatic using struct tags:

```go
type RegisterRequest struct {
    Email       string  `json:"email" binding:"required,email"`
    Password    string  `json:"password" binding:"required,min=8,max=72"`
    DisplayName *string `json:"display_name,omitempty" binding:"omitempty,max=255"`
}
```

Custom validators are available:
- `slug` - validates slug format
- `password_strength` - validates password complexity

## CORS Configuration

### Development
```go
// Permissive CORS for local development
app.Use(middleware.DevelopmentCORS())
```

### Production
```go
// Strict CORS with specific origins
app.Use(middleware.ProductionCORS([]string{
    "https://app.example.com",
    "https://admin.example.com",
}))
```

## Testing

Run handler tests:

```bash
# All handler tests
go test ./internal/modules/identity/handler/http/...

# Specific handler
go test ./internal/modules/identity/handler/http -run TestAuthHandler_Register

# With coverage
go test -cover ./internal/modules/identity/handler/http/...
```

## Health Check

Check service health:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "service": "identity",
  "timestamp": "2024-01-15T10:30:00Z",
  "uptime": "2h15m30s"
}
```

## Metrics

Access metrics dashboard (development only):

```
http://localhost:8080/metrics
```

## Documentation

For detailed documentation, see:
- [Handler Layer Guide](../../../docs/HANDLER_LAYER_GUIDE.md)
- [API Documentation](../../../docs/API_DOCUMENTATION.md) (TODO)
- [Service Layer](../service/README.md)
- [Repository Layer](../repository/README.md)

## Status

- ✅ Auth Handler - Complete
- ✅ User Handler - Complete
- ✅ Tenant Handler - Complete
- ✅ Session Handler - Complete
- ✅ Middleware - Complete
- ✅ Router Configuration - Complete
- ⏳ Unit Tests - In Progress
- ⏳ Integration Tests - In Progress
- ⏳ API Documentation (Swagger) - TODO

## Next Steps

1. Add comprehensive unit tests
2. Add integration tests
3. Generate OpenAPI/Swagger documentation
4. Add request/response logging
5. Implement email service integration
6. Add Redis for session storage
7. Add monitoring and observability
8. Performance testing and optimization

## Contributing

When adding new handlers:
1. Follow existing handler patterns
2. Add proper validation
3. Use appropriate middleware
4. Add comprehensive error handling
5. Write unit and integration tests
6. Update API documentation
7. Add rate limiting if needed

## License

Copyright © 2024 WibuSystem. All rights reserved.
