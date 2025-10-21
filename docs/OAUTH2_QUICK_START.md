# OAuth2 Quick Start Guide

## 🚀 Create Your First OAuth2 Client (30 seconds)

### Step 1: Start the Database
```bash
make db-up
```

### Step 2: Start the Server
```bash
make run
# or
go run ./cmd/server/main.go
```

### Step 3: Create an OAuth2 Client
```bash
make oauth2-create-client \
  ID=my-app \
  NAME="My Application" \
  SECRET="my-secret-key-change-in-production"
```

### Step 4: Verify
```bash
make oauth2-list-clients
```

## 📋 Common Commands

| Command | Description |
|---------|-------------|
| `make oauth2-create-client ID=app NAME="App" SECRET=secret` | Create confidential client |
| `make oauth2-create-client ID=spa NAME="SPA" PUBLIC=true` | Create public client |
| `make oauth2-list-clients` | List all clients |
| `make oauth2-delete-client ID=app` | Delete a client |

## 📝 Common Use Cases

### Web Application (Next.js, Rails, Django)
```bash
make oauth2-create-client \
  ID=web-app \
  NAME="Web Application" \
  SECRET="$(openssl rand -base64 32)" \
  REDIRECT_URIS="http://localhost:3000/auth/callback"
```

### Single Page Application (React, Vue, Angular)
```bash
make oauth2-create-client \
  ID=spa-app \
  NAME="SPA Application" \
  PUBLIC=true \
  REDIRECT_URIS="http://localhost:3000/callback"
```

### Mobile App (iOS, Android, React Native)
```bash
make oauth2-create-client \
  ID=mobile-app \
  NAME="Mobile Application" \
  SECRET="mobile-secret" \
  REDIRECT_URIS="myapp://oauth/callback" \
  GRANT_TYPES="authorization_code,refresh_token"
```

### Backend Service (M2M)
```bash
make oauth2-create-client \
  ID=backend-service \
  NAME="Backend Service" \
  SECRET="service-secret" \
  GRANT_TYPES="client_credentials"
```

## 🔑 Client Types

### Confidential Client
- Can securely store secrets
- Requires `SECRET` parameter
- Examples: Backend apps, server-side apps

```bash
make oauth2-create-client ID=app NAME="App" SECRET=secret
```

### Public Client
- Cannot securely store secrets
- Set `PUBLIC=true`
- Examples: SPAs, mobile apps (with PKCE)

```bash
make oauth2-create-client ID=spa NAME="SPA" PUBLIC=true
```

## 🎯 Important Notes

1. **Save Your Secret**: The client secret is shown only once during creation
2. **Use HTTPS**: Always use HTTPS redirect URIs in production
3. **Environment Variables**: Store secrets in `.env` file, never in code
4. **Rotate Secrets**: Change secrets regularly in production

## 🔗 Your Client Credentials

After creating a client, you'll receive:

```
✅ OAuth2 Client Created Successfully!
=====================================
Client ID:        my-app
Client Name:      My Application
Client Secret:    my-secret-key-change-in-production
Public Client:    false
Redirect URIs:    http://localhost:3000/auth/callback
Grant Types:      authorization_code, refresh_token
Scopes:           openid, profile, email, offline_access
```

**Save these credentials securely!**

## 🛠️ Next Steps

1. ✅ **Created a client** - You're here!
2. 📖 Read [Full OAuth2 Client Management Guide](./OAUTH2_CLIENT_MANAGEMENT.md)
3. 🔐 Implement authentication in your app
4. 🧪 Test the OAuth2 flow

## ❓ Need Help?

- Full documentation: [OAUTH2_CLIENT_MANAGEMENT.md](./OAUTH2_CLIENT_MANAGEMENT.md)
- Getting started: [START_HERE.md](../START_HERE.md)
- Makefile commands: `make help`
