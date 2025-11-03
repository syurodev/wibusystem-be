# OAuth 2.0 Authorization Server - Kế hoạch Implementation Chi tiết

## 📋 Tổng quan

Tài liệu này chi tiết hóa kế hoạch implementation cho OAuth 2.0 Authorization Server dựa trên kiến trúc đã được định nghĩa trong `docs/business/oauth2.md`.

**Mục tiêu**: Xây dựng một Authorization Server đầy đủ tính năng, tuân thủ OAuth 2.0 và OpenID Connect, sử dụng:
- **Go 1.25** với **Gin Framework**
- **ORY Fosite** SDK
- **PostgreSQL 18** (với UUID v7)
- **Redis 8.2.1** (hybrid storage)
- Multi-tenant architecture với RBAC

---

## 🎯 Các OAuth 2.0 Endpoints Cần Implement

### 0. OpenID Connect Discovery Endpoint ⭐ **CRITICAL**
```
GET /.well-known/openid-configuration
```

**Mục đích**: Publish server metadata để clients có thể auto-discover configuration

**Responsibilities**:
- Trả về JSON document chứa tất cả endpoint URLs
- List các features được support (grant types, scopes, algorithms)
- Enable automatic client configuration

**Request**: Không có parameters

**Response** (JSON):
```json
{
  "issuer": "https://auth.example.com",
  "authorization_endpoint": "https://auth.example.com/oauth2/auth",
  "token_endpoint": "https://auth.example.com/oauth2/token",
  "userinfo_endpoint": "https://auth.example.com/oauth2/userinfo",
  "jwks_uri": "https://auth.example.com/.well-known/jwks.json",
  "introspection_endpoint": "https://auth.example.com/oauth2/introspect",
  "revocation_endpoint": "https://auth.example.com/oauth2/revoke",
  "end_session_endpoint": "https://auth.example.com/oauth2/logout",
  "response_types_supported": ["code", "token", "id_token"],
  "grant_types_supported": ["authorization_code", "refresh_token", "client_credentials"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"],
  "scopes_supported": ["openid", "profile", "email", "offline_access"],
  "token_endpoint_auth_methods_supported": ["client_secret_basic", "client_secret_post"],
  "code_challenge_methods_supported": ["S256"],
  "claims_supported": ["sub", "iss", "aud", "exp", "iat", "email", "name"]
}
```

**Implementation**:
```go
func (h *Handler) Discovery(c *gin.Context) {
    baseURL := h.config.Issuer // "https://auth.example.com"

    discovery := map[string]any{
        "issuer": baseURL,
        "authorization_endpoint": baseURL + "/oauth2/auth",
        "token_endpoint": baseURL + "/oauth2/token",
        "userinfo_endpoint": baseURL + "/oauth2/userinfo",
        "jwks_uri": baseURL + "/.well-known/jwks.json",
        "introspection_endpoint": baseURL + "/oauth2/introspect",
        "revocation_endpoint": baseURL + "/oauth2/revoke",
        "end_session_endpoint": baseURL + "/oauth2/logout",

        "response_types_supported": []string{"code", "token", "id_token", "code id_token"},
        "grant_types_supported": []string{"authorization_code", "refresh_token", "client_credentials"},
        "subject_types_supported": []string{"public"},
        "id_token_signing_alg_values_supported": []string{"RS256"},

        "scopes_supported": []string{
            "openid", "profile", "email", "offline_access",
        },

        "token_endpoint_auth_methods_supported": []string{
            "client_secret_basic",
            "client_secret_post",
        },

        "code_challenge_methods_supported": []string{"S256"},

        "claims_supported": []string{
            "sub", "iss", "aud", "exp", "iat",
            "email", "email_verified", "name", "picture",
        },
    }

    c.JSON(http.StatusOK, discovery)
}
```

**Storage Operations**: Không có (static configuration)

**Why Critical**:
- ✅ OpenID Connect compliance requirement
- ✅ Enables automatic client configuration
- ✅ Industry standard (Google, Microsoft, Auth0 đều có)
- ✅ Giảm thiểu manual setup cho developers

---

### 0.1. JWKS Endpoint (JSON Web Key Set)
```
GET /.well-known/jwks.json
```

**Mục đích**: Publish public keys để clients verify JWT signatures

**Response** (JSON):
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "2024-10-26",
      "n": "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx...",
      "e": "AQAB"
    }
  ]
}
```

**Implementation**:
```go
func (h *Handler) JWKS(c *gin.Context) {
    // Get public key từ RSA private key
    publicKey := &h.privateKey.PublicKey

    jwks := map[string]any{
        "keys": []map[string]any{
            {
                "kty": "RSA",
                "use": "sig",
                "kid": h.config.KeyID, // "2024-10-26"
                "n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
                "e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
            },
        },
    }

    c.JSON(http.StatusOK, jwks)
}
```

**Why Needed**:
- Clients cần verify ID token signatures
- Không cần share private key
- Support key rotation

---

### 0.2. UserInfo Endpoint
```
GET /oauth2/userinfo
```

**Mục đích**: Trả về user profile claims (OpenID Connect standard)

**Authentication**: Requires valid access_token (Bearer token)

**Request Headers**:
```
Authorization: Bearer {access_token}
```

**Response** (JSON):
```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440001",
  "email": "user@example.com",
  "email_verified": true,
  "name": "John Doe",
  "picture": "https://example.com/avatar.jpg",
  "updated_at": 1735123456
}
```

**Implementation**:
```go
func (h *Handler) UserInfo(c *gin.Context) {
    ctx := c.Request.Context()

    // Extract access token từ Authorization header
    token := c.GetHeader("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")

    // Validate token (introspect internally)
    session, err := h.validateAccessToken(ctx, token)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
        return
    }

    // Get user info từ database
    userID := session.Subject
    user, err := h.store.GetUserByID(ctx, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
        return
    }

    // Return claims based on granted scopes
    claims := h.buildUserInfoClaims(user, session.GetGrantedScopes())
    c.JSON(http.StatusOK, claims)
}
```

**Scope → Claims Mapping**:
- `openid` scope → `sub` claim (always)
- `profile` scope → `name`, `picture`, `updated_at`
- `email` scope → `email`, `email_verified`

---

### 1. Authorization Endpoint
```
GET/POST /oauth2/auth
```

**Mục đích**: Endpoint cho user authorization (front-channel)

**Responsibilities**:
- Nhận authorization request từ client (qua browser redirect)
- Validate client_id, redirect_uri, scope, response_type
- Xác thực người dùng (nếu chưa authenticated)
- Hiển thị consent screen (nếu chưa có consent)
- Tạo authorization code
- Redirect về client với authorization code

**Request Parameters** (Query String hoặc POST body):
```
- client_id (required): UUID của OAuth2 client
- response_type (required): "code" (authorization code flow)
- redirect_uri (required): URL để redirect sau khi authorize
- scope (optional): Space-separated list of scopes
- state (optional): Opaque value dùng để maintain state
- code_challenge (required for PKCE): SHA256 hash của code_verifier
- code_challenge_method (required for PKCE): "S256" hoặc "plain"
```

**Response** (Redirect):
```
Success: 302 Found -> {redirect_uri}?code={authorization_code}&state={state}
Error: 302 Found -> {redirect_uri}?error={error_code}&error_description={desc}&state={state}
```

**Flow Details**:
1. Parse request với `Provider.NewAuthorizeRequest()`
2. Check user authentication (via session cookie)
3. Nếu chưa auth → redirect to login page
4. Check user consent
5. Nếu chưa consent → show consent screen
6. Create Fosite session với user ID
7. Grant requested scopes
8. Call `Provider.NewAuthorizeResponse()` để generate code
9. Fosite sẽ call `CreateAuthorizeCodeSession()` trên storage (Redis)
10. Write redirect response với `Provider.WriteAuthorizeResponse()`

**Storage Operations**:
- PostgreSQL: `GetClient()` - validate client exists
- Redis: `CreateAuthorizeCodeSession()` - lưu authorization code với TTL ~60s

**Security Considerations**:
- PKCE là mandatory cho public clients
- State parameter để prevent CSRF
- Validate redirect_uri khớp với registered URIs
- Authorization code chỉ dùng 1 lần

---

### 2. Token Endpoint
```
POST /oauth2/token
```

**Mục đích**: Exchange authorization code (hoặc credentials) lấy tokens (back-channel)

**Responsibilities**:
- Validate client credentials
- Exchange authorization code cho access_token + refresh_token
- Refresh access_token bằng refresh_token
- Validate PKCE code_verifier

**Request Parameters** (application/x-www-form-urlencoded):

**Grant Type: authorization_code**
```
- grant_type (required): "authorization_code"
- code (required): Authorization code từ /authorize
- redirect_uri (required): Phải khớp với request ban đầu
- client_id (required): Client identifier
- client_secret (required for confidential clients): Client secret
- code_verifier (required for PKCE): Original random string
```

**Grant Type: refresh_token**
```
- grant_type (required): "refresh_token"
- refresh_token (required): Refresh token đã được cấp trước đó
- client_id (required)
- client_secret (required for confidential clients)
- scope (optional): Không được vượt quá scope ban đầu
```

**Grant Type: client_credentials** (for service-to-service)
```
- grant_type (required): "client_credentials"
- client_id (required)
- client_secret (required)
- scope (optional)
```

**Response** (JSON):
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def50200...",
  "scope": "openid profile email"
}
```

**Error Response** (JSON):
```json
{
  "error": "invalid_grant",
  "error_description": "The authorization code has expired"
}
```

**Flow Details (Authorization Code)**:
1. Parse request với `Provider.NewAccessRequest()`
2. Fosite validates client_id + client_secret (PostgreSQL)
3. Fosite calls `GetAuthorizeCodeSession()` từ Redis
4. Validate code_verifier matches code_challenge
5. Fosite calls `InvalidateAuthorizeCodeSession()` - xóa code khỏi Redis
6. Generate access_token + refresh_token
7. Fosite calls `CreateAccessTokenSession()` và `CreateRefreshTokenSession()`
8. Write JSON response với `Provider.WriteAccessResponse()`

**Flow Details (Refresh Token)**:
1. Parse request với `Provider.NewAccessRequest()`
2. Fosite validates client credentials
3. Fosite calls `GetRefreshTokenSession()`
4. Validate refresh token chưa expired và chưa bị revoked
5. Grant scopes từ original request (không được vượt quá)
6. Generate access_token mới (và optionally refresh_token mới nếu rotation enabled)
7. Nếu rotation enabled: invalidate refresh_token cũ
8. Write response

**Storage Operations**:
- PostgreSQL: `GetClient()` - validate client credentials
- Redis: `GetAuthorizeCodeSession()`, `InvalidateAuthorizeCodeSession()`
- Redis/PostgreSQL: `CreateAccessTokenSession()`, `CreateRefreshTokenSession()`

**Security Considerations**:
- Client secret phải được hash với BCrypt (cost factor 12)
- Authorization code chỉ dùng được 1 lần
- PKCE code_verifier validation
- Refresh token rotation nên được enable

---

### 3. Token Introspection Endpoint
```
POST /oauth2/introspect
```

**Mục đích**: Validate và lấy metadata của token (for resource servers)

**Responsibilities**:
- Validate token còn active không
- Trả về metadata: user_id, client_id, scopes, expiration
- Check token revocation

**Authentication**: Requires client authentication (Basic Auth hoặc client credentials trong body)

**Request Parameters** (application/x-www-form-urlencoded):
```
- token (required): Token cần introspect (access_token hoặc refresh_token)
- token_type_hint (optional): "access_token" hoặc "refresh_token"
- client_id (required)
- client_secret (required)
```

**Response** (JSON):

**Active Token**:
```json
{
  "active": true,
  "scope": "openid profile email",
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john@example.com",
  "token_type": "Bearer",
  "exp": 1735123456,
  "iat": 1735119856,
  "sub": "550e8400-e29b-41d4-a716-446655440001",
  "aud": ["api.example.com"],
  "iss": "https://auth.example.com",
  "jti": "unique-token-id"
}
```

**Inactive Token**:
```json
{
  "active": false
}
```

**Flow Details**:
1. Authenticate client (middleware hoặc trong handler)
2. Parse request với `Provider.NewIntrospectionRequest()`
3. Fosite validates token signature
4. Check token trong Redis blacklist (revoked tokens)
5. Fosite calls `GetAccessTokenSession()` hoặc `GetRefreshTokenSession()`
6. Check expiration
7. Write response với `Provider.WriteIntrospectionResponse()`

**Storage Operations**:
- Redis: Check blacklist `SISMEMBER revoked:tokens {token_signature}`
- Redis/PostgreSQL: `GetAccessTokenSession()` hoặc `GetRefreshTokenSession()`

**Performance Optimization**:
- Redis check trước (blacklist + active sessions)
- PostgreSQL fallback nếu không tìm thấy trong Redis

---

### 4. Token Revocation Endpoint
```
POST /oauth2/revoke
```

**Mục đích**: Thu hồi access_token hoặc refresh_token

**Responsibilities**:
- Invalidate token ngay lập tức
- Add token signature vào blacklist
- Optionally revoke related tokens

**Authentication**: Requires client authentication

**Request Parameters** (application/x-www-form-urlencoded):
```
- token (required): Token cần revoke
- token_type_hint (optional): "access_token" hoặc "refresh_token"
- client_id (required)
- client_secret (required)
```

**Response**:
```
200 OK (body trống, theo RFC 7009)
```

**Flow Details**:
1. Authenticate client
2. Parse request với `Provider.NewRevocationRequest()`
3. Fosite validates token
4. Add token signature vào Redis blacklist với TTL = token expiration
5. Optionally: revoke related tokens (cascade)
6. Write response với `Provider.WriteRevocationResponse()`

**Storage Operations**:
- Redis: `RevokeAccessToken()` - add to blacklist
- Redis: `RevokeRefreshToken()` - add to blacklist
- Redis: Update session status to revoked

**Revocation Strategies**:
- **Single token revocation**: Chỉ revoke token được chỉ định
- **Cascade revocation**: Revoke tất cả tokens của cùng session
- **User logout**: Revoke tất cả tokens của user (all sessions, all clients)

---

### 5. Dynamic Client Registration Endpoint (RFC 7591) ⭐ **RECOMMENDED**
```
POST /oauth2/register
```

**Mục đích**: Cho phép clients tự động register programmatically (không cần manual admin setup)

**Responsibilities**:
- Validate registration request
- Generate client_id và client_secret
- Store client configuration trong database
- Return client credentials + registration_access_token

**Use Cases**:
- Multi-tenant SaaS: Tenants tự register apps
- Developer portals: Developers tự tạo OAuth apps
- CI/CD automation: Programmatic client creation

**Authentication**:
- **Option 1**: Open registration (với rate limiting)
- **Option 2**: Require initial_access_token (recommended for production)
- **Option 3**: Require tenant authentication (for multi-tenant)

**Request Body** (application/json):
```json
{
  "client_name": "My Application",
  "redirect_uris": [
    "https://app.example.com/callback",
    "https://app.example.com/callback2"
  ],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "scope": "openid profile email",
  "token_endpoint_auth_method": "client_secret_basic",
  "logo_uri": "https://app.example.com/logo.png",
  "policy_uri": "https://app.example.com/policy",
  "tos_uri": "https://app.example.com/tos",
  "contacts": ["admin@example.com"]
}
```

**Response** (201 Created):
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "8kPi9fL2mN3qR5sT7vW1xY0zA4bC6dE8",
  "client_id_issued_at": 1735123456,
  "client_secret_expires_at": 0,
  "client_name": "My Application",
  "redirect_uris": [
    "https://app.example.com/callback",
    "https://app.example.com/callback2"
  ],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "scope": "openid profile email",
  "token_endpoint_auth_method": "client_secret_basic",
  "registration_access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "registration_client_uri": "https://auth.example.com/oauth2/register/550e8400-e29b-41d4-a716-446655440000"
}
```

**Implementation** (Multi-tenant):
```go
func (h *Handler) RegisterClient(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Authenticate tenant (from JWT, API key, or initial_access_token)
    tenantID, err := h.getTenantFromAuth(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    // 2. Parse registration request
    var req ClientRegistrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid_request"})
        return
    }

    // 3. Validate request (redirect URIs, grant types, scopes)
    if err := h.validateRegistrationRequest(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid_client_metadata"})
        return
    }

    // 4. Generate client credentials
    clientID := uuid.NewV7()
    clientSecret := generateSecureRandom(32) // crypto/rand
    secretHash, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), 12)

    // 5. Create client in database
    client := &oauth2_clients{
        ID:           clientID,
        TenantID:     &tenantID, // Associate with tenant
        ClientName:   req.ClientName,
        SecretHash:   string(secretHash),
        RedirectURIs: req.RedirectURIs,
        GrantTypes:   req.GrantTypes,
        // ... other fields
    }

    if err := h.store.CreateClient(ctx, client); err != nil {
        c.JSON(500, gin.H{"error": "server_error"})
        return
    }

    // 6. Generate registration_access_token (JWT)
    regToken, err := h.generateRegistrationAccessToken(clientID, tenantID)

    // 7. Return response
    c.JSON(201, gin.H{
        "client_id": clientID.String(),
        "client_secret": clientSecret, // Plain text, ONLY returned once
        "client_id_issued_at": time.Now().Unix(),
        "client_secret_expires_at": 0,
        "registration_access_token": regToken,
        "registration_client_uri": fmt.Sprintf("%s/oauth2/register/%s", h.config.Issuer, clientID),
        // ... echo back all metadata
    })
}
```

**Storage Operations**:
- PostgreSQL: `INSERT INTO identify.oauth2_clients`
- Validate tenant quota (e.g., max 10 clients per tenant)

**Security Considerations**:
- Rate limiting (prevent abuse)
- Validate redirect URIs (no wildcards, HTTPS only for production)
- Limit scopes to tenant's allowed scopes
- Store registration_access_token hash in database

---

### 6. Client Configuration Management Endpoint (RFC 7592)
```
GET    /oauth2/register/{client_id}
PUT    /oauth2/register/{client_id}
DELETE /oauth2/register/{client_id}
```

**Mục đích**: Manage client configuration sau khi registration

**Authentication**: Requires `registration_access_token` (from registration response)

#### 6.1. Get Client Configuration
```
GET /oauth2/register/{client_id}
Authorization: Bearer {registration_access_token}
```

**Response**:
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_name": "My Application",
  "redirect_uris": ["https://app.example.com/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  // ... all metadata (KHÔNG return client_secret)
}
```

#### 6.2. Update Client Configuration
```
PUT /oauth2/register/{client_id}
Authorization: Bearer {registration_access_token}
Content-Type: application/json

{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_name": "My Application (Updated)",
  "redirect_uris": [
    "https://app.example.com/callback",
    "https://app.example.com/new-callback"
  ]
}
```

**Response**: Same as registration response (200 OK)

**Implementation**:
```go
func (h *Handler) UpdateClient(c *gin.Context) {
    clientID := c.Param("client_id")

    // Validate registration_access_token
    claims, err := h.validateRegistrationToken(c.GetHeader("Authorization"))
    if err != nil || claims.ClientID != clientID {
        c.JSON(401, gin.H{"error": "invalid_token"})
        return
    }

    // Parse and validate update request
    var req ClientRegistrationRequest
    c.ShouldBindJSON(&req)

    // Update in database
    h.store.UpdateClient(ctx, clientID, req)

    // Return updated config
}
```

#### 6.3. Delete Client
```
DELETE /oauth2/register/{client_id}
Authorization: Bearer {registration_access_token}
```

**Response**: 204 No Content

**Side Effects**:
- Revoke tất cả active tokens của client này
- Delete tất cả sessions
- Cannot be undone

---

### 7. Pushed Authorization Requests (PAR) Endpoint (RFC 9126) 🔒 **SECURITY**
```
POST /oauth2/par
```

**Mục đích**: Pre-register authorization request parameters (security enhancement)

**Why PAR**:
- ✅ **Security**: Authorization params không exposed trong browser URL
- ✅ **Privacy**: Sensitive data không visible trong browser history/logs
- ✅ **No URL limits**: Support large authorization requests
- ✅ **Tamper-proof**: Client signs request, server validates
- ✅ **FAPI 2.0 requirement**: Required for Financial-grade API security profile

**Flow**:
```
Step 1: Client pushes authorization params to server (back-channel)
   POST /oauth2/par

Step 2: Server validates and returns request_uri

Step 3: Client redirects user với short request_uri
   GET /oauth2/auth?client_id=xxx&request_uri=urn:...
```

**Authentication**: Requires client authentication (same as /token)

**Request** (application/x-www-form-urlencoded):
```
client_id=550e8400-e29b-41d4-a716-446655440000
&client_secret=8kPi9fL2mN3qR5sT7vW1xY0zA4bC6dE8
&response_type=code
&redirect_uri=https://app.example.com/callback
&scope=openid profile email
&state=xyz123
&code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM
&code_challenge_method=S256
&nonce=abc456
```

**Response** (201 Created):
```json
{
  "request_uri": "urn:ietf:params:oauth:request_uri:bwc4JK-ESC0w8acc191e-Y1LTC2",
  "expires_in": 90
}
```

**Usage** (Step 3):
```
# Client redirects user to authorization endpoint với request_uri
GET /oauth2/auth?client_id=550e8400-e29b-41d4-a716-446655440000&request_uri=urn:ietf:params:oauth:request_uri:bwc4JK-ESC0w8acc191e-Y1LTC2
```

**Implementation**:
```go
func (h *Handler) PushedAuthorizationRequest(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Authenticate client (Basic Auth hoặc client_secret trong body)
    client, err := h.authenticateClient(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "invalid_client"})
        return
    }

    // 2. Parse authorization parameters
    params := make(map[string]any)
    for key, values := range c.Request.Form {
        if len(values) > 0 {
            params[key] = values[0]
        }
    }

    // 3. Validate parameters (same as /authorize)
    if err := h.validateAuthorizationParams(params, client); err != nil {
        c.JSON(400, gin.H{"error": "invalid_request"})
        return
    }

    // 4. Generate request_uri (unique, opaque)
    requestURI := "urn:ietf:params:oauth:request_uri:" + generateSecureRandom(32)

    // 5. Store params in Redis với TTL 90 seconds
    parData, _ := json.Marshal(params)
    h.redis.Set(ctx, "par:"+requestURI, parData, 90*time.Second)

    // 6. Return request_uri
    c.JSON(201, gin.H{
        "request_uri": requestURI,
        "expires_in": 90,
    })
}

// Modify Authorization Endpoint to support request_uri
func (h *Handler) Authorize(c *gin.Context) {
    // Check if request_uri is present
    if requestURI := c.Query("request_uri"); requestURI != "" {
        // Retrieve params from Redis
        parData, err := h.redis.Get(ctx, "par:"+requestURI).Bytes()
        if err != nil {
            // request_uri invalid or expired
            c.JSON(400, gin.H{"error": "invalid_request_uri"})
            return
        }

        // Parse stored params
        var params map[string]string
        json.Unmarshal(parData, &params)

        // Delete request_uri (single use)
        h.redis.Del(ctx, "par:"+requestURI)

        // Continue with authorization using stored params
        // ...
    }

    // Normal authorization flow
    // ...
}
```

**Storage Operations**:
- Redis: `SET par:{request_uri} {params_json} EX 90`
- Redis: `DEL par:{request_uri}` (after use)

**Security Benefits**:
- Authorization request parameters không visible trong browser
- Prevents parameter tampering
- Reduces attack surface
- Required for FAPI (Financial-grade API)

---

## 🏗️ Components Cần Xây Dựng

### Phase 1: Core Infrastructure

#### 1.1. Storage Layer - PostgreSQL Store
**File**: `internal/oauth2/storage/sql_store.go`

**Struct**:
```go
type SQLStore struct {
    db     *sqlx.DB
    hasher fosite.Hasher
}
```

**Interfaces cần implement**:
- ✅ `fosite.ClientManager`
  - `GetClient(ctx, id) (fosite.Client, error)`

- ✅ `fosite.Storage` (composite interface)
  - Embeds: ClientManager + các storage interfaces khác

**Methods**:
```go
// Client Management
func (s *SQLStore) GetClient(ctx context.Context, id string) (fosite.Client, error)

// Helper: map database row to fosite.DefaultClient
func (s *SQLStore) rowToClient(row *clientRow) (*fosite.DefaultClient, error)
```

**Database Operations**:
```sql
-- Get Client
SELECT id, client_name, secret_hash, redirect_uris, grant_types,
       response_types, scopes, is_public, tenant_id
FROM identify.oauth2_clients
WHERE id = $1
```

**Responsibilities**:
- Load client configuration từ PostgreSQL
- Validate client credentials (BCrypt hash comparison)
- Tenant isolation (check tenant_id)

---

#### 1.2. Storage Layer - Redis Store
**File**: `internal/oauth2/storage/redis_store.go`

**Struct**:
```go
type RedisStore struct {
    client *redis.Client
}
```

**Interfaces cần implement**:
- ✅ `oauth2.AuthorizeCodeStorage`
  - `CreateAuthorizeCodeSession(ctx, signature, requester)`
  - `GetAuthorizeCodeSession(ctx, signature, session) (requester, error)`
  - `InvalidateAuthorizeCodeSession(ctx, signature)`
  - `DeleteAuthorizeCodeSession(ctx, signature)`

- ✅ `oauth2.AccessTokenStorage`
  - `CreateAccessTokenSession(ctx, signature, requester)`
  - `GetAccessTokenSession(ctx, signature, session) (requester, error)`
  - `DeleteAccessTokenSession(ctx, signature)`

- ✅ `oauth2.RefreshTokenStorage`
  - `CreateRefreshTokenSession(ctx, signature, requester)`
  - `GetRefreshTokenSession(ctx, signature, session) (requester, error)`
  - `DeleteRefreshTokenSession(ctx, signature)`

- ✅ `oauth2.PKCERequestStorage`
  - `CreatePKCERequestSession(ctx, signature, requester)`
  - `GetPKCERequestSession(ctx, signature, session) (requester, error)`
  - `DeletePKCERequestSession(ctx, signature)`

- ✅ `oauth2.TokenRevocationStorage`
  - `RevokeAccessToken(ctx, requestID)`
  - `RevokeRefreshToken(ctx, requestID)`
  - `RevokeRefreshTokenMaybeGracePeriod(ctx, requestID, signature)`

**Redis Key Patterns**:
```
fosite:auth_code:{signature}           → Authorization code session (TTL: 60s)
fosite:access_token:{signature}        → Access token session (TTL: 1h)
fosite:refresh_token:{signature}       → Refresh token session (TTL: 30d)
fosite:pkce:{signature}                → PKCE session (TTL: 10m)
fosite:blacklist:{signature}           → Revoked token (TTL: token expiration)
```

**Methods Example**:
```go
func (r *RedisStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, req fosite.Requester) error {
    key := "fosite:auth_code:" + signature

    // Serialize requester to JSON
    data, err := json.Marshal(req)
    if err != nil {
        return fosite.ErrServerError.WithWrap(err)
    }

    // Get TTL from session
    lifespan := req.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now())

    // Store in Redis with expiration
    return r.client.Set(ctx, key, data, lifespan).Err()
}

func (r *RedisStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
    key := "fosite:auth_code:" + signature

    data, err := r.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, fosite.ErrNotFound
    }
    if err != nil {
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    var req fosite.Request
    if err := json.Unmarshal(data, &req); err != nil {
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    return &req, nil
}
```

---

#### 1.3. Hybrid Storage Strategy
**File**: `internal/oauth2/storage/hybrid_store.go`

**Purpose**: Router để quyết định dùng PostgreSQL hay Redis cho từng operation

**Struct**:
```go
type HybridStore struct {
    sql   *SQLStore
    redis *RedisStore
}
```

**Strategy**:
```go
// Client data → PostgreSQL (persistent, infrequent reads)
func (h *HybridStore) GetClient(ctx, id) → h.sql.GetClient(ctx, id)

// Authorization codes → Redis (short-lived, high-frequency)
func (h *HybridStore) CreateAuthorizeCodeSession(ctx, sig, req) → h.redis.CreateAuthorizeCodeSession(...)

// Access tokens → Redis primary, PostgreSQL fallback
func (h *HybridStore) GetAccessTokenSession(ctx, sig, sess) {
    req, err := h.redis.GetAccessTokenSession(ctx, sig, sess)
    if err == fosite.ErrNotFound {
        return h.sql.GetAccessTokenSession(ctx, sig, sess) // Fallback to PostgreSQL
    }
    return req, err
}

// Refresh tokens → PostgreSQL (long-lived, less frequent)
func (h *HybridStore) CreateRefreshTokenSession(ctx, sig, req) {
    // Store in both for redundancy
    h.redis.CreateRefreshTokenSession(ctx, sig, req)
    return h.sql.CreateRefreshTokenSession(ctx, sig, req)
}
```

---

#### 1.4. Fosite Provider Setup
**File**: `internal/oauth2/provider.go`

**Responsibilities**:
- Initialize Fosite OAuth2Provider
- Configure strategies (JWT, HMAC, etc.)
- Set token lifespans
- Enable/disable flows

**Code**:
```go
package oauth2

import (
    "crypto/rsa"
    "time"

    "github.com/ory/fosite"
    "github.com/ory/fosite/compose"
    "github.com/ory/fosite/handler/openid"
    "github.com/ory/fosite/token/jwt"
)

type Config struct {
    AccessTokenLifespan  time.Duration // 1 hour
    RefreshTokenLifespan time.Duration // 30 days
    AuthCodeLifespan     time.Duration // 60 seconds
    IDTokenLifespan      time.Duration // 1 hour
    Issuer               string        // "https://auth.example.com"
    PrivateKey           *rsa.PrivateKey
}

func NewProvider(store fosite.Storage, config *Config) fosite.OAuth2Provider {
    fc := &fosite.Config{
        AccessTokenLifespan:  config.AccessTokenLifespan,
        RefreshTokenLifespan: config.RefreshTokenLifespan,
        AuthorizeCodeLifespan: config.AuthCodeLifespan,
        IDTokenLifespan:      config.IDTokenLifespan,

        // Global secret for HMAC strategy
        GlobalSecret: []byte("your-very-secret-key-min-32-bytes"),

        // Enable PKCE
        EnforcePKCE: true,
        EnablePKCEPlainChallengeMethod: false, // Only S256

        // Token format
        AccessTokenIssuer: config.Issuer,

        // Refresh token rotation
        RefreshTokenScopes: []string{"offline_access"},

        // Hash cost for BCrypt
        HashCost: 12,
    }

    // Create Fosite provider với strategies
    return compose.Compose(
        fc,
        store,
        &compose.CommonStrategy{
            CoreStrategy: compose.NewOAuth2HMACStrategy(fc),
            OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(fc, privateKey),
        },

        // Enable OAuth2 flows
        compose.OAuth2AuthorizeExplicitFactory,
        compose.OAuth2RefreshTokenGrantFactory,
        compose.OAuth2ClientCredentialsGrantFactory,

        // Enable introspection & revocation
        compose.OAuth2TokenIntrospectionFactory,
        compose.OAuth2TokenRevocationFactory,

        // Enable PKCE
        compose.OAuth2PKCEFactory,

        // Enable OpenID Connect
        compose.OpenIDConnectExplicitFactory,
    )
}
```

---

### Phase 2: HTTP Handlers (Gin)

#### 2.1. Authorization Handler
**File**: `internal/app/handler/v1/oauth2/authorize.go`

```go
package oauth2

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/ory/fosite"
)

type Handler struct {
    provider fosite.OAuth2Provider
    store    storage.Store
}

func NewHandler(provider fosite.OAuth2Provider, store storage.Store) *Handler {
    return &Handler{
        provider: provider,
        store:    store,
    }
}

func (h *Handler) Authorize(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Parse authorization request
    ar, err := h.provider.NewAuthorizeRequest(ctx, c.Request)
    if err != nil {
        h.provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
        return
    }

    // 2. Check user authentication
    userID, err := h.getUserFromSession(c)
    if err != nil {
        // Redirect to login page with return_to parameter
        loginURL := "/login?return_to=" + url.QueryEscape(c.Request.URL.String())
        c.Redirect(http.StatusFound, loginURL)
        return
    }

    // 3. Check consent
    hasConsent, err := h.checkConsent(ctx, userID, ar.GetClient().GetID(), ar.GetRequestedScopes())
    if err != nil {
        h.provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
        return
    }

    if !hasConsent {
        // Show consent screen
        h.showConsentScreen(c, ar)
        return
    }

    // 4. Create session
    session := &fosite.DefaultSession{
        Subject:  userID.String(),
        Username: "user@example.com", // Get from database
    }

    // 5. Grant scopes
    for _, scope := range ar.GetRequestedScopes() {
        ar.GrantScope(scope)
    }

    // 6. Generate authorization response
    response, err := h.provider.NewAuthorizeResponse(ctx, ar, session)
    if err != nil {
        h.provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
        return
    }

    // 7. Write redirect response (với authorization code)
    h.provider.WriteAuthorizeResponse(ctx, c.Writer, ar, response)
}

func (h *Handler) getUserFromSession(c *gin.Context) (*uuid.UUID, error) {
    // Get session cookie
    sessionID, err := c.Cookie("session_id")
    if err != nil {
        return nil, err
    }

    // Validate session and get user ID
    // TODO: Implement session management
    return nil, nil
}

func (h *Handler) checkConsent(ctx context.Context, userID *uuid.UUID, clientID string, scopes []string) (bool, error) {
    // Check if user đã consent cho client này với scopes này
    // TODO: Implement consent storage
    return false, nil
}

func (h *Handler) showConsentScreen(c *gin.Context, ar fosite.AuthorizeRequester) {
    // Render consent page
    c.HTML(http.StatusOK, "consent.html", gin.H{
        "ClientName": ar.GetClient().GetName(),
        "Scopes":     ar.GetRequestedScopes(),
        "RequestURL": c.Request.URL.String(),
    })
}
```

---

#### 2.2. Token Handler
**File**: `internal/app/handler/v1/oauth2/token.go`

```go
func (h *Handler) Token(c *gin.Context) {
    ctx := c.Request.Context()

    // 1. Parse access request
    accessRequest, err := h.provider.NewAccessRequest(ctx, c.Request, &fosite.DefaultSession{})
    if err != nil {
        h.provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
        return
    }

    // 2. Handle refresh token grant
    if accessRequest.GetGrantTypes().ExactOne("refresh_token") {
        // Grant original scopes
        for _, scope := range accessRequest.GetGrantedScopes() {
            accessRequest.GrantScope(scope)
        }
    }

    // 3. Generate token response
    response, err := h.provider.NewAccessResponse(ctx, accessRequest)
    if err != nil {
        h.provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
        return
    }

    // 4. Write JSON response
    h.provider.WriteAccessResponse(ctx, c.Writer, accessRequest, response)
}
```

---

#### 2.3. Introspection Handler
**File**: `internal/app/handler/v1/oauth2/introspect.go`

```go
func (h *Handler) Introspect(c *gin.Context) {
    ctx := c.Request.Context()

    // Parse introspection request
    ir, err := h.provider.NewIntrospectionRequest(ctx, c.Request, &fosite.DefaultSession{})
    if err != nil {
        h.provider.WriteIntrospectionError(c.Writer, err)
        return
    }

    // Write response
    h.provider.WriteIntrospectionResponse(c.Writer, ir)
}
```

---

#### 2.4. Revocation Handler
**File**: `internal/app/handler/v1/oauth2/revoke.go`

```go
func (h *Handler) Revoke(c *gin.Context) {
    ctx := c.Request.Context()

    // Parse revocation request
    err := h.provider.NewRevocationRequest(ctx, c.Request)
    if err != nil {
        h.provider.WriteRevocationResponse(c.Writer, err)
        return
    }

    // Write success response
    h.provider.WriteRevocationResponse(c.Writer, nil)
}
```

---

### Phase 3: Middleware

#### 3.1. Client Authentication Middleware
**File**: `internal/app/middleware/oauth_client.go`

**Purpose**: Authenticate OAuth2 clients cho /introspect và /revoke endpoints

```go
package middleware

import (
    "encoding/base64"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

func ClientAuth(store storage.Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()

        // Try Basic Auth first
        clientID, clientSecret, hasBasicAuth := c.Request.BasicAuth()

        if !hasBasicAuth {
            // Try form parameters
            clientID = c.PostForm("client_id")
            clientSecret = c.PostForm("client_secret")
        }

        if clientID == "" || clientSecret == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "invalid_client",
                "error_description": "Client authentication failed",
            })
            c.Abort()
            return
        }

        // Validate client credentials
        client, err := store.GetClient(ctx, clientID)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "invalid_client",
            })
            c.Abort()
            return
        }

        // Compare secret hash
        hasher := &fosite.BCrypt{WorkFactor: 12}
        if err := hasher.Compare(ctx, client.GetHashedSecret(), []byte(clientSecret)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "invalid_client",
            })
            c.Abort()
            return
        }

        // Store client in context
        c.Set("oauth_client", client)
        c.Next()
    }
}
```

---

### Phase 4: Routing Setup

#### 4.1. Register OAuth2 Routes
**File**: `cmd/server/main.go` hoặc `internal/app/router.go`

```go
func setupOAuth2Routes(router *gin.Engine, handler *oauth2.Handler, store storage.Store) {
    // ========================================
    // OpenID Connect Discovery Endpoints
    // ========================================
    wellKnown := router.Group("/.well-known")
    {
        // OpenID Connect Discovery (CRITICAL - MUST be first)
        wellKnown.GET("/openid-configuration", handler.Discovery)

        // JWKS endpoint for public key distribution
        wellKnown.GET("/jwks.json", handler.JWKS)
    }

    // ========================================
    // OAuth2 / OpenID Connect Endpoints
    // ========================================
    oauth2Group := router.Group("/oauth2")
    {
        // Authorization endpoint (user-facing, front-channel)
        oauth2Group.GET("/auth", handler.Authorize)
        oauth2Group.POST("/auth", handler.Authorize)

        // Token endpoint (back-channel)
        oauth2Group.POST("/token", handler.Token)

        // UserInfo endpoint (OpenID Connect)
        oauth2Group.GET("/userinfo", handler.UserInfo)
        oauth2Group.POST("/userinfo", handler.UserInfo)

        // Pushed Authorization Requests (PAR) - RFC 9126
        // Requires client authentication
        oauth2Group.POST("/par", middleware.ClientAuth(store), handler.PushedAuthorizationRequest)

        // Protected endpoints (require client authentication)
        protected := oauth2Group.Group("")
        protected.Use(middleware.ClientAuth(store))
        {
            protected.POST("/introspect", handler.Introspect)
            protected.POST("/revoke", handler.Revoke)
        }

        // Dynamic Client Registration - RFC 7591/7592
        // Registration might be open or protected depending on configuration
        oauth2Group.POST("/register", handler.RegisterClient)

        // Client configuration management (requires registration_access_token)
        oauth2Group.GET("/register/:client_id", middleware.RegistrationTokenAuth(), handler.GetClient)
        oauth2Group.PUT("/register/:client_id", middleware.RegistrationTokenAuth(), handler.UpdateClient)
        oauth2Group.DELETE("/register/:client_id", middleware.RegistrationTokenAuth(), handler.DeleteClient)

        // Logout endpoint (optional, for session management)
        oauth2Group.GET("/logout", handler.Logout)
        oauth2Group.POST("/logout", handler.Logout)
    }
}
```

**Endpoint Summary**:
```
# Discovery & Metadata (OpenID Connect)
/.well-known/openid-configuration   GET     Discovery metadata
/.well-known/jwks.json              GET     Public keys (JWKS)

# Core OAuth2/OIDC Endpoints
/oauth2/auth                        GET     Authorization (with login/consent UI)
/oauth2/auth                        POST    Authorization (form submit)
/oauth2/token                       POST    Token exchange
/oauth2/userinfo                    GET     User profile claims

# Token Management
/oauth2/introspect                  POST    Token validation (client auth required)
/oauth2/revoke                      POST    Token revocation (client auth required)

# Session Management
/oauth2/logout                      GET     End session

# Dynamic Client Registration (RFC 7591/7592) - RECOMMENDED for SaaS
/oauth2/register                    POST    Register new OAuth2 client
/oauth2/register/{client_id}        GET     Get client configuration
/oauth2/register/{client_id}        PUT     Update client configuration
/oauth2/register/{client_id}        DELETE  Delete client

# Pushed Authorization Requests (RFC 9126) - Security Enhancement
/oauth2/par                         POST    Push authorization request parameters
```

---

## 📐 Implementation Order

### Week 1: Foundation
- [x] Database schema đã có (migrations 000001-000006)
- [ ] **Task 1.1**: Implement `SQLStore` (PostgreSQL storage)
  - `GetClient()` method
  - Client validation logic
  - BCrypt hasher integration

- [ ] **Task 1.2**: Implement `RedisStore` (Redis storage)
  - `AuthorizeCodeStorage` interface
  - `AccessTokenStorage` interface
  - `RefreshTokenStorage` interface
  - `PKCERequestStorage` interface
  - `TokenRevocationStorage` interface

- [ ] **Task 1.3**: Implement `HybridStore`
  - Routing logic between SQL và Redis
  - Fallback strategy

- [ ] **Task 1.4**: Setup Fosite Provider
  - Configuration
  - Strategy selection (HMAC vs JWT)
  - Enable flows

### Week 2: Core Endpoints
- [ ] **Task 2.0**: Discovery & JWKS Endpoints ⭐ **START HERE**
  - `Discovery()` handler - return OpenID configuration JSON
  - `JWKS()` handler - return public keys
  - RSA key pair generation (or load from file)
  - Configuration struct (issuer, endpoints)
  - Test với OIDC client libraries

- [ ] **Task 2.1**: Authorization Endpoint
  - Parse request handler
  - User authentication check
  - Consent screen (HTML template)
  - Session creation
  - Redirect response

- [ ] **Task 2.2**: Token Endpoint
  - Authorization code exchange
  - Refresh token grant
  - Client credentials grant (optional)
  - Response generation

- [ ] **Task 2.3**: UserInfo Endpoint
  - Access token validation
  - User data retrieval from PostgreSQL
  - Scope-based claims filtering
  - Response generation

- [ ] **Task 2.4**: Testing Authorization Code Flow
  - End-to-end test
  - PKCE validation test
  - Error cases test

### Week 3: Advanced Features
- [ ] **Task 3.1**: Introspection Endpoint
  - Request parsing
  - Token validation
  - Blacklist check
  - Response generation

- [ ] **Task 3.2**: Revocation Endpoint
  - Token invalidation
  - Blacklist management
  - Cascade revocation

- [ ] **Task 3.3**: Client Authentication Middleware
  - Basic Auth support
  - Form parameters support
  - BCrypt validation

### Week 4: Security & Polish
- [ ] **Task 4.1**: Security Hardening
  - Rate limiting
  - CORS configuration
  - HTTPS enforcement
  - Secret rotation

- [ ] **Task 4.2**: Monitoring & Logging
  - Request logging
  - Performance metrics
  - Error tracking

- [ ] **Task 4.3**: Documentation
  - API documentation
  - Integration guide
  - Deployment guide

---

## 🧪 Testing Strategy

### Unit Tests
- [ ] SQLStore methods
- [ ] RedisStore methods
- [ ] HybridStore routing logic
- [ ] Handler logic (with mocked storage)

### Integration Tests
- [ ] **OpenID Discovery**: Test `.well-known/openid-configuration` response
- [ ] **JWKS Endpoint**: Verify public key format
- [ ] **Full Authorization Code Flow**: End-to-end với PKCE
- [ ] **Refresh Token Flow**: Token rotation test
- [ ] **Token Introspection**: Active/inactive token validation
- [ ] **Token Revocation**: Immediate invalidation
- [ ] **UserInfo Endpoint**: Scope-based claim filtering
- [ ] **Error scenarios**: Invalid client, expired tokens, etc.

### E2E Tests
- [ ] **Auto-discovery**: Test OIDC client library auto-configuration
- [ ] **Browser-based flow**: Selenium/Playwright authorization
- [ ] **Multi-client scenarios**: Different clients, different scopes
- [ ] **Concurrent requests**: Load testing
- [ ] **Token expiration**: Refresh flow after access token expires
- [ ] **ID Token verification**: JWT signature validation với JWKS

---

## 📚 Dependencies

### Go Modules (cần thêm vào go.mod)
```bash
go get github.com/ory/fosite@latest
go get github.com/gofrs/uuid/v5
go get github.com/jmoiron/sqlx
go get golang.org/x/crypto/bcrypt
```

### Fosite Documentation
- [Fosite GitHub](https://github.com/ory/fosite)
- [OAuth 2.0 RFC 6749](https://tools.ietf.org/html/rfc6749)
- [PKCE RFC 7636](https://tools.ietf.org/html/rfc7636)
- [Token Introspection RFC 7662](https://tools.ietf.org/html/rfc7662)
- [Token Revocation RFC 7009](https://tools.ietf.org/html/rfc7009)

---

## ⚠️ Critical Considerations

### Security
1. **Client Secret**: NEVER store plain text, always BCrypt hash (cost factor 12)
2. **RSA Private Key**:
   - Generate 2048-bit RSA key pair for JWT signing
   - Store private key securely (encrypted file, Vault, AWS Secrets Manager)
   - NEVER commit private key to git
   - Rotate keys periodically (publish new key in JWKS, keep old key for verification)
   - Use `kid` (key ID) in JWT header để support multiple keys
3. **PKCE**: Mandatory for public clients, recommended for all (S256 only, disable plain)
4. **HTTPS**: Required for production (all endpoints)
5. **Token Lifetime**:
   - Authorization code: 60 seconds
   - Access token: 15-60 minutes
   - Refresh token: 7-30 days
   - ID token: 1 hour
6. **Refresh Token Rotation**: Enable để prevent replay attacks
7. **OpenID Discovery**: Serve over HTTPS only, cache response for performance

### Performance
1. **Redis First**: Authorization codes, PKCE sessions → Redis only
2. **Hybrid Access Tokens**: Redis primary, PostgreSQL fallback
3. **PostgreSQL for Persistent**: Refresh tokens, client configs
4. **Connection Pooling**: pgxpool (max 25), redis pool (max 10)
5. **Indexing**: UUID v7 for optimal B-tree performance

### Multi-tenancy
1. **Tenant Isolation**: Every query must filter by `tenant_id`
2. **Global Clients**: `tenant_id = NULL` for platform clients
3. **RBAC Integration**: Check user permissions before granting scopes
4. **Audit Logging**: Track token grants per tenant

---

## 🔑 RSA Key Generation (Required Setup)

Before starting development, generate RSA key pair:

```bash
# Generate 2048-bit RSA private key
openssl genrsa -out private_key.pem 2048

# Extract public key
openssl rsa -in private_key.pem -pubout -out public_key.pem

# Verify keys
openssl rsa -in private_key.pem -noout -text
```

**In Go code**:
```go
import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "os"
)

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
    keyData, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    block, _ := pem.Decode(keyData)
    if block == nil {
        return nil, errors.New("failed to decode PEM block")
    }

    privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return nil, err
    }

    return privateKey, nil
}
```

**Configuration** (`.env`):
```bash
# OAuth2 / OpenID Connect Configuration
OAUTH2_ISSUER=https://auth.example.com
OAUTH2_PRIVATE_KEY_PATH=/secrets/private_key.pem
OAUTH2_KEY_ID=2024-10-26

# Token Lifespans (seconds)
OAUTH2_ACCESS_TOKEN_LIFESPAN=3600        # 1 hour
OAUTH2_REFRESH_TOKEN_LIFESPAN=2592000    # 30 days
OAUTH2_AUTH_CODE_LIFESPAN=60             # 60 seconds
OAUTH2_ID_TOKEN_LIFESPAN=3600            # 1 hour
```

---

## 🚀 Next Steps

1. **Generate RSA Keys**: Follow steps above to create key pair
2. **Review Plan**: Đọc kỹ plan này và confirm với team
3. **Setup Dependencies**: Install Fosite và related libraries
4. **Start Week 1**: Begin với SQLStore implementation
5. **Daily Standups**: Track progress theo weekly tasks
6. **Code Reviews**: Mỗi phase cần review trước khi proceed

---

**Last Updated**: 2025-10-26
**Status**: 📝 Planning Phase
**Next Milestone**: Week 1 - Foundation Layer
