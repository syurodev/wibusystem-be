// Package http contains HTTP handlers for the Identity module.
package http

import (
	"time"

	"wibusystem/internal/modules/identity/handler/middleware"
	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// RouterConfig holds configuration for the router
type RouterConfig struct {
	AuthService    service.AuthService
	UserService    service.UserService
	TenantService  service.TenantService
	Environment    string // "development" or "production"
	AllowedOrigins []string
}

// SetupRouter sets up all routes for the Identity module
func SetupRouter(app *fiber.App, config RouterConfig) {
	// Create handlers
	authHandler := NewAuthHandler(config.AuthService)
	userHandler := NewUserHandler(config.UserService)
	tenantHandler := NewTenantHandler(config.TenantService)
	sessionHandler := NewSessionHandler(config.AuthService)

	// Global middleware
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(middleware.RecoverMiddleware())

	// CORS middleware
	if config.Environment == "development" {
		app.Use(middleware.DevelopmentCORS())
	} else {
		app.Use(middleware.ProductionCORS(config.AllowedOrigins))
	}

	// Health check endpoint
	app.Get("/health", healthCheck)
	app.Get("/metrics", monitor.New(monitor.Config{
		Title: "WibuSystem Metrics",
	}))

	// API versioning
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Setup routes
	setupAuthRoutes(v1, authHandler, config.AuthService)
	setupUserRoutes(v1, userHandler, config.AuthService)
	setupTenantRoutes(v1, tenantHandler, config.AuthService, config.TenantService)
	setupSessionRoutes(v1, sessionHandler, config.AuthService)

	// 404 handler
	app.Use(middleware.NotFoundHandler())
}

// setupAuthRoutes sets up authentication routes
func setupAuthRoutes(v1 fiber.Router, handler *AuthHandler, authService service.AuthService) {
	auth := v1.Group("/auth")

	// Public routes (no authentication required)
	auth.Post("/register", middleware.RegistrationRateLimiter(), handler.Register)
	auth.Post("/login", middleware.AuthRateLimiter(), handler.Login)
	auth.Post("/verify-email", handler.VerifyEmail)
	auth.Post("/forgot-password", middleware.PasswordResetRateLimiter(), handler.ForgotPassword)
	auth.Post("/reset-password", handler.ResetPassword)

	// Protected routes (authentication required)
	authProtected := auth.Group("", middleware.AuthMiddleware(authService))
	authProtected.Post("/logout", handler.Logout)
	authProtected.Post("/refresh", handler.RefreshSession)
	authProtected.Post("/change-password", handler.ChangePassword)
	authProtected.Post("/resend-verification", handler.ResendVerificationEmail)
	authProtected.Get("/validate", handler.ValidateSession)
	authProtected.Get("/me", handler.GetMe)
}

// setupUserRoutes sets up user management routes
func setupUserRoutes(v1 fiber.Router, handler *UserHandler, authService service.AuthService) {
	users := v1.Group("/users")

	// Protected routes (authentication required)
	usersProtected := users.Group("", middleware.AuthMiddleware(authService))

	// Current user profile
	usersProtected.Get("/profile", handler.GetProfile)
	usersProtected.Put("/profile", handler.UpdateProfile)
	usersProtected.Patch("/profile", handler.UpdateProfile)
	usersProtected.Delete("/account", handler.DeleteAccount)

	// Session management
	usersProtected.Get("/sessions", handler.ListSessions)
	usersProtected.Delete("/sessions/:sessionId", handler.RevokeSession)
	usersProtected.Delete("/sessions", handler.RevokeAllSessions)

	// Admin routes (would require admin middleware in production)
	// For now, just require authentication
	usersProtected.Get("", middleware.APIRateLimiter(), handler.ListUsers)
	usersProtected.Get("/search", handler.SearchUsers)
	usersProtected.Get("/:userId", handler.GetUserByID)
	usersProtected.Get("/:userId/stats", handler.GetUserStats)
}

// setupTenantRoutes sets up tenant management routes
func setupTenantRoutes(v1 fiber.Router, handler *TenantHandler, authService service.AuthService, tenantService service.TenantService) {
	tenants := v1.Group("/tenants")

	// Protected routes (authentication required)
	tenantsProtected := tenants.Group("", middleware.AuthMiddleware(authService))

	// Tenant CRUD
	tenantsProtected.Post("", handler.CreateTenant)
	tenantsProtected.Get("", middleware.APIRateLimiter(), handler.ListTenants)
	tenantsProtected.Get("/my-tenants", handler.GetUserTenants)
	tenantsProtected.Get("/my-owned-tenants", handler.GetUserOwnedTenants)

	// Single tenant operations (require tenant membership)
	tenant := tenantsProtected.Group("/:tenantId", middleware.RequireTenantMembership(tenantService))
	tenant.Get("", handler.GetTenant)
	tenant.Get("/stats", handler.GetTenantStats)

	// Tenant modification (require admin or owner role)
	tenantAdmin := tenantsProtected.Group("/:tenantId", middleware.RequireTenantRole(tenantService, "owner", "admin"))
	tenantAdmin.Put("", handler.UpdateTenant)
	tenantAdmin.Patch("", handler.UpdateTenant)
	tenantAdmin.Get("/members", handler.ListMembers)
	tenantAdmin.Post("/members", handler.AddMember)
	tenantAdmin.Put("/members/:userId", handler.UpdateMember)
	tenantAdmin.Patch("/members/:userId", handler.UpdateMember)
	tenantAdmin.Delete("/members/:userId", handler.RemoveMember)

	// Owner-only operations
	tenantOwner := tenantsProtected.Group("/:tenantId", middleware.RequireTenantRole(tenantService, "owner"))
	tenantOwner.Delete("", handler.DeleteTenant)
	tenantOwner.Post("/transfer-ownership", handler.TransferOwnership)
}

// setupSessionRoutes sets up session management routes
func setupSessionRoutes(v1 fiber.Router, handler *SessionHandler, authService service.AuthService) {
	sessions := v1.Group("/sessions")

	// Protected routes (authentication required)
	sessionsProtected := sessions.Group("", middleware.AuthMiddleware(authService))

	// Session management
	sessionsProtected.Get("", handler.ListUserSessions)
	sessionsProtected.Get("/count", handler.GetActiveSessionsCount)
	sessionsProtected.Get("/:sessionId", handler.GetSession)
	sessionsProtected.Delete("/:sessionId", handler.RevokeSession)
	sessionsProtected.Delete("", handler.RevokeAllSessions)
	sessionsProtected.Delete("/current", handler.RevokeCurrentSession)
}

// healthCheck returns the health status of the service
func healthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "ok",
		"service":   "identity",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

// startTime is used to track service uptime
var startTime = time.Now()

// SetupMinimalRouter sets up a minimal router for testing
func SetupMinimalRouter(app *fiber.App, authService service.AuthService) {
	authHandler := NewAuthHandler(authService)

	// Minimal middleware
	app.Use(requestid.New())
	app.Use(middleware.RecoverMiddleware())
	app.Use(middleware.DevelopmentCORS())

	// Health check
	app.Get("/health", healthCheck)

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1")
	auth := v1.Group("/auth")

	// Basic auth routes
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", middleware.AuthMiddleware(authService), authHandler.Logout)

	// 404 handler
	app.Use(middleware.NotFoundHandler())
}

// RouterOptions provides additional configuration options
type RouterOptions struct {
	// EnableMetrics enables the /metrics endpoint
	EnableMetrics bool
	// EnableMonitor enables the /monitor endpoint
	EnableMonitor bool
	// EnablePprof enables pprof endpoints for profiling
	EnablePprof bool
	// RateLimitEnabled enables global rate limiting
	RateLimitEnabled bool
	// MaxRequestsPerMinute sets the global rate limit
	MaxRequestsPerMinute int
	// CustomMiddleware allows adding custom middleware
	CustomMiddleware []fiber.Handler
}

// SetupRouterWithOptions sets up routes with additional options
func SetupRouterWithOptions(app *fiber.App, config RouterConfig, options RouterOptions) {
	// Apply custom middleware first
	for _, mw := range options.CustomMiddleware {
		app.Use(mw)
	}

	// Global rate limiting
	if options.RateLimitEnabled {
		if options.MaxRequestsPerMinute > 0 {
			app.Use(middleware.CustomRateLimiter(
				options.MaxRequestsPerMinute,
				1*time.Minute,
				nil,
			))
		} else {
			app.Use(middleware.GlobalRateLimiter())
		}
	}

	// Setup standard router
	SetupRouter(app, config)

	// Additional endpoints
	if options.EnableMonitor {
		app.Get("/monitor", monitor.New(monitor.Config{
			Title: "WibuSystem Monitor",
		}))
	}
}

// GetAPIVersion returns the current API version
func GetAPIVersion() string {
	return "v1"
}

// GetServiceInfo returns information about the service
func GetServiceInfo() map[string]any {
	return map[string]any{
		"name":        "Identity Service",
		"version":     "1.0.0",
		"api_version": GetAPIVersion(),
		"description": "User authentication and tenant management service",
	}
}
