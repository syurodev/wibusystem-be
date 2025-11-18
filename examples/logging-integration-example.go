package main

// ============================================================
// LOGGING INTEGRATION EXAMPLE
// ============================================================
// File này show cách integrate logging system vào application.
// Copy các đoạn code cần thiết vào project của bạn.

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"system/configs"
	"system/internal/app/middleware"
	"system/internal/platform/database"
	"system/internal/platform/i18n"
	"system/internal/platform/logger"
)

// ============================================================
// EXAMPLE 1: Main.go Setup
// ============================================================
func mainExample() {
	// 1. Load configuration
	cfg, err := configs.LoadConfig(".env")
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 2. ⚠️ CRITICAL: Initialize logger FIRST
	appLogger, err := logger.InitLogger(cfg.Server.IsProd)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.SyncLogger()

	// Test log - should appear in Grafana
	appLogger.Info("Application starting...",
		zap.String("environment", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// 3. Initialize I18n
	if err := i18n.InitI18n(appLogger); err != nil {
		appLogger.Fatal("Failed to initialize i18n", zap.Error(err))
	}

	// 4. Initialize databases
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(ctx, &cfg.DB, cfg.Log.DBLogQueries, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	rdb, err := database.NewRedisClient(ctx, &cfg.Redis, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer rdb.Close()

	// 5. ⚠️ CRITICAL: Pass logger to router
	router := setupRouter(appLogger, cfg, db, rdb)

	// 6. Start HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		appLogger.Info("HTTP Server started",
			zap.String("port", cfg.Server.Port),
			zap.String("category", "application"),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// 7. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	appLogger.Info("Application ready. Press Ctrl+C to shutdown.")
	<-quit

	appLogger.Info("Shutting down application gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Application shutdown completed.")
}

// ============================================================
// EXAMPLE 2: Router Setup với Logging Middleware
// ============================================================
func setupRouter(appLogger *zap.Logger, cfg *configs.Config, db *database.PostgresDB, rdb *database.RedisClient) *gin.Engine {
	// Disable Gin's default logger (we use structured logging)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// ⚠️ CRITICAL: Add logging middleware
	perfLogger := logger.NewPerformanceLogger(appLogger)
	router.Use(middleware.RecoveryMiddleware(appLogger))
	router.Use(middleware.LoggingMiddleware(appLogger, perfLogger))

	// CORS, I18n, etc.
	// router.Use(cors.New(...))
	// router.Use(i18n.GinI18n(...))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Your routes...
	// setupAuthRoutes(router, appLogger)
	// setupOAuth2Routes(router, appLogger)

	return router
}

// ============================================================
// EXAMPLE 3: Handler với Audit Logging
// ============================================================
type ExampleAuthHandler struct {
	authService interface{} // Replace with your service
	auditLogger *logger.AuditLogger
}

func NewExampleAuthHandler(authService interface{}, auditLogger *logger.AuditLogger) *ExampleAuthHandler {
	return &ExampleAuthHandler{
		authService: authService,
		auditLogger: auditLogger,
	}
}

func (h *ExampleAuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Simulate authentication
	// user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
	var err error = nil // Simulate success

	if err != nil {
		// ⚠️ LOG FAILED LOGIN ATTEMPT
		h.auditLogger.LogLoginAttempt(
			ctx,
			req.Email,
			c.ClientIP(),
			c.UserAgent(),
			false, // failure
			err.Error(),
		)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// ⚠️ LOG SUCCESSFUL LOGIN
	h.auditLogger.LogLoginAttempt(
		ctx,
		req.Email,
		c.ClientIP(),
		c.UserAgent(),
		true, // success
		"",
	)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

// ============================================================
// EXAMPLE 4: Performance Tracking
// ============================================================
type ExampleUserService struct {
	perfLogger *logger.PerformanceLogger
}

func (s *ExampleUserService) CreateUser(ctx context.Context, email string) error {
	// ⚠️ TRACK PERFORMANCE
	timer := s.perfLogger.NewTimer("create_user", logger.OpTypeDatabase)
	defer timer.End(ctx, nil)

	// Simulate database operation
	time.Sleep(50 * time.Millisecond)

	// Business logic here...

	return nil
}

func (s *ExampleUserService) GetUser(ctx context.Context, userID string) error {
	// Manual performance tracking
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		s.perfLogger.LogDatabaseQuery(
			ctx,
			"SELECT",
			"SELECT * FROM users WHERE id = $1",
			duration,
			1, // rows affected
			nil,
		)
	}()

	// Business logic here...

	return nil
}

// ============================================================
// EXAMPLE 5: Context-Aware Logging
// ============================================================
func exampleContextLogging(appLogger *zap.Logger) {
	// In your handler
	ctx := context.Background()

	// Add context
	requestID := "req-123456"
	ctx = logger.WithRequestID(ctx, requestID)

	// Create context-aware logger
	ctxLogger := logger.NewLoggerWithContext(ctx, appLogger)

	// All logs will automatically include request_id
	ctxLogger.Info("Processing user request")
	ctxLogger.Error("Failed to process request", zap.Error(nil))
}

// ============================================================
// EXAMPLE 6: OAuth2 Handler với Audit Logging
// ============================================================
func exampleOAuth2TokenHandler(c *gin.Context, auditLogger *logger.AuditLogger) {
	ctx := c.Request.Context()

	// After successful token issuance
	// userID, _ := uuid.FromString("550e8400-e29b-41d4-a716-446655440000")
	// clientID, _ := uuid.FromString("660e8400-e29b-41d4-a716-446655440000")

	// ⚠️ LOG TOKEN ISSUED
	// auditLogger.LogTokenIssued(
	// 	ctx,
	// 	&userID,
	// 	&clientID,
	// 	"authorization_code",
	// 	"access_token",
	// 	[]string{"openid", "profile", "email"},
	// 	c.ClientIP(),
	// )
}

// ============================================================
// EXAMPLE 7: Custom Audit Event
// ============================================================
func exampleCustomAuditEvent(auditLogger *logger.AuditLogger) {
	ctx := context.Background()

	// Create custom audit event
	event := &logger.AuditEvent{
		EventType: logger.EventAdminAction,
		Status:    logger.StatusSuccess,
		Message:   "Admin deleted user account",
		TargetType: "user",
		TargetID:   "user-123",
		IPAddress:  "192.168.1.100",
		Metadata: map[string]interface{}{
			"reason": "User requested account deletion",
			"admin_id": "admin-456",
		},
	}

	auditLogger.LogEvent(ctx, event)
}

// ============================================================
// INTEGRATION CHECKLIST
// ============================================================
/*
☐ 1. Update main.go:
   - Initialize logger with logger.InitLogger()
   - Pass logger to router
   - Add defer logger.SyncLogger()

☐ 2. Update router:
   - Create perfLogger
   - Add middleware.RecoveryMiddleware()
   - Add middleware.LoggingMiddleware()

☐ 3. Update handlers:
   - Add auditLogger to struct
   - Log authentication events
   - Log OAuth2 events
   - Log security events

☐ 4. Update services:
   - Add perfLogger for tracking
   - Use timers for operations

☐ 5. Test:
   - Run application
   - Make test requests
   - Check docker logs
   - Check Grafana Explore

☐ 6. Production:
   - Switch to production Loki config
   - Setup alert rules
   - Configure retention
*/
