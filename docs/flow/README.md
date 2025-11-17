# OAuth2 Flow Documentation

Tài liệu chi tiết về các luồng OAuth2 được implement trong hệ thống.

---

## 📚 Danh sách Flow

### 1. [Authorization Code Flow](./01-authorization-code-flow.md)
**Mục đích:** User authentication và authorization cho web/mobile apps

**Khi nào dùng:**
- ✅ Web applications (React, Vue, Angular)
- ✅ Mobile applications (iOS, Android)
- ✅ Single Page Applications (SPA)
- ✅ Bất kỳ ứng dụng nào cần user context

**Đặc điểm:**
- User login thông qua browser redirect
- Support PKCE (Proof Key for Code Exchange)
- Trả về authorization code → đổi lấy tokens
- Bao gồm consent screen
- Hỗ trợ refresh tokens (với scope `offline_access`)

**Tokens nhận được:**
- ✅ Access Token (1 hour)
- ✅ Refresh Token (30 days)
- ✅ ID Token (OIDC)

---

### 2. [Client Credentials Flow](./02-client-credentials-flow.md)
**Mục đích:** Server-to-server authentication (không có user)

**Khi nào dùng:**
- ✅ Backend services gọi API của nhau
- ✅ Microservices communication
- ✅ Scheduled jobs / Cron jobs
- ✅ CLI tools / Scripts

**Đặc điểm:**
- Không có user involvement
- Chỉ dùng client credentials (client_id + client_secret)
- Đơn giản và nhanh nhất
- Không có browser redirect
- Không có refresh token (không cần thiết)

**Tokens nhận được:**
- ✅ Access Token (1 hour)
- ❌ Refresh Token (không có)
- ❌ ID Token (không có user)

---

### 3. [Refresh Token Flow](./03-refresh-token-flow.md)
**Mục đích:** Lấy access token mới mà không cần user login lại

**Khi nào dùng:**
- ✅ Access token hết hạn
- ✅ Muốn duy trì user session
- ✅ Background token refresh
- ✅ Silent authentication

**Đặc điểm:**
- Không cần user interaction
- Sử dụng refresh token để lấy access token mới
- Support token rotation (security best practice)
- Có thể request scope ít hơn original
- Detect token reuse (security)

**Tokens nhận được:**
- ✅ Access Token mới (1 hour)
- ✅ Refresh Token mới (30 days, rotated)
- ❌ ID Token (không re-issue)

---

### 4. [Logout Flow](./04-logout-flow.md)
**Mục đích:** User logout khỏi Authorization Server (RP-Initiated Logout)

**Khi nào dùng:**
- ✅ User muốn logout khỏi application
- ✅ Security requirement: force logout tất cả sessions
- ✅ Revoke tokens khi account bị compromise
- ✅ End session sau khi hoàn thành task

**Đặc điểm:**
- Support cả GET và POST methods
- Xóa session cookie
- Optionally revoke OAuth2 tokens
- Redirect về client app sau logout
- Tuân thủ OpenID Connect RP-Initiated Logout spec

**Actions thực hiện:**
- ✅ Delete user session (Redis)
- ✅ Clear session cookie
- ✅ Optionally mark tokens inactive (PostgreSQL)
- ✅ Redirect to post_logout_redirect_uri

---

## 🔄 Flow Comparison

| Aspect | Authorization Code | Client Credentials | Refresh Token | Logout |
|--------|-------------------|-------------------|---------------|--------|
| **User Context** | ✅ Yes | ❌ No | ✅ Yes (from original) | ✅ Yes |
| **User Interaction** | ✅ Required | ❌ Not required | ❌ Not required | Optional |
| **Browser Redirect** | ✅ Yes | ❌ No | ❌ No | ✅ Yes (optional) |
| **PKCE** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Consent Screen** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Access Token** | ✅ 1 hour | ✅ 1 hour | ✅ 1 hour | ❌ N/A |
| **Refresh Token** | ✅ 30 days | ❌ No | ✅ 30 days (rotated) | ❌ N/A |
| **ID Token** | ✅ Yes | ❌ No | ❌ No | ❌ N/A |
| **Client Secret** | Optional* | ✅ Required | ✅ Required* | ❌ Not required |
| **Use Case** | User login | Service auth | Silent refresh | User logout |

\* Confidential clients require secret, public clients don't

---

## 🎯 Chọn Flow Nào?

### Flowchart

```
                    ┌─────────────────────────┐
                    │  Cần user context?      │
                    └───────┬─────────────────┘
                            │
                    ┌───────┴────────┐
                    │                │
                   Yes              No
                    │                │
                    ▼                ▼
        ┌──────────────────┐  ┌──────────────────┐
        │ User đã login    │  │ Client           │
        │ và có refresh    │  │ Credentials Flow │
        │ token?           │  └──────────────────┘
        └────┬─────────────┘
             │
        ┌────┴─────┐
       Yes        No
        │          │
        ▼          ▼
┌──────────────┐ ┌──────────────────┐
│ Refresh      │ │ Authorization    │
│ Token Flow   │ │ Code Flow        │
└──────────────┘ └──────────────────┘
```

### Decision Tree

1. **Bạn đang xây dựng gì?**
   
   - **Web/Mobile App với user login** 
     → Authorization Code Flow
   
   - **Backend service không có user**
     → Client Credentials Flow
   
   - **Cần renew token cho user đã login**
     → Refresh Token Flow

2. **Bạn có client secret không?**
   
   - **Yes (Confidential client)** 
     → Có thể dùng tất cả flows
   
   - **No (Public client - SPA/Mobile)**
     → Authorization Code Flow + PKCE

3. **Bạn cần maintain session lâu dài?**
   
   - **Yes (> 1 hour)**
     → Request scope `offline_access` → nhận refresh token
   
   - **No (≤ 1 hour)**
     → Chỉ cần access token

---

## 🔐 Security Best Practices

### Áp dụng cho tất cả flows

1. **Always use HTTPS**
   - ⚠️ Never send tokens over HTTP
   - ✅ All endpoints phải HTTPS trong production

2. **Validate redirect URIs**
   - ✅ Chỉ redirect về URIs đã đăng ký
   - ⚠️ Không cho phép wildcard URIs

3. **Use short-lived access tokens**
   - ✅ Access token: 1 hour (default)
   - ✅ Force token rotation thường xuyên

4. **Implement rate limiting**
   - ✅ Token endpoint: 100 req/hour per client
   - ✅ API endpoints: 1000 req/min per token

5. **Log security events**
   - ✅ Failed authentication attempts
   - ✅ Token reuse attempts
   - ✅ Unusual access patterns

### Specific flows

#### Authorization Code Flow
- ✅ Always use PKCE (even for confidential clients)
- ✅ Validate state parameter (CSRF protection)
- ✅ Single-use authorization codes (60 seconds TTL)
- ✅ Bind code to client via code_challenge

#### Client Credentials Flow
- ✅ Rotate client secrets every 90 days
- ✅ Use environment-specific secrets
- ✅ Never log client secrets
- ✅ Implement client authentication rate limiting

#### Refresh Token Flow
- ✅ Implement token rotation
- ✅ Detect and block token reuse
- ✅ Revoke tokens on password change
- ✅ Absolute expiration (30 days max)

---

## 📊 Data Flow Overview

### Authorization Code Flow

```
User Browser
     │
     │ 1. GET /oauth2/auth
     ▼
┌─────────────────────┐
│ Authorization       │
│ Server              │──► 2. Redirect to login
│ (Gin + Fosite)      │
└──────┬──────────────┘
       │ 3. Check: PostgreSQL (client, user)
       │ 4. Save: Redis (auth_request, code)
       │
       │ 5. Redirect with code
       ▼
Client App
     │
     │ 6. POST /oauth2/token
     ▼
┌─────────────────────┐
│ Authorization       │──► 7. Validate code (Redis)
│ Server              │──► 8. Generate tokens
└──────┬──────────────┘
       │ 9. Save: Redis (token sessions)
       │
       │ 10. Return tokens
       ▼
Client App stores tokens
```

### Client Credentials Flow

```
Client Service
     │
     │ 1. POST /oauth2/token
     │    + client credentials
     ▼
┌─────────────────────┐
│ Authorization       │──► 2. Validate client (PostgreSQL)
│ Server              │──► 3. Generate access token
└──────┬──────────────┘
       │ 4. Save: Redis (token session)
       │
       │ 5. Return token
       ▼
Client Service stores token
```

### Refresh Token Flow

```
Client App
     │ (has expired access_token + valid refresh_token)
     │
     │ 1. POST /oauth2/token
     │    + refresh_token
     ▼
┌─────────────────────┐
│ Authorization       │──► 2. Validate refresh token (Redis)
│ Server              │──► 3. Generate new tokens
└──────┬──────────────┘
       │ 4. Invalidate old refresh token
       │ 5. Save new token sessions
       │
       │ 6. Return new tokens
       ▼
Client App updates stored tokens
```

---

## 🗄️ Storage Strategy

### Redis (Temporary Data)

| Data Type | Key Pattern | TTL | Purpose |
|-----------|-------------|-----|---------|
| Authorization Request | `auth_request:{request_id}` | 10 min | Store request state during login |
| User Session | `session:{session_id}` | Configurable | User authentication session |
| Authorization Code | `auth_code:{signature}` | 60 sec | Single-use code for token exchange |
| Access Token Session | `access_token:{signature}` | 1 hour | Token validation & introspection |
| Refresh Token Session | `refresh_token:{signature}` | 30 days | Refresh token validation |
| Token Blacklist | `revoked:{token_type}:{sig}` | Match original TTL | Immediate token revocation |

### PostgreSQL (Persistent Data)

| Table | Purpose | Key Data |
|-------|---------|----------|
| `users` | User accounts | email, password_hash, email_verified |
| `oauth2_clients` | Registered clients | client_id, secret_hash, redirect_uris, scopes |
| `oauth2_consents` | User consent records | user_id, client_id, granted_scopes |
| `sessions` | (Optional) Long-lived sessions | user_id, created_at, last_used_at |

---

## 🔍 Common Scenarios

### Scenario 1: User Login (First Time)

**Flow:** Authorization Code Flow

1. User clicks "Login" → Redirect to `/oauth2/auth`
2. User sees login page (server-rendered)
3. User enters credentials → POST `/oauth2/login`
4. User sees consent page (first time) → Approves
5. Redirect back to app with `code`
6. App exchanges `code` for tokens
7. App stores tokens
8. User logged in ✅

**Duration:** ~30 seconds (with user interaction)

---

### Scenario 2: User Login (Returning User)

**Flow:** Authorization Code Flow

1. User clicks "Login" → Redirect to `/oauth2/auth`
2. Server checks session → User already logged in ✅
3. Server checks consent → Already granted ✅
4. Redirect back with `code` immediately (no UI)
5. App exchanges `code` for tokens
6. User logged in ✅

**Duration:** ~2 seconds (no user interaction needed)

---

### Scenario 3: API Call with Expired Token

**Flow:** Refresh Token Flow

1. App prepares API call
2. Check access token → Expired ❌
3. Use refresh token → Get new access token
4. Retry API call with new token
5. Success ✅

**Duration:** ~500ms (transparent to user)

---

### Scenario 4: Service-to-Service Communication

**Flow:** Client Credentials Flow

1. Service A needs to call Service B
2. Service A requests token (cached if still valid)
3. Service A calls Service B with token
4. Service B validates token
5. Success ✅

**Duration:** ~100ms (with caching)

---

### Scenario 5: User Logout

**Flow:** Logout Flow (RP-Initiated Logout)

1. User clicks "Logout" in client app
2. Client redirects to `/oauth2/logout?post_logout_redirect_uri=...`
3. Server deletes session from Redis
4. Server optionally marks all tokens inactive
5. Server clears session cookie
6. Redirect back to client app
7. Client clears stored tokens
8. User logged out ✅

**Duration:** ~1 second (transparent redirect)

**Flows affected:**
- ❌ Authorization Code Flow (session cleared)
- ❌ Refresh Token Flow (tokens optionally revoked)
- ✅ Access Token (may still be valid ≤ 1 hour if not revoked)

---

## 📝 Implementation Checklist

### Authorization Code Flow
- [ ] Login page (server-rendered HTML)
- [ ] Consent page (server-rendered HTML)
- [ ] `/oauth2/auth` endpoint
- [ ] `/oauth2/login` endpoint
- [ ] `/oauth2/consent` endpoint
- [ ] `/oauth2/token` endpoint (code exchange)
- [ ] PKCE validation
- [ ] State parameter validation
- [ ] Authorization code storage (Redis)
- [ ] Session management
- [ ] Consent storage (PostgreSQL)

### Client Credentials Flow
- [ ] `/oauth2/token` endpoint (client credentials grant)
- [ ] Client authentication (Basic Auth)
- [ ] Client validation (PostgreSQL)
- [ ] Client secret hashing (BCrypt)
- [ ] Scope validation
- [ ] Token generation
- [ ] Token storage (Redis)
- [ ] Rate limiting per client

### Refresh Token Flow
- [ ] `/oauth2/token` endpoint (refresh token grant)
- [ ] Refresh token validation
- [ ] Token rotation
- [ ] Token reuse detection
- [ ] Scope downgrade support
- [ ] Token family revocation
- [ ] Grace period (optional)

### Logout Flow
- [x] `/oauth2/logout` endpoint (GET/POST)
- [x] Session cleanup (Redis)
- [x] Session cookie clearing
- [x] Optional token revocation (PostgreSQL)
- [x] Redirect URI handling
- [x] State parameter preservation
- [ ] Redirect URI validation (security)
- [x] Support id_token_hint parameter
- [x] Error handling
- [x] Audit logging

### Common Requirements
- [x] `/oauth2/token` endpoint (handles all grant types)
- [x] `/oauth2/revoke` endpoint
- [x] `/oauth2/introspect` endpoint
- [x] `/oauth2/userinfo` endpoint (OIDC)
- [x] `/oauth2/logout` endpoint (OIDC RP-Initiated Logout)
- [x] `/.well-known/openid-configuration` (OIDC Discovery)
- [x] `/.well-known/jwks.json` (OIDC JWKS)
- [ ] Token validation middleware
- [x] Error handling (RFC 6749 format)
- [ ] Audit logging
- [ ] Monitoring & metrics

---

## 🧪 Testing

### Test Authorization Code Flow

```bash
# Generate PKCE
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -binary -sha256 | openssl base64 | tr -d '=+/')

# Open authorization URL in browser
open "http://localhost:8080/oauth2/auth?client_id=xxx&redirect_uri=http://localhost:3000/callback&response_type=code&scope=openid+profile+email+offline_access&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256&state=$(openssl rand -hex 16)"

# After login, exchange code for tokens
curl -X POST http://localhost:8080/oauth2/token \
  -u "client_id:client_secret" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "code_verifier=$CODE_VERIFIER"
```

### Test Client Credentials Flow

```bash
curl -X POST http://localhost:8080/oauth2/token \
  -u "client_id:client_secret" \
  -d "grant_type=client_credentials" \
  -d "scope=api:read api:write"
```

### Test Refresh Token Flow

```bash
curl -X POST http://localhost:8080/oauth2/token \
  -u "client_id:client_secret" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=REFRESH_TOKEN"
```

### Test Logout Flow

```bash
# Simple logout (no redirect)
curl -X GET http://localhost:8080/oauth2/logout \
  -b "session_id=YOUR_SESSION_ID"

# Logout with redirect
curl -X GET "http://localhost:8080/oauth2/logout?post_logout_redirect_uri=http://localhost:3000/goodbye&state=xyz789" \
  -b "session_id=YOUR_SESSION_ID" \
  -L  # Follow redirects

# From browser (recommended)
open "http://localhost:8080/oauth2/logout?post_logout_redirect_uri=http://localhost:3000/"
```

---

## 📚 Additional Resources

### Specifications
- [RFC 6749 - OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 6750 - Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [RFC 7662 - Token Introspection](https://datatracker.ietf.org/doc/html/rfc7662)
- [RFC 7009 - Token Revocation](https://datatracker.ietf.org/doc/html/rfc7009)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)

### Security Best Practices
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [OAuth 2.0 for Browser-Based Apps](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps)
- [OAuth 2.0 for Native Apps](https://datatracker.ietf.org/doc/html/rfc8252)

### Project Documentation
- [Business Documentation](../business/oauth2.md) - Architecture overview
- [Technical Documentation](../tech/fosite.md) - Fosite integration details
- [Implementation Plan](../implementation/oauth2-implementation-plan.md)
- [Middleware Guide](../MIDDLEWARE_USAGE_GUIDE.md)
- [Production Improvements](../OAUTH2_PRODUCTION_READY_IMPROVEMENTS.md)

---

## 🤝 Contributing

Khi update flows hoặc thêm flows mới:

1. Tạo file mới trong `docs/flow/`
2. Follow format của existing flows
3. Include diagrams (ASCII art)
4. Add examples và test cases
5. Update README này với links
6. Add to comparison table

---

## 📧 Support

Nếu có câu hỏi về flows, tham khảo:
1. Documentation này
2. [RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
3. Source code trong `internal/app/handler/v1/oauth2/`
4. Unit tests trong `*_test.go` files
