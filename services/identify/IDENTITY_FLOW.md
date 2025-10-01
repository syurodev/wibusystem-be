# Identity Service – End-to-End Flow (Admin Bootstrap)

This guide walks through running the Identity Service from scratch when you only have one admin account, then creating OAuth2 clients and obtaining admin tokens to manage Dynamic Client Registration (DCR).

Contents
- Prerequisites
- Start databases and Redis
- Run the Identity Service
- Verify OIDC discovery surfaces
- Admin account (seed) and login
- Obtain ADMIN_ACCESS_TOKEN
  - Option A: client_credentials (admin-cli)
  - Option B: authorization_code + PKCE (spa-client)
- Issue Initial Access Token (IAT) for DCR
- Register OAuth2 clients (DCR)
- Manage registered clients (RAT)
- Common OAuth2/OIDC endpoints
- Troubleshooting
- Security notes

## Prerequisites
- Go toolchain installed (Go 1.22+ recommended).
- Docker available (for Postgres, TimescaleDB, Redis via `docker-compose`).
- Ports available locally:
  - Postgres: 5432 (identify), 5433, 5434, 5435 (other services)
  - Identity HTTP: 8080

## Start databases and Redis
From repository root:

```bash
docker compose up -d identify-db redis
```

Notes
- Identity DB container name: `identify-db`, exposed on `localhost:5432`.
- Default DB creds in this repo (dev only):
  - DB name: `identify_service`
  - User: `identify_service`
  - Password: `c2f960ed5f802b6acac2d4d928f21ada`

## Run the Identity Service
Identity Service defaults are set for local dev. Optional overrides:

```bash
export OAUTH2_ISSUER="http://localhost:8080"
export ENVIRONMENT="development"
# DB_* are already defaulted to localhost:5432 and above creds
```

Run the service (from repo root):
```bash
  openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out oidc-signing.pem
```

```bash
go run ./services/identify
```

Behavior
- On startup, the service runs migrations from `./db-init/account-init.sql`.
- A default admin user is seeded by `003_seed_data.up.sql`.
- OAuth2/OIDC endpoints are exposed under `http://localhost:8080`.

## Verify OIDC discovery surfaces

```bash
curl http://localhost:8080/.well-known/openid_configuration | jq .
curl http://localhost:8080/.well-known/jwks.json | jq .
```

You should see issuer, authorization endpoint, token endpoint, scopes, and JWKS.

## Admin account (seed) and login
Seeded admin user (dev only — change in production):
- Email: `syuro.dev@gmail.com`
- Password: `Vv19082001@#`

Login to establish a session cookie (used by the authorize flow during dev):

```bash
curl -i -c cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"email":"syuro.dev@gmail.com","password":"Vv19082001@#"}' \
  http://localhost:8080/api/v1/auth/login
```

If successful, `cookies.txt` is created with a session cookie.

## Obtain ADMIN_ACCESS_TOKEN
There are two supported paths to obtain an access token with `admin` scope.

### Option A: client_credentials (admin-cli)
Create a confidential client that has the `admin` scope and allows client_credentials.

1) Connect to Postgres (identify DB):

```bash
psql "host=localhost port=5432 dbname=identify_service user=identify_service password=c2f960ed5f802b6acac2d4d928f21ada"
```

2) Generate a bcrypt hash for your client secret:

```sql
SELECT crypt('db953e2a62fd4c6aa4599e4087d8dec3cab5f5b8badcfbd42dd4d0d6bfacc2f5', gen_salt('bf', 12));
-- Copy the resulting hash
```

3) Create the client:

```sql
INSERT INTO oauth2_clients (
  id, client_secret_hash, redirect_uris, grant_types, response_types,
  scopes, audience, public, client_name, token_endpoint_auth_method
) VALUES (
  'wibutime_web_client',
  '$2a$12$fYfdxMRUzZAaMAraOq6xS.h0qD4JuPilURg4kavI4gJPFpRlcCrjq',
  ARRAY['http://localhost:3000/api/auth/callback/oidc'],
  ARRAY['authorization_code','refresh_token'],
  ARRAY['code'],
  ARRAY['openid','profile','email','offline_access'],
  ARRAY[]::text[],
  FALSE,
  'Wibutime Web Client',
  'client_secret_basic'
);
```

4) Request an admin access token using client_credentials:

```bash
curl -u admin-cli:30b9ece359121606c0cbd18daa2f9163ff643bc642f2a072475c366ad3f9c214 \
  -d 'grant_type=client_credentials&scope=admin' \
  http://localhost:8080/oauth2/token
```

Response contains `access_token` with `admin` scope. Use it as `ADMIN_ACCESS_TOKEN`.

### Option B: authorization_code + PKCE (spa-client)
Use the seeded public SPA client (dev) and request the `admin` scope.

1) Ensure `spa-client` has the `admin` scope (one-time):

```sql
UPDATE oauth2_clients
SET scopes = array_append(scopes, 'admin')
WHERE id = 'spa-client' AND NOT ('admin' = ANY(scopes));
```

2) Prepare PKCE values:
- `CODE_VERIFIER`: random URL-safe string (43–128 chars)
- `CODE_CHALLENGE`: base64url(SHA256(CODE_VERIFIER)) without padding

3) Start authorization in browser (or curl follow redirects), ensure you’re logged in (cookies.txt):

```bash
AUTH_URL="http://localhost:8080/oauth2/authorize?response_type=code&client_id=spa-client&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email%20offline_access%20admin&state=xyz&code_challenge=${CODE_CHALLENGE}&code_challenge_method=S256"

curl -L -b cookies.txt -o /dev/null -w "%{url_effective}\n" "$AUTH_URL"
# The final URL contains ?code=...
```

4) Exchange code for tokens (no client secret for public client):

```bash
curl -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=<CODE>&redirect_uri=http://localhost:3000/callback&client_id=spa-client&code_verifier=${CODE_VERIFIER}" \
  http://localhost:8080/oauth2/token
```

Response includes an `access_token` that contains `admin` if requested.

## Issue Initial Access Token (IAT) for DCR
With `ADMIN_ACCESS_TOKEN`, create an Initial Access Token used to register clients dynamically:

```bash
curl -X POST http://localhost:8080/admin/registration/iat \
  -H "Authorization: Bearer ${ADMIN_ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"expires_in_seconds": 86400, "description": "bootstrap DCR"}'
```

Response contains `token` (IAT). Keep it safe.

## Register OAuth2 clients (DCR)
Use the IAT as a Bearer token for the registration endpoint.

Confidential Web App example:

```bash
curl -X POST http://localhost:8080/register \
  -H "Authorization: Bearer <IAT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My Web",
    "redirect_uris": ["http://localhost:3000/callback"],
    "grant_types": ["authorization_code","refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile email offline_access",
    "token_endpoint_auth_method": "client_secret_basic"
  }'
```

SPA / Mobile (public + PKCE) example:

```bash
curl -X POST http://localhost:8080/register \
  -H "Authorization: Bearer <IAT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My SPA",
    "redirect_uris": ["http://localhost:3000/callback"],
    "grant_types": ["authorization_code","refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile email offline_access",
    "token_endpoint_auth_method": "none"
  }'
```

Response contains `client_id`, `client_secret` (if confidential), and a `registration_access_token` (RAT) with `registration_client_uri`.

## Manage registered clients (RAT)
Use the `registration_access_token` to read/update/delete the client:

Get:
```bash
curl -H "Authorization: Bearer <RAT>" \
  http://localhost:8080/register/<client_id>
```

Update (example changing redirect URIs):
```bash
curl -X PUT -H "Authorization: Bearer <RAT>" -H "Content-Type: application/json" \
  -d '{
    "client_name": "My Web (updated)",
    "redirect_uris": ["http://localhost:3000/callback","http://localhost:3000/alt-callback"],
    "grant_types": ["authorization_code","refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile email offline_access",
    "token_endpoint_auth_method": "client_secret_basic"
  }' \
  http://localhost:8080/register/<client_id>
```

Delete:
```bash
curl -X DELETE -H "Authorization: Bearer <RAT>" \
  http://localhost:8080/register/<client_id>
```

## Common OAuth2/OIDC endpoints
- Authorize: `GET/POST /oauth2/authorize` (Authorization Code, Implicit, Hybrid; PKCE enforced for public clients)
- Token: `POST /oauth2/token` (authorization_code, refresh_token, client_credentials)
- Introspect: `POST /oauth2/introspect`
- Revoke: `POST /oauth2/revoke`
- UserInfo: `GET /api/v1/userinfo` (requires `openid` scope)
- OIDC Discovery: `GET /.well-known/openid_configuration`
- JWKS: `GET /.well-known/jwks.json`

Example: Token introspection
```bash
curl -X POST http://localhost:8080/oauth2/introspect \
  -u admin-cli:your-super-secret \
  -d "token=<ACCESS_TOKEN>"
```

## Troubleshooting
- 401/invalid_token during DCR: ensure you used IAT for `/register`, or RAT for `/register/:client_id`.
- 403/insufficient_scope on admin routes: your token lacks `admin`. Use Option A or B to include `admin` scope.
- Redirect URI mismatch: your `redirect_uri` in the authorize/token request must exactly match the client’s registered URIs.
- PKCE errors: verify `code_challenge_method=S256` and `code_verifier` pairs correctly.
- DB connection: verify `identify-db` is healthy and the service can reach `localhost:5432`.

## Security notes
- Rotate and secure all secrets; never store plaintext client secrets. This repo uses bcrypt-hashed client secrets in DB.
- In production, set strong values for:
  - `JWT_SECRET`, `OAUTH2_JWT_SIGNING_KEY`, `DCR_REG_ACCESS_TOKEN_SECRET`
  - Database credentials and SSL (`DB_SSL_MODE`), CORS policy, rate limits
- Consent: dev mode may auto-approve. In production, require explicit consent.
- Limit admin endpoints by auth scope and optionally network allowlists.
