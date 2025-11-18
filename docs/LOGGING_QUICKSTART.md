# Logging Quick Start Guide

## 🚀 Khởi Động Nhanh

### 1. Start Logging Stack

```bash
# Start tất cả services (PostgreSQL, Redis, Loki, Promtail, Grafana)
make docker-up

# Hoặc
docker-compose up -d
```

### 2. Verify Services

```bash
# Check all services are running
docker ps

# Should see:
# - system_dev (PostgreSQL)
# - redis
# - loki
# - promtail
# - grafana
```

### 3. Access Grafana

1. Mở browser: http://localhost:5555
2. Login:
   - Username: `admin`
   - Password: `admin`
3. Loki datasource đã được tự động cấu hình

### 4. Import Dashboard

**Option 1: Via UI**
1. Grafana → Dashboards → Import
2. Upload file: `grafana-dashboards/oauth2-overview.json`
3. Click "Import"

**Option 2: Via API**
```bash
curl -X POST http://admin:admin@localhost:5555/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @grafana-dashboards/oauth2-overview.json
```

### 5. Start Your Application

```bash
go run ./cmd/server/main.go
```

Logs sẽ tự động được gửi đến Loki qua Promtail.

---

## 🔍 Quick Queries

### View All Logs
```logql
{container_name="wibusystem-be"}
```

### View Audit Logs
```logql
{category="audit"}
```

### View Errors
```logql
{level="error"}
```

### View Specific User Activity
```logql
{category="audit"} | json | user_id="YOUR-USER-ID"
```

---

## 📝 Integration trong Code

### 1. Setup Logger trong main.go

```go
package main

import (
    "system/configs"
    "system/internal/platform/logger"
)

func main() {
    // Load config
    cfg, _ := configs.LoadConfig(".env")

    // Initialize logger
    appLogger, _ := logger.InitLogger(cfg.Server.IsProd)
    defer logger.SyncLogger()

    // Initialize audit và performance loggers
    auditLogger := logger.NewAuditLogger(appLogger)
    perfLogger := logger.NewPerformanceLogger(appLogger)

    // Pass to your router/handlers
    router := setupRouter(appLogger, auditLogger, perfLogger)
    router.Run(":8080")
}
```

### 2. Add Logging Middleware

Update `internal/app/router/router.go`:

```go
package router

import (
    "github.com/gin-gonic/gin"
    "system/internal/app/middleware"
    "system/internal/platform/logger"
)

func NewRouter(
    cfg *configs.Config,
    appLogger *zap.Logger,
    perfLogger *logger.PerformanceLogger,
) *gin.Engine {
    router := gin.New()

    // Add logging middleware
    router.Use(middleware.RecoveryMiddleware(appLogger))
    router.Use(middleware.LoggingMiddleware(appLogger, perfLogger))

    // ... rest of your routes
    return router
}
```

### 3. Log Audit Events trong Handlers

**Example: Login Handler**

```go
package auth

import (
    "system/internal/platform/logger"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    auditLogger *logger.AuditLogger
    authService AuthService
}

func (h *Handler) Login(c *gin.Context) {
    var req LoginRequest
    c.ShouldBindJSON(&req)

    // Get context info
    ctx := c.Request.Context()
    ipAddress := c.ClientIP()
    userAgent := c.UserAgent()

    // Attempt authentication
    user, err := h.authService.Authenticate(ctx, req.Email, req.Password)

    if err != nil {
        // Log failed attempt
        h.auditLogger.LogLoginAttempt(
            ctx,
            req.Email,
            ipAddress,
            userAgent,
            false,  // failure
            err.Error(),
        )

        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }

    // Log successful attempt
    h.auditLogger.LogLoginAttempt(
        ctx,
        req.Email,
        ipAddress,
        userAgent,
        true,  // success
        "",
    )

    c.JSON(200, gin.H{"user": user})
}
```

**Example: OAuth2 Token Handler**

```go
func (h *Handler) Token(c *gin.Context) {
    ctx := c.Request.Context()

    // ... OAuth2 token logic with Fosite ...

    // After successful token issuance
    h.auditLogger.LogTokenIssued(
        ctx,
        &userID,
        &clientID,
        grantType,
        "access_token",
        scopes,
        c.ClientIP(),
    )
}
```

### 4. Performance Tracking

**Option 1: Manual**
```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    startTime := time.Now()
    defer func() {
        duration := time.Since(startTime)
        s.perfLogger.LogDatabaseQuery(
            ctx,
            "INSERT",
            "INSERT INTO users ...",
            duration,
            1, // rows affected
            err,
        )
    }()

    // ... create user logic ...
}
```

**Option 2: Timer (Recommended)**
```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    timer := s.perfLogger.NewTimer("create_user", logger.OpTypeDatabase)
    defer timer.End(ctx, nil)

    // ... create user logic ...

    return user, nil
}
```

---

## 🎯 Common Use Cases

### Use Case 1: Tracking Failed Logins

```go
// In your auth handler
if authFailed {
    h.auditLogger.LogEvent(ctx, &logger.AuditEvent{
        EventType: logger.EventLoginFailure,
        Status:    logger.StatusFailure,
        Username:  email,
        IPAddress: c.ClientIP(),
        UserAgent: c.UserAgent(),
        ErrorDetail: "Invalid password",
    })
}
```

**Query in Grafana:**
```logql
{event_type="auth.login.failure"} | json
```

### Use Case 2: Monitoring Slow Endpoints

```go
// Middleware automatically logs all requests
// Query slow ones:
```

**Query in Grafana:**
```logql
{category="performance", operation_type="http"}
| json
| duration_ms > 2000
```

### Use Case 3: Investigating Security Incident

```logql
# Find all activity from suspicious IP
{ip_address="192.168.1.100"} | json

# Find all failed operations for a user
{category="audit", user_id="USER-UUID", status="failure"} | json

# Timeline of events for a request
{request_id="REQUEST-ID"} | json
```

### Use Case 4: OAuth2 Client Monitoring

```go
// When creating client
h.auditLogger.LogClientCreated(
    ctx,
    &adminUserID,
    clientID,
    "My OAuth2 Client",
    c.ClientIP(),
)
```

**Query:**
```logql
{event_type=~"oauth2.client.*"} | json
```

---

## 📊 Monitoring Checklist

### Daily
- [ ] Check error rate dashboard
- [ ] Review failed authentication attempts
- [ ] Monitor slow query alerts

### Weekly
- [ ] Review security alerts
- [ ] Check disk usage for Loki
- [ ] Analyze performance trends

### Monthly
- [ ] Review retention policies
- [ ] Audit user access patterns
- [ ] Update alert thresholds if needed

---

## 🐛 Troubleshooting

### Problem: Logs not showing in Grafana

**Solution 1: Check application logs are JSON**
```bash
docker logs wibusystem-be | head -1 | jq .
```

Expected output:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "msg": "Application started"
}
```

**Solution 2: Check Promtail is collecting logs**
```bash
# Check Promtail logs
docker logs promtail

# Check targets in Promtail
curl http://localhost:9080/targets
```

**Solution 3: Check Loki is receiving data**
```bash
# Check Loki health
curl http://localhost:3100/ready

# Check ingester stats
curl http://localhost:3100/metrics | grep loki_ingester
```

### Problem: Alert not firing

**Solution 1: Verify rule syntax**
```bash
# Check Loki ruler
curl http://localhost:3100/loki/api/v1/rules
```

**Solution 2: Test query in Grafana Explore**
```logql
sum(rate({category="audit", event_type="auth.login.failure"}[5m]))
```

---

## 🔗 Next Steps

1. **Read full documentation**: [LOGGING.md](./LOGGING.md)
2. **Customize dashboards**: Add panels for your specific needs
3. **Set up alerting**: Configure Alertmanager integration
4. **Add more audit events**: Cover all security-sensitive operations
5. **Performance optimization**: Tune Loki/Promtail for your load

---

## 📚 Resources

- **Loki Docs**: https://grafana.com/docs/loki/latest/
- **LogQL**: https://grafana.com/docs/loki/latest/logql/
- **Grafana Dashboards**: https://grafana.com/grafana/dashboards/
- **Promtail**: https://grafana.com/docs/loki/latest/clients/promtail/

---

## 💡 Tips

1. **Use labels wisely**: Don't create too many unique label combinations (causes high cardinality)

   ```logql
   # Good ✅
   {category="audit", event_type="auth.login.success"}

   # Bad ❌ (user_id creates millions of unique streams)
   {user_id="123"}
   ```

2. **Structured metadata for high-cardinality data**: Use structured metadata (Loki 3.x) for user IDs, request IDs

   ```go
   // Promtail automatically extracts these as structured metadata
   logger.Info("User action",
       zap.String("user_id", userID),  // → structured metadata
       zap.String("action", "login"),  // → label
   )
   ```

3. **Aggregate before visualizing**: For dashboards, use aggregations

   ```logql
   # Instead of raw logs
   {category="audit"}

   # Use aggregation
   sum by (event_type) (rate({category="audit"}[5m]))
   ```

4. **Set appropriate retention**: Balance between storage cost and compliance

   ```yaml
   # loki-config-production.yml
   limits_config:
     retention_period: 720h  # 30 days
   ```
