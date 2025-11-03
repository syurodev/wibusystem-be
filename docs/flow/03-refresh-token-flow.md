# Refresh Token Flow

## Tổng quan

Refresh Token Flow cho phép client lấy access token mới **mà không cần user đăng nhập lại**. Đây là cơ chế quan trọng để duy trì user session lâu dài trong khi vẫn giữ access token có thời gian sống ngắn (short-lived).

**Đặc điểm:**
- ✅ Duy trì user session lâu dài
- ✅ Không cần user interaction
- ✅ Access token ngắn hạn (1 hour) + Refresh token dài hạn (30 days)
- ✅ Support token rotation (optional)
- ⚠️ Chỉ available khi request scope `offline_access`

---

## Actors

- **User**: Người dùng đã login trước đó
- **Client App**: Application đã có refresh token
- **Authorization Server**: OAuth2 server (Gin + Fosite)
- **Redis**: Lưu trữ token sessions
- **PostgreSQL**: Verify client credentials

---

## Flow Diagram

```
┌──────────────────┐
│   Client App     │
│  (React/Mobile)  │
└────────┬─────────┘
         │
         │  Initial State:
         │  - access_token: expired
         │  - refresh_token: valid
         │
         │  1. Detect access_token expired
         │     (exp claim < current time)
         │
         │  2. POST /oauth2/token
         │     Authorization: Basic base64(client_id:client_secret)
         │     Content-Type: application/x-www-form-urlencoded
         │     
         │     grant_type=refresh_token
         │     refresh_token=eyJhbGc...
         │     scope=openid profile email (optional)
         │
         ▼
┌──────────────────────────────────────────────────────────────┐
│              Authorization Server (Gin + Fosite)             │
└──────────────────────────────────────────────────────────────┘
         │                              │
         │  3. Parse Authorization      │
         │     header                   │
         │                              │
         │  4. Validate client          │
         ├──────────────────────────────┼─────────────────────►
         │                              │         ┌────────────▼──────────┐
         │                              │         │     PostgreSQL        │
         │                              │         │   oauth2_clients      │
         │                              │         └───────────────────────┘
         │                              │                     │
         │  5. Retrieve refresh token   │                     │
         │     session                  │                     │
         ├──────────────────────────────►                     │
         │                              ┌─────────────────────▼───────┐
         │                              │         Redis               │
         │                              │  refresh_token:{signature}  │
         │                              │  Contains:                  │
         │                              │   - user_id                 │
         │                              │   - client_id               │
         │                              │   - granted_scopes          │
         │                              │   - issued_at               │
         │                              │   - expires_at              │
         │                              └─────────────────────────────┘
         │                              │
         │  6. Validate refresh token:  │
         │     ✓ Not expired            │
         │     ✓ Not revoked            │
         │     ✓ Belongs to client      │
         │     ✓ Active = true          │
         │                              │
         │  7. Validate requested       │
         │     scopes ≤ original scopes │
         │                              │
         │  8. Generate new tokens:     │
         │     - New access_token       │
         │     - New refresh_token      │
         │       (token rotation)       │
         │                              │
         │  9. Invalidate old           │
         │     refresh_token            │
         ├──────────────────────────────►
         │                              ┌─────────────────────────────┐
         │                              │         Redis               │
         │  SET refresh_token:{old_sig} │                             │
         │      { active: false }       │  Old token marked inactive  │
         │                              └─────────────────────────────┘
         │                              │
         │  10. Save new token sessions │
         ├──────────────────────────────►
         │                              ┌─────────────────────────────┐
         │                              │         Redis               │
         │                              │  access_token:{new_sig}     │
         │                              │  refresh_token:{new_sig}    │
         │                              └─────────────────────────────┘
         │                              │
         │  11. Return token response   │
         │  {                           │
         │    "access_token": "eyJhbGc...",
         │    "token_type": "Bearer",   │
         │    "expires_in": 3600,       │
         │    "refresh_token": "eyJhbGc...",
         │    "scope": "openid profile email"
         │  }                           │
         │                              │
         ▼                              │
┌──────────────────┐                   │
│   Client App     │                   │
│                  │                   │
│  12. Update      │                   │
│      stored      │                   │
│      tokens      │                   │
│                  │                   │
│  13. Retry       │                   │
│      original    │                   │
│      API call    │                   │
└──────────────────┘                   │
```

---

## Chi tiết từng bước

### **Bước 1: Detect Token Expiration**

Client kiểm tra access token expiration:

```javascript
// Check token expiration
function isTokenExpired(token) {
    if (!token) return true;
    
    try {
        // Decode JWT (without verification)
        const payload = JSON.parse(atob(token.split('.')[1]));
        const expiresAt = payload.exp * 1000; // Convert to milliseconds
        const now = Date.now();
        
        // Check if expired or will expire in next 5 minutes
        return expiresAt < (now + 5 * 60 * 1000);
    } catch (e) {
        return true;
    }
}

// Before API call
if (isTokenExpired(accessToken)) {
    // Refresh token
    await refreshAccessToken();
}
```

**Proactive Refresh Strategy:**
- ✅ Refresh 5 minutes before expiration
- ✅ Avoid "token expired" errors during user activity
- ✅ Better user experience (no interruption)

---

### **Bước 2: Request Token Refresh**

Client gửi refresh token request:

```http
POST /oauth2/token HTTP/1.1
Host: localhost:8080
Authorization: Basic Y2xpZW50X2lkOmNsaWVudF9zZWNyZXQ=
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...&scope=openid%20profile%20email
```

**Request Parameters:**
- `grant_type`: Must be "refresh_token"
- `refresh_token`: The refresh token received earlier
- `scope` (optional): Can request subset of original scopes

**Client Authentication:**
- **Confidential clients**: Require client_secret (Authorization header)
- **Public clients**: No secret needed (client_id in body)

---

### **Bước 3-4: Validate Client**

Server validates client credentials:

```go
// Parse Authorization header
clientID, clientSecret, err := parseBasicAuth(r.Header.Get("Authorization"))

// Query client from database
client, err := store.GetClient(ctx, clientID)

// Verify client secret (if confidential client)
if !client.Public {
    err := bcrypt.CompareHashAndPassword(
        []byte(client.SecretHash),
        []byte(clientSecret),
    )
    if err != nil {
        return ErrInvalidClient
    }
}

// Check grant_types includes "refresh_token"
if !contains(client.GrantTypes, "refresh_token") {
    return ErrUnauthorizedClient
}
```

---

### **Bước 5-6: Validate Refresh Token**

Server retrieves và validates refresh token:

#### 5.1. Retrieve Token Session

```redis
GET refresh_token:{signature}
```

Response:
```json
{
  "user_id": "00000000-0000-0000-0000-000000000001",
  "client_id": "10000000-0000-0000-0000-000000000001",
  "granted_scopes": ["openid", "profile", "email", "offline_access"],
  "issued_at": "2024-11-15T10:00:00Z",
  "expires_at": "2024-12-15T10:00:00Z",
  "active": true
}
```

#### 6.2. Validate Token

**Validation Checks:**

1. **Exists**: Token session found in Redis
   ```go
   if session == nil {
       return fosite.ErrInvalidGrant.WithHint("Refresh token not found")
   }
   ```

2. **Not Expired**: Current time < expires_at
   ```go
   if time.Now().After(session.ExpiresAt) {
       return fosite.ErrInvalidGrant.WithHint("Refresh token expired")
   }
   ```

3. **Active**: Token not revoked
   ```go
   if !session.Active {
       return fosite.ErrInvalidGrant.WithHint("Refresh token revoked")
   }
   ```

4. **Client Match**: Token belongs to requesting client
   ```go
   if session.ClientID != requestingClientID {
       return fosite.ErrInvalidGrant.WithHint("Token belongs to different client")
   }
   ```

5. **Not Blacklisted**: Check global revocation list
   ```redis
   EXISTS revoked:refresh_token:{signature}
   # If exists → token was explicitly revoked
   ```

---

### **Bước 7: Validate Requested Scopes**

Client có thể request subset của original scopes:

```go
// Original scopes from refresh token session
originalScopes := session.GrantedScopes // ["openid", "profile", "email", "offline_access"]

// Requested scopes (from request body)
requestedScopes := parseScopes(r.Form.Get("scope"))

// If no scope requested, grant all original scopes
if len(requestedScopes) == 0 {
    requestedScopes = originalScopes
}

// Validate: requested ⊆ original
for _, scope := range requestedScopes {
    if !contains(originalScopes, scope) {
        return fosite.ErrInvalidScope.WithHintf(
            "Scope '%s' was not granted in original authorization", scope)
    }
}
```

**Scope Downgrade Rules:**
- ✅ Can request fewer scopes than original
- ❌ Cannot request MORE scopes than original
- ✅ If no scope requested → grant all original scopes

**Examples:**

| Original Scopes | Requested Scopes | Result |
|----------------|------------------|--------|
| `openid profile email` | `openid profile` | ✅ Allowed |
| `openid profile` | `openid profile email` | ❌ Invalid scope |
| `openid profile email` | (empty) | ✅ Grant all: `openid profile email` |

---

### **Bước 8: Generate New Tokens**

Server generates new access token và optionally new refresh token:

#### 8.1. Generate New Access Token

```go
// Create new access token
accessToken := &jwt.Token{
    Header: map[string]interface{}{
        "alg": "HS256",
        "typ": "JWT",
    },
    Claims: jwt.MapClaims{
        "iss":      "http://localhost:8080",
        "sub":      session.UserID,
        "aud":      []string{"api"},
        "exp":      time.Now().Add(1 * time.Hour).Unix(),
        "iat":      time.Now().Unix(),
        "jti":      uuid.New().String(),
        "client_id": session.ClientID,
        "scope":    strings.Join(grantedScopes, " "),
    },
}

signedAccessToken, _ := accessToken.SignedString([]byte(secretKey))
```

#### 8.2. Generate New Refresh Token (Token Rotation)

**Token Rotation** là security best practice:

```go
// Generate new refresh token
newRefreshToken := generateOpaqueToken() // Random 32-byte string

// Create new refresh token session
newSession := &RefreshTokenSession{
    UserID:        session.UserID,
    ClientID:      session.ClientID,
    GrantedScopes: grantedScopes,
    IssuedAt:      time.Now(),
    ExpiresAt:     time.Now().Add(30 * 24 * time.Hour), // 30 days
    Active:        true,
}
```

**Why Token Rotation?**
- 🔒 Limits exposure window nếu token bị stolen
- 🔒 Detects token theft (if both old and new tokens used)
- 🔒 Automatic cleanup of unused tokens

**Configuration Options:**

| Strategy | Description | Security | Compatibility |
|----------|-------------|----------|---------------|
| **Rotation** | New refresh token every refresh | ⭐⭐⭐⭐⭐ High | May break some clients |
| **Reuse** | Same refresh token | ⭐⭐⭐ Medium | Better compatibility |
| **Hybrid** | Rotate after N uses or X days | ⭐⭐⭐⭐ High | Good balance |

Project của bạn sử dụng **Rotation** strategy (khuyến nghị).

---

### **Bước 9: Invalidate Old Refresh Token**

Server marks old refresh token as inactive:

```redis
# Update old token session
SET refresh_token:{old_signature} {
    "user_id": "00000000-0000-0000-0000-000000000001",
    "client_id": "10000000-0000-0000-0000-000000000001",
    "granted_scopes": [...],
    "issued_at": "2024-11-15T10:00:00Z",
    "expires_at": "2024-12-15T10:00:00Z",
    "active": false,  # ← Changed to false
    "rotated_at": "2024-11-15T12:00:00Z",
    "rotated_to": "{new_signature}"
}
```

**Grace Period (Optional):**

Để handle race conditions (simultaneous refresh requests):

```go
// Allow old token for 30 seconds after rotation
if !session.Active {
    if session.RotatedAt != nil {
        gracePeriod := 30 * time.Second
        if time.Since(*session.RotatedAt) < gracePeriod {
            // Return the NEW token that was issued
            return cachedNewToken, nil
        }
    }
    return fosite.ErrInvalidGrant.WithHint("Refresh token already used")
}
```

---

### **Bước 10: Save New Token Sessions**

Server saves new token sessions to Redis:

#### Access Token Session
```redis
SET access_token:{new_access_signature} {
    "user_id": "00000000-0000-0000-0000-000000000001",
    "client_id": "10000000-0000-0000-0000-000000000001",
    "granted_scopes": ["openid", "profile", "email"],
    "expires_at": "2024-11-15T13:00:00Z",
    "active": true
}
EX 3600  # 1 hour TTL
```

#### Refresh Token Session
```redis
SET refresh_token:{new_refresh_signature} {
    "user_id": "00000000-0000-0000-0000-000000000001",
    "client_id": "10000000-0000-0000-0000-000000000001",
    "granted_scopes": ["openid", "profile", "email", "offline_access"],
    "issued_at": "2024-11-15T12:00:00Z",
    "expires_at": "2024-12-15T12:00:00Z",
    "active": true,
    "parent_token": "{old_signature}"  # For audit trail
}
EX 2592000  # 30 days TTL
```

---

### **Bước 11: Return Token Response**

Server returns new tokens:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAiLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDEiLCJhdWQiOlsiYXBpIl0sImV4cCI6MTY5OTk5OTk5OSwiaWF0IjoxNjk5OTk2Mzk5LCJqdGkiOiJ1bmlxdWUtdG9rZW4taWQiLCJjbGllbnRfaWQiOiIxMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDEiLCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIn0...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwODAiLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDEiLCJleHAiOjE3MDI1ODgzOTl9...",
  "scope": "openid profile email offline_access"
}
```

**Response Fields:**
- `access_token`: New JWT access token
- `token_type`: Always "Bearer"
- `expires_in`: Seconds until access token expires (3600)
- `refresh_token`: **New** refresh token (rotated)
- `scope`: Granted scopes (may be less than requested)

---

### **Bước 12-13: Client Updates Tokens**

Client replaces old tokens with new ones:

```javascript
async function refreshAccessToken() {
    const refreshToken = localStorage.getItem('refresh_token');
    
    const response = await fetch('http://localhost:8080/oauth2/token', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Authorization': 'Basic ' + btoa(CLIENT_ID + ':' + CLIENT_SECRET)
        },
        body: new URLSearchParams({
            grant_type: 'refresh_token',
            refresh_token: refreshToken
        })
    });
    
    if (!response.ok) {
        // Refresh failed → redirect to login
        redirectToLogin();
        return;
    }
    
    const tokens = await response.json();
    
    // Update stored tokens
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);  // Important: Update refresh token too!
    
    return tokens.access_token;
}

// Usage
async function callAPI() {
    let accessToken = localStorage.getItem('access_token');
    
    // Check expiration
    if (isTokenExpired(accessToken)) {
        accessToken = await refreshAccessToken();
    }
    
    // Make API call
    const response = await fetch('https://api.example.com/data', {
        headers: {
            'Authorization': 'Bearer ' + accessToken
        }
    });
    
    // Handle 401 → retry with refresh
    if (response.status === 401) {
        accessToken = await refreshAccessToken();
        return fetch('https://api.example.com/data', {
            headers: { 'Authorization': 'Bearer ' + accessToken }
        });
    }
    
    return response;
}
```

**Critical:** Client MUST update BOTH access_token and refresh_token!

---

## Security Considerations

### 1. **Token Rotation**

**Benefits:**
- 🔒 Detects token theft
- 🔒 Limits token lifetime
- 🔒 Automatic cleanup

**Detection of Token Theft:**

```
Timeline:
T0: User gets tokens (access + refresh)
T1: Attacker steals refresh_token
T2: User refreshes → gets new tokens (old refresh invalidated)
T3: Attacker tries to use stolen refresh_token → ERROR!
```

When old refresh token used:
```go
if !session.Active && session.RotatedTo != "" {
    // Token was rotated → possible theft detected!
    
    // 1. Log security event
    logSecurityAlert("refresh_token_reuse", session.UserID, session.ClientID)
    
    // 2. Revoke entire token family
    revokeTokenFamily(session.UserID, session.ClientID)
    
    // 3. Notify user
    sendSecurityEmail(session.UserID, "Suspicious activity detected")
    
    return fosite.ErrInvalidGrant.WithHint("Token reuse detected")
}
```

---

### 2. **Refresh Token Binding**

Bind refresh token to specific client:

```go
// During validation
if session.ClientID != request.ClientID {
    return fosite.ErrInvalidGrant.WithHint("Token belongs to different client")
}
```

**Benefits:**
- Stolen refresh token cannot be used by different client
- Even if client_id known, client_secret required

---

### 3. **Expiration Strategy**

| Token Type | Lifetime | Rationale |
|-----------|----------|-----------|
| Access Token | 1 hour | Short enough to limit damage if stolen |
| Refresh Token | 30 days | Long enough for good UX, short enough to enforce re-auth |
| Remember Me | 90 days | Optional extended session |

**Absolute Expiration:**

```go
// Refresh token has maximum lifetime
if time.Now().After(session.ExpiresAt) {
    return fosite.ErrInvalidGrant.WithHint("Refresh token expired")
}
```

**Sliding Expiration (Alternative):**

```go
// Extend refresh token lifetime on each use
session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
```

Project của bạn dùng **Absolute Expiration** (khuyến nghị).

---

### 4. **Revocation Scenarios**

#### User Logout
```go
// Revoke all tokens for user
func RevokeAllUserTokens(userID string) {
    // Add to revocation list
    redis.Set("revoked:user:" + userID, "all", 30*24*time.Hour)
    
    // Delete all sessions
    keys := redis.Keys("*_token:*:user:" + userID)
    redis.Del(keys...)
}
```

#### Password Change
```go
// Revoke all tokens issued before password change
func OnPasswordChange(userID string) {
    user := getUser(userID)
    
    // Update password_changed_at
    user.PasswordChangedAt = time.Now()
    saveUser(user)
    
    // Tokens issued before this time are invalid
    redis.Set("revoked:user:" + userID + ":before", user.PasswordChangedAt, 30*24*time.Hour)
}

// During token validation
if user.PasswordChangedAt.After(session.IssuedAt) {
    return ErrInvalidGrant.WithHint("Password changed, please re-authenticate")
}
```

#### Account Suspension
```go
// Immediate revocation
func SuspendAccount(userID string) {
    // Mark user as suspended
    updateUserStatus(userID, "suspended")
    
    // Revoke all tokens
    RevokeAllUserTokens(userID)
}
```

---

## Error Handling

### Common Errors

| Error | Cause | Client Action |
|-------|-------|---------------|
| `invalid_grant` - "Refresh token expired" | Token lifetime exceeded | Redirect to login |
| `invalid_grant` - "Refresh token revoked" | User logged out / password changed | Redirect to login |
| `invalid_grant` - "Token reuse detected" | Security violation | Redirect to login + alert |
| `invalid_client` | Wrong client credentials | Fix client config |
| `invalid_scope` | Requested scope not originally granted | Request valid scopes |

### Error Response Examples

#### Expired Token
```json
{
  "error": "invalid_grant",
  "error_description": "The provided authorization grant is invalid, expired, or revoked",
  "error_hint": "Refresh token expired. Please re-authenticate."
}
```

#### Token Reuse
```json
{
  "error": "invalid_grant",
  "error_description": "Token reuse detected",
  "error_hint": "This refresh token was already used. All tokens have been revoked for security."
}
```

---

## Best Practices

### Client-Side

#### ✅ DO

1. **Store tokens securely**
   - Web: httpOnly cookies (preferred) or localStorage
   - Mobile: Keychain (iOS) / KeyStore (Android)

2. **Refresh proactively**
   - Refresh 5 minutes before expiration
   - Avoid token expiration during user activity

3. **Handle refresh failures gracefully**
   - Redirect to login on refresh failure
   - Don't retry infinitely

4. **Update both tokens**
   - Replace BOTH access and refresh tokens after refresh

5. **Implement retry logic**
   - Retry API calls with new token after 401

#### ❌ DON'T

1. **Don't expose refresh tokens**
   - Never log refresh tokens
   - Never send in URL parameters

2. **Don't ignore rotation**
   - Must use new refresh token from response

3. **Don't refresh on every request**
   - Check expiration first
   - Use in-memory cache

---

### Server-Side

#### ✅ DO

1. **Implement token rotation**
   - Invalidate old refresh token
   - Issue new refresh token

2. **Detect token reuse**
   - Log security events
   - Revoke token family

3. **Set appropriate TTLs**
   - Access: 1 hour
   - Refresh: 30 days

4. **Implement revocation**
   - User logout
   - Password change
   - Account suspension

5. **Log all refresh operations**
   - User ID, client ID, timestamp
   - Success/failure
   - IP address

#### ❌ DON'T

1. **Don't reuse refresh tokens**
   - Always rotate (unless good reason)

2. **Don't skip validation**
   - Verify client
   - Check expiration
   - Verify not revoked

3. **Don't allow scope expansion**
   - New scopes must be ≤ original

---

## Monitoring

### Key Metrics

1. **Refresh Token Usage**
   - Refresh requests per minute
   - Success rate
   - Average token lifetime before refresh

2. **Security Events**
   - Token reuse attempts
   - Failed refresh attempts
   - Revocation events

3. **User Experience**
   - Average session duration
   - Re-authentication frequency
   - Silent refresh success rate

### Alerts

- 🚨 Token reuse detected
- ⚠️ High refresh failure rate (> 5%)
- ⚠️ Unusual refresh patterns
- ⚠️ High revocation rate

---

## Testing

### Test Cases

```javascript
// Test 1: Normal refresh flow
test('should refresh tokens successfully', async () => {
    const response = await refreshToken(validRefreshToken);
    expect(response.access_token).toBeDefined();
    expect(response.refresh_token).toBeDefined();
    expect(response.refresh_token).not.toEqual(validRefreshToken); // Rotation
});

// Test 2: Expired refresh token
test('should reject expired refresh token', async () => {
    await expect(refreshToken(expiredRefreshToken))
        .rejects.toThrow('invalid_grant');
});

// Test 3: Token reuse detection
test('should detect token reuse', async () => {
    // First refresh
    await refreshToken(validRefreshToken);
    
    // Try to use same token again
    await expect(refreshToken(validRefreshToken))
        .rejects.toThrow('Token reuse detected');
});

// Test 4: Scope downgrade
test('should allow scope downgrade', async () => {
    const response = await refreshToken(validRefreshToken, {
        scope: 'openid profile'  // Original: openid profile email
    });
    expect(response.scope).toEqual('openid profile');
});

// Test 5: Invalid scope expansion
test('should reject scope expansion', async () => {
    await expect(refreshToken(validRefreshToken, {
        scope: 'openid profile email admin'  // admin not in original
    })).rejects.toThrow('invalid_scope');
});
```

---

## References

- [RFC 6749 Section 6 - Refreshing an Access Token](https://datatracker.ietf.org/doc/html/rfc6749#section-6)
- [OAuth 2.0 Security Best Current Practice - Refresh Tokens](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics#section-4.13)
- [OAuth 2.0 for Browser-Based Apps - Refresh Tokens](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps#section-8)
