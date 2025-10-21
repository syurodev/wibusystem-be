# Scripts Directory

This directory contains utility scripts for database management and system operations.

## Available Scripts

### OAuth2 Client Creation (`create_oauth2_client.sql`)

Creates an OAuth2 client directly via SQL.

**Usage:**

```bash
# Via docker
docker exec -i system_dev psql -U system_dev -d system_dev < scripts/create_oauth2_client.sql

# Or if you have psql installed locally
psql -U system_dev -d system_dev -h localhost -f scripts/create_oauth2_client.sql
```

**Note:** You need to edit the script to customize:
- Client ID
- Client name
- Client secret (generate bcrypt hash)
- Redirect URIs
- Grant types
- Scopes

## Recommended Approach

For most use cases, use the Makefile commands instead:

```bash
# Create client
make oauth2-create-client ID=my-app NAME="My App" SECRET=secret

# List clients
make oauth2-list-clients

# Delete client
make oauth2-delete-client ID=my-app
```

See [OAuth2 Client Management Guide](../docs/OAUTH2_CLIENT_MANAGEMENT.md) for detailed documentation.
