# OAuth2 Authentication Middleware - Usage Guide

Hướng dẫn sử dụng OAuth2 Authentication Middleware để bảo vệ API endpoints.

---

## 📋 Tổng quan

Middleware này cung cấp xác thực Bearer Token cho protected resources sử dụng OAuth2 Token Introspection.

**Files:**
- `internal/app/middleware/auth.go` - Authentication middleware implementation

**Middlewares có sẵn:**
- `RequireAuth()` - Xác thực Bearer Token (bắt buộc có token hợp lệ)
- `RequireScope()` - Kiểm tra token có scope cụ thể
- `RequireAnyScope()` - Kiểm tra token có ít nhất một trong các scopes

**Helper Functions:**
- `GetUserID()` - Lấy user ID từ token
- `GetClientID()` - Lấy client ID từ token
- `GetScopes()` - Lấy granted scopes từ token

---

## 🚀 Cách sử dụng

### 1. Basic Authentication - Require Valid Token

Bảo vệ endpoint yêu cầu valid access token:

```go
package handler

import (
    "net/http"
    "system/internal/app/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, oauth2Provider fosite.OAuth2Provider) {
    // Protected routes group
    protected := router.Group("/api")
    protected.Use(middleware.RequireAuth(oauth2Provider))
    {
        protected.GET("/profile", GetProfile)
        protected.GET("/data", GetData)
    }
}

func GetProfile(c *gin.Context) {
    // Token đã được validate bởi middleware
    // Lấy user ID từ context
    userID, _ := middleware.GetUserID(c)
    
    c.JSON(http.StatusOK, gin.H{
        "user_id": userID,
        "message": "This is protected data",
    })
}
```

### 2. Scope-based Authorization

Bảo vệ endpoint với scope requirements:

```go
func SetupRoutes(router *gin.Engine, oauth2Provider fosite.OAuth2Provider) {
    protected := router.Group("/api")
    protected.Use(middleware.RequireAuth(oauth2Provider))
    
    // Endpoint yêu cầu scope "admin"
    adminRoutes := protected.Group("/admin")
    adminRoutes.Use(middleware.RequireScope("admin"))
    {
        adminRoutes.GET("/users", ListUsers)
        adminRoutes.DELETE("/users/:id", DeleteUser)
    }
    
    // Endpoint yêu cầu scope "read" hoặc "write"
    dataRoutes := protected.Group("/data")
    dataRoutes.Use(middleware.RequireAnyScope("read", "write"))
    {
        dataRoutes.GET("/list", ListData)
    }
}
```

### 3. Multiple Scope Checks

Kết hợp nhiều middleware cho fine-grained access control:

```go
func SetupRoutes(router *gin.Engine, oauth2Provider fosite.OAuth2Provider) {
    protected := router.Group("/api")
    protected.Use(middleware.RequireAuth(oauth2Provider))
    
    // Public API - chỉ cần token valid
    protected.GET("/public/data", GetPublicData)
    
    // Admin API - cần token + scope "admin"
    adminRoutes := protected.Group("/admin")
    adminRoutes.Use(middleware.RequireScope("admin"))
    {
        adminRoutes.GET("/stats", GetAdminStats)
    }
    
    // Premium API - cần token + scope "premium"
    premiumRoutes := protected.Group("/premium")
    premiumRoutes.Use(middleware.RequireScope("premium"))
    {
        premiumRoutes.GET("/features", GetPremiumFeatures)
    }
}
```

### 4. Using Helper Functions in Handlers

Truy xuất thông tin từ token trong handlers:

```go
func GetProfile(c *gin.Context) {
    // Lấy user ID
    userID, ok := middleware.GetUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "No user ID in token"})
        return
    }
    
    // Lấy client ID
    clientID, _ := middleware.GetClientID(c)
    
    // Lấy granted scopes
    scopes, _ := middleware.GetScopes(c)
    
    c.JSON(http.StatusOK, gin.H{
        "user_id":   userID,
        "client_id": clientID,
        "scopes":    scopes,
    })
}
```

### 5. Conditional Authorization

Xử lý authorization logic phức tạp trong handler:

```go
func UpdateUser(c *gin.Context) {
    targetUserID := c.Param("id")
    currentUserID, _ := middleware.GetUserID(c)
    scopes, _ := middleware.GetScopes(c)
    
    // Check if user can update: either own profile or has admin scope
    hasAdminScope := false
    for _, scope := range scopes {
        if scope == "admin" {
            hasAdminScope = true
            break
        }
    }
    
    if currentUserID != targetUserID && !hasAdminScope {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "You can only update your own profile",
        })
        return
    }
    
    // Proceed with update...
    c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}
```

---

## 🔒 Error Responses

Middleware trả về responses theo **Standard Response Format** với i18n support:

### 1. Missing Authorization Header
```json
{
  "success": false,
  "code": "E4011",
  "message": "Missing Authorization header."
}
```
**HTTP Status:** 401 Unauthorized  
**i18n Key:** `auth.missing_authorization_header`

### 2. Invalid Authorization Header Format
```json
{
  "success": false,
  "code": "E4012",
  "message": "Invalid Authorization header format. Expected: Bearer <token>"
}
```
**HTTP Status:** 401 Unauthorized  
**i18n Key:** `auth.invalid_authorization_format`

### 3. Invalid/Expired Token
```json
{
  "success": false,
  "code": "E4013",
  "message": "Access token is invalid or expired."
}
```
**HTTP Status:** 401 Unauthorized  
**i18n Key:** `auth.invalid_token`

### 4. Insufficient Scope
```json
{
  "success": false,
  "code": "E4031",
  "message": "The access token does not have the required scope."
}
```
**HTTP Status:** 403 Forbidden  
**i18n Key:** `auth.insufficient_scope`

### 5. Middleware Error
```json
{
  "success": false,
  "code": "E5001",
  "message": "Authentication middleware error."
}
```
**HTTP Status:** 500 Internal Server Error  
**i18n Key:** `auth.middleware_error`

**Note:** Response messages tự động được dịch theo `Accept-Language` header (en/vi).

---

## 📊 Integration Example

Complete example với router setup:

```go
package router

import (
    "net/http"
    "system/configs"
    "system/internal/app/middleware"
    "system/internal/oauth2"
    "github.com/gin-gonic/gin"
)

func NewRouter(cfg *configs.Config, oauth2Provider fosite.OAuth2Provider) *gin.Engine {
    router := gin.New()
    
    // Public endpoints - no authentication required
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "UP"})
    })
    
    // Protected API endpoints
    api := router.Group("/api/v1")
    api.Use(middleware.RequireAuth(oauth2Provider))
    {
        // Public endpoints (require valid token only)
        api.GET("/profile", GetProfile)
        api.GET("/settings", GetSettings)
        
        // Admin endpoints (require valid token + "admin" scope)
        admin := api.Group("/admin")
        admin.Use(middleware.RequireScope("admin"))
        {
            admin.GET("/users", ListUsers)
            admin.POST("/users", CreateUser)
            admin.DELETE("/users/:id", DeleteUser)
        }
        
        // Premium endpoints (require valid token + "premium" scope)
        premium := api.Group("/premium")
        premium.Use(middleware.RequireScope("premium"))
        {
            premium.GET("/features", GetPremiumFeatures)
            premium.GET("/analytics", GetAnalytics)
        }
        
        // Flexible endpoints (require valid token + any of multiple scopes)
        data := api.Group("/data")
        data.Use(middleware.RequireAnyScope("read", "write", "admin"))
        {
            data.GET("/list", ListData)
            data.GET("/export", ExportData)
        }
    }
    
    return router
}
```

---

## 🧪 Testing with cURL

### 1. Request without token
```bash
curl -X GET http://localhost:8080/api/v1/profile

# Response: 401 Unauthorized
{
  "success": false,
  "code": "E4011",
  "message": "Missing Authorization header."
}
```

### 2. Request with valid token
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..."

# Response: 200 OK
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

### 3. Request with token but insufficient scope
```bash
curl -X GET http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <token_without_admin_scope>"

# Response: 403 Forbidden
{
  "success": false,
  "code": "E4031",
  "message": "The access token does not have the required scope."
}
```

### 4. Request with Vietnamese language
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Accept-Language: vi"

# Response: 401 Unauthorized (in Vietnamese)
{
  "success": false,
  "code": "E4011",
  "message": "Thiếu Authorization header."
}
```

---

## 📝 Best Practices

1. **Always use RequireAuth first:**
   ```go
   protected := router.Group("/api")
   protected.Use(middleware.RequireAuth(oauth2Provider)) // Must be first
   protected.Use(middleware.RequireScope("admin"))       // Then scope checks
   ```

2. **Scope naming convention:**
   - Use lowercase: `admin`, `read`, `write`
   - Use dots for namespacing: `user.read`, `user.write`
   - Be specific: `users:delete` better than just `delete`

3. **Error handling:**
   - Middleware handles authentication errors automatically
   - Focus on business logic errors in handlers

4. **Performance:**
   - Token introspection is called on every request
   - Consider implementing caching if needed
   - Redis can cache introspection results with short TTL

5. **Security:**
   - Always use HTTPS in production
   - Set secure token expiration times
   - Implement rate limiting for protected endpoints
   - Log authentication failures for security monitoring

---

## 🔗 Related Documentation

- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749)
- [Token Introspection RFC 7662](https://datatracker.ietf.org/doc/html/rfc7662)
- [OAUTH2_IMPLEMENTATION_CHECKLIST.md](./OAUTH2_IMPLEMENTATION_CHECKLIST.md)
- [OAUTH2_PRODUCTION_READY_IMPROVEMENTS.md](./OAUTH2_PRODUCTION_READY_IMPROVEMENTS.md)

---

## ❓ FAQ

**Q: Có thể sử dụng middleware cho WebSocket không?**  
A: Không trực tiếp. WebSocket cần custom authentication logic vì không support headers theo cách tương tự HTTP requests.

**Q: Performance impact của token introspection mỗi request?**  
A: Introspection gọi Fosite provider internal logic (không qua network), rất nhanh. Nếu cần optimize hơn, implement caching với Redis.

**Q: Có thể custom error responses không?**  
A: Có, fork middleware code và customize JSON responses theo business requirements.

**Q: Client credentials grant có subject không?**  
A: Không, client credentials grant không có user context. `GetUserID()` sẽ return empty string.

**Q: Làm sao để test middleware?**  
A: Xem file `scripts/oauth2_curl_examples.md` để có examples về cách request tokens và test protected endpoints.
