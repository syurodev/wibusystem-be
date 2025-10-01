# Tenant API Testing Guide

## Prerequisites
1. Start the Identity Service: `go run main.go`
2. Ensure database is running and migrations completed

## API Endpoints

### 1. Create Tenant (POST /api/v1/tenants)
**Required**: Admin scope

```bash
# Create tenant with all fields
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "My Company",
    "slug": "my-company",
    "description": "A sample company tenant",
    "settings": {
      "theme": "dark",
      "notifications": true
    }
  }'
```

```bash
# Create tenant with minimal fields
curl -X POST http://localhost:8080/api/v1/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "Basic Tenant"
  }'
```

### 2. List Tenants (GET /api/v1/tenants)

```bash
# List all tenants (admin only)
curl -X GET http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# List user's tenants only
curl -X GET "http://localhost:8080/api/v1/tenants?user_only=true" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# List with pagination
curl -X GET "http://localhost:8080/api/v1/tenants?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 3. Get Tenant by ID (GET /api/v1/tenants/:id)

```bash
curl -X GET http://localhost:8080/api/v1/tenants/TENANT_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 4. Update Tenant (PUT /api/v1/tenants/:id)

```bash
# Update tenant name and description
curl -X PUT http://localhost:8080/api/v1/tenants/TENANT_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "name": "Updated Company Name",
    "description": "Updated description"
  }'

# Update tenant slug
curl -X PUT http://localhost:8080/api/v1/tenants/TENANT_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "slug": "new-company-slug"
  }'

# Update tenant settings
curl -X PUT http://localhost:8080/api/v1/tenants/TENANT_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "settings": {
      "theme": "light",
      "notifications": false,
      "max_users": 100
    }
  }'
```

### 5. Delete Tenant (DELETE /api/v1/tenants/:id)
**Required**: Admin scope

```bash
curl -X DELETE http://localhost:8080/api/v1/tenants/TENANT_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 6. Get Tenant Members (GET /api/v1/tenants/:id/members)

```bash
# Get all members of a tenant
curl -X GET http://localhost:8080/api/v1/tenants/TENANT_ID/members \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Get members with pagination
curl -X GET "http://localhost:8080/api/v1/tenants/TENANT_ID/members?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 7. Tenant Dashboard (GET /api/v1/t/:tenant_id/dashboard)

```bash
curl -X GET http://localhost:8080/api/v1/t/TENANT_ID/dashboard \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Expected Response Format

### Success Response
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "id": "uuid",
    "name": "My Company",
    "slug": "my-company", 
    "description": "A sample company tenant",
    "settings": {
      "theme": "dark",
      "notifications": true
    },
    "status": "active",
    "created_at": "2025-09-28T10:41:00Z",
    "updated_at": "2025-09-28T10:41:00Z"
  },
  "error": null,
  "meta": {}
}
```

### Error Response
```json
{
  "success": false,
  "message": "Invalid request",
  "data": null,
  "error": {
    "code": "invalid_request",
    "description": "Tenant name is required"
  },
  "meta": {}
}
```

## Authentication

To get an access token for testing:

1. **Register/Login first** to get authenticated
2. **Use OAuth2 flow** to get access token
3. **Or use existing session** if testing from browser

## Validation Rules

### Tenant Name
- Required
- 2-100 characters
- Can contain letters, numbers, spaces, and common symbols

### Tenant Slug
- Optional
- 2-50 characters if provided
- Must be lowercase letters, numbers, and hyphens only
- Cannot start or end with hyphen
- Must be unique across all tenants

### Description
- Optional
- Text field, no length limit

### Settings
- Optional
- JSON object with any structure
- Stored as JSONB in database

## Common Test Scenarios

1. **Create tenant without authentication** → 401 Unauthorized
2. **Create tenant without admin scope** → 403 Access Denied  
3. **Create tenant with duplicate slug** → 400 Bad Request
4. **Create tenant with invalid slug format** → 400 Bad Request
5. **Update non-existent tenant** → 404 Not Found
6. **Access tenant without membership** → 403 Access Denied
7. **List tenants with pagination** → 200 OK with meta info