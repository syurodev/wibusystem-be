# Logging Troubleshooting Guide

## Vấn Đề: Không Thấy Logs Trên Grafana

### Bước 1: Chạy Diagnostic Script

```bash
./scripts/debug-logging.sh
```

Script này sẽ kiểm tra:
- ✅ Docker services đang chạy
- ✅ Loki health status
- ✅ Promtail targets
- ✅ Application log format
- ✅ Log ingestion metrics

---

## Checklist Từng Bước

### 1️⃣ Kiểm Tra Services Đang Chạy

```bash
docker ps
```

**Phải thấy các containers**:
- ✅ `loki` - Port 3100
- ✅ `promtail` - Port 9080
- ✅ `grafana` - Port 5555 (3000)
- ✅ Application container

**Nếu thiếu service**:
```bash
docker-compose up -d
```

---

### 2️⃣ Kiểm Tra Loki Đang Hoạt Động

```bash
# Check Loki ready
curl http://localhost:3100/ready

# Expected: "ready"
```

**Nếu không ready**:
```bash
# Xem Loki logs
docker logs loki

# Restart Loki
docker restart loki
```

---

### 3️⃣ Kiểm Tra Promtail Đang Thu Thập Logs

```bash
# Check Promtail targets
curl http://localhost:9080/targets | jq .

# Check Promtail metrics
curl http://localhost:9080/metrics | grep promtail_sent_entries_total
```

**Nếu `promtail_sent_entries_total` = 0**:
- Promtail không đọc được logs từ Docker
- Kiểm tra Docker socket mount trong `docker-compose.yml`

**Fix**:
```yaml
# docker-compose.yml
promtail:
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock  # ✅ Must have this
    - ./promtail-config.yml:/etc/promtail/promtail-config.yml
```

---

### 4️⃣ Kiểm Tra Application Đang Output Logs

```bash
# Xem logs của application
docker logs <your-app-container-name>

# Hoặc nếu chạy local
tail -f /var/log/your-app.log
```

**Expected**: Phải thấy logs dạng JSON
```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "msg": "Application started",
  "category": "application"
}
```

**Nếu không thấy logs hoặc logs không phải JSON**:
→ Application chưa được configure đúng (xem Bước 5)

---

### 5️⃣ **QUAN TRỌNG**: Application Chưa Sử Dụng Logging Middleware

Đây là nguyên nhân phổ biến nhất!

#### Check: Application có đang dùng middleware không?

Xem file `internal/app/router/router.go`:

```go
package router

import (
    "github.com/gin-gonic/gin"
    "system/internal/app/middleware"
    "system/internal/platform/logger"
    "go.uber.org/zap"
)

func NewRouter(
    cfg *configs.Config,
    i18nBundle *i18n.Bundle,
    appLogger *zap.Logger,
    db *database.PostgresDB,
    rdb *database.RedisClient,
) *gin.Engine {
    router := gin.New()

    // ⚠️ PHẢI CÓ 3 DÒNG NÀY
    perfLogger := logger.NewPerformanceLogger(appLogger)
    router.Use(middleware.RecoveryMiddleware(appLogger))
    router.Use(middleware.LoggingMiddleware(appLogger, perfLogger))

    // ... rest of your routes
}
```

#### Check: main.go có khởi tạo logger đúng không?

```go
package main

import (
    "system/configs"
    "system/internal/platform/logger"
)

func main() {
    cfg, _ := configs.LoadConfig(".env")

    // ⚠️ PHẢI CÓ DÒNG NÀY
    appLogger, _ := logger.InitLogger(cfg.Server.IsProd)
    defer logger.SyncLogger()

    appLogger.Info("Application starting...")  // Test log

    // Pass logger to router
    router := router.NewRouter(cfg, i18nBundle, appLogger, db, rdb)
    // ...
}
```

**Nếu chưa có → Xem hướng dẫn integration ở cuối file này**

---

### 6️⃣ Kiểm Tra Container Name Trong Promtail Config

Promtail cần biết container name để đọc logs.

**Check**: `promtail-config.yml`
```yaml
scrape_configs:
  - job_name: docker
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
```

**Check container names**:
```bash
docker ps --format "{{.Names}}"
```

**Promtail tự động phát hiện tất cả containers**, nhưng nếu muốn filter:
```yaml
docker_sd_configs:
  - host: unix:///var/run/docker.sock
    filters:
      - name: label
        values: ["com.docker.compose.project=wibusystem"]
```

---

### 7️⃣ Kiểm Tra Log Format Là JSON

Promtail config hiện tại expects JSON logs.

**Test**:
```bash
docker logs <container-name> --tail 1 | jq .
```

**Nếu lỗi "parse error"** → Logs không phải JSON

**Fix**: Đảm bảo application sử dụng Zap logger với JSON encoder:

```go
// internal/platform/logger/zap.go
func InitLogger(isProd bool) (*zap.Logger, error) {
    var config zap.Config
    if isProd {
        config = zap.NewProductionConfig()  // ✅ JSON encoder
    } else {
        config = zap.NewDevelopmentConfig()
        config.Encoding = "json"  // ✅ Force JSON in dev too
    }
    // ...
}
```

---

### 8️⃣ Test Loki Query Trực Tiếp

```bash
# Query tất cả logs
curl -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={container_name=~".+"}' \
  --data-urlencode 'limit=10' | jq .

# Query specific container
curl -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={container_name="wibusystem-be"}' \
  --data-urlencode 'limit=10' | jq .
```

**Expected**: Phải thấy `data.result` có entries

**Nếu `data.result` rỗng**:
- Loki chưa nhận được logs từ Promtail
- Check Promtail logs: `docker logs promtail`

---

### 9️⃣ Kiểm Tra Grafana Datasource

1. Mở Grafana: http://localhost:5555
2. Login: admin/admin
3. Configuration → Data Sources → Loki
4. Click "Test" button

**Expected**: "Data source is working"

**Nếu failed**:
```yaml
# grafana-datasource.yml
datasources:
  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100  # ✅ Must use container name in Docker network
```

---

### 🔟 Test Query Trong Grafana Explore

1. Grafana → Explore (compass icon)
2. Select "Loki" datasource
3. Query: `{container_name=~".+"}`
4. Run Query

**Expected**: Phải thấy logs

**Common queries**:
```logql
# All logs
{container_name=~".+"}

# Specific container
{container_name="your-app-name"}

# JSON parsed
{container_name="your-app-name"} | json

# Specific level
{container_name="your-app-name"} | json | level="info"
```

---

## Integration Guide (Nếu Chưa Setup)

### Step 1: Update main.go

```go
package main

import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"
    "system/configs"
    "system/internal/app/router"
    "system/internal/platform/database"
    "system/internal/platform/i18n"
    "system/internal/platform/logger"
)

func main() {
    // 1. Load config
    cfg, err := configs.LoadConfig(".env")
    if err != nil {
        panic("Failed to load config: " + err.Error())
    }

    // 2. Initialize logger ← ⚠️ QUAN TRỌNG
    appLogger, err := logger.InitLogger(cfg.Server.IsProd)
    if err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }
    defer logger.SyncLogger()

    appLogger.Info("Application starting...")

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

    // 5. Initialize router với logger ← ⚠️ QUAN TRỌNG
    appRouter := router.NewRouter(cfg, i18n.GetInstance(), appLogger, db, rdb)

    // 6. Start server
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: appRouter,
    }

    go func() {
        appLogger.Info("HTTP server starting", zap.String("port", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatal("Failed to start server", zap.Error(err))
        }
    }()

    // 7. Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    appLogger.Info("Shutting down server...")
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        appLogger.Error("Server forced to shutdown", zap.Error(err))
    }

    appLogger.Info("Server exited")
}
```

### Step 2: Update router/router.go

```go
package router

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "system/configs"
    "system/internal/app/handler/v1/auth"
    "system/internal/app/handler/v1/oauth2"
    "system/internal/app/handler/v1/oauth2_admin"
    "system/internal/app/handler/v1/user"
    "system/internal/app/middleware"
    "system/internal/oauth2/provider"
    "system/internal/oauth2/storage"
    "system/internal/pkg/repository"
    "system/internal/pkg/service"
    "system/internal/platform/database"
    "system/internal/platform/i18n"
    "system/internal/platform/logger"
)

func NewRouter(
    cfg *configs.Config,
    i18nBundle *i18n.Bundle,
    appLogger *zap.Logger,
    db *database.PostgresDB,
    rdb *database.RedisClient,
) *gin.Engine {
    // Disable Gin's default logger (we use our own)
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()

    // ⚠️ MIDDLEWARE - PHẢI CÓ 3 DÒNG NÀY
    perfLogger := logger.NewPerformanceLogger(appLogger)
    router.Use(middleware.RecoveryMiddleware(appLogger))
    router.Use(middleware.LoggingMiddleware(appLogger, perfLogger))

    // CORS
    router.Use(cors.New(cors.Config{
        AllowOrigins:     cfg.CORS.AllowOrigins,
        AllowMethods:     cfg.CORS.AllowMethods,
        AllowHeaders:     cfg.CORS.AllowHeaders,
        ExposeHeaders:    cfg.CORS.ExposeHeaders,
        AllowCredentials: cfg.CORS.AllowCredentials,
        MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
    }))

    // I18n middleware
    router.Use(i18n.GinI18n(i18nBundle))

    // Initialize repositories
    userRepo := repository.NewUserRepository(db)
    oauth2ClientRepo := repository.NewOAuth2ClientRepository(db)
    // ... other repos

    // Initialize services
    authService := service.NewAuthService(userRepo, /* other deps */)
    // ... other services

    // Initialize OAuth2 provider
    sqlStore := storage.NewSQLStore(oauth2ClientRepo, oauth2SessionRepo)
    redisStore := storage.NewRedisStore(rdb)
    hybridStore := storage.NewHybridStore(sqlStore, redisStore)
    oauth2Provider := provider.NewOAuth2Provider(hybridStore, &cfg.OAuth2)

    // Initialize audit logger ← ⚠️ FOR HANDLERS
    auditLogger := logger.NewAuditLogger(appLogger)

    // Initialize handlers (pass audit logger)
    authHandler := auth.NewHandler(authService, auditLogger)
    oauth2Handler := oauth2.NewHandler(&cfg.OAuth2, oauth2Provider, authService, authRequestRepo)
    // ... other handlers

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API routes
    v1 := router.Group("/api/v1")
    {
        authHandler.RegisterRoutes(v1.Group("/auth"))
        userHandler.RegisterRoutes(v1.Group("/users"))
        oauth2AdminHandler.RegisterRoutes(v1.Group("/oauth2_admin"))
    }

    // OAuth2 routes
    oauth2Group := router.Group("/oauth2")
    oauth2Handler.RegisterRoutes(oauth2Group)

    // Well-known routes
    wellKnown := router.Group("/.well-known")
    oauth2Handler.RegisterWellKnownRoutes(wellKnown)

    return router
}
```

### Step 3: Update Handlers để Log Audit Events

Example: `internal/app/handler/v1/auth/handler.go`

```go
package auth

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "system/internal/platform/logger"
    "system/pkg/util/response"
)

type Handler struct {
    authService service.AuthService
    auditLogger *logger.AuditLogger  // ← ⚠️ Add this
}

func NewHandler(authService service.AuthService, auditLogger *logger.AuditLogger) *Handler {
    return &Handler{
        authService: authService,
        auditLogger: auditLogger,  // ← ⚠️ Add this
    }
}

func (h *Handler) Login(c *gin.Context) {
    ctx := c.Request.Context()
    var req LoginRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation.failed", err.Error())
        return
    }

    // Authenticate
    user, err := h.authService.Authenticate(ctx, req.Email, req.Password)

    if err != nil {
        // ← ⚠️ LOG FAILED ATTEMPT
        h.auditLogger.LogLoginAttempt(
            ctx,
            req.Email,
            c.ClientIP(),
            c.UserAgent(),
            false,  // failure
            err.Error(),
        )

        response.Error(c, http.StatusUnauthorized, "AUTH_FAILED", "auth.invalid_credentials", nil)
        return
    }

    // ← ⚠️ LOG SUCCESSFUL ATTEMPT
    h.auditLogger.LogLoginAttempt(
        ctx,
        req.Email,
        c.ClientIP(),
        c.UserAgent(),
        true,  // success
        "",
    )

    // Create session...
    response.Success(c, http.StatusOK, "auth.login_success", gin.H{"user": user}, nil)
}
```

### Step 4: Restart Application

```bash
# If using Docker
docker-compose restart <your-app-service>

# If running locally
go run ./cmd/server/main.go
```

### Step 5: Test

```bash
# Make a request
curl http://localhost:8080/health

# Check logs appeared
docker logs <your-app-container> --tail 10

# Check Grafana
# Go to Explore → Query: {container_name="your-app"}
```

---

## Quick Fix Commands

```bash
# Restart all logging services
docker restart loki promtail grafana

# Clear Loki data (if corrupted)
docker-compose down
docker volume rm wibusystem-backend_loki_data
docker-compose up -d

# View live logs
docker logs -f promtail
docker logs -f loki
docker logs -f <your-app>

# Test Loki ingestion
curl -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={job="docker"}' | jq .
```

---

## Common Issues & Solutions

### Issue: "No data" in Grafana

**Cause**: Time range mismatch
**Solution**:
- Click time picker (top right)
- Select "Last 15 minutes"
- Make a test request to generate logs

### Issue: Logs in console but not in Loki

**Cause**: Promtail cannot parse logs
**Solution**:
- Ensure logs are JSON format
- Check `promtail-config.yml` pipeline_stages

### Issue: "Connection refused" to Loki

**Cause**: Services not on same Docker network
**Solution**:
```yaml
# docker-compose.yml
networks:
  wibusystem_backend:
    driver: bridge

# All services must use:
networks:
  - wibusystem_backend
```

---

## Need Help?

1. Run diagnostic: `./scripts/debug-logging.sh`
2. Check logs: `docker logs promtail`, `docker logs loki`
3. Test queries manually (curl commands above)
4. Review `docs/LOGGING_QUICKSTART.md`

**Still stuck?** Share output của diagnostic script.
