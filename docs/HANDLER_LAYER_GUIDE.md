# Handler Layer Implementation Guide

## Overview

This document provides a comprehensive guide for the Handler Layer implementation in the Identity module. The Handler Layer is responsible for handling HTTP requests, validation, and returning appropriate responses.

## Architecture

The Handler Layer follows these principles:

- **Separation of Concerns**: Handlers only deal with HTTP-specific logic
- **Middleware Pattern**: Cross-cutting concerns (auth, validation, CORS) are handled via middleware
- **Error Handling**: Standardized error responses across all endpoints
- **Validation**: Request validation using struct tags and custom validators
- **Rate Limiting**: Protection against abuse and DDoS attacks

## Directory Structure

```
internal/modules/identity/handler/
├── http/
│   ├── auth_handler.go       # Authentication endpoints
│   ├── user_handler.go        # User management endpoints
│   ├── tenant_handler.go      # Tenant management endpoints
│   ├── session_handler.go     # Session management endpoints
│   └── router.go              # Route configuration
└── middleware/
    ├── auth_middleware.go     # Authentication & authorization
    ├── error_handler.go       # Error handling
    ├── validation.go          # Request validation
    ├── cors.go               # CORS configuration
    └── rate_limiter.go       # Rate limiting
```

## Dependencies

### Required Go Modules

Add the following dependencies to your project:

```bash
# Fiber web framework (high-performance alternative to Gin)
go get github.com/gofiber/fiber/v2

# Fiber middleware
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/limiter
go get github.com/gofiber/fiber/v2/middleware/compress
go get github.com/gofiber/fiber/v2/middleware/logger
go get github.com/gofiber/fiber/v2/middleware/requestid
go get github.com/gofiber/fiber/v2/middleware/monitor

# Fiber storage for rate limiting
go get github.com/gofiber/storage/memory/v2

# Validator (already in project)
# github.com/go-playground/validator/v10

# UUID (already in project)
# github.com/google/uuid
```

Run:
```bash
cd wibusystem-be
go mod tidy
```

### Alternative: Using Existing Gin Framework

If you prefer to keep Gin instead of Fiber, you can convert the handlers. See "Converting to Gin" section below.

## Components

### 1. Handlers

#### Auth Handler
Handles authentication-related operations:
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh session token
- `POST /api/v1/auth/verify-email` - Verify email address
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password
- `POST /api/v1/auth/change-password` - Change password (authenticated)
- `GET /api/v1/auth/validate` - Validate current session
- `GET /api/v1/auth/me` - Get current user info

#### User Handler
Handles user profile and account operations:
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `DELETE /api/v1/users/account` - Delete user account
- `GET /api/v1/users` - List users (admin)
- `GET /api/v1/users/search` - Search users (admin)
- `GET /api/v1/users/:userId` - Get user by ID (admin)
- `GET /api/v1/users/:userId/stats` - Get user statistics (admin)
- `GET /api/v1/users/sessions` - List user sessions
- `DELETE /api/v1/users/sessions/:sessionId` - Revoke specific session
- `DELETE /api/v1/users/sessions` - Revoke all sessions

#### Tenant Handler
Handles tenant and multi-tenancy operations:
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List tenants
- `GET /api/v1/tenants/my-tenants` - Get user's tenants
- `GET /api/v1/tenants/my-owned-tenants` - Get owned tenants
- `GET /api/v1/tenants/:tenantId` - Get tenant by ID
- `PUT /api/v1/tenants/:tenantId` - Update tenant
- `DELETE /api/v1/tenants/:tenantId` - Delete tenant
- `GET /api/v1/tenants/:tenantId/stats` - Get tenant statistics
- `POST /api/v1/tenants/:tenantId/members` - Add member
- `GET /api/v1/tenants/:tenantId/members` - List members
- `PUT /api/v1/tenants/:tenantId/members/:userId` - Update member role
- `DELETE /api/v1/tenants/:tenantId/members/:userId` - Remove member
- `POST /api/v1/tenants/:tenantId/transfer-ownership` - Transfer ownership

#### Session Handler
Handles session management:
- `GET /api/v1/sessions` - List user sessions
- `GET /api/v1/sessions/count` - Get active sessions count
- `GET /api/v1/sessions/:sessionId` - Get session details
- `DELETE /api/v1/sessions/:sessionId` - Revoke specific session
- `DELETE /api/v1/sessions` - Revoke all sessions except current
- `DELETE /api/v1/sessions/current` - Revoke current session (logout)

### 2. Middleware

#### Authentication Middleware
- **AuthMiddleware**: Validates session tokens and populates user context
- **OptionalAuthMiddleware**: Like AuthMiddleware but doesn't fail if no token
- **RequireTenantMembership**: Ensures user is a member of the tenant
- **RequireTenantRole**: Ensures user has specific role in tenant

#### Error Handler Middleware
- **ErrorHandler**: Catches and formats all errors
- **RecoverMiddleware**: Recovers from panics
- **NotFoundHandler**: Handles 404 errors
- **MethodNotAllowedHandler**: Handles 405 errors

#### Validation Middleware
- **ValidateRequest**: Validates request body against struct
- **ValidateQuery**: Validates query parameters
- Custom validators: email, slug, password strength

#### CORS Middleware
- **DevelopmentCORS**: Permissive CORS for development
- **ProductionCORS**: Strict CORS for production
- **CustomOriginValidator**: Custom origin validation logic

#### Rate Limiter Middleware
- **GlobalRateLimiter**: Global rate limit (1000 req/min)
- **AuthRateLimiter**: Strict limit for auth endpoints (5 req/15min)
- **RegistrationRateLimiter**: Registration limit (3 req/hour)
- **PasswordResetRateLimiter**: Password reset limit (3 req/hour)
- **APIRateLimiter**: API rate limit (100 req/min)

## Integration Steps

### Step 1: Update Main Entry Point

Update `cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wibusystem/internal/infrastructure/database"
	"wibusystem/internal/modules/identity/handler/http"
	identityRepo "wibusystem/internal/modules/identity/repository/postgres"
	identityService "wibusystem/internal/modules/identity/service"
	"wibusystem/internal/platform/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, cfg.Database.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := identityRepo.NewUserRepository(db)
	tenantRepo := identityRepo.NewTenantRepository(db)
	tenantMemberRepo := identityRepo.NewTenantMemberRepository(db)
	sessionRepo := identityRepo.NewSessionRepository(db)

	// Initialize services
	authService := identityService.NewAuthService(userRepo, sessionRepo)
	userService := identityService.NewUserService(userRepo)
	tenantService := identityService.NewTenantService(
		tenantRepo,
		tenantMemberRepo,
		userRepo,
	)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Setup routes
	http.SetupRouter(app, http.RouterConfig{
		AuthService:    authService,
		UserService:    userService,
		TenantService:  tenantService,
		Environment:    cfg.Environment,
		AllowedOrigins: cfg.CORS.AllowedOrigins,
	})

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

### Step 2: Update Configuration

Ensure `internal/platform/config/config.go` includes:

```go
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	CORS        CORSConfig
	// ... other configs
}

type ServerConfig struct {
	Port string
	Host string
}

type CORSConfig struct {
	AllowedOrigins []string
}
```

### Step 3: Testing the Handler Layer

#### Unit Testing Example

Create `internal/modules/identity/handler/http/auth_handler_test.go`:

```go
package http

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"wibusystem/internal/modules/identity/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req dto.RegisterRequest) (*domain.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// ... implement other methods

func TestAuthHandler_Register(t *testing.T) {
	// Setup
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	app := fiber.New()
	app.Post("/register", handler.Register)

	// Test data
	reqBody := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// Mock service response
	expectedUser := &domain.User{
		ID:    uuid.New(),
		Email: reqBody.Email,
	}
	mockService.On("Register", mock.Anything, reqBody).Return(expectedUser, nil)

	// Execute
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// Assert
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
	mockService.AssertExpectations(t)
}
```

#### Integration Testing

Create `tests/integration/identity_api_test.go`:

```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IdentityAPITestSuite struct {
	suite.Suite
	app *fiber.App
	db  *sql.DB
}

func (suite *IdentityAPITestSuite) SetupSuite() {
	// Setup test database, services, and app
}

func (suite *IdentityAPITestSuite) TearDownSuite() {
	// Cleanup
}

func (suite *IdentityAPITestSuite) TestRegisterAndLogin() {
	// Test complete registration and login flow
}

func TestIdentityAPISuite(t *testing.T) {
	suite.Run(t, new(IdentityAPITestSuite))
}
```

## Error Handling

All errors are standardized using the `AppError` type:

```go
return middleware.NewAppError(
    middleware.ErrBadRequest,
    "Invalid email format",
    fiber.StatusBadRequest,
).WithCode("INVALID_EMAIL").WithDetails(map[string]any{
    "field": "email",
})
```

Standard error codes:
- `REGISTRATION_FAILED`
- `LOGIN_FAILED`
- `INVALID_SESSION`
- `NOT_TENANT_MEMBER`
- `INSUFFICIENT_PERMISSIONS`
- `USER_NOT_FOUND`
- `TENANT_NOT_FOUND`

## Rate Limiting Strategy

| Endpoint Type | Limit | Window |
|--------------|-------|--------|
| Global | 1000 requests | 1 minute |
| Authentication | 5 requests | 15 minutes |
| Registration | 3 requests | 1 hour |
| Password Reset | 3 requests | 1 hour |
| API Calls | 100 requests | 1 minute |

## Security Considerations

1. **Authentication**: Session tokens in `Authorization: Bearer <token>` header
2. **CSRF Protection**: Not needed for stateless API (token-based)
3. **Rate Limiting**: Prevents brute force and DDoS attacks
4. **Input Validation**: All inputs validated before processing
5. **Error Messages**: Don't leak sensitive information
6. **CORS**: Strict origin validation in production
7. **HTTPS Only**: Force HTTPS in production (configure reverse proxy)

## Performance Optimizations

1. **Compression**: Gzip compression enabled for responses
2. **Connection Pooling**: Database connection pooling configured
3. **Rate Limiting Storage**: In-memory storage for development, Redis for production
4. **Caching**: Consider adding Redis cache for session validation
5. **Pagination**: All list endpoints support pagination

## Converting to Gin Framework

If you prefer to use the existing Gin framework instead of Fiber, here are the main changes:

### Replace Fiber with Gin

```go
// Fiber
app := fiber.New()
app.Get("/path", handler)
return c.JSON(data)

// Gin equivalent
router := gin.Default()
router.GET("/path", handler)
c.JSON(http.StatusOK, data)
```

### Context differences

```go
// Fiber
userID := c.Locals("user_id").(uuid.UUID)
params := c.Params("id")
query := c.Query("page")

// Gin equivalent
userID := c.MustGet("user_id").(uuid.UUID)
params := c.Param("id")
query := c.Query("page")
```

## Troubleshooting

### Common Issues

1. **Import path errors**: Ensure module path matches `go.mod`
2. **Missing dependencies**: Run `go mod tidy`
3. **Port already in use**: Change port in config or kill existing process
4. **Database connection fails**: Check PostgreSQL is running and config is correct
5. **CORS errors**: Ensure frontend origin is in `AllowedOrigins`

### Debug Mode

Enable debug logging in development:

```go
app := fiber.New(fiber.Config{
    Debug: true,
})
```

## Next Steps

1. ✅ Handler Layer implemented
2. ⏳ Add Redis for session storage (optional)
3. ⏳ Add email service integration
4. ⏳ Implement remaining service layer methods
5. ⏳ Add comprehensive tests
6. ⏳ Add API documentation (Swagger/OpenAPI)
7. ⏳ Add monitoring and observability
8. ⏳ Deploy to production

## API Documentation

Consider adding Swagger/OpenAPI documentation:

```bash
go get github.com/gofiber/swagger
go get github.com/swaggo/swag/cmd/swag
```

## Conclusion

The Handler Layer is now complete with:
- ✅ 4 HTTP handlers (Auth, User, Tenant, Session)
- ✅ 5 middleware components (Auth, Error, Validation, CORS, Rate Limiting)
- ✅ Comprehensive route configuration
- ✅ Error handling and validation
- ✅ Security features (rate limiting, CORS, auth)
- ✅ Performance optimizations

You can now start the server and test all endpoints using tools like Postman, curl, or your frontend application.
