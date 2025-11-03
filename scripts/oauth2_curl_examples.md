# OAuth2 cURL Examples

Quick reference for testing OAuth2 endpoints with cURL.

---

## 🔍 Discovery & Metadata

### OpenID Connect Discovery
```bash
curl http://localhost:8080/.well-known/openid-configuration | jq '.'
```

### JWKS (Public Keys)
```bash
curl http://localhost:8080/.well-known/jwks.json | jq '.'
```

---

## 🔐 Authorization Code Flow

### Step 1: Generate PKCE values
```bash
# Generate code verifier (random 43-char string)
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)
echo "Code Verifier: $CODE_VERIFIER"

# Generate code challenge (SHA256 hash of verifier)
CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -binary -sha256 | openssl base64 | tr -d '=+/' | cut -c1-43)
echo "Code Challenge: $CODE_CHALLENGE"

# Generate state
STATE=$(openssl rand -hex 16)
echo "State: $STATE"
```

### Step 2: Build Authorization URL
```bash
CLIENT_ID="10000000-0000-0000-0000-000000000001"
REDIRECT_URI="http://localhost:3000/callback"

AUTH_URL="http://localhost:8080/oauth2/auth?client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&response_type=code&scope=openid+profile+email+offline_access&state=${STATE}&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256"

echo "Open this URL in browser:"
echo "$AUTH_URL"
```

### Step 3: Login (via browser or curl simulation)
```bash
# This would be done in browser, but for testing:
REQUEST_ID="<from_redirect>"

# Login
curl -X POST http://localhost:8080/oauth2/login \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "request_id=${REQUEST_ID}" \
  -d "email=john@example.com" \
  -d "password=password123" \
  -c cookies.txt \
  -L
```

### Step 4: Exchange Authorization Code for Tokens
```bash
# After getting the code from redirect
AUTH_CODE="<authorization_code_from_redirect>"
CLIENT_ID="10000000-0000-0000-0000-000000000001"
CLIENT_SECRET="test-client-secret"
REDIRECT_URI="http://localhost:3000/callback"

curl -X POST http://localhost:8080/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=${AUTH_CODE}" \
  -d "redirect_uri=${REDIRECT_URI}" \
  -d "code_verifier=${CODE_VERIFIER}" \
  | jq '.'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1...",
  "token_type": "bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1...",
  "id_token": "eyJhbGciOiJSUzI1...",
  "scope": "openid profile email offline_access"
}
```

---

## 👤 UserInfo Endpoint

```bash
ACCESS_TOKEN="<your_access_token>"

curl http://localhost:8080/oauth2/userinfo \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  | jq '.'
```

**Response:**
```json
{
  "sub": "00000000-0000-0000-0000-000000000001",
  "name": "John Doe",
  "email": "john@example.com",
  "email_verified": true,
  "picture": "https://i.pravatar.cc/150?img=1"
}
```

---

## 🔄 Refresh Token

```bash
REFRESH_TOKEN="<your_refresh_token>"
CLIENT_ID="10000000-0000-0000-0000-000000000001"
CLIENT_SECRET="test-client-secret"

curl -X POST http://localhost:8080/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=${REFRESH_TOKEN}" \
  | jq '.'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1... (new)",
  "token_type": "bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJIUzI1... (may be rotated)",
  "scope": "openid profile email offline_access"
}
```

---

## 🤖 Client Credentials Grant

For server-to-server authentication (no user context).

```bash
CLIENT_ID="10000000-0000-0000-0000-000000000004"
CLIENT_SECRET="test-client-secret"

curl -X POST http://localhost:8080/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "scope=api:read+api:write" \
  | jq '.'
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1...",
  "token_type": "bearer",
  "expires_in": 3600,
  "scope": "api:read api:write"
}
```

---

## 🧪 Complete Flow Example

### 1. Setup
```bash
# Export variables
export CLIENT_ID="10000000-0000-0000-0000-000000000001"
export CLIENT_SECRET="test-client-secret"
export REDIRECT_URI="https://oauth.pstmn.io/v1/callback"
export BASE_URL="http://localhost:8080"

# Generate PKCE
export CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)
export CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -binary -sha256 | openssl base64 | tr -d '=+/' | cut -c1-43)
export STATE=$(openssl rand -hex 16)
```

### 2. Get Authorization URL
```bash
cat << EOF
Authorization URL:
${BASE_URL}/oauth2/auth?client_id=${CLIENT_ID}&redirect_uri=${REDIRECT_URI}&response_type=code&scope=openid+profile+email+offline_access&state=${STATE}&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256

Login with:
Email: john@example.com
Password: password123

After redirect, copy the 'code' parameter
EOF
```

### 3. Exchange Code (after manual auth)
```bash
# Set the code you received
export AUTH_CODE="<paste_code_here>"

# Exchange for tokens
TOKENS=$(curl -s -X POST ${BASE_URL}/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=${AUTH_CODE}" \
  -d "redirect_uri=${REDIRECT_URI}" \
  -d "code_verifier=${CODE_VERIFIER}")

echo "$TOKENS" | jq '.'

# Extract tokens
export ACCESS_TOKEN=$(echo "$TOKENS" | jq -r '.access_token')
export REFRESH_TOKEN=$(echo "$TOKENS" | jq -r '.refresh_token')
export ID_TOKEN=$(echo "$TOKENS" | jq -r '.id_token')
```

### 4. Get User Info
```bash
curl -s ${BASE_URL}/oauth2/userinfo \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  | jq '.'
```

### 5. Refresh Token
```bash
NEW_TOKENS=$(curl -s -X POST ${BASE_URL}/oauth2/token \
  -u "${CLIENT_ID}:${CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=${REFRESH_TOKEN}")

echo "$NEW_TOKENS" | jq '.'
export ACCESS_TOKEN=$(echo "$NEW_TOKENS" | jq -r '.access_token')
```

---

## 🔍 Decode JWT Tokens

### Decode ID Token (without verification)
```bash
# Extract payload (2nd part of JWT)
echo "$ID_TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq '.'
```

### Decode Access Token
```bash
echo "$ACCESS_TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq '.'
```

---

## 📝 Test Credentials

### Users
| Email | Password | User ID |
|-------|----------|---------|
| john@example.com | password123 | 00000000-0000-0000-0000-000000000001 |
| jane@example.com | password456 | 00000000-0000-0000-0000-000000000002 |
| test@example.com | test123 | 00000000-0000-0000-0000-000000000003 |

### OAuth2 Clients
| Client | ID | Secret | Type |
|--------|-----|--------|------|
| Web App | 10000000-0000-0000-0000-000000000001 | test-client-secret | Confidential |
| Mobile App | 10000000-0000-0000-0000-000000000002 | (none) | Public |
| SPA | 10000000-0000-0000-0000-000000000003 | (none) | Public |
| API Service | 10000000-0000-0000-0000-000000000004 | test-client-secret | Confidential |

---

## 🛠️ Troubleshooting

### Check Server Status
```bash
curl http://localhost:8080/health
```

### View Recent Logs
```bash
tail -f /tmp/server.log
```

### Clear Session Cookies
```bash
rm -f cookies.txt
```

### Check Redis Sessions
```bash
redis-cli keys "session:*"
redis-cli get "session:<session_id>"
```

### Check Database
```bash
psql -U system_dev -d system_dev -c "SELECT * FROM identify.oauth2_consents;"
```
