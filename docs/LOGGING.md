# Hệ Thống Logging - OAuth2 Server

## Tổng Quan

Hệ thống logging được thiết kế để cung cấp:
- **Structured logging** với JSON format
- **Audit trail** cho tất cả security events
- **Performance monitoring** cho HTTP requests và database queries
- **Centralized logging** với Loki
- **Visualization** với Grafana dashboards
- **Alerting** cho critical events

## Kiến Trúc

```
Application (Go + Zap)
    ↓ JSON logs to stdout
Promtail (Log collector)
    ↓ Parse & Label
Loki (Log aggregation)
    ↓ Query
Grafana (Visualization)
```

---

## Log Categories

### 1. Application Logs (category="application")

Logs chung của application lifecycle và business logic.

**Fields**:
- `level`: debug, info, warn, error, fatal
- `msg`: Log message
- `timestamp`: ISO8601 timestamp
- `caller`: Source file và line number

**Example**:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "caller": "router/router.go:95",
  "msg": "HTTP Server started",
  "category": "application",
  "port": "8080"
}
```

---

### 2. Audit Logs (category="audit")

Logs cho tất cả security-sensitive events.

**Fields**:
- `event_type`: Loại event (auth.login.success, oauth2.token.issued, etc.)
- `status`: success, failure, denied, pending
- `user_id`: UUID của user
- `client_id`: UUID của OAuth2 client
- `ip_address`: IP address của request
- `user_agent`: Browser/client user agent
- `request_id`: Unique request ID

**Event Types**:

#### Authentication Events
- `auth.login.attempt` - Login attempt
- `auth.login.success` - Successful login
- `auth.login.failure` - Failed login
- `auth.logout` - User logout
- `auth.password.change` - Password changed
- `auth.password.reset.request` - Password reset requested
- `auth.password.reset` - Password reset completed
- `auth.email.verification` - Email verified
- `auth.account.created` - New account created
- `auth.account.deleted` - Account deleted

#### OAuth2 Events
- `oauth2.authorize` - Authorization request
- `oauth2.token.issued` - Token issued
- `oauth2.token.refreshed` - Token refreshed
- `oauth2.token.revoked` - Token revoked
- `oauth2.token.introspected` - Token introspected
- `oauth2.consent.granted` - User granted consent
- `oauth2.consent.revoked` - Consent revoked
- `oauth2.consent.denied` - User denied consent

#### OAuth2 Client Management
- `oauth2.client.created` - Client created
- `oauth2.client.updated` - Client updated
- `oauth2.client.deleted` - Client deleted
- `oauth2.client.secret.regenerated` - Client secret regenerated

#### Security Events
- `security.unauthorized.access` - Unauthorized access attempt
- `security.suspicious.activity` - Suspicious activity detected
- `security.rate_limit.exceeded` - Rate limit exceeded
- `security.invalid.token` - Invalid token used

**Example**:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:31:20.456Z",
  "category": "audit",
  "event_type": "oauth2.token.issued",
  "status": "success",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_id": "660e8400-e29b-41d4-a716-446655440000",
  "grant_type": "authorization_code",
  "token_type": "access_token",
  "scopes": ["openid", "profile", "email"],
  "ip_address": "192.168.1.100",
  "request_id": "01JQRST7EXAMPLE"
}
```

---

### 3. Performance Logs (category="performance")

Logs cho performance metrics và timing information.

**Fields**:
- `operation_type`: http, database, redis, oauth2, external_api
- `operation`: Tên operation
- `duration_ms`: Duration in milliseconds
- `success`: true/false

**Operation Types**:

#### HTTP Operations
```json
{
  "category": "performance",
  "operation_type": "http",
  "operation": "http_request",
  "method": "POST",
  "path": "/oauth2/token",
  "status_code": 200,
  "duration_ms": 145,
  "success": true
}
```

#### Database Operations
```json
{
  "category": "performance",
  "operation_type": "database",
  "operation": "db_query",
  "query_type": "SELECT",
  "query": "SELECT * FROM users WHERE id = $1",
  "duration_ms": 23,
  "rows_affected": 1,
  "success": true
}
```

#### Cache Operations
```json
{
  "category": "performance",
  "operation_type": "redis",
  "operation": "cache_get",
  "cache_hit": true,
  "duration_ms": 5,
  "success": true,
  "metadata": {
    "key": "session:abc123"
  }
}
```

---

### 4. HTTP Logs (category="http")

Automatic logs cho mỗi HTTP request.

**Fields**:
- `method`: HTTP method (GET, POST, etc.)
- `path`: Request path
- `status_code`: HTTP status code
- `duration_ms`: Request duration
- `response_size`: Response size in bytes
- `ip_address`: Client IP
- `user_agent`: Client user agent
- `request_id`: Unique request ID

**Example**:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:32:00.789Z",
  "category": "http",
  "msg": "HTTP request completed",
  "method": "POST",
  "path": "/oauth2/token",
  "status_code": 200,
  "duration_ms": 145,
  "response_size": 1234,
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "request_id": "01JQRST7EXAMPLE"
}
```

---

### 5. Error Logs (category="error")

Logs cho errors và exceptions.

**Fields**:
- `error`: Error message
- `error_code`: Application error code
- `error_detail`: Detailed error information
- `stack_trace`: Stack trace (for panics)

**Example**:
```json
{
  "level": "error",
  "ts": "2024-01-15T10:33:00.123Z",
  "category": "error",
  "msg": "Database connection failed",
  "error": "connection refused",
  "error_code": "DB_CONNECTION_ERROR",
  "error_detail": "Failed to connect to PostgreSQL at localhost:5432",
  "request_id": "01JQRST7EXAMPLE"
}
```

---

## Usage Examples

### 1. Audit Logging

```go
import (
    "context"
    "system/internal/platform/logger"
)

// Initialize audit logger
auditLogger := logger.NewAuditLogger(appLogger)

// Log successful login
auditLogger.LogLoginAttempt(
    ctx,
    "user@example.com",
    "192.168.1.100",
    "Mozilla/5.0...",
    true,  // success
    "",    // no error
)

// Log token issued
userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
clientID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")
auditLogger.LogTokenIssued(
    ctx,
    &userID,
    &clientID,
    "authorization_code",
    "access_token",
    []string{"openid", "profile"},
    "192.168.1.100",
)

// Log consent granted
auditLogger.LogConsentGranted(
    ctx,
    userID,
    clientID,
    []string{"openid", "profile", "email"},
    "192.168.1.100",
    "Mozilla/5.0...",
)
```

### 2. Performance Logging

```go
// Initialize performance logger
perfLogger := logger.NewPerformanceLogger(appLogger)

// Using timer for automatic duration tracking
timer := perfLogger.NewTimer("user_authentication", logger.OpTypeAuthentication)
defer timer.End(ctx, err)

// ... perform authentication ...

// Manual performance logging
perfLogger.LogHTTPRequest(
    ctx,
    "POST",
    "/oauth2/token",
    200,
    145*time.Millisecond,
    nil,
)

perfLogger.LogDatabaseQuery(
    ctx,
    "SELECT",
    "SELECT * FROM users WHERE email = $1",
    23*time.Millisecond,
    1, // rows affected
    nil,
)
```

### 3. Context-aware Logging

```go
import "system/internal/platform/logger"

// Add context information
ctx = logger.WithRequestID(ctx, requestID)
ctx = logger.WithUserID(ctx, userID)
ctx = logger.WithIPAddress(ctx, ipAddress)

// Create logger with context
ctxLogger := logger.NewLoggerWithContext(ctx, appLogger)

// All logs will include context fields
ctxLogger.Info("Processing user request")
// Output includes: request_id, user_id, ip_address
```

---

## Loki Query Examples (LogQL)

### 1. Search Audit Logs

```logql
# All OAuth2 token events
{category="audit", event_type=~"oauth2.token.*"}

# Failed login attempts
{category="audit", event_type="auth.login.failure"}

# All events for a specific user
{category="audit"} | json | user_id="550e8400-e29b-41d4-a716-446655440000"

# Suspicious activities
{category="audit", event_type="security.suspicious.activity"}
```

### 2. Performance Queries

```logql
# HTTP requests slower than 2 seconds
{category="performance", operation_type="http"} | json | duration_ms > 2000

# Database queries by type
{category="performance", operation_type="database"} | json | line_format "{{.query_type}}: {{.duration_ms}}ms"

# Average response time (5min window)
avg_over_time({category="performance", operation_type="http"} | json | unwrap duration_ms [5m])

# P95 latency
quantile_over_time(0.95, {category="performance", operation_type="http"} | json | unwrap duration_ms [5m])
```

### 3. Error Tracking

```logql
# All errors
{level="error"}

# Errors by category
{level="error"} | json | line_format "{{.category}}: {{.msg}}"

# Rate of errors (per minute)
rate({level="error"}[1m])

# Errors for specific request
{request_id="01JQRST7EXAMPLE"}
```

### 4. Security Monitoring

```logql
# Unauthorized access attempts
{event_type="security.unauthorized.access"}

# Rate limit violations
{event_type="security.rate_limit.exceeded"}

# Failed authentication rate (last hour)
sum(rate({category="audit", status="failure"}[1h]))

# Top IPs with failed logins
topk(10, sum by (ip_address) (count_over_time({event_type="auth.login.failure"}[1h])))
```

---

## Grafana Dashboards

### OAuth2 Overview Dashboard

**Location**: `grafana-dashboards/oauth2-overview.json`

**Panels**:
1. **OAuth2 Events Rate** - Rate of OAuth2 events over time
2. **Tokens Issued** - Count of tokens issued in last hour
3. **Failed Authentications** - Count of failed auth attempts
4. **Recent Audit Events** - Table of recent OAuth2 events
5. **HTTP Request Duration** - P95, P99, Average latency
6. **Error Logs** - Live error log stream

**Import to Grafana**:
```bash
# Via UI: Dashboards → Import → Upload JSON file
# Or use API
curl -X POST http://localhost:5555/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @grafana-dashboards/oauth2-overview.json
```

---

## Alert Rules

### Example Alert Rules (Loki Ruler)

Create file: `loki-rules/alerts.yml`

```yaml
groups:
  - name: oauth2_alerts
    interval: 1m
    rules:
      # Alert on high failed login rate
      - alert: HighFailedLoginRate
        expr: |
          sum(rate({category="audit", event_type="auth.login.failure"}[5m])) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High failed login rate detected"
          description: "More than 10 failed logins per minute in the last 5 minutes"

      # Alert on slow requests
      - alert: SlowHTTPRequests
        expr: |
          quantile_over_time(0.95, {category="performance", operation_type="http"}
            | json | unwrap duration_ms [5m]) > 2000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High HTTP latency detected"
          description: "P95 latency is above 2 seconds"

      # Alert on error rate
      - alert: HighErrorRate
        expr: |
          sum(rate({level="error"}[5m])) > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "More than 5 errors per minute"

      # Alert on unauthorized access attempts
      - alert: UnauthorizedAccessSpike
        expr: |
          sum(rate({event_type="security.unauthorized.access"}[5m])) > 20
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Spike in unauthorized access attempts"
          description: "Possible security attack detected"
```

---

## Configuration

### Development

Use existing configs:
- `loki-config.yml`
- `promtail-config.yml`

### Production

Use production configs with better performance and retention:
- `loki-config-production.yml`
- `promtail-config-production.yml`

**Key differences**:
- Structured metadata enabled
- 30-day retention
- Better rate limits
- Compaction enabled
- Query caching
- Alert rules support

### Switching to Production Config

Update `docker-compose.yml`:

```yaml
loki:
  image: grafana/loki:3.5
  command: -config.file=/etc/loki/loki-config-production.yml
  volumes:
    - loki_data:/loki
    - ./loki-config-production.yml:/etc/loki/loki-config-production.yml

promtail:
  image: grafana/promtail:3.5
  command: -config.file=/etc/promtail/promtail-config-production.yml
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock
    - ./promtail-config-production.yml:/etc/promtail/promtail-config-production.yml
```

---

## Best Practices

### 1. Always Include Request Context

```go
// At the start of HTTP handler
requestID := uuid.NewV7().String()
ctx = logger.WithRequestID(ctx, requestID)
ctx = logger.WithIPAddress(ctx, c.ClientIP())
ctx = logger.WithUserAgent(ctx, c.UserAgent())
```

### 2. Log Security Events

```go
// Always log authentication attempts
auditLogger.LogLoginAttempt(ctx, email, ip, userAgent, success, errorDetail)

// Always log token operations
auditLogger.LogTokenIssued(ctx, userID, clientID, grantType, tokenType, scopes, ip)

// Always log consent decisions
auditLogger.LogConsentGranted(ctx, userID, clientID, scopes, ip, userAgent)
```

### 3. Track Performance

```go
// For important operations
timer := perfLogger.NewTimer("create_user", logger.OpTypeDatabase)
defer timer.End(ctx, err)

user, err := userService.Create(ctx, req)
// Timer automatically logs duration and success/failure
```

### 4. Use Structured Fields

```go
// Good ✅
logger.Info("User created",
    zap.String("user_id", userID.String()),
    zap.String("email", email),
)

// Bad ❌
logger.Info(fmt.Sprintf("User %s created with email %s", userID, email))
```

### 5. Appropriate Log Levels

- `DEBUG`: Detailed information for debugging (disabled in production)
- `INFO`: General information (default level)
- `WARN`: Warning messages (slow queries, deprecated features)
- `ERROR`: Error messages (failed operations, exceptions)
- `FATAL`: Critical errors that cause application to exit

---

## Troubleshooting

### Logs not appearing in Loki

1. Check Promtail is running:
   ```bash
   docker logs promtail
   ```

2. Check Promtail can reach Loki:
   ```bash
   curl http://localhost:3100/ready
   ```

3. Check log format is JSON:
   ```bash
   docker logs <container-name> | head -1 | jq .
   ```

### Query performance issues

1. Use specific label filters:
   ```logql
   # Good ✅
   {category="audit", event_type="auth.login.success"}

   # Bad ❌ (scans all logs)
   {} | json | event_type="auth.login.success"
   ```

2. Limit time range:
   - Use shorter time ranges for ad-hoc queries
   - Use aggregations for long time ranges

3. Use pre-aggregated metrics for dashboards

### High disk usage

1. Check retention settings in `loki-config-production.yml`:
   ```yaml
   limits_config:
     retention_period: 720h  # 30 days
   ```

2. Manually compact:
   ```bash
   # Check compactor logs
   docker logs loki | grep compactor
   ```

---

## Monitoring the Logging System

### Loki Metrics

Loki exposes Prometheus metrics at `http://localhost:3100/metrics`

**Key metrics**:
- `loki_ingester_chunks_created_total` - Chunks created
- `loki_distributor_bytes_received_total` - Bytes received
- `loki_query_frontend_queries_total` - Queries executed

### Promtail Metrics

Promtail exposes metrics at `http://localhost:9080/metrics`

**Key metrics**:
- `promtail_sent_entries_total` - Log entries sent
- `promtail_dropped_entries_total` - Dropped entries
- `promtail_read_bytes_total` - Bytes read

---

## References

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
- [Promtail Configuration](https://grafana.com/docs/loki/latest/clients/promtail/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [Zap Logger](https://github.com/uber-go/zap)
