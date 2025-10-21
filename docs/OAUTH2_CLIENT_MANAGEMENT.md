# OAuth2 Client Management Guide

This guide explains how to create and manage OAuth2 clients in WibuSystem.

## Overview

OAuth2 clients are applications that can authenticate users and access the API on their behalf. There are two types of clients:

- **Confidential Clients**: Can securely store secrets (backend apps, mobile apps)
- **Public Clients**: Cannot securely store secrets (SPAs, JavaScript apps)

## Quick Start

### Create a Client Using Makefile

The easiest way to create an OAuth2 client:

```bash
make oauth2-create-client \
  ID=my-app \
  NAME="My Application" \
  SECRET="my-secret-key"
```

### List All Clients

```bash
make oauth2-list-clients
```

### Delete a Client

```bash
make oauth2-delete-client ID=my-app
```

## Detailed Usage

### 1. Create Confidential Client (Web/Backend App)

```bash
make oauth2-create-client \
  ID=wibusystem-web \
  NAME="WibuSystem Web App" \
  SECRET="super-secret-change-in-production" \
  REDIRECT_URIS="http://localhost:3000/auth/callback,https://myapp.com/callback" \
  GRANT_TYPES="authorization_code,refresh_token"
```

**Parameters:**
- `ID` (required): Unique client identifier
- `NAME` (required): Human-readable client name
- `SECRET` (required for confidential): Client secret (will be hashed with bcrypt)
- `REDIRECT_URIS` (optional): Comma-separated redirect URIs (default: `http://localhost:3000/auth/callback`)
- `GRANT_TYPES` (optional): Comma-separated grant types (default: `authorization_code,refresh_token`)
- `SCOPES` (optional): Comma-separated scopes (default: `openid,profile,email,offline_access`)

### 2. Create Public Client (SPA/JavaScript App)

```bash
make oauth2-create-client \
  ID=wibusystem-spa \
  NAME="WibuSystem SPA" \
  PUBLIC=true \
  REDIRECT_URIS="http://localhost:3000/callback"
```

**Note:** Public clients don't require a secret.

### 3. Create Mobile App Client

```bash
make oauth2-create-client \
  ID=wibusystem-mobile \
  NAME="WibuSystem Mobile" \
  SECRET="mobile-secret" \
  REDIRECT_URIS="myapp://oauth/callback" \
  GRANT_TYPES="authorization_code,refresh_token,password"
```

## Using the CLI Tool Directly

For more control, use the Go CLI tool directly:

```bash
go run ./cmd/oauth2-client \
  -id="my-client" \
  -name="My Client" \
  -secret="my-secret" \
  -redirect-uris="http://localhost:3000/callback,http://localhost:3000/silent-renew" \
  -grant-types="authorization_code,refresh_token" \
  -scopes="openid,profile,email" \
  -public=false
```

### CLI Options

| Flag | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `-id` | string | Yes | - | Client ID |
| `-name` | string | Yes | - | Client name |
| `-secret` | string | Conditional* | - | Client secret (hashed with bcrypt) |
| `-redirect-uris` | string | No | `http://localhost:3000/auth/callback` | Comma-separated URIs |
| `-grant-types` | string | No | `authorization_code,refresh_token` | Comma-separated grant types |
| `-scopes` | string | No | `openid,profile,email,offline_access` | Comma-separated scopes |
| `-public` | bool | No | `false` | Is this a public client? |

*Required for confidential clients (when `-public=false`)

## Using SQL Script

For manual insertion or scripting:

```bash
docker exec -i system_dev psql -U system_dev -d system_dev < scripts/create_oauth2_client.sql
```

Edit `scripts/create_oauth2_client.sql` to customize the client details.

## Examples

### Example 1: Next.js Web Application

```bash
make oauth2-create-client \
  ID=nextjs-app \
  NAME="Next.js Application" \
  SECRET="$(openssl rand -base64 32)" \
  REDIRECT_URIS="http://localhost:3000/api/auth/callback,https://myapp.com/api/auth/callback"
```

### Example 2: React SPA

```bash
make oauth2-create-client \
  ID=react-spa \
  NAME="React Single Page App" \
  PUBLIC=true \
  REDIRECT_URIS="http://localhost:3000/callback" \
  GRANT_TYPES="authorization_code"
```

### Example 3: Mobile App with Password Grant

```bash
make oauth2-create-client \
  ID=ios-app \
  NAME="iOS Mobile App" \
  SECRET="ios-secret-key" \
  REDIRECT_URIS="wibusystem://oauth/callback" \
  GRANT_TYPES="authorization_code,refresh_token,password"
```

### Example 4: Machine-to-Machine (M2M)

```bash
make oauth2-create-client \
  ID=backend-service \
  NAME="Backend Service" \
  SECRET="service-secret" \
  GRANT_TYPES="client_credentials" \
  SCOPES="api:read,api:write"
```

## Grant Types

Available grant types:

- `authorization_code`: Standard OAuth2 flow for web/mobile apps
- `refresh_token`: Allow token refresh
- `password`: Resource Owner Password Credentials (use with caution)
- `client_credentials`: Machine-to-machine authentication
- `implicit`: Legacy flow (not recommended)

## Scopes

Default scopes:

- `openid`: OpenID Connect authentication
- `profile`: User profile information
- `email`: User email address
- `offline_access`: Refresh tokens

## Security Best Practices

1. **Store Secrets Securely**: Never commit client secrets to version control
2. **Use Environment Variables**: Store secrets in `.env` files
3. **Rotate Secrets Regularly**: Change client secrets periodically
4. **Use HTTPS**: Always use HTTPS in production redirect URIs
5. **Validate Redirect URIs**: Only whitelist exact redirect URIs
6. **Public Clients**: Use PKCE (Proof Key for Code Exchange) for public clients
7. **Limit Scopes**: Only grant necessary scopes

## Viewing Client Details

### Via Database

```bash
docker exec system_dev psql -U system_dev -d system_dev -c \
  "SELECT id, client_name, public, redirect_uris, grant_types, scopes FROM identity.oauth2_clients;"
```

### Via Make Command

```bash
make oauth2-list-clients
```

## Updating a Client

To update a client, simply run the create command again with the same ID:

```bash
make oauth2-create-client \
  ID=existing-client \
  NAME="Updated Name" \
  SECRET="new-secret"
```

This will update the client details while preserving the creation timestamp.

## Deleting a Client

```bash
make oauth2-delete-client ID=my-client
```

**Warning:** This will cascade delete all associated tokens, authorization codes, and refresh tokens.

## Troubleshooting

### "Client secret is required" Error

**Solution:** Add the `SECRET` parameter or set `PUBLIC=true` for public clients.

### "Failed to connect to database" Error

**Solution:** Make sure the database is running:
```bash
make db-up
```

### Cannot Find Client After Creation

**Solution:** List all clients to verify:
```bash
make oauth2-list-clients
```

## Integration Examples

### Node.js (Passport.js)

```javascript
const OAuth2Strategy = require('passport-oauth2');

passport.use('wibusystem', new OAuth2Strategy({
    authorizationURL: 'http://localhost:8080/oauth2/authorize',
    tokenURL: 'http://localhost:8080/oauth2/token',
    clientID: 'wibusystem-web',
    clientSecret: 'super-secret-change-in-production',
    callbackURL: 'http://localhost:3000/auth/callback'
  },
  function(accessToken, refreshToken, profile, cb) {
    // Handle authentication
  }
));
```

### Python (Authlib)

```python
from authlib.integrations.flask_client import OAuth

oauth = OAuth(app)
oauth.register(
    name='wibusystem',
    client_id='wibusystem-web',
    client_secret='super-secret-change-in-production',
    authorize_url='http://localhost:8080/oauth2/authorize',
    access_token_url='http://localhost:8080/oauth2/token',
    redirect_uri='http://localhost:3000/auth/callback',
    client_kwargs={'scope': 'openid profile email'}
)
```

## Related Documentation

- [OAuth2 Flow Documentation](./OAUTH2_FLOW.md)
- [API Authentication Guide](./API_AUTH.md)
- [Identity Module Guide](./HANDLER_LAYER_GUIDE.md)

## Support

For issues or questions, please check:
- [GitHub Issues](https://github.com/yourusername/wibusystem-be/issues)
- [Documentation](../START_HERE.md)
