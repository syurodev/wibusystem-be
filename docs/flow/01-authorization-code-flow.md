# Authorization Code Flow

## Tổng quan

Authorization Code Flow là luồng OAuth2 chuẩn và an toàn nhất, được khuyến nghị cho web applications và mobile applications. Flow này kết hợp với PKCE (Proof Key for Code Exchange) để tăng cường bảo mật.

---

## Actors

- **User**: Người dùng cuối sử dụng browser
- **Client App**: Ứng dụng third-party (React/Mobile app) muốn truy cập tài nguyên
- **Authorization Server**: OAuth2 server của bạn (Gin + Fosite)
- **Redis**: Lưu trữ temporary data (auth request, session, authorization code)
- **PostgreSQL**: Lưu trữ persistent data (user, client, consent)

---

## Flow Diagram

```
┌──────┐                                                          ┌─────────────┐
│ User │                                                          │ Client App  │
└──┬───┘                                                          └──────┬──────┘
   │                                                                     │
   │  1. Click "Login"                                                  │
   │ ◄──────────────────────────────────────────────────────────────────┤
   │                                                                     │
   │                                                                     │
   │  2. Generate PKCE (code_verifier, code_challenge)                  │
   │                                                                     │
   │                                                                     │
   │  3. Redirect to /oauth2/auth                                       │
   │ ────────────────────────────────────────────────────────────────► │
   │                                                                     │
   ▼                                                                     │
┌──────────────────────────────────────────────────────────────────────┴──────┐
│                        Authorization Server (Gin)                            │
└──────────────────────────────────────────────────────────────────────────────┘
   │                                │                              │
   │  4. GET /oauth2/auth           │                              │
   │  (client_id, scope,            │                              │
   │   code_challenge, ...)         │                              │
   │                                │                              │
   ├─► Parse request (Fosite)       │                              │
   │                                │                              │
   │  5. Validate client_id         │                              │
   ├────────────────────────────────┼─────────────────────────────►│
   │                                │                   ┌──────────▼─────────┐
   │                                │                   │    PostgreSQL      │
   │                                │                   │  oauth2_clients    │
   │  6. Check session cookie       │                   └────────────────────┘
   │  → User chưa đăng nhập         │                              │
   │                                │                              │
   │  7. Save auth request to Redis │                              │
   ├────────────────────────────────►                              │
   │                                ┌──────────────────▼──────────┐│
   │                                │        Redis                ││
   │                                │  auth_request:{request_id}  ││
   │                                │  TTL: 10 minutes            ││
   │                                └─────────────────────────────┘│
   │                                                                │
   │  8. Redirect to /oauth2/login?request_id=xxx                  │
   │                                                                │
   ▼                                                                │
┌────────────────────────────────────────────────────────────────┐ │
│                   Login Page (Server-rendered HTML)            │ │
│  ┌──────────────────────────────────────────────────────────┐  │ │
│  │  Sign In                                                  │  │ │
│  │                                                           │  │ │
│  │  Email:    [________________]                            │  │ │
│  │  Password: [________________]                            │  │ │
│  │                                                           │  │ │
│  │             [  Sign In  ]                                │  │ │
│  └──────────────────────────────────────────────────────────┘  │ │
└────────────────────────────────────────────────────────────────┘ │
   │                                                                │
   │  9. User enters email/password                                │
   │  POST /oauth2/login                                           │
   │  (request_id, email, password)                                │
   │                                                                │
   ├─► Validate credentials        │                               │
   │                                │                               │
   │  10. Query user by email       │                               │
   ├────────────────────────────────┼───────────────────────────────►
   │                                │                   ┌───────────▼────────┐
   │                                │                   │    PostgreSQL      │
   │                                │                   │   identify.users   │
   │                                │                   └────────────────────┘
   │  11. Verify password hash      │                               │
   │                                │                               │
   │  12. Create user session       │                               │
   ├────────────────────────────────►                               │
   │                                ┌───────────────────▼──────────┐│
   │                                │        Redis                 ││
   │                                │  session:{session_id}        ││
   │                                │  Value: user_id              ││
   │                                │  TTL: configurable           ││
   │                                └──────────────────────────────┘│
   │                                                                │
   │  13. Set cookie: session_id                                   │
   │                                                                │
   │  14. Redirect to /oauth2/auth (original request)              │
   │                                                                │
   ├─► GET /oauth2/auth (lần 2)    │                               │
   │                                │                               │
   │  15. Check session cookie      │                               │
   │  → User ĐÃ đăng nhập ✓         │                               │
   │                                │                               │
   │  16. Check consent             │                               │
   ├────────────────────────────────┼───────────────────────────────►
   │                                │                   ┌───────────▼────────┐
   │                                │                   │    PostgreSQL      │
   │                                │                   │ oauth2_consents    │
   │  → User CHƯA consent           │                   └────────────────────┘
   │                                │                               │
   │  17. Save auth request + user_id to Redis                     │
   ├────────────────────────────────►                               │
   │                                                                │
   │  18. Redirect to /oauth2/consent?request_id=xxx               │
   │                                                                │
   ▼                                                                │
┌────────────────────────────────────────────────────────────────┐ │
│                 Consent Page (Server-rendered HTML)            │ │
│  ┌──────────────────────────────────────────────────────────┐  │ │
│  │  Authorization Request                                    │  │ │
│  │                                                           │  │ │
│  │  "My App" wants to access:                               │  │ │
│  │   ☑ Read your profile                                    │  │ │
│  │   ☑ Read your email                                      │  │ │
│  │   ☑ Keep you signed in                                   │  │ │
│  │                                                           │  │ │
│  │   [  Cancel  ]        [  Allow  ]                        │  │ │
│  └──────────────────────────────────────────────────────────┘  │ │
└────────────────────────────────────────────────────────────────┘ │
   │                                                                │
   │  19. User clicks "Allow"                                      │
   │  POST /oauth2/consent                                         │
   │  (request_id, granted_scopes)                                 │
   │                                                                │
   │  20. Save consent to database   │                             │
   ├────────────────────────────────┼─────────────────────────────►│
   │                                │                   ┌──────────▼─────────┐
   │                                │                   │    PostgreSQL      │
   │                                │                   │ oauth2_consents    │
   │                                │                   └────────────────────┘
   │                                │                               │
   │  21. Redirect to /oauth2/auth (lần 3)                         │
   │                                                                │
   ├─► GET /oauth2/auth (lần 3)    │                               │
   │                                │                               │
   │  22. Check session → ✓         │                               │
   │  23. Check consent → ✓         │                               │
   │                                │                               │
   │  24. Generate authorization code                              │
   │  (Fosite generates unique code)                               │
   │                                │                               │
   │  25. Save code session to Redis                               │
   ├────────────────────────────────►                               │
   │                                ┌───────────────────▼──────────┐│
   │                                │        Redis                 ││
   │                                │  auth_code:{signature}       ││
   │                                │  Contains:                   ││
   │                                │   - user_id                  ││
   │                                │   - client_id                ││
   │                                │   - code_challenge           ││
   │                                │   - granted_scopes           ││
   │                                │  TTL: 60 seconds             ││
   │                                └──────────────────────────────┘│
   │                                                                │
   │  26. Redirect back to client                                  │
   │  Location: http://localhost:3000/callback?code=ABC123&state=xyz
   │                                                                │
   ▼                                                                │
┌──────────────────────────────────────────────────────────────────┴──────┐
│                            Client App                                    │
└──────────────────────────────────────────────────────────────────────────┘
   │
   │  27. Parse authorization code from URL
   │
   │  28. Retrieve code_verifier from localStorage
   │
   │  29. POST /oauth2/token (Backend-to-Backend)
   │      Authorization: Basic base64(client_id:client_secret)
   │      Body:
   │        grant_type=authorization_code
   │        code=ABC123
   │        redirect_uri=http://localhost:3000/callback
   │        code_verifier=...
   │
   ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        Authorization Server (Gin)                        │
└──────────────────────────────────────────────────────────────────────────┘
   │                                │                              │
   │  30. Validate client credentials                              │
   ├────────────────────────────────┼─────────────────────────────►│
   │                                │                   ┌──────────▼─────────┐
   │                                │                   │    PostgreSQL      │
   │                                │                   │  oauth2_clients    │
   │                                │                   └────────────────────┘
   │  31. Retrieve auth code session                               │
   ├────────────────────────────────►                              │
   │                                ┌──────────────────▼──────────┐│
   │                                │        Redis                ││
   │                                │  auth_code:{signature}      ││
   │                                └─────────────────────────────┘│
   │                                                                │
   │  32. Validate code_verifier against code_challenge            │
   │      SHA256(code_verifier) == code_challenge ?                │
   │                                                                │
   │  33. Invalidate authorization code (delete from Redis)        │
   │      → Code chỉ dùng 1 lần                                    │
   │                                                                │
   │  34. Generate tokens:                                         │
   │      - access_token (JWT, expires in 1 hour)                  │
   │      - refresh_token (opaque, expires in 30 days)             │
   │      - id_token (JWT, OIDC)                                   │
   │                                                                │
   │  35. Save token sessions to Redis                             │
   ├────────────────────────────────►                              │
   │                                ┌──────────────────▼──────────┐│
   │                                │        Redis                ││
   │                                │  access_token:{sig}         ││
   │                                │  refresh_token:{sig}        ││
   │                                └─────────────────────────────┘│
   │                                                                │
   │  36. Return token response                                    │
   │  {                                                            │
   │    "access_token": "eyJhbGc...",                              │
   │    "token_type": "Bearer",                                    │
   │    "expires_in": 3600,                                        │
   │    "refresh_token": "eyJhbGc...",                             │
   │    "id_token": "eyJhbGc...",                                  │
   │    "scope": "openid profile email offline_access"            │
   │  }                                                            │
   │                                                                │
   ▼                                                                │
┌──────────────────────────────────────────────────────────────────┴──────┐
│                            Client App                                    │
└──────────────────────────────────────────────────────────────────────────┘
   │
   │  37. Store tokens securely
   │      - localStorage / secure cookies
   │
   │  38. Cleanup temporary data
   │      - Remove code_verifier
   │      - Remove state
   │
   │  39. Redirect user to dashboard
   │
   ▼
   Success!
```

---

## Chi tiết từng bước

### **Phase 1: Initialization (Client App)**

#### Bước 1-2: Generate PKCE
Client app tạo PKCE để bảo vệ authorization code khỏi bị đánh cắp:
- `code_verifier`: Chuỗi ngẫu nhiên 43-128 ký tự
- `code_challenge`: SHA256(code_verifier) được base64url encode

#### Bước 3: Redirect to Authorization Endpoint
Client chuyển hướng user browser đến `/oauth2/auth` với parameters:
- `client_id`: ID của client app
- `redirect_uri`: URL để nhận authorization code
- `response_type`: "code"
- `scope`: Các quyền được yêu cầu (openid, profile, email, etc.)
- `state`: Random string để chống CSRF
- `code_challenge`: PKCE challenge
- `code_challenge_method`: "S256" (SHA256)

---

### **Phase 2: Authentication (Authorization Server)**

#### Bước 4-6: Validate Request & Check Authentication
- Server parse và validate authorization request
- Kiểm tra `client_id` có tồn tại trong database
- Kiểm tra `redirect_uri` có được đăng ký với client
- Kiểm tra session cookie → User chưa đăng nhập

#### Bước 7-8: Store Request & Redirect to Login
- Lưu toàn bộ authorization request vào Redis với key là `request_id`
- TTL: 10 phút
- Redirect user đến `/oauth2/login?request_id=xxx`

#### Bước 9-14: Login Process
**Login page được render từ server** (không phải JSON API):
- Server render HTML form với `request_id` hidden field
- User nhập email/password
- POST đến `/oauth2/login`
- Server validate credentials với database
- Tạo session trong Redis với TTL configurable
- Set cookie `session_id` trong response
- Redirect về `/oauth2/auth` để tiếp tục flow

---

### **Phase 3: Consent (Authorization Server)**

#### Bước 15-18: Check Consent
- Server kiểm tra session cookie → User đã đăng nhập ✓
- Kiểm tra database xem user đã consent cho client + scopes này chưa
- Nếu chưa: lưu request + user_id vào Redis, redirect đến consent page

#### Bước 19-21: Grant Consent
**Consent page được render từ server**:
- Hiển thị tên client và các quyền được yêu cầu
- User click "Allow" → POST đến `/oauth2/consent`
- Server lưu consent vào PostgreSQL
- Redirect về `/oauth2/auth` lần 3

---

### **Phase 4: Authorization Code Generation (Authorization Server)**

#### Bước 22-26: Generate & Return Code
- Kiểm tra lại session ✓ và consent ✓
- Fosite tạo authorization code duy nhất
- Lưu code session vào Redis bao gồm:
  - `user_id`
  - `client_id`
  - `code_challenge` (để verify sau)
  - `granted_scopes`
  - TTL: 60 giây
- Redirect về client callback URL với code

---

### **Phase 5: Token Exchange (Client App + Authorization Server)**

#### Bước 27-29: Exchange Code for Tokens
Client gửi POST request đến `/oauth2/token`:
- **Backend-to-backend call** (không qua browser)
- Client credentials trong `Authorization` header
- Body chứa:
  - `grant_type`: "authorization_code"
  - `code`: Authorization code nhận được
  - `redirect_uri`: Phải match với request ban đầu
  - `code_verifier`: Để verify PKCE

#### Bước 30-36: Validate & Issue Tokens
Server thực hiện các validation:
1. Verify client credentials (client_id + client_secret)
2. Retrieve authorization code session từ Redis
3. **Validate PKCE**: `SHA256(code_verifier) == code_challenge`
4. Check code chưa expired (< 60s)
5. **Invalidate code** (xóa khỏi Redis) - mỗi code chỉ dùng 1 lần
6. Generate tokens:
   - **Access Token**: JWT ký bằng HMAC/RSA, expires 1 hour
   - **Refresh Token**: Opaque token, expires 30 days
   - **ID Token**: JWT chứa user claims (OIDC)
7. Save token sessions vào Redis
8. Return token response

#### Bước 37-39: Store Tokens & Complete
Client app:
- Lưu tokens an toàn (localStorage hoặc httpOnly cookies)
- Xóa dữ liệu tạm (`code_verifier`, `state`)
- Redirect user đến dashboard

---

## Security Features

### 1. **PKCE (Proof Key for Code Exchange)**
- Bảo vệ authorization code khỏi bị intercept
- Ngay cả khi code bị đánh cắp, attacker không thể đổi lấy token vì không có `code_verifier`

### 2. **State Parameter**
- Chống CSRF attacks
- Client verify state nhận được match với state đã gửi

### 3. **Single-Use Authorization Code**
- Code bị invalidate ngay sau khi dùng
- Chống replay attacks

### 4. **Short-Lived Tokens**
- Access token expires sau 1 giờ
- Giảm thiểu rủi ro nếu token bị lộ

### 5. **Client Authentication**
- Client secret không bao giờ expose ra browser
- Token exchange diễn ra backend-to-backend

### 6. **Redirect URI Validation**
- Server chỉ redirect về URI đã đăng ký
- Chống authorization code injection

---

## Data Storage Strategy

### Redis (Temporary Data)
- **Authorization Request**: 10 minutes TTL
- **User Session**: Configurable TTL
- **Authorization Code**: 60 seconds TTL
- **Access Token Session**: Match token expiry
- **Refresh Token Session**: 30 days TTL

### PostgreSQL (Persistent Data)
- **Users**: User accounts và credentials
- **OAuth2 Clients**: Client registrations
- **Consents**: User consent records (permanent)
- **RBAC**: Roles và permissions

---

## Error Handling

### Common Errors

| Error | Code | Khi nào xảy ra |
|-------|------|----------------|
| `invalid_request` | 400 | Missing/invalid parameters |
| `invalid_client` | 401 | Client authentication failed |
| `invalid_grant` | 400 | Invalid/expired authorization code |
| `unauthorized_client` | 400 | Client không được phép dùng grant type này |
| `access_denied` | 403 | User từ chối consent |
| `server_error` | 500 | Internal server error |

### Error Response Format (RFC 6749)
```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code has expired",
  "error_hint": "The authorization code is only valid for 60 seconds"
}
```

---

## Performance Considerations

### Bottlenecks và Solutions

1. **Database Lookups**
   - ✅ Client info được cache trong Redis
   - ✅ Session validation chỉ hit Redis, không hit PostgreSQL
   - ✅ Consent check có index trên (user_id, client_id)

2. **Token Generation**
   - ✅ JWT signing dùng HMAC (faster than RSA) cho access token
   - ✅ RSA chỉ dùng cho ID token (ít frequent hơn)

3. **Redis Load**
   - ✅ Appropriate TTLs tự động cleanup expired data
   - ✅ Keys có prefix để dễ quản lý

---

## Monitoring Points

### Metrics cần track
- Authorization request rate
- Login success/failure rate
- Token generation rate
- Authorization code expiration rate
- Average flow completion time
- Redis hit/miss ratio

### Logs cần ghi
- Authorization request với client_id và scope
- Login attempts (success/failure)
- Consent grants
- Token issuance
- Token validation failures
- Suspicious activities (code reuse, invalid state, etc.)

---

## Variants

### Authorization Code Flow without PKCE
- Không khuyến nghị cho modern applications
- Chỉ dùng cho legacy clients không support PKCE

### Implicit Flow (Deprecated)
- Token trả về trực tiếp trong redirect URL
- Không còn được khuyến nghị sử dụng
- Thay thế bằng Authorization Code Flow + PKCE

---

## References

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
