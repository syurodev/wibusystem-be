# Client Credentials Flow

## Tổng quan

Client Credentials Flow là luồng OAuth2 dành cho **server-to-server authentication**, không có sự tham gia của user. Flow này dùng cho các trường hợp:

- Backend services gọi API của nhau
- Microservices authentication
- Scheduled jobs/cron jobs cần access API
- CLI tools/scripts

**Đặc điểm:**

- ✅ Đơn giản nhất trong các OAuth2 flows
- ✅ Không có user context
- ✅ Token không chứa user information
- ✅ Chỉ dùng client credentials (client_id + client_secret)
- ⚠️ Chỉ dùng cho confidential clients (backend servers)

---

## Actors

- **Client Application**: Backend service/script muốn access API
- **Authorization Server**: OAuth2 server (Gin + Fosite)
- **Redis**: Lưu trữ token sessions
- **PostgreSQL**: Lưu trữ client credentials và permissions

---

## Flow Diagram

```
┌────────────────────┐
│  Client Service    │
│  (Backend/Script)  │
└─────────┬──────────┘
          │
          │  1. POST /oauth2/token
          │     Authorization: Basic base64(client_id:client_secret)
          │     Content-Type: application/x-www-form-urlencoded
          │
          │     grant_type=client_credentials
          │     scope=api:read api:write
          │
          ▼
┌──────────────────────────────────────────────────────────────┐
│              Authorization Server (Gin + Fosite)             │
└──────────────────────────────────────────────────────────────┘
          │                              │
          │  2. Parse Authorization      │
          │     header                   │
          │                              │
          │  3. Decode Base64            │
          │     → client_id              │
          │     → client_secret          │
          │                              │
          │  4. Validate client          │
          ├──────────────────────────────┼─────────────────────►
          │                              │         ┌────────────▼──────────┐
          │                              │         │     PostgreSQL        │
          │                              │         │   oauth2_clients      │
          │                              │         │   - id                │
          │                              │         │   - secret_hash       │
          │                              │         │   - grant_types       │
          │                              │         │   - scopes            │
          │                              │         │   - is_internal       │
          │  5. Verify grant_types       │         └───────────────────────┘
          │     contains                 │                     │
          │     "client_credentials"     │                     │
          │                              │                     │
          │  6. Compare client_secret    │                     │
          │     with secret_hash         │                     │
          │     (BCrypt verify)          │                     │
          │                              │                     │
          │  7. Validate requested       │                     │
          │     scopes                   │                     │
          │     → Check client allowed   │                     │
          │       scopes                 │                     │
          │                              │                     │
          │  8. Generate access_token    │                     │
          │     - Type: JWT              │                     │
          │     - Algorithm: HMAC-SHA256 │                     │
          │     - Expires: 1 hour        │                     │
          │     - Claims:                │                     │
          │       * client_id            │                     │
          │       * scope                │                     │
          │       * exp, iat, jti        │                     │
          │       * iss                  │                     │
          │                              │                     │
          │  9. Save token session       │                     │
          ├──────────────────────────────►                     │
          │                              ┌─────────────────────▼───────┐
          │                              │         Redis               │
          │                              │  access_token:{signature}   │
          │                              │  Value: {                   │
          │                              │    client_id,               │
          │                              │    granted_scopes,          │
          │                              │    expires_at               │
          │                              │  }                          │
          │                              │  TTL: 3600 seconds          │
          │                              └─────────────────────────────┘
          │                              │
          │  10. Return token response   │
          │  {                           │
          │    "access_token": "eyJhbGc...",
          │    "token_type": "Bearer",
          │    "expires_in": 3600,
          │    "scope": "api:read api:write"
          │  }                           │
          │                              │
          ▼                              │
┌────────────────────┐                  │
│  Client Service    │                  │
│                    │                  │
│  11. Store token   │                  │
│      in memory     │                  │
│      or cache      │                  │
└─────────┬──────────┘                  │
          │                              │
          │  12. Use token for API calls │
          │  GET /api/resource           │
          │  Authorization: Bearer eyJhbGc...
          │                              │
          ▼                              │
┌────────────────────────────────────────┴──────────────────┐
│                    Resource Server                        │
│                                                            │
│  13. Validate token:                                      │
│      - Check signature                                    │
│      - Check expiration                                   │
│      - Check revocation (Redis blacklist)                 │
│      - Check scopes                                       │
│                                                            │
│  14. Process request                                      │
│                                                            │
│  15. Return response                                      │
└───────────────────────────────────────────────────────────┘
```

---

## Chi tiết từng bước

### **Bước 1: Token Request**

Client gửi POST request đến `/oauth2/token` với:

**Headers:**

```
Authorization: Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ=
Content-Type: application/x-www-form-urlencoded
```

**Body:**

```
grant_type=client_credentials
scope=api:read api:write
```

**Authorization Header Format:**

```
Basic base64_encode(client_id + ":" + client_secret)
```

---

### **Bước 2-3: Parse Client Credentials**

Server parse `Authorization` header:

1. Check header starts with "Basic "
2. Decode Base64 string
3. Split by ":" to get client_id và client_secret

---

### **Bước 4-6: Validate Client**

Server thực hiện các validation:

#### 4.1. Query Client từ Database

```sql
SELECT id, client_name, secret_hash, grant_types, scopes
FROM oauth2.oauth2_clients
WHERE id = $1 AND active = true;
```

Kiểm tra:

- ✅ Client tồn tại
- ✅ Client active (không bị disabled)
- ✅ Client không expired

#### 4.2. Verify Grant Type

```go
// Check grant_types contains "client_credentials"
if !contains(client.GrantTypes, "client_credentials") {
    return error("unauthorized_client")
}
```

#### 4.3. Verify Client Secret

```go
// Compare provided secret with hashed secret
err := bcrypt.CompareHashAndPassword(
    client.SecretHash,
    []byte(providedSecret)
)
if err != nil {
    return error("invalid_client")
}
```

**Security Note:** Secret được hash bằng BCrypt với work factor 12.

---

### **Bước 7: Validate Scopes**

Server kiểm tra requested scopes:

```go
// Check mỗi requested scope có trong allowed scopes của client
for _, requestedScope := range requestedScopes {
    if !contains(client.AllowedScopes, requestedScope) {
        return error("invalid_scope")
    }
}
```

**Scope Validation Rules:**

- ✅ Client chỉ được request scopes đã được cấp phép
- ✅ Nếu không request scope nào → grant default scopes
- ✅ Invalid scope → reject toàn bộ request

---

### **Bước 8: Generate Access Token**

Server tạo JWT access token:

#### Token Structure

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "iss": "http://localhost:8080",
    "sub": "10000000-0000-0000-0000-000000000004",
    "aud": ["api"],
    "exp": 1699999999,
    "iat": 1699996399,
    "jti": "unique-token-id",
    "client_id": "10000000-0000-0000-0000-000000000004",
    "scope": "api:read api:write"
  }
}
```

#### Claims Explanation

- `iss` (Issuer): Authorization server URL
- `sub` (Subject): Client ID (không phải user ID!)
- `aud` (Audience): Resource servers có thể accept token này
- `exp` (Expiration): Unix timestamp khi token hết hạn
- `iat` (Issued At): Unix timestamp khi token được issue
- `jti` (JWT ID): Unique identifier cho token
- `client_id`: ID của client
- `scope`: Các scopes được grant

#### Token Signing

```go
// Sign JWT with HMAC-SHA256
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken, err := token.SignedString([]byte(secretKey))
```

---

### **Bước 9: Save Token Session**

Server lưu token session vào Redis:

```redis
SET access_token:{signature} {
    "client_id": "10000000-0000-0000-0000-000000000004",
    "granted_scopes": ["api:read", "api:write"],
    "expires_at": "2024-11-15T10:00:00Z",
    "active": true
}
EX 3600
```

**Key Components:**

- **Key**: `access_token:` prefix + token signature
- **Value**: JSON với token metadata
- **TTL**: Match với token expiration (3600 seconds)

**Purpose:**

- Fast token introspection
- Immediate revocation capability
- Tracking active tokens

---

### **Bước 10: Return Token Response**

Server trả về token theo RFC 6749:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

**Response Fields:**

- `access_token`: JWT token string
- `token_type`: Luôn là "Bearer"
- `expires_in`: Seconds until expiration
- `scope`: Space-separated granted scopes

**Note:** Client Credentials Flow **KHÔNG** có refresh_token vì:

- Không có user session cần maintain
- Client có thể request token mới bất cứ lúc nào
- Token thường short-lived (1 hour)

---

### **Bước 11: Store Token (Client Side)**

Client lưu token:

#### In-Memory (Recommended)

```go
// Lưu trong RAM, tự động mất khi process restart
var accessToken string
var tokenExpiry time.Time

func getToken() string {
    if time.Now().After(tokenExpiry) {
        // Refresh token
        accessToken, tokenExpiry = requestNewToken()
    }
    return accessToken
}
```

#### Cache/Redis (For Distributed Systems)

```go
// Share token across multiple instances
redis.Set("service_token", accessToken, tokenExpiry)
```

**Security Best Practices:**

- ⚠️ Never log token values
- ⚠️ Never commit tokens to source control
- ✅ Store in environment variables or secret management systems
- ✅ Rotate tokens periodically
- ✅ Use different tokens per environment (dev/staging/prod)

---

### **Bước 12-15: Use Token**

Client sử dụng token để gọi APIs:

```bash
curl -X GET https://api.example.com/resource \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

Resource Server validates token:

1. **Parse JWT** và verify signature
2. **Check expiration** (exp claim)
3. **Check revocation** (query Redis blacklist)
4. **Check scopes** match required permissions
5. Process request hoặc return 401/403

---

## Security Considerations

### 1. **Client Secret Management**

#### Storage

- ✅ Store hashed (BCrypt) trong database
- ✅ Never log client secrets
- ✅ Use secret management systems (Vault, AWS Secrets Manager)

#### Rotation

```sql
-- Regular secret rotation
UPDATE oauth2.oauth2_clients
SET secret_hash = $1, updated_at = NOW()
WHERE id = $2;
```

**Rotation Strategy:**

- Rotate every 90 days
- Rotate immediately nếu nghi ngờ bị compromise
- Support multiple active secrets during rotation period

---

### 2. **Scope-Based Access Control**

#### Scope Naming Convention

```
resource:action
```

Examples:

- `api:read` - Read API resources
- `api:write` - Create/update API resources
- `api:delete` - Delete API resources
- `users:manage` - Manage users
- `reports:generate` - Generate reports

#### Scope Hierarchy

```
api:admin (implies all below)
  ├─ api:write (implies api:read)
  │   └─ api:read
  └─ api:delete
```

---

### 3. **Rate Limiting**

#### Token Request Rate Limiting

```
Limit: 100 requests / hour per client_id
```

**Implementation:**

```redis
INCR rate_limit:token:{client_id}:{hour}
EXPIRE rate_limit:token:{client_id}:{hour} 3600

GET rate_limit:token:{client_id}:{hour}
→ If > 100: return 429 Too Many Requests
```

#### API Rate Limiting per Token

```
Limit: 1000 requests / minute per token
```

---

### 4. **Token Revocation**

#### Immediate Revocation

```redis
SET revoked:access_token:{signature} "revoked" EX 3600
```

Client secret changed/rotated → invalidate ALL tokens:

```redis
-- Add client to revoked list
SET revoked:client:{client_id} "all_tokens" EX 86400

-- Resource server checks:
IF EXISTS revoked:client:{client_id} THEN
    CHECK token issued_at > client.secret_updated_at
END
```

---

### 5. **Audit Logging**

Log mọi token requests:

```json
{
  "event": "token_issued",
  "timestamp": "2024-11-15T10:00:00Z",
  "client_id": "10000000-0000-0000-0000-000000000004",
  "grant_type": "client_credentials",
  "scopes": ["api:read", "api:write"],
  "ip_address": "192.168.1.100",
  "user_agent": "MyService/1.0"
}
```

---

## Common Use Cases

### 1. **Microservices Communication**

Service A cần gọi Service B:

```go
// Service A
func callServiceB() {
    token := getClientCredentialsToken()

    req, _ := http.NewRequest("GET", "http://service-b/api/data", nil)
    req.Header.Set("Authorization", "Bearer " + token)

    resp, _ := http.DefaultClient.Do(req)
    // Process response
}

func getClientCredentialsToken() string {
    // Check cache
    if cachedToken != "" && !isExpired(cachedToken) {
        return cachedToken
    }

    // Request new token
    data := url.Values{
        "grant_type": {"client_credentials"},
        "scope": {"service-b:read"},
    }

    req, _ := http.NewRequest("POST", "http://auth-server/oauth2/token",
        strings.NewReader(data.Encode()))
    req.SetBasicAuth(clientID, clientSecret)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, _ := http.DefaultClient.Do(req)
    var tokenResp TokenResponse
    json.NewDecoder(resp.Body).Decode(&tokenResp)

    // Cache token
    cachedToken = tokenResp.AccessToken
    tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

    return cachedToken
}
```

---

### 2. **Scheduled Jobs/Cron**

```bash
#!/bin/bash
# Daily report generation script

# Get token
TOKEN_RESPONSE=$(curl -s -X POST http://auth-server/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -d "grant_type=client_credentials" \
  -d "scope=reports:generate")

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

# Call API
curl -X POST http://api-server/reports/daily \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"date": "2024-11-15"}'
```

---

### 3. **CLI Tools**

```go
// CLI tool for admin operations
package main

import (
    "github.com/spf13/cobra"
    "golang.org/x/oauth2/clientcredentials"
)

func main() {
    // Configure OAuth2 client credentials
    config := &clientcredentials.Config{
        ClientID:     os.Getenv("CLIENT_ID"),
        ClientSecret: os.Getenv("CLIENT_SECRET"),
        TokenURL:     "http://auth-server/oauth2/token",
        Scopes:       []string{"admin:manage"},
    }

    // Auto-handles token refresh
    httpClient := config.Client(context.Background())

    // Use client for API calls
    rootCmd := &cobra.Command{
        Use: "admin-cli",
        Run: func(cmd *cobra.Command, args []string) {
            resp, _ := httpClient.Get("http://api-server/admin/status")
            // Process response
        },
    }

    rootCmd.Execute()
}
```

---

## Error Handling

### Common Errors

| Error Code            | Description                          | Cause                     | Solution                                   |
| --------------------- | ------------------------------------ | ------------------------- | ------------------------------------------ |
| `invalid_client`      | Client authentication failed         | Wrong client_id/secret    | Verify credentials                         |
| `unauthorized_client` | Client not authorized for grant type | Grant type not allowed    | Enable client_credentials in client config |
| `invalid_scope`       | Requested scope invalid              | Scope not in allowed list | Check client allowed scopes                |
| `server_error`        | Internal server error                | Database/Redis down       | Check server logs                          |

### Error Response Example

```json
{
  "error": "invalid_client",
  "error_description": "Client authentication failed",
  "error_hint": "Verify your client_id and client_secret are correct"
}
```

---

## Performance Optimization

### 1. **Token Caching**

#### Client-Side Caching

```go
type TokenCache struct {
    mu          sync.RWMutex
    accessToken string
    expiresAt   time.Time
}

func (c *TokenCache) GetToken() string {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if time.Now().Before(c.expiresAt.Add(-5 * time.Minute)) {
        return c.accessToken
    }

    // Token expired or near expiry, need refresh
    return ""
}
```

**Benefits:**

- Reduce token requests to auth server
- Lower latency for API calls
- Less load on Redis/database

---

### 2. **Connection Pooling**

```go
// HTTP client with connection pooling
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}
```

---

### 3. **Batch Operations**

Thay vì request token cho mỗi operation:

```go
// ❌ Bad - Multiple token requests
for i := 0; i < 100; i++ {
    token := getToken()  // 100 token requests!
    callAPI(token, data[i])
}

// ✅ Good - Reuse token
token := getToken()  // 1 token request
for i := 0; i < 100; i++ {
    callAPI(token, data[i])
}
```

---

## Monitoring & Metrics

### Key Metrics

1. **Token Request Rate**

   - Requests per second
   - Success rate
   - Error rate by error type

2. **Token Lifetime**

   - Average token usage duration
   - Token reuse count
   - Premature token refresh rate

3. **Client Activity**

   - Active clients count
   - Top clients by request volume
   - Inactive clients

4. **Security Metrics**
   - Failed authentication attempts
   - Invalid scope requests
   - Rate limit hits
   - Revoked tokens usage attempts

### Alerting

Set up alerts for:

- ⚠️ Authentication failure rate > 10%
- ⚠️ Rate limit hits > 100/hour
- 🚨 Revoked token usage attempts
- 🚨 Unusual token request patterns

---

## Comparison with Other Flows

| Feature                | Client Credentials | Authorization Code | Refresh Token |
| ---------------------- | ------------------ | ------------------ | ------------- |
| User context           | ❌ No              | ✅ Yes             | ✅ Yes        |
| Refresh token          | ❌ No              | ✅ Yes             | N/A           |
| PKCE                   | ❌ No              | ✅ Yes             | ❌ No         |
| Client secret required | ✅ Yes             | Depends            | ✅ Yes        |
| Browser redirect       | ❌ No              | ✅ Yes             | ❌ No         |
| Use case               | Service-to-service | User login         | Token refresh |

---

## Best Practices

### ✅ DO

1. **Use HTTPS only** for token endpoints
2. **Rotate client secrets** regularly (every 90 days)
3. **Request minimum scopes** needed
4. **Cache tokens** until near expiry
5. **Log all token operations** for audit
6. **Implement rate limiting** per client
7. **Monitor token usage** patterns
8. **Use environment variables** for credentials
9. **Implement retry logic** with exponential backoff
10. **Validate token** before each API call

### ❌ DON'T

1. **Never log tokens** or client secrets
2. **Don't hardcode credentials** in source code
3. **Don't share client credentials** between environments
4. **Don't request excessive scopes** "just in case"
5. **Don't ignore token expiration**
6. **Don't skip TLS certificate verification**
7. **Don't reuse tokens** after revocation
8. **Don't store tokens** in version control
9. **Don't expose client secrets** in client-side code
10. **Don't ignore rate limit responses**

---

## Testing

### Unit Tests

```go
func TestClientCredentialsFlow(t *testing.T) {
    // Setup mock auth server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/oauth2/token", r.URL.Path)

        // Verify Authorization header
        auth := r.Header.Get("Authorization")
        assert.True(t, strings.HasPrefix(auth, "Basic "))

        // Return mock token
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "access_token": "mock_token",
            "token_type":   "Bearer",
            "expires_in":   3600,
            "scope":        "api:read",
        })
    }))
    defer server.Close()

    // Test client
    token, err := requestToken(server.URL, "client_id", "client_secret", []string{"api:read"})
    assert.NoError(t, err)
    assert.Equal(t, "mock_token", token.AccessToken)
}
```

### Integration Tests

```bash
# Test token request
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth2/token \
  -u "test_client:test_secret" \
  -d "grant_type=client_credentials" \
  -d "scope=api:read")

# Verify response
echo $TOKEN_RESPONSE | jq '.access_token' # Should not be null
echo $TOKEN_RESPONSE | jq '.expires_in'   # Should be 3600

# Test token usage
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')
curl -X GET http://localhost:8080/api/protected \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n"  # Should be 200
```

---

## References

- [RFC 6749 Section 4.4 - Client Credentials Grant](https://datatracker.ietf.org/doc/html/rfc6749#section-4.4)
- [RFC 6750 - Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
