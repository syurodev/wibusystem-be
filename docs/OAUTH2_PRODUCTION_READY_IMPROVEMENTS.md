# OAuth2 Production-Ready Improvements

Tài liệu này mô tả các cải tiến đã được thực hiện để làm cho OAuth2 Authorization Server sẵn sàng cho production.

---

## 📋 Tổng quan

Các cải tiến này chuyển đổi implementation từ **prototype/demo** sang **production-ready** với:
- ✅ Bảo mật tốt hơn (bcrypt, secure sessions)
- ✅ Persistence đúng đắn (Redis cho temporary data, PostgreSQL cho permanent data)
- ✅ Error handling tốt hơn
- ✅ Consent management đầy đủ

---

## 1️⃣ Session Management (Redis-based)

### **Vấn đề trước đây:**
- Session được lưu tạm thời bằng cách return `userID` làm sessionID (NOT SECURE)
- Không có TTL, không có invalidation
- Không có Redis integration

### **Giải pháp:**

#### **Domain Model** (`internal/domain/session.go`)
```go
type SessionRepository interface {
    CreateSession(ctx context.Context, sessionID string, userID string, ttl time.Duration) error
    GetSession(ctx context.Context, sessionID string) (string, error)
    DeleteSession(ctx context.Context, sessionID string) error
    RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error
}
```

#### **Implementation** (`internal/pkg/repository/session_repo.go`)
- Sử dụng Redis với key pattern: `session:{sessionID}`
- TTL: 1 hour (configurable)
- Session ID: Secure random 32-byte string (base64 encoded)

#### **Benefits:**
- ✅ Secure random session IDs
- ✅ Automatic expiration với Redis TTL
- ✅ Fast lookup (O(1) trong Redis)
- ✅ Easy invalidation (logout)

---

## 2️⃣ Authorization Request Storage (Redis-based)

### **Vấn đề trước đây:**
- Authorization requests không được lưu trữ giữa các bước (login → consent → finalize)
- Không thể resume OAuth2 flow sau khi user đăng nhập

### **Giải pháp:**

#### **Domain Model** (`internal/domain/auth_request.go`)
```go
type AuthRequestRepository interface {
    SaveAuthRequest(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, ttl time.Duration) error
    GetAuthRequest(ctx context.Context, requestID string) (fosite.AuthorizeRequester, error)
    DeleteAuthRequest(ctx context.Context, requestID string) error
    SaveAuthRequestWithUserID(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, userID string, ttl time.Duration) error
    GetUserIDFromAuthRequest(ctx context.Context, requestID string) (string, error)
}
```

#### **Implementation** (`internal/pkg/repository/auth_request_repo.go`)
- Redis storage với key pattern: 
  - `auth_request:{requestID}` - Authorization request data (JSON)
  - `auth_request_user:{requestID}` - UserID mapping
- TTL: 10 minutes
- JSON serialization của `fosite.AuthorizeRequester`

#### **Benefits:**
- ✅ Stateful OAuth2 flow hoạt động chính xác
- ✅ Automatic cleanup sau 10 phút (không còn orphaned requests)
- ✅ Hỗ trợ multi-step flow (auth → login → consent → finalize)

---

## 3️⃣ Password Verification (bcrypt)

### **Vấn đề trước đây:**
- Password verification bị comment out
- Không có security cho authentication

### **Giải pháp:**

#### **Utility** (`pkg/util/crypto/password.go`)
```go
func HashPassword(password string) (string, error)
func VerifyPassword(hashedPassword, password string) bool
```

#### **Implementation:**
- Sử dụng `golang.org/x/crypto/bcrypt`
- Cost factor: 12 (balanced security/performance)
- Constant-time comparison

#### **Usage trong LoginSubmit:**
```go
if !crypto.VerifyPassword(user.PasswordHash, password) {
    return error
}
```

#### **Benefits:**
- ✅ Industry-standard password hashing
- ✅ Protection against rainbow table attacks
- ✅ Configurable cost factor
- ✅ Timing attack resistant

---

## 4️⃣ Consent Management (PostgreSQL)

### **Vấn đề trước đây:**
- Consent luôn được hỏi lại mỗi lần (bad UX)
- Không có persistent consent storage
- Không có revocation mechanism

### **Giải pháp:**

#### **Database Migration** (`migrations/000007_create_oauth2_consents_table.up.sql`)
```sql
CREATE TABLE identify.oauth2_consents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    client_id UUID NOT NULL,
    granted_scopes TEXT[],
    revoked BOOLEAN DEFAULT FALSE,
    granted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    consent_method VARCHAR(50), -- explicit, implicit, remembered
    ...
    UNIQUE (user_id, client_id)
);
```

**Functions:**
- `cleanup_expired_consents()` - Cleanup job
- `revoke_consent(user_id, client_id)` - Single revocation
- `revoke_all_user_consents(user_id)` - Global revocation
- `get_active_consent(user_id, client_id)` - Check consent

#### **Domain Model** (`internal/domain/consent.go`)
```go
type OAuth2Consent struct {
    ID            uuid.UUID
    UserID        uuid.UUID
    ClientID      uuid.UUID
    GrantedScopes []string
    Revoked       bool
    GrantedAt     time.Time
    ExpiresAt     *time.Time
    ConsentMethod ConsentMethod
}

type ConsentRepository interface {
    GetActiveConsent(ctx, userID, clientID) (*OAuth2Consent, error)
    CreateConsent(ctx, consent) error
    RevokeConsent(ctx, userID, clientID) error
    ...
}
```

#### **Implementation** (`internal/pkg/repository/consent_repo.go`)
- PostgreSQL storage
- Upsert logic (ON CONFLICT DO UPDATE)
- Query optimization với indexes

#### **Flow Integration:**
1. **Check consent** trước khi hiển thị consent page
2. **Skip consent page** nếu đã có active consent
3. **Save consent** khi user click "Allow"
4. **Revoke support** cho user settings

#### **Benefits:**
- ✅ Better UX: không hỏi lại consent mỗi lần
- ✅ Audit trail: track khi nào consent được cấp
- ✅ Revocation: user có thể thu hồi quyền
- ✅ Scopes tracking: biết chính xác scopes nào đã được grant
- ✅ Expiration support: consent có thể hết hạn

---

## 5️⃣ Secure Random String Generation

### **Utility** (`pkg/util/random/string.go`)
```go
func GenerateRandomString(n int) (string, error)
func GenerateSessionID() (string, error)
```

#### **Implementation:**
- Sử dụng `crypto/rand` (cryptographically secure)
- Base64 URL encoding
- 32 bytes default cho session IDs

#### **Benefits:**
- ✅ Cryptographically secure randomness
- ✅ Không predictable
- ✅ URL-safe encoding

---

## 6️⃣ Error Handling Improvements

### **Improvements:**

1. **OAuth2 Error Responses:**
   - Sử dụng `writeOAuth2Error()` helper
   - Proper RFC 6749 error format
   - Redirect về client với error params

2. **User Deny Consent:**
   ```go
   redirectURI := ar.GetRedirectURI().String()
   state := ar.GetState()
   errorURL := redirectURI + "?error=access_denied&error_description=User+denied+consent&state=" + state
   c.Redirect(http.StatusFound, errorURL)
   ```

3. **Session Expiration:**
   - Clear error messages
   - HTTP 401 Unauthorized
   - Proper JSON responses

#### **Benefits:**
- ✅ OAuth2 spec compliant
- ✅ Better client integration
- ✅ Clear error messages cho debugging

---

## 7️⃣ Handler Updates

### **Handler Dependencies** (`internal/app/handler/v1/oauth2/handler.go`)
```go
type Handler struct {
    config          *configs.OAuthConfig
    provider        fosite.OAuth2Provider
    store           Store
    sessionRepo     domain.SessionRepository          // NEW
    authRequestRepo domain.AuthRequestRepository      // NEW
    consentRepo     domain.ConsentRepository          // NEW
    userRepo        domain.UserRepository             // NEW
}
```

### **Updated Methods:**
- ✅ `checkUserAuthentication()` - Sử dụng sessionRepo
- ✅ `checkUserConsent()` - Sử dụng consentRepo
- ✅ `redirectToLogin()` - Lưu auth request vào Redis
- ✅ `redirectToConsent()` - Lưu auth request + userID
- ✅ `LoginSubmit()` - Bcrypt verification + session creation
- ✅ `createUserSession()` - Generate secure session ID + Redis storage
- ✅ `ConsentSubmit()` - Save consent + proper error handling

---

## 8️⃣ Router Integration

### **Router Updates** (`internal/app/router/router.go`)
```go
// Initialize all repositories
userRepo := repository.NewUserRepository(db.Pool)
sessionRepo := repository.NewSessionRepository(rdb)
authRequestRepo := repository.NewAuthRequestRepository(rdb, oauth2ClientRepo)
consentRepo := repository.NewConsentRepository(db.Pool)

// Inject into OAuth2 handler
oauth2Handler := oauth2_handler.NewHandler(
    &cfg.OAuth2,
    oauth2Provider,
    mockStore,
    sessionRepo,
    authRequestRepo,
    consentRepo,
    userRepo,
)
```

---

## 📊 Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    OAuth2 Handler                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  checkUserAuthentication()  → SessionRepo (Redis)   │  │
│  │  checkUserConsent()          → ConsentRepo (PG)     │  │
│  │  redirectToLogin()           → AuthRequestRepo (Redis)│ │
│  │  LoginSubmit()               → UserRepo (PG) + bcrypt│ │
│  │  ConsentSubmit()             → ConsentRepo (PG)     │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Redis     │     │ PostgreSQL  │     │   Fosite    │
│             │     │             │     │  Provider   │
│ - Sessions  │     │ - Users     │     │             │
│ - AuthReqs  │     │ - Consents  │     │ (Hybrid     │
│             │     │ - Clients   │     │  Storage)   │
└─────────────┘     └─────────────┘     └─────────────┘
```

---

## ✅ Checklist: Production-Ready Features

- [x] **Session Management**: Redis-based, secure, auto-expiring
- [x] **Auth Request Storage**: Redis-based, TTL 10min
- [x] **Password Hashing**: bcrypt with cost 12
- [x] **Consent Persistence**: PostgreSQL, revocable, trackable
- [x] **Random Generation**: Cryptographically secure
- [x] **Error Handling**: OAuth2 spec compliant
- [x] **Database Migration**: oauth2_consents table
- [x] **Repository Pattern**: Clean separation of concerns
- [x] **Dependency Injection**: All repos injected via constructor

---

## 🚀 Next Steps (Optional Enhancements)

1. **Session Refresh**: Extend session TTL on activity
2. **Remember Me**: Long-lived sessions option
3. **Multi-device Sessions**: Track sessions per device
4. **Consent Scopes Upgrade**: Handle scope changes
5. **Audit Logging**: Log all consent changes
6. **Admin UI**: Manage consents and sessions
7. **Rate Limiting**: Protect login endpoint
8. **2FA Integration**: Optional two-factor authentication

---

## 📝 Migration Instructions

### 1. Install golang-migrate (if not installed)
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

### 2. Run Migration
```bash
make migrate-up
```

### 3. Verify Migration
```bash
psql -U system_dev -d system_dev -c "\dt identify.oauth2_consents"
```

### 4. Test OAuth2 Flow
1. Start application: `go run cmd/server/main.go`
2. Navigate to authorization endpoint
3. Login with test credentials
4. Grant consent
5. Verify consent is saved (second login should skip consent)

---

## 🔒 Security Considerations

1. **HTTPS in Production**: Set `secure: true` for cookies
2. **SameSite Cookies**: Consider setting `SameSite=Strict`
3. **CSRF Protection**: Implement for login/consent forms
4. **Rate Limiting**: Add to prevent brute force
5. **Session Rotation**: Rotate session ID after privilege escalation
6. **Consent Audit**: Log all consent grants/revokes
7. **Key Rotation**: Rotate HMAC secret periodically

---

## 📚 References

- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
- [OpenID Connect Core](https://openid.net/specs/openid-connect-core-1_0.html)
- [Fosite Documentation](https://github.com/ory/fosite)
- [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
