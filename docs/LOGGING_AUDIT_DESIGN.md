# Logging & Audit Trail Design - Loki-Based Architecture

**Thiết kế tối ưu sử dụng Grafana Loki stack có sẵn**

---

## 🎯 Design Goals

- ✅ **Lightweight**: Tối ưu resource usage
- ✅ **Cost-effective**: Dùng infrastructure có sẵn (Loki + Promtail + Grafana)
- ✅ **Performant**: Không impact API response time
- ✅ **Compliance-ready**: Đáp ứng audit requirements
- ✅ **Simple**: Dễ maintain, dễ query
- ✅ **Scalable**: Loki scale horizontally dễ dàng

---

## 📐 Architecture Overview

### Stack Sử Dụng (Already in docker-compose!)

```
┌──────────────────────────────────────────────────────────┐
│                     Existing Stack                       │
├──────────────────────────────────────────────────────────┤
│  PostgreSQL  │  Redis  │  Loki  │  Promtail  │  Grafana │
└──────────────────────────────────────────────────────────┘
```

**✨ Lợi ích:**
- Không cần thêm MongoDB container
- Tận dụng Loki stack đã có
- Grafana sẵn sàng để query & visualize
- Promtail tự động collect logs

### 2-Tier Logging Strategy

```
                    ┌──────────────────────┐
                    │    OAuth2 Server     │
                    │   (Zap Logger)       │
                    └──────────┬───────────┘
                               │
                    ┌──────────┴───────────┐
                    │                      │
           ┌────────▼─────────┐   ┌───────▼──────────┐
           │  Application Logs │   │   Audit Logs     │
           │  (Debug/Perf)     │   │ (Security/Audit) │
           │  level=info/debug │   │  level=info      │
           └────────┬───────────┘   └───────┬──────────┘
                    │                       │
                    └──────────┬────────────┘
                               │
                    ┌──────────▼───────────┐
                    │    Stdout/File       │
                    │   (JSON format)      │
                    └──────────┬───────────┘
                               │
                    ┌──────────▼───────────┐
                    │      Promtail        │
                    │   (Log Collector)    │
                    └──────────┬───────────┘
                               │
                    ┌──────────▼───────────┐
                    │        Loki          │
                    │  (Log Aggregation)   │
                    │  Retention: 90 days  │
                    └──────────┬───────────┘
                               │
                    ┌──────────▼───────────┐
                    │      Grafana         │
                    │   (Query/Visualize)  │
                    └──────────────────────┘
```

**Cách hoạt động:**
1. **Application** log events với Zap (structured JSON)
2. **Promtail** collect logs từ stdout/files
3. **Loki** store và index logs với labels
4. **Grafana** query và visualize logs

---

## Layer 1: Application Logs (Zap → Stdout/File → Promtail → Loki)

### Purpose
- Runtime debugging
- Performance monitoring
- Error tracking
- Request/Response logging

### Output Strategy

**Dual Output (Recommended):**
```
Zap Logger
  ├─ Stdout (primary) → Promtail → Loki
  └─ File (backup)    → Local retention 7 days
```

**File backup rotation:**
```
logs/
  ├── app.log              # Current log
  ├── app-2024-11-02.log   # Rotated by date
  ├── app-2024-11-01.log
  └── app-2024-10-31.log
```

**Rotation Policy (File backup only):**
- Rotate daily
- Keep last 7 days locally
- Max size: 100MB per file
- Auto compress old logs (gzip)
- Loki retention: 90 days (configurable)

**Implementation:**
```go
// Use: go.uber.org/zap + lumberjack for file rotation
import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() *zap.Logger {
    // Encoder config (JSON for Loki compatibility)
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "timestamp",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

    // Dual output: stdout + file
    stdoutWriter := zapcore.AddSync(os.Stdout)  // Primary: Promtail collects từ stdout
    fileWriter := zapcore.AddSync(&lumberjack.Logger{  // Backup: Local file
        Filename:   "logs/app.log",
        MaxSize:    100, // MB
        MaxBackups: 7,   // Keep 7 files
        MaxAge:     7,   // Days
        Compress:   true,
    })

    // Multi writer (both stdout and file)
    multiWriter := zapcore.NewMultiWriteSyncer(stdoutWriter, fileWriter)

    core := zapcore.NewCore(jsonEncoder, multiWriter, zapcore.InfoLevel)
    return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
```

### Log Format (JSON)

```json
{
  "timestamp": "2024-11-02T10:30:45.123Z",
  "level": "info",
  "logger": "oauth2",
  "caller": "handler/token.go:192",
  "msg": "Token issued",
  "request_id": "req_abc123",
  "client_id": "10000000-0000-0000-0000-000000000001",
  "grant_type": "authorization_code",
  "user_id": "00000000-0000-0000-0000-000000000001",
  "scopes": ["openid", "profile", "email"],
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "duration_ms": 125
}
```

### Querying (Simple grep)

```bash
# Find all token requests from a client
grep "client_id.*10000000" logs/app.log

# Find errors in last 7 days
grep "\"level\":\"error\"" logs/app-*.log

# Find slow requests (>1000ms)
grep -E "duration_ms\":[0-9]{4,}" logs/app.log

# Track specific user activity
grep "user_id.*00000000" logs/app.log | jq .
```

**Cost:** ✅ FREE (disk space only: ~100MB × 7 days = 700MB)

---

## Layer 2: Security Audit Logs (Loki)

### Purpose
- Security investigation
- Compliance (GDPR, SOC2, HIPAA)
- Forensics
- Long-term retention (90 days default, configurable lên 1 year+)

### Why Loki (Not MongoDB/PostgreSQL)?

| Criteria | Loki | MongoDB | PostgreSQL |
|----------|------|---------|------------|
| **Infrastructure** | ✅ Đã có sẵn | ❌ Cần thêm container | ⚠️ Có rồi (nhưng cho transactional) |
| **Write Performance** | ✅ Excellent (append-only) | ✅ Excellent | ⚠️ Good |
| **Log-optimized** | ✅ Designed cho logs | ⚠️ General purpose | ⚠️ General purpose |
| **Query Language** | ✅ LogQL (simple) | ✅ Rich aggregation | ✅ SQL |
| **Compression** | ✅ Excellent (chunking) | ✅ Built-in | ⚠️ Need TOAST |
| **Retention** | ✅ Time-based | ✅ TTL indexes | ⚠️ Manual partitioning |
| **Horizontal Scaling** | ✅ Easy | ✅ Easy sharding | ⚠️ Complex |
| **Integration** | ✅ Grafana native | ⚠️ Need tools | ⚠️ Need tools |
| **Cost** | ✅ Free (existing) | ⚠️ Resource overhead | ⚠️ Schema overhead |

**Decision:** Loki cho audit logs (tận dụng stack có sẵn), PostgreSQL cho transactional data

---

## Loki Label Strategy

### Label Design (Low Cardinality!)

**⚠️ CRITICAL:** Loki labels phải có **low cardinality** để tránh performance issues.

**Good Labels (Low Cardinality):**
```
- log_type: "audit" | "application"
- event_category: "auth" | "authz" | "token" | "admin" | "security"
- event_type: "login_success" | "login_failed" | "token_issued" | ...
- severity: "info" | "warn" | "error" | "critical"
- environment: "dev" | "staging" | "prod"
```

**Bad Labels (High Cardinality - AVOID!):**
```
❌ user_id (millions of unique values)
❌ client_id (thousands of unique values)
❌ ip_address (millions of unique values)
❌ request_id (unique every request)
```

**Solution:** Put high-cardinality data vào **JSON log body**, không phải labels!

### Log Structure for Loki

**Labels:** Categorical data (low cardinality)
**Log Body (JSON):** Detailed data (high cardinality OK)

```json
{
  "timestamp": "2024-11-03T10:30:45.123Z",
  "level": "info",
  "log_type": "audit",
  "event_category": "auth",
  "event_type": "login_success",
  "severity": "info",

  "user_id": "00000000-0000-0000-0000-000000000001",
  "email": "user@example.com",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "session_id": "sess_xyz789",
  "location": {
    "country": "VN",
    "city": "Hanoi"
  },
  "metadata": {
    "request_id": "req_abc123",
    "duration_ms": 145
  }
}
```

### Event Categories

```
┌─────────────────────────────────────────────┐
│          event_category (Label)             │
├─────────────────────────────────────────────┤
│  auth       │  Authentication events        │
│  authz      │  Authorization/Consent        │
│  token      │  Token lifecycle              │
│  admin      │  Admin actions                │
│  security   │  Security violations          │
└─────────────────────────────────────────────┘
```

### Event Category: Authentication (`event_category="auth"`)

**Purpose:** Track authentication attempts

**Event Types:**
- `login_success`
- `login_failed` (with reason)
- `logout`
- `password_changed`
- `email_verified`
- `password_reset_requested`
- `password_reset_completed`

**Example Log:**
```json
{
  "timestamp": "2024-11-03T10:30:45.123Z",
  "level": "info",
  "log_type": "audit",
  "event_category": "auth",
  "event_type": "login_success",
  "severity": "info",

  "user_id": "00000000-0000-0000-0000-000000000001",
  "email": "user@example.com",
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
  "session_id": "sess_xyz789",
  "location": {
    "country": "VN",
    "city": "Hanoi"
  },
  "device": {
    "type": "browser",
    "os": "Windows",
    "browser": "Chrome"
  },
  "metadata": {
    "request_id": "req_abc123",
    "duration_ms": 145
  }
}
```

---

### Event Category: Authorization (`event_category="authz"`)

**Purpose:** Track authorization grants and consents

**Event Types:**
- `consent_granted`
- `consent_denied`
- `consent_revoked`
- `authorization_code_issued`
- `authorization_code_used`
- `authorization_failed` (with reason)

**Example Log:**
```json
{
  "timestamp": "2024-11-03T10:35:20.456Z",
  "level": "info",
  "log_type": "audit",
  "event_category": "authz",
  "event_type": "consent_granted",
  "severity": "info",

  "user_id": "00000000-0000-0000-0000-000000000001",
  "client_id": "10000000-0000-0000-0000-000000000001",
  "client_name": "My App",
  "scopes_requested": ["openid", "profile", "email", "offline_access"],
  "scopes_granted": ["openid", "profile", "email"],
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "authorization_code": "auth_code_hash_abc123",
  "metadata": {
    "request_id": "req_def456"
  }
}
```

---

### Event Category: Token Lifecycle (`event_category="token"`)

**Purpose:** Track token lifecycle

**Event Types:**
- `access_token_issued`
- `access_token_revoked`
- `refresh_token_issued`
- `refresh_token_used`
- `refresh_token_revoked`
- `refresh_token_rotated`
- `id_token_issued`
- `token_introspected`

**Example Log:**
```json
{
  "timestamp": "2024-11-03T10:35:25.789Z",
  "level": "info",
  "log_type": "audit",
  "event_category": "token",
  "event_type": "access_token_issued",
  "severity": "info",

  "user_id": "00000000-0000-0000-0000-000000000001",
  "client_id": "10000000-0000-0000-0000-000000000001",
  "grant_type": "authorization_code",
  "token_signature": "sha256_hash_of_token",
  "scopes": ["openid", "profile", "email"],
  "expires_at": "2024-11-03T11:35:25.789Z",
  "ip_address": "192.168.1.100",
  "metadata": {
    "request_id": "req_def456",
    "pkce_used": true
  }
}
```

---

### Event Category: Admin Actions (`event_category="admin"`)

**Purpose:** Track admin/privileged actions

**Event Types:**
- `client_created`
- `client_updated`
- `client_deleted`
- `client_secret_rotated`
- `user_created`
- `user_updated`
- `user_deleted`
- `user_suspended`
- `consent_revoked_by_admin`
- `tokens_revoked_by_admin`

**Example Log:**
```json
{
  "timestamp": "2024-11-03T09:00:00.000Z",
  "level": "info",
  "log_type": "audit",
  "event_category": "admin",
  "event_type": "client_created",
  "severity": "info",

  "admin_user_id": "admin_00000000",
  "admin_email": "admin@example.com",
  "action": "create_oauth2_client",
  "resource_type": "oauth2_client",
  "resource_id": "10000000-0000-0000-0000-000000000005",
  "changes": {
    "client_name": "New App",
    "redirect_uris": ["https://newapp.com/callback"],
    "grant_types": ["authorization_code", "refresh_token"]
  },
  "ip_address": "10.0.1.50",
  "user_agent": "Mozilla/5.0...",
  "metadata": {
    "request_id": "req_admin123"
  }
}
```

---

### Event Category: Security Events (`event_category="security"`)

**Purpose:** Track security violations and anomalies

**Event Types:**
- `token_reuse_detected`
- `rate_limit_exceeded`
- `invalid_client_credentials`
- `pkce_validation_failed`
- `suspicious_login_location`
- `brute_force_detected`
- `account_takeover_attempt`
- `invalid_redirect_uri`

**Severity Levels:**
- `info`: Informational
- `warn`: Suspicious activity
- `error`: Likely attack
- `critical`: Confirmed breach/attack

**Example Log:**
```json
{
  "timestamp": "2024-11-03T14:30:00.000Z",
  "level": "error",
  "log_type": "audit",
  "event_category": "security",
  "event_type": "token_reuse_detected",
  "severity": "critical",

  "user_id": "00000000-0000-0000-0000-000000000001",
  "client_id": "10000000-0000-0000-0000-000000000001",
  "description": "Refresh token was reused after rotation",
  "token_signature": "sha256_hash",
  "ip_address": "192.168.1.200",
  "user_agent": "curl/7.68.0",
  "action_taken": "revoked_token_family",
  "metadata": {
    "original_token_issued_at": "2024-11-03T10:00:00.000Z",
    "rotated_at": "2024-11-03T12:00:00.000Z",
    "reuse_attempt_at": "2024-11-03T14:30:00.000Z"
  }
}
```

---

## Code Structure

```
internal/
  ├── platform/
  │   └── logger/
  │       ├── logger.go           # Zap logger setup (stdout + file)
  │       ├── audit.go            # Audit logger wrapper
  │       ├── fields.go           # Common log fields & labels
  │       └── middleware.go       # Request logging middleware
  │
  └── domain/
      └── audit.go                # Audit event types & interfaces
```

**Key Changes from MongoDB design:**
- No MongoDB client/connection needed
- AuditLogger chỉ là wrapper around Zap logger
- Logs đi qua stdout → Promtail → Loki (automatic)
- Simpler codebase, less dependencies

---

## Implementation

### 1. Logger Setup (`internal/platform/logger/logger.go`)

```go
package logger

import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

var (
    AppLogger   *zap.Logger
    AuditLogger *AuditLogger
)

// Initialize loggers
func Initialize() error {
    // Application logger (Zap + dual output: stdout + file)
    AppLogger = newAppLogger()

    // Audit logger (wrapper around Zap)
    AuditLogger = NewAuditLogger(AppLogger)
    return nil
}

func newAppLogger() *zap.Logger {
    // Encoder config (JSON for Loki/Promtail compatibility)
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "timestamp",
        LevelKey:       "level",
        MessageKey:     "msg",
        CallerKey:      "caller",
        StacktraceKey:  "stacktrace",
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

    // Dual output: stdout (primary for Promtail) + file (backup)
    stdoutWriter := zapcore.AddSync(os.Stdout)
    fileWriter := zapcore.AddSync(&lumberjack.Logger{
        Filename:   "logs/app.log",
        MaxSize:    100, // MB
        MaxBackups: 7,
        MaxAge:     7, // days
        Compress:   true,
    })
    multiWriter := zapcore.NewMultiWriteSyncer(stdoutWriter, fileWriter)

    core := zapcore.NewCore(jsonEncoder, multiWriter, zapcore.InfoLevel)
    return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
```

---

### 2. Audit Logger (`internal/platform/logger/audit.go`)

```go
package logger

import (
    "context"
    "go.uber.org/zap"
)

type AuditLogger struct {
    logger *zap.Logger
}

func NewAuditLogger(baseLogger *zap.Logger) *AuditLogger {
    // Create specialized audit logger with default fields
    auditLogger := baseLogger.With(
        zap.String("log_type", "audit"),
    )
    return &AuditLogger{logger: auditLogger}
}

// Authentication Events
func (l *AuditLogger) LogLoginSuccess(ctx context.Context, userID, email, ip, userAgent, sessionID string) {
    l.logger.Info("User login successful",
        zap.String("event_category", "auth"),
        zap.String("event_type", "login_success"),
        zap.String("severity", "info"),
        zap.String("user_id", userID),
        zap.String("email", email),
        zap.String("ip_address", ip),
        zap.String("user_agent", userAgent),
        zap.String("session_id", sessionID),
    )
}

func (l *AuditLogger) LogLoginFailed(ctx context.Context, email, reason, ip, userAgent string) {
    l.logger.Warn("User login failed",
        zap.String("event_category", "auth"),
        zap.String("event_type", "login_failed"),
        zap.String("severity", "warn"),
        zap.String("email", email),
        zap.String("reason", reason),
        zap.String("ip_address", ip),
        zap.String("user_agent", userAgent),
    )
}

func (l *AuditLogger) LogLogout(ctx context.Context, userID, sessionID string) {
    l.logger.Info("User logout",
        zap.String("event_category", "auth"),
        zap.String("event_type", "logout"),
        zap.String("severity", "info"),
        zap.String("user_id", userID),
        zap.String("session_id", sessionID),
    )
}

// Authorization Events
func (l *AuditLogger) LogConsentGranted(ctx context.Context, userID, clientID, clientName string, scopesGranted []string, ip, userAgent string) {
    l.logger.Info("User granted consent",
        zap.String("event_category", "authz"),
        zap.String("event_type", "consent_granted"),
        zap.String("severity", "info"),
        zap.String("user_id", userID),
        zap.String("client_id", clientID),
        zap.String("client_name", clientName),
        zap.Strings("scopes_granted", scopesGranted),
        zap.String("ip_address", ip),
        zap.String("user_agent", userAgent),
    )
}

func (l *AuditLogger) LogConsentDenied(ctx context.Context, userID, clientID, clientName string, ip, userAgent string) {
    l.logger.Info("User denied consent",
        zap.String("event_category", "authz"),
        zap.String("event_type", "consent_denied"),
        zap.String("severity", "info"),
        zap.String("user_id", userID),
        zap.String("client_id", clientID),
        zap.String("client_name", clientName),
        zap.String("ip_address", ip),
        zap.String("user_agent", userAgent),
    )
}

// Token Events
func (l *AuditLogger) LogTokenIssued(ctx context.Context, tokenType, userID, clientID, grantType string, scopes []string, ip string) {
    l.logger.Info("Token issued",
        zap.String("event_category", "token"),
        zap.String("event_type", tokenType+"_issued"), // e.g., "access_token_issued"
        zap.String("severity", "info"),
        zap.String("user_id", userID),
        zap.String("client_id", clientID),
        zap.String("grant_type", grantType),
        zap.Strings("scopes", scopes),
        zap.String("ip_address", ip),
    )
}

func (l *AuditLogger) LogTokenRevoked(ctx context.Context, tokenType, tokenSig, userID, clientID, reason string) {
    l.logger.Info("Token revoked",
        zap.String("event_category", "token"),
        zap.String("event_type", tokenType+"_revoked"),
        zap.String("severity", "info"),
        zap.String("token_signature", tokenSig),
        zap.String("user_id", userID),
        zap.String("client_id", clientID),
        zap.String("reason", reason),
    )
}

// Admin Events
func (l *AuditLogger) LogClientCreated(ctx context.Context, adminID, clientID, clientName string) {
    l.logger.Info("OAuth2 client created",
        zap.String("event_category", "admin"),
        zap.String("event_type", "client_created"),
        zap.String("severity", "info"),
        zap.String("admin_user_id", adminID),
        zap.String("resource_type", "oauth2_client"),
        zap.String("resource_id", clientID),
        zap.String("client_name", clientName),
    )
}

func (l *AuditLogger) LogClientUpdated(ctx context.Context, adminID, clientID string, changes map[string]interface{}) {
    l.logger.Info("OAuth2 client updated",
        zap.String("event_category", "admin"),
        zap.String("event_type", "client_updated"),
        zap.String("severity", "info"),
        zap.String("admin_user_id", adminID),
        zap.String("resource_id", clientID),
        zap.Any("changes", changes),
    )
}

// Security Events
func (l *AuditLogger) LogSecurityEvent(ctx context.Context, eventType, severity, description string, userID, clientID, ip string, metadata map[string]interface{}) {
    l.logger.Error("Security event detected",
        zap.String("event_category", "security"),
        zap.String("event_type", eventType),
        zap.String("severity", severity),
        zap.String("description", description),
        zap.String("user_id", userID),
        zap.String("client_id", clientID),
        zap.String("ip_address", ip),
        zap.Any("metadata", metadata),
    )
}

func (l *AuditLogger) LogTokenReuse(ctx context.Context, tokenSig, userID, clientID, ip string) {
    l.LogSecurityEvent(ctx, "token_reuse_detected", "critical",
        "Refresh token was reused after rotation",
        userID, clientID, ip,
        map[string]interface{}{
            "token_signature": tokenSig,
            "action_taken":    "revoked_token_family",
        },
    )
}
```

**Key Changes:**
- ❌ No MongoDB connection needed
- ✅ Simple wrapper around Zap logger
- ✅ Structured logging with consistent fields
- ✅ Automatic labels: `log_type="audit"`, `event_category`, `event_type`, `severity`
- ✅ Logs go to stdout → Promtail → Loki automatically

---

### 3. Event Types (`internal/domain/audit.go`)

```go
package domain

// Event categories for audit logs (used as Loki labels)
const (
    EventCategoryAuth     = "auth"
    EventCategoryAuthz    = "authz"
    EventCategoryToken    = "token"
    EventCategoryAdmin    = "admin"
    EventCategorySecurity = "security"
)

// Authentication event types
const (
    EventLoginSuccess            = "login_success"
    EventLoginFailed             = "login_failed"
    EventLogout                  = "logout"
    EventPasswordChanged         = "password_changed"
    EventEmailVerified           = "email_verified"
    EventPasswordResetRequested  = "password_reset_requested"
    EventPasswordResetCompleted  = "password_reset_completed"
)

// Authorization event types
const (
    EventConsentGranted         = "consent_granted"
    EventConsentDenied          = "consent_denied"
    EventConsentRevoked         = "consent_revoked"
    EventAuthCodeIssued         = "authorization_code_issued"
    EventAuthCodeUsed           = "authorization_code_used"
    EventAuthorizationFailed    = "authorization_failed"
)

// Token event types
const (
    EventAccessTokenIssued      = "access_token_issued"
    EventAccessTokenRevoked     = "access_token_revoked"
    EventRefreshTokenIssued     = "refresh_token_issued"
    EventRefreshTokenUsed       = "refresh_token_used"
    EventRefreshTokenRevoked    = "refresh_token_revoked"
    EventRefreshTokenRotated    = "refresh_token_rotated"
    EventIDTokenIssued          = "id_token_issued"
    EventTokenIntrospected      = "token_introspected"
)

// Admin event types
const (
    EventClientCreated          = "client_created"
    EventClientUpdated          = "client_updated"
    EventClientDeleted          = "client_deleted"
    EventClientSecretRotated    = "client_secret_rotated"
    EventUserCreated            = "user_created"
    EventUserUpdated            = "user_updated"
    EventUserDeleted            = "user_deleted"
)

// Security event types
const (
    EventTokenReuseDetected     = "token_reuse_detected"
    EventRateLimitExceeded      = "rate_limit_exceeded"
    EventInvalidClientCreds     = "invalid_client_credentials"
    EventPKCEValidationFailed   = "pkce_validation_failed"
    EventBruteForceDetected     = "brute_force_detected"
    EventInvalidRedirectURI     = "invalid_redirect_uri"
)

// Severity levels
const (
    SeverityInfo     = "info"
    SeverityWarn     = "warn"
    SeverityError    = "error"
    SeverityCritical = "critical"
)
```

**Simpler Design:**
- ❌ No complex struct definitions needed
- ✅ Use constants for event types và categories
- ✅ Type-safe trong code
- ✅ Easy to maintain và extend

---

### 4. Usage in Handlers

```go
// internal/app/handler/v1/oauth2/handler.go

func (h *Handler) LoginSubmit(c *gin.Context) {
    email := c.PostForm("email")
    password := c.PostForm("password")

    // Application log (debug/performance)
    logger.AppLogger.Info("Login attempt",
        zap.String("email", email),
        zap.String("ip", c.ClientIP()),
        zap.String("user_agent", c.Request.UserAgent()),
    )

    // Authenticate
    user, err := h.oauth2Service.AuthenticateUser(ctx, email, password)

    if err != nil {
        // Audit log: Failed login
        logger.AuditLogger.LogLoginFailed(ctx,
            email,
            err.Error(), // reason
            c.ClientIP(),
            c.Request.UserAgent(),
        )

        // Application log: Error
        logger.AppLogger.Error("Login failed",
            zap.String("email", email),
            zap.Error(err),
        )

        c.JSON(401, gin.H{"error": "Invalid credentials"})
        return
    }

    // Create session
    sessionID, _ := h.oauth2Service.CreateUserSession(ctx, user.ID, time.Hour)

    // Audit log: Successful login
    logger.AuditLogger.LogLoginSuccess(ctx,
        user.ID.String(),
        user.Email,
        c.ClientIP(),
        c.Request.UserAgent(),
        sessionID,
    )

    // Application log: Success
    logger.AppLogger.Info("Login successful",
        zap.String("user_id", user.ID.String()),
        zap.String("email", email),
        zap.String("session_id", sessionID),
    )

    // Set cookie and redirect
    c.SetCookie("session_id", sessionID, 3600, "/", "", false, true)
    c.Redirect(302, "/oauth2/auth?request_id="+requestID)
}
```

**Simpler API:**
- Pass parameters directly, không cần struct
- Type-safe và clear
- IDE autocomplete works better

---

## Promtail Configuration

### Update `promtail-config.yml`

Cấu hình Promtail để extract labels từ JSON logs:

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Docker logs
  - job_name: system
    docker_sd_configs:
      - host: unix:///var/run/docker.sock
        refresh_interval: 5s
    relabel_configs:
      - source_labels: ['__meta_docker_container_name']
        regex: '/(.*)'
        target_label: 'container'
    pipeline_stages:
      # Parse JSON logs
      - json:
          expressions:
            timestamp: timestamp
            level: level
            log_type: log_type
            event_category: event_category
            event_type: event_type
            severity: severity
            msg: msg

      # Extract labels (low cardinality only!)
      - labels:
          log_type:
          event_category:
          event_type:
          severity:
          level:

      # Set timestamp
      - timestamp:
          source: timestamp
          format: RFC3339

  # File-based logs (backup)
  - job_name: app_files
    static_configs:
      - targets:
          - localhost
        labels:
          job: oauth2_server
          __path__: /var/log/oauth2/app*.log
    pipeline_stages:
      - json:
          expressions:
            timestamp: timestamp
            level: level
            log_type: log_type
            event_category: event_category
            event_type: event_type
            severity: severity
      - labels:
          log_type:
          event_category:
          event_type:
          severity:
          level:
      - timestamp:
          source: timestamp
          format: RFC3339
```

**Key Points:**
- Extract labels từ JSON fields
- Chỉ promote low-cardinality fields thành labels
- High-cardinality data (user_id, ip, etc.) ở trong JSON body
- Automatic timestamp parsing

---

## Querying Audit Logs với LogQL

### LogQL Basics

LogQL syntax: `{label_selector} | line_filter | json_parser | filter_expression`

### Common Queries

#### 1. Find all login attempts for a user (filter trong JSON)
```logql
{log_type="audit", event_category="auth", event_type="login_success"}
| json
| user_id="00000000-0000-0000-0000-000000000001"
```

#### 2. Find failed login attempts in last 24h
```logql
{log_type="audit", event_category="auth", event_type="login_failed"} [24h]
```

#### 3. Find all consents granted to a client
```logql
{log_type="audit", event_category="authz", event_type="consent_granted"}
| json
| client_id="10000000-0000-0000-0000-000000000001"
```

#### 4. Count tokens issued today (aggregation)
```logql
sum(count_over_time({log_type="audit", event_category="token", event_type="access_token_issued"} [24h]))
```

Hoặc group by client:
```logql
sum by (client_id) (
  count_over_time({log_type="audit", event_category="token", event_type="access_token_issued"} [24h] | json)
)
```

#### 5. Find security events by severity
```logql
{log_type="audit", event_category="security", severity=~"error|critical"} [7d]
```

#### 6. User activity timeline (all events for a user)
```logql
{log_type="audit"}
| json
| user_id="00000000-0000-0000-0000-000000000001"
```

#### 7. Failed logins by email (detect brute force)
```logql
sum by (email) (
  count_over_time({log_type="audit", event_type="login_failed"} [15m] | json)
) > 5
```

#### 8. Token reuse detection events
```logql
{log_type="audit", event_category="security", event_type="token_reuse_detected", severity="critical"}
```

#### 9. All admin actions
```logql
{log_type="audit", event_category="admin"}
```

#### 10. Logs with specific IP address
```logql
{log_type="audit"}
| json
| ip_address="192.168.1.100"
```

### Advanced Queries

#### Rate of login failures (per minute)
```logql
rate({log_type="audit", event_type="login_failed"} [5m])
```

#### Top 10 users with most logins
```logql
topk(10,
  sum by (user_id) (
    count_over_time({log_type="audit", event_type="login_success"} [24h] | json)
  )
)
```

#### Logins from suspicious countries
```logql
{log_type="audit", event_category="auth", event_type="login_success"}
| json
| location_country!="VN"
```

### Grafana Dashboard Queries

#### Panel: Login Success Rate (over time)
```logql
sum(rate({log_type="audit", event_type="login_success"} [5m])) /
sum(rate({log_type="audit", event_category="auth"} [5m]))
```

#### Panel: Security Events by Severity
```logql
sum by (severity) (count_over_time({log_type="audit", event_category="security"} [24h]))
```

#### Panel: Token Issuance by Grant Type
```logql
sum by (grant_type) (
  count_over_time({log_type="audit", event_type="access_token_issued"} [1h] | json)
)
```

---

## Admin API for Querying

### Option 1: Use Grafana (Recommended)

**Simplest approach:** Sử dụng Grafana UI để query logs

- ✅ No custom API needed
- ✅ Rich visualization
- ✅ Built-in access control
- ✅ Shareable dashboards

### Option 2: Loki HTTP API (Programmatic Access)

Nếu cần API cho automation:

```go
// GET /admin/audit/auth?user_id=xxx&from=2024-11-01&to=2024-11-02
func (h *AdminHandler) GetAuthEvents(c *gin.Context) {
    userID := c.Query("user_id")
    from := c.Query("from") // RFC3339 timestamp
    to := c.Query("to")

    // Build LogQL query
    query := `{log_type="audit", event_category="auth"}`
    if userID != "" {
        query += ` | json | user_id="` + userID + `"`
    }

    // Call Loki API
    lokiURL := "http://loki:3100/loki/api/v1/query_range"
    params := url.Values{}
    params.Set("query", query)
    params.Set("start", from)
    params.Set("end", to)
    params.Set("limit", "100")

    resp, err := http.Get(lokiURL + "?" + params.Encode())
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to query Loki"})
        return
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    c.JSON(200, result)
}

// GET /admin/audit/security?severity=critical
func (h *AdminHandler) GetSecurityEvents(c *gin.Context) {
    severity := c.Query("severity")
    from := c.Query("from")
    to := c.Query("to")

    query := `{log_type="audit", event_category="security"`
    if severity != "" {
        query += `, severity="` + severity + `"`
    }
    query += `}`

    // Call Loki API (same pattern as above)
    // ...
}
```

**Loki Query API Endpoints:**
- `/loki/api/v1/query` - Instant query
- `/loki/api/v1/query_range` - Range query
- `/loki/api/v1/labels` - Get available labels
- `/loki/api/v1/label/<name>/values` - Get label values

**Documentation:** https://grafana.com/docs/loki/latest/api/

### Option 3: Grafana HTTP API (Best of Both Worlds)

Query Grafana programmatically với API tokens:

```bash
# Get datasource info
curl -H "Authorization: Bearer <token>" \
  http://localhost:5555/api/datasources

# Query Loki via Grafana
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"queries":[{"expr":"{log_type=\"audit\"}","refId":"A"}]}' \
  http://localhost:5555/api/ds/query
```

---

## Resource Usage Estimation

### All Logs (Loki)

Tất cả logs (application + audit) đều đi qua Loki:

| Metric | Value |
|--------|-------|
| **Log rate** | ~100 requests/sec × 800 bytes = 80 KB/sec |
| **Daily volume** | 80 KB/sec × 86400 sec = ~6.9 GB/day (raw) |
| **Loki compression** | ~700 MB/day (Loki chunks ~10:1 compression) |
| **90-day retention** | 700 MB × 90 = **~63 GB** |

### Breakdown by Category

| Log Type | Events/Day | Avg Size | Daily (Raw) | Daily (Compressed) |
|----------|-----------|----------|-------------|---------------------|
| Application | 8,640,000 | 400 bytes | ~3.5 GB | ~350 MB |
| Audit: Auth | 10,000 | 500 bytes | 5 MB | 0.5 MB |
| Audit: Authz | 5,000 | 600 bytes | 3 MB | 0.3 MB |
| Audit: Token | 20,000 | 400 bytes | 8 MB | 0.8 MB |
| Audit: Admin | 100 | 800 bytes | 80 KB | 8 KB |
| Audit: Security | 500 | 700 bytes | 350 KB | 35 KB |
| **Total** | **~8.6M** | - | **~6.9 GB/day** | **~700 MB/day** |

### Local File Backup (Optional)

| Metric | Value |
|--------|-------|
| **Rotation** | Daily |
| **Retention** | 7 days (local disk) |
| **Compressed** | ~700 MB × 7 = **~5 GB** |

**Total Storage (Loki + Local backup):** ~68 GB (90 days)

---

## Cost Analysis

### Infrastructure Cost (Monthly)

| Resource | Cost | Notes |
|----------|------|-------|
| **Disk space** | $6-10 | 70 GB × $0.10/GB (cheap HDD) |
| **Loki** | $0 | Already running in docker-compose |
| **Promtail** | $0 | Already running |
| **Grafana** | $0 | Already running |
| **CPU overhead** | <2% | Promtail + Loki (lightweight) |
| **Memory** | ~200 MB | Loki + Promtail + buffers |
| **Total** | **~$10/month** | Extremely cheap! |

**Comparisons:**
- MongoDB-based design: ~$5/month (but need to add MongoDB)
- ELK Stack: ~$100-500/month (dedicated servers, heavy resources)
- Commercial SaaS logging: $50-200/month (Datadog, Splunk, etc.)

**Winner:** Loki ✅ (cheap + already have infrastructure)

---

## Maintenance

### Daily Tasks (Automated)
- ✅ Log rotation (automatic via lumberjack for local files)
- ✅ Old log compression (automatic)
- ✅ Loki retention cleanup (automatic based on config)
- ✅ Promtail log collection (automatic)

### Weekly Tasks
- Check Loki disk space usage
- Review security events in Grafana
- Check Promtail/Loki health status
- Monitor label cardinality

### Monthly Tasks
- Review Loki retention policies (90 days default)
- Analyze log patterns in Grafana
- Review dashboard performance
- Update Grafana dashboards if needed

### Loki Config for Retention

Update `loki-config.yml`:

```yaml
limits_config:
  retention_period: 2160h  # 90 days

compactor:
  working_directory: /loki/compactor
  shared_store: filesystem
  compaction_interval: 10m
  retention_enabled: true
  retention_delete_delay: 2h
  retention_delete_worker_count: 150
```

---

## Migration Path (Future)

Nếu sau này cần scale up:

```
Current Setup (Loki)       →    Future Options
─────────────────────────       ────────────────────
Single Loki instance      →     Loki cluster (HA)
                          →     or
                          →     Grafana Cloud Logs
                          →     or
                          →     Export to object storage (S3)
                          →     or
                          →     Hybrid: Loki + Long-term DB
```

**Migration Paths:**

1. **Scale Loki horizontally:**
   - Multiple Loki instances với load balancer
   - S3/GCS backend for chunks
   - Separate read/write paths

2. **Export to Cloud:**
   - Grafana Cloud Logs (managed Loki)
   - AWS CloudWatch Logs
   - Google Cloud Logging

3. **Long-term Archive:**
   - Export old logs to PostgreSQL/ClickHouse
   - Keep hot data in Loki, cold data in DB
   - Best of both worlds

**Easy migration:** Logs vẫn là JSON format, có thể export và import vào bất kỳ system nào.

---

## Summary

### ✅ What You Get

1. **Unified logging system:**
   - All logs (application + audit) qua Loki
   - Structured JSON logs
   - Automatic collection via Promtail

2. **Low resource usage:**
   - ~68 GB storage (90 days retention)
   - <2% CPU overhead
   - ~$10/month cost

3. **Compliance-ready:**
   - Immutable audit trail
   - Configurable retention (90 days default, có thể lên 1 year+)
   - LogQL query capabilities

4. **Zero new infrastructure:**
   - Uses existing Loki + Promtail + Grafana stack
   - No MongoDB needed
   - Simpler architecture

5. **Built-in visualization:**
   - Grafana dashboards ready to use
   - Real-time log exploration
   - Alerting capabilities

6. **Future-proof:**
   - Easy to scale horizontally
   - Export to cloud (Grafana Cloud, S3)
   - Migrate to other systems (still JSON)

### 📋 Implementation Checklist

#### Phase 1: Logger Setup (Week 1)
- [ ] Setup Zap with dual output (stdout + file via lumberjack)
- [ ] Add `log_type`, `event_category`, `event_type`, `severity` fields
- [ ] Test JSON output format
- [ ] Verify logs appear in stdout

#### Phase 2: Promtail Configuration (Week 1)
- [ ] Update `promtail-config.yml` với JSON parsing
- [ ] Configure label extraction (low cardinality only!)
- [ ] Test Promtail collecting logs from Docker
- [ ] Verify logs arrive in Loki

#### Phase 3: Audit Logger Implementation (Week 2)
- [ ] Create `internal/platform/logger/audit.go` (Zap wrapper)
- [ ] Define event constants in `internal/domain/audit.go`
- [ ] Implement audit methods (LogLoginSuccess, LogTokenIssued, etc.)
- [ ] Add unit tests

#### Phase 4: Integration (Week 2-3)
- [ ] Add audit logging to login flow
- [ ] Add audit logging to token endpoint
- [ ] Add audit logging to consent flow
- [ ] Add audit logging to admin actions
- [ ] Add security event logging (token reuse, etc.)

#### Phase 5: Loki Configuration (Week 3)
- [ ] Update `loki-config.yml` với retention policy (90 days)
- [ ] Enable compactor for cleanup
- [ ] Configure storage limits
- [ ] Test retention cleanup

#### Phase 6: Grafana Dashboards (Week 3-4)
- [ ] Create dashboard: Authentication Events
- [ ] Create dashboard: Token Lifecycle
- [ ] Create dashboard: Security Events
- [ ] Create dashboard: Admin Actions
- [ ] Setup alerts (failed logins, security events)

#### Phase 7: Testing & Documentation (Week 4)
- [ ] Test all log flows end-to-end
- [ ] Verify LogQL queries work
- [ ] Document common queries
- [ ] Test retention cleanup
- [ ] Performance testing (ensure <5ms overhead)

### 🎯 Success Criteria

- [ ] Tất cả audit events được log đầy đủ
- [ ] Logs xuất hiện trong Grafana realtime
- [ ] LogQL queries return correct data
- [ ] Retention policy cleanup works
- [ ] No performance degradation (<5ms overhead)
- [ ] Label cardinality acceptable (<1000 unique values per label)
- [ ] Dashboards useful và dễ hiểu

---

**Recommended:** Start implementation theo order trên, test từng phase trước khi move to next.

**Timeline:** 3-4 tuần để complete full implementation với testing.
