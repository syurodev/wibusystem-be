# Checklist Triển khai OAuth2 Server

Tài liệu này theo dõi tiến độ triển khai các tính năng cho OAuth2 Authorization Server.

---

## ✅ Đã hoàn thành

### Phase 0: Thiết lập Nền tảng & Cấu hình
- [x] **Fosite Provider**: Khởi tạo và cấu hình Fosite provider.
  - [x] Tắt grant type `ResourceOwnerPasswordCredentials` (ROPC) không an toàn.
  - [x] Cấu hình đúng các strategy và dependency cho Fosite v0.49.0.
- [x] **Storage Layer**: Thiết lập kiến trúc lưu trữ kết hợp (Hybrid).
  - [x] **SQLStore**: Dùng cho dữ liệu bền vững.
    - [x] Implement `fosite.ClientManager` (quản lý client).
    - [x] Implement các phương thức JTI (`ClientAssertionJWTValid`, `SetClientAssertionJWT`).
  - [x] **RedisStore**: Dùng cho dữ liệu tạm thời.
    - [x] Implement `AuthorizeCodeStorage` (authorization code).
    - [x] Implement `PKCERequestStorage` (PKCE sessions).
    - [x] Implement `AccessTokenStorage` (access token).
    - [x] Implement `OpenIDConnectRequestStorage` (OIDC sessions).
    - [x] Implement `TokenRevocationStorage` (danh sách token bị thu hồi).
  - [x] **HybridStore**: Kết hợp SQL và Redis.
    - [x] Implement chiến lược **Cache-Aside** cho `RefreshTokenStorage` (lưu ở SQL, cache ở Redis) để đảm bảo an toàn và hiệu năng.
- [x] **Repository Pattern**: Tách biệt logic truy vấn khỏi lớp storage.
  - [x] Tạo `OAuth2ClientRepository` để quản lý `oauth2_clients`.
  - [x] Tạo `OAuth2SessionRepository` để quản lý `oauth2_sessions`.
  - [x] Tái cấu trúc để sử dụng `pgx` thuần, loại bỏ `sqlx` và `lib/pq`.
- [x] **Cấu hình**: Thêm các cấu hình cần thiết cho OAuth2 (`Issuer`, `PrivateKey`, `HMACSecret`).
- [x] **Endpoints Cơ bản**:
  - [x] Implement OpenID Connect Discovery Endpoint (`/.well-known/openid-configuration`).
  - [x] Implement JWKS Endpoint (`/.well-known/jwks.json`).
- [x] **Refactoring & Code Quality**:
  - [x] Sử dụng `enum` (hằng số) cho các giá trị scope.

- [x] **User Domain Model & Repository**: Tạo User entity và UserRepository.
  - [x] Tạo `domain.User` với đầy đủ fields (ID, Email, PasswordHash, FullName, etc.).
  - [x] Tạo `UserRepository` interface và implementation với pgx.
  - [x] Các phương thức: `GetByID`, `GetByEmail`, `Create`, `Update`, `UpdateLastLogin`.

### Phase 1: Luồng Ủy quyền Cốt lõi (Authorization Code Flow)
- [x] **Authorization Endpoint (`/oauth2/auth`)**
  - [x] Tạo handler `Authorize`.
  - [x] Xử lý việc xác thực người dùng (cần tạo trang/logic đăng nhập).
  - [x] Xử lý việc người dùng chấp thuận (cần tạo trang/logic consent).
  - [x] Tích hợp với `provider.NewAuthorizeRequest` và `provider.NewAuthorizeResponse` của Fosite.
  - [x] Tạo Login Page với HTML template đẹp.
  - [x] Tạo Consent Page với HTML template đẹp.
  - [x] Implement flow: check authentication → login → check consent → consent → finalize.
- [x] **Token Endpoint (`/oauth2/token`)**
  - [x] Tạo handler `Token`.
  - [x] Xử lý grant type `authorization_code`.
  - [x] Xử lý grant type `refresh_token`.
  - [x] Xử lý grant type `client_credentials`.
  - [x] Tích hợp với `provider.NewAccessRequest` và `provider.NewAccessResponse` của Fosite.
- [x] **Routes Registration**: Đăng ký tất cả routes mới vào router.

### Phase 1: Production-ready Improvements
- [x] **Session Management**:
  - [x] Implement Redis-based session storage thay vì temporary mock.
  - [x] Tạo `SessionRepository` để quản lý user sessions.
  - [x] Secure session cookies (HttpOnly, Secure, SameSite).
  - **Files**: `internal/domain/session.go`, `internal/pkg/repository/session_repo.go`
- [x] **Authorization Request Storage**:
  - [x] Lưu authorization requests vào Redis với TTL 10 phút.
  - [x] Implement methods: `SaveAuthRequest`, `GetAuthRequest`, `DeleteAuthRequest`, `SaveAuthRequestWithUserID`.
  - **Files**: `internal/domain/auth_request.go`, `internal/pkg/repository/auth_request_repo.go`
- [x] **Password Verification**:
  - [x] Implement bcrypt password verification trong `LoginSubmit`.
  - **Files**: `pkg/util/crypto/password.go`
- [x] **Consent Management**:
  - [x] Tạo bảng `oauth2_consents` để lưu user consents.
  - [x] Implement `ConsentRepository` với methods: `GetActiveConsent`, `CreateConsent`, `RevokeConsent`, `GetUserConsents`.
  - [x] Skip consent page nếu user đã consent trước đó (logic đã implement).
  - [x] Save consent khi user click "Allow".
  - **Files**: `migrations/000007_create_oauth2_consents_table.up.sql`, `internal/domain/consent.go`, `internal/pkg/repository/consent_repo.go`
- [x] **Error Handling**:
  - [x] Xử lý các error cases trong authorization flow.
  - [x] Redirect về client với proper OAuth2 error responses khi user deny consent.
- [x] **Utilities**:
  - [x] Implement secure random string generation cho session IDs.
  - **Files**: `pkg/util/random/string.go`

### Phase 2: UserInfo Endpoint (Production-ready)
- [x] **UserInfo Endpoint (`/oauth2/userinfo`)** ✅ **HOÀN THÀNH**
  - [x] **Loại bỏ `MockStore`**: Đã xóa hoàn toàn file `store.go` và `MockStore` implementation.
  - [x] **Sử dụng `UserRepository`**: UserInfo endpoint giờ lấy dữ liệu từ database thông qua `UserRepository`.
  - [x] **Mapping với `domain.User`**:
    - Parse `userID` string sang UUID với validation.
    - Map fields: `user.FullName`, `user.AvatarURL`, `user.EmailVerified`.
    - Handle nil pointer cho `AvatarURL`.
  - [x] **Error Handling**: Xử lý invalid UUID và user not found cases.
  - **Files**: `internal/app/handler/v1/oauth2/handler.go` (lines 76-120), `internal/app/router/router.go`
  - **Verified**: Code compiles successfully và ready for production.

### Phase 4: OAuth2 Client Management API
- [x] **Admin API cho quản lý OAuth2 Clients**
  - [x] **Domain Model Updates**:
    - [x] Thêm fields vào `OAuth2Client`: `TokenEndpointAuth`, `ClientURI`, `LogoURL`, `Active`, `CreatedAt`, `UpdatedAt`.
    - [x] Thêm db struct tags cho tất cả fields để mapping đúng với database columns.
    - [x] Extend `OAuth2ClientRepository` interface với CRUD methods.
    - **Files**: `internal/domain/oauth2_client.go`
  - [x] **Repository Implementation**:
    - [x] Implement `GetByID` - Lấy client theo ID với đầy đủ fields.
    - [x] Implement `Create` - Tạo client mới với UUID v7.
    - [x] Implement `Update` - Cập nhật thông tin client.
    - [x] Implement `Delete` - Xóa client.
    - [x] Implement `List` - Liệt kê clients với dynamic filtering (tenant_id, active) và pagination.
    - **Files**: `internal/pkg/repository/oauth2_client_repo.go`
  - [x] **DTOs**:
    - [x] `CreateClientRequest` - Validate client creation với Gin binding tags.
    - [x] `UpdateClientRequest` - Optional fields cho partial updates.
    - [x] `ClientResponse` - Response DTO với client_secret chỉ hiện khi tạo mới/regenerate.
    - [x] `ClientListResponse` - Pagination metadata (total, limit, offset).
    - **Files**: `internal/app/handler/v1/oauth2_admin/dto.go`
  - [x] **Handler Implementation**:
    - [x] `CreateClient` - POST `/admin/oauth2/clients` - Tạo client với secret generation và bcrypt hashing.
    - [x] `GetClient` - GET `/admin/oauth2/clients/:id` - Lấy thông tin client.
    - [x] `ListClients` - GET `/admin/oauth2/clients` - Liệt kê với query params (tenant_id, active, limit, offset).
    - [x] `UpdateClient` - PUT `/admin/oauth2/clients/:id` - Cập nhật client (partial update supported).
    - [x] `DeleteClient` - DELETE `/admin/oauth2/clients/:id` - Xóa client.
    - [x] `RegenerateSecret` - POST `/admin/oauth2/clients/:id/regenerate-secret` - Tạo lại secret (chỉ cho confidential clients).
    - [x] Validation cho grant_types và response_types.
    - [x] Sử dụng `response.Success()` và `response.Error()` với i18n support.
    - **Files**: `internal/app/handler/v1/oauth2_admin/handler.go`
  - [x] **I18n Support**:
    - [x] Tạo domain `oauth2` trong i18n với message keys cho client management.
    - [x] Messages: `client.created`, `client.updated`, `client.deleted`, `client.retrieved`, `client.listed`, `client.secret_regenerated`.
    - [x] Error messages: `client.not_found`, `client.create_failed`, `client.update_failed`, etc.
    - [x] Validation messages: `validation.invalid_grant_type`, `validation.invalid_response_type`, `validation.invalid_client_id`, `validation.invalid_tenant_id`.
    - [x] Support tiếng Anh và tiếng Việt.
    - **Files**: `internal/platform/i18n/locales/oauth2/en.json`, `internal/platform/i18n/locales/oauth2/vi.json`, `internal/platform/i18n/i18n.go`
  - [x] **Router Registration**:
    - [x] Tạo router file cho oauth2_admin package.
    - [x] Đăng ký tất cả 6 endpoints vào `/api/v1/admin/oauth2/*`.
    - [x] Integrate vào main router với dependency injection.
    - **Files**: `internal/app/handler/v1/oauth2_admin/router.go`, `internal/app/router/router.go`
  - [x] **Technical Notes**:
    - [x] Sử dụng UUID v7 (time-ordered) cho client IDs.
    - [x] Sử dụng `github.com/gofrs/uuid/v5` package thay vì `google/uuid`.
    - [x] Bcrypt cost factor 12 cho client secret hashing.
    - [x] Secure random generation (32 characters) cho client secrets.
    - [x] Public clients không có secret và không thể regenerate secret.

### Phase 2: Token Management Endpoints (Production-ready)
- [x] **Token Revocation Endpoint (`/oauth2/revoke`)** ✅ **HOÀN THÀNH**
  - [x] Tạo handler `Revoke` theo RFC 7009.
  - [x] Tích hợp với `provider.NewRevocationRequest`.
  - [x] Xác thực client credentials tự động bởi Fosite.
  - [x] Trả về 200 OK theo RFC 7009 spec.
  - [x] Đăng ký route `POST /oauth2/revoke`.
  - **Files**: `internal/app/handler/v1/oauth2/handler.go` (lines 718-735), `internal/app/handler/v1/oauth2/router.go`
  - **Verified**: Code compiles successfully.

- [x] **Token Introspection Endpoint (`/oauth2/introspect`)** ✅ **HOÀN THÀNH**
  - [x] Tạo handler `Introspect` theo RFC 7662.
  - [x] Xác thực client credentials (chỉ authorized clients được introspect).
  - [x] Sử dụng `provider.IntrospectToken` của Fosite.
  - [x] Trả về token metadata: active, scope, client_id, exp, iat, sub, aud.
  - [x] Handle invalid tokens (trả về `{"active": false}`).
  - [x] Đăng ký route `POST /oauth2/introspect`.
  - **Files**: `internal/app/handler/v1/oauth2/handler.go` (lines 737-781), `internal/app/handler/v1/oauth2/router.go`
  - **Verified**: Code compiles successfully.

### Phase 3: Authentication & Authorization (Production-ready)
- [x] **Authentication Middleware** ✅ **HOÀN THÀNH**
  - [x] `RequireAuth()` middleware - Xác thực Bearer Token.
  - [x] `RequireScope()` middleware - Kiểm tra specific scope.
  - [x] `RequireAnyScope()` middleware - Kiểm tra any of multiple scopes.
  - [x] Helper functions: `GetUserID()`, `GetClientID()`, `GetScopes()`.
  - [x] Sử dụng token introspection để validate tokens.
  - [x] OAuth2-compliant error responses (401, 403).
  - [x] Context injection cho handlers (user_id, client_id, scopes).
  - [x] Documentation đầy đủ với usage examples.
  - **Files**: `internal/app/middleware/auth.go`, `docs/MIDDLEWARE_USAGE_GUIDE.md`
  - **Verified**: Code compiles successfully.


## 🚧 Đang Triển Khai

### Phase 5: Logging & Audit Trail (CRITICAL - Security & Compliance)

**Reference:** `docs/LOGGING_AUDIT_DESIGN.md`

**Architecture Change:** ✅ Using **Loki** (already in docker-compose) instead of MongoDB

#### 5.1. Infrastructure Setup (DONE ✅)
- [x] **Loki Stack Already Available**
  - [x] Loki container (grafana/loki:3.5) - Already running
  - [x] Promtail container (grafana/promtail:3.5) - Already running
  - [x] Grafana container (grafana/grafana:12.3.0) - Already running
  - [x] No new containers needed!
  - **Files**: `docker-compose.yml` (no changes needed)

- [ ] **Loki Configuration**
  - [ ] Update `loki-config.yml` với retention policy (90 days)
  - [ ] Enable compactor for automatic cleanup
  - [ ] Configure storage limits
  - [ ] No connection config needed in app (logs go via stdout → Promtail)
  - **Files**: `loki-config.yml`

- [ ] **Promtail Configuration**
  - [ ] Update `promtail-config.yml` để parse JSON logs
  - [ ] Configure label extraction (event_category, event_type, severity)
  - [ ] Ensure low cardinality labels only
  - **Files**: `promtail-config.yml`

#### 5.2. Logging Infrastructure (Unified: Application + Audit)
- [ ] **Zap Logger Setup (Dual Output)**
  - [ ] Install dependencies: `go.uber.org/zap`, `gopkg.in/natefinch/lumberjack.v2`
  - [ ] Configure structured JSON logging
  - [ ] **Dual output:**
    - Primary: stdout (cho Promtail collection)
    - Backup: file rotation với lumberjack (7 days local)
  - [ ] Add fields: `log_type`, `event_category`, `event_type`, `severity`
  - [ ] Log levels: Info, Warn, Error, Debug
  - [ ] ISO8601 timestamps
  - **Files**: `internal/platform/logger/logger.go`

- [ ] **Request Logging Middleware**
  - [ ] Log all incoming HTTP requests
  - [ ] Fields: method, path, status, duration, IP, user_agent, request_id
  - [ ] Exclude healthcheck endpoints
  - **Files**: `internal/platform/logger/middleware.go`

#### 5.3. Event Types & Constants
- [ ] **Event Constants** (Simpler than structs)
  - [ ] Define event categories: `auth`, `authz`, `token`, `admin`, `security`
  - [ ] Define event types per category
  - [ ] Define severity levels: `info`, `warn`, `error`, `critical`
  - [ ] No complex struct definitions needed (just string constants)
  - **Files**: `internal/domain/audit.go`

#### 5.4. Audit Logger (Zap Wrapper)
- [ ] **AuditLogger Implementation**
  - [ ] Create wrapper around Zap logger
  - [ ] Add default field: `log_type="audit"`
  - [ ] High-level methods cho từng event type:
    - **Authentication:**
      - `LogLoginSuccess(ctx, userID, email, ip, userAgent, sessionID)`
      - `LogLoginFailed(ctx, email, reason, ip, userAgent)`
      - `LogLogout(ctx, userID, sessionID)`
      - `LogPasswordChanged(ctx, userID, ip)`
      - `LogEmailVerified(ctx, userID, email)`
    - **Authorization:**
      - `LogConsentGranted(ctx, userID, clientID, clientName, scopes, ip, userAgent)`
      - `LogConsentDenied(ctx, userID, clientID, clientName, ip, userAgent)`
      - `LogConsentRevoked(ctx, userID, clientID, revokedBy)`
      - `LogAuthCodeIssued(ctx, userID, clientID, code, ip)`
    - **Tokens:**
      - `LogTokenIssued(ctx, tokenType, userID, clientID, grantType, scopes, ip)`
      - `LogTokenRevoked(ctx, tokenType, signature, userID, clientID, reason)`
      - `LogTokenRotated(ctx, oldSig, newSig, userID, clientID)`
    - **Admin:**
      - `LogClientCreated(ctx, adminID, clientID, clientName)`
      - `LogClientUpdated(ctx, adminID, clientID, changes)`
      - `LogClientDeleted(ctx, adminID, clientID)`
      - `LogSecretRotated(ctx, adminID, clientID)`
    - **Security:**
      - `LogTokenReuse(ctx, tokenSig, userID, clientID, ip)`
      - `LogRateLimitExceeded(ctx, identifier, endpoint, ip)`
      - `LogInvalidClientCredentials(ctx, clientID, ip)`
      - `LogSuspiciousActivity(ctx, userID, description, ip, severity)`
  - [ ] All methods use Zap structured logging
  - [ ] No async needed (Zap is fast, Promtail handles buffering)
  - **Files**: `internal/platform/logger/audit.go`

#### 5.5. Integration với OAuth2 Handlers
- [ ] **Login Flow Logging**
  - [ ] `LoginSubmit`: Log success/failure
  - [ ] Include: email, IP, user_agent, session_id
  - [ ] Failed login count tracking (security)
  - **Files**: `internal/app/handler/v1/oauth2/handler.go`

- [ ] **Authorization Flow Logging**
  - [ ] `Authorize`: Log authorization request
  - [ ] `ConsentSubmit`: Log consent granted/denied
  - [ ] Include: client_id, scopes, user decision
  - **Files**: `internal/app/handler/v1/oauth2/handler.go`

- [ ] **Token Endpoint Logging**
  - [ ] `Token`: Log all token issuance
  - [ ] Different log cho mỗi grant_type
  - [ ] Include: grant_type, client_id, scopes, IP
  - [ ] Log token refresh và rotation
  - **Files**: `internal/app/handler/v1/oauth2/handler.go`

- [ ] **Token Management Logging**
  - [ ] `Revoke`: Log token revocation
  - [ ] Include: token_type, revoked_by, reason
  - **Files**: `internal/app/handler/v1/oauth2/handler.go`

- [ ] **Admin Actions Logging**
  - [ ] All CRUD operations on clients
  - [ ] Log admin user ID và changes made
  - [ ] Include old values và new values (audit trail)
  - **Files**: `internal/app/handler/v1/oauth2_admin/handler.go`

#### 5.6. Security Event Detection
- [ ] **Token Reuse Detection**
  - [ ] Detect khi refresh token đã rotated được dùng lại
  - [ ] Log với severity: CRITICAL
  - [ ] Auto-revoke entire token family
  - [ ] Send alert (optional)
  - **Files**: `internal/oauth2/storage/hybrid_store.go`

- [ ] **Brute Force Detection** (Future - Phase 6)
  - [ ] Track failed login attempts per email + IP
  - [ ] Threshold: 5 failures in 15 minutes
  - [ ] Log security event khi threshold exceeded
  - **Files**: `internal/app/middleware/rate_limiter.go`

#### 5.7. Querying Audit Logs (Grafana + Loki API)
- [ ] **Option 1: Use Grafana Explore (Recommended)**
  - [ ] Access Grafana at http://localhost:5555
  - [ ] Use Explore tab với Loki datasource
  - [ ] Write LogQL queries directly
  - [ ] Create dashboards for common queries
  - [ ] No custom API needed!

- [ ] **Option 2: Loki HTTP API (Optional - for automation)**
  - [ ] `GET /admin/audit/query` - Generic query endpoint
    - Proxy requests to Loki HTTP API
    - Build LogQL queries from params
    - Query params: `query`, `from`, `to`, `limit`
  - [ ] Loki API endpoints:
    - `/loki/api/v1/query_range` - Range queries
    - `/loki/api/v1/query` - Instant queries
    - `/loki/api/v1/labels` - Available labels
  - [ ] Response format: Loki JSON format
  - **Files**: `internal/app/handler/v1/audit/handler.go` (optional)

- [ ] **Common LogQL Queries Documentation**
  - [ ] Document query patterns in `docs/LOGGING_AUDIT_DESIGN.md`
  - [ ] Authentication events queries
  - [ ] Token events queries
  - [ ] Security events queries
  - [ ] Admin action queries

#### 5.8. Grafana Dashboards
- [ ] **Create Dashboard: Authentication Events**
  - [ ] Login success/failure rate over time
  - [ ] Failed logins by email (brute force detection)
  - [ ] Login attempts by country/location
  - [ ] Top users by login frequency

- [ ] **Create Dashboard: Token Lifecycle**
  - [ ] Token issuance rate by grant type
  - [ ] Token revocations over time
  - [ ] Refresh token rotation events
  - [ ] Token introspection requests

- [ ] **Create Dashboard: Security Events**
  - [ ] Security events by severity
  - [ ] Token reuse detection alerts
  - [ ] Rate limit exceeded events
  - [ ] Invalid credentials attempts

- [ ] **Create Dashboard: Admin Actions**
  - [ ] Client CRUD operations
  - [ ] Admin user activity
  - [ ] Secret rotation events

- [ ] **Setup Alerts**
  - [ ] Alert on critical security events
  - [ ] Alert on high failed login rate
  - [ ] Alert on token reuse detection
  - [ ] Alert on Loki/Promtail down

#### 5.9. Testing & Verification
- [ ] **Unit Tests**
  - [ ] Test AuditLogger methods
  - [ ] Test log field extraction
  - [ ] Mock Zap logger for testing
  - **Files**: `internal/platform/logger/*_test.go`

- [ ] **Integration Tests**
  - [ ] Test logs appear in Loki
  - [ ] Test LogQL queries return correct data
  - [ ] Test Promtail parsing
  - [ ] Test label extraction
  - **Files**: `internal/platform/logger/integration_test.go`

- [ ] **Manual Testing**
  - [ ] Login flow → verify in Grafana Explore
  - [ ] Token issuance → verify token_events
  - [ ] Admin actions → verify admin_events
  - [ ] Security violations → verify security_events
  - [ ] Run LogQL queries → verify results

- [ ] **Performance Testing**
  - [ ] Measure logging overhead (<5ms target)
  - [ ] Test under load (1000 req/sec)
  - [ ] Check label cardinality
  - [ ] Monitor Loki resource usage

#### 5.10. Documentation
- [x] **Update LOGGING_AUDIT_DESIGN.md** ✅
  - [x] Updated to Loki architecture
  - [x] Added LogQL query examples
  - [x] Added Promtail configuration
  - [x] Updated retention policy (90 days)

- [ ] **Create GRAFANA_DASHBOARD_GUIDE.md**
  - [ ] How to create dashboards
  - [ ] Common LogQL query patterns
  - [ ] Alert configuration
  - [ ] Best practices

#### 5.11. Deployment Preparation
- [ ] **Loki Configuration**
  - [ ] Update `loki-config.yml` với production settings
  - [ ] Configure retention (90 days)
  - [ ] Enable compactor
  - [ ] Set storage limits
  - **Files**: `loki-config.yml`

- [ ] **Promtail Configuration**
  - [ ] Update `promtail-config.yml` với final settings
  - [ ] Verify JSON parsing rules
  - [ ] Verify label extraction
  - **Files**: `promtail-config.yml`

- [ ] **Monitoring Setup**
  - [ ] Loki metrics trong Grafana
  - [ ] Promtail metrics
  - [ ] Log ingestion rate
  - [ ] Disk usage tracking
  - [ ] Alert on service down

---

## 📋 Implementation Order (Recommended)

**Week 1: Infrastructure & Logger Setup**
1. ✅ Verify Loki/Promtail/Grafana running (already done!)
2. Configure Loki retention policy (5.1)
3. Configure Promtail JSON parsing (5.1)
4. Setup Zap logger với dual output (5.2)
5. Test logs appear in Grafana Explore

**Week 2: Audit Logger & Event Types**
6. Define event constants (5.3)
7. Implement AuditLogger wrapper (5.4)
8. Request logging middleware (5.2)
9. Unit tests for logger
10. Verify structured logs in Loki

**Week 3: Integration**
11. Login flow logging (5.5)
12. Token endpoint logging (5.5)
13. Consent flow logging (5.5)
14. Admin actions logging (5.5)
15. Security event detection (5.6)

**Week 4: Dashboards, Testing & Docs**
16. Create Grafana dashboards (5.8)
17. Setup alerts (5.8)
18. Integration testing (5.9)
19. Performance testing (5.9)
20. Documentation (5.10)

---

## ✅ Definition of Done

**Logging:**
- [ ] Tất cả audit events được log đầy đủ (auth, authz, token, admin, security)
- [ ] Logs xuất hiện trong Grafana realtime
- [ ] JSON format với đúng fields: log_type, event_category, event_type, severity
- [ ] Structured logging với Zap working correctly

**Loki & Promtail:**
- [ ] Loki retention policy configured (90 days)
- [ ] Promtail parsing JSON logs correctly
- [ ] Labels extracted correctly (low cardinality verified)
- [ ] Logs queryable với LogQL

**Grafana:**
- [ ] Dashboards created cho mỗi event category
- [ ] Alerts configured cho critical events
- [ ] LogQL queries work correctly
- [ ] Performance acceptable (query time <2s)

**Testing:**
- [ ] Unit tests pass cho AuditLogger
- [ ] Integration tests pass (logs → Loki → Grafana)
- [ ] Manual testing completed
- [ ] Performance testing: <5ms logging overhead

**Documentation:**
- [x] LOGGING_AUDIT_DESIGN.md updated với Loki architecture ✅
- [ ] Common LogQL queries documented
- [ ] Grafana dashboard guide created
- [ ] Deployment guide updated

**Performance & Quality:**
- [ ] No performance degradation (<5ms overhead per request)
- [ ] Label cardinality acceptable (<1000 unique values per label)
- [ ] Loki disk usage reasonable (~700MB/day compressed)
- [ ] No errors in Promtail/Loki logs


---

## ⏳ Tùy chọn (Optional Features)

- [ ] **Dynamic Client Registration (`/oauth2/register`)**: Cho phép clients tự đăng ký động.
- [ ] **Pushed Authorization Requests (PAR) (`/oauth2/par`)**: Bảo mật cao hơn cho authorization requests.
