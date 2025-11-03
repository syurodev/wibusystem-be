# Error Codes Package

Package `errcode` định nghĩa tất cả error codes được sử dụng trong hệ thống dưới dạng typed constants (enum).

## 🎯 Mục đích

- **Type-safe**: Sử dụng custom type `ErrorCode` thay vì hardcoded strings
- **Maintainable**: Tập trung tất cả error codes ở một nơi
- **Documentable**: Dễ dàng document và tra cứu
- **Refactorable**: IDE hỗ trợ refactoring và find usages

## 📋 Error Code Format

Error codes theo format: `Exxxx`
- **E4xxx**: Client Errors (400-499)
- **E5xxx**: Server Errors (500-599)

## 📖 Error Code Categories

### 400x: Bad Request / Validation Errors
```go
errcode.ValidationRequired       // E4001 - Validation: field required
errcode.ValidationInvalidFormat  // E4002 - Validation: invalid format
errcode.ValidationInvalidUserID  // E4003 - Validation: invalid user ID
```

### 401x: Unauthorized / Authentication Errors
```go
errcode.AuthInvalidToken                // E4011 - Invalid or expired token
errcode.AuthInvalidAuthorizationFormat  // E4012 - Invalid Authorization header format
errcode.AuthInvalidCredentials          // E4013 - Invalid username/password
errcode.AuthMissingAuthHeader           // E4014 - Missing Authorization header
```

### 403x: Forbidden / Authorization Errors
```go
errcode.AuthInsufficientScope     // E4031 - Insufficient scope
errcode.AuthInsufficientAnyScope  // E4032 - No required scopes
errcode.AuthForbidden             // E4033 - General forbidden
```

### 404x: Not Found Errors
```go
errcode.ResourceNotFound  // E4041 - Resource not found
errcode.UserNotFound      // E4042 - User not found
errcode.ClientNotFound    // E4043 - OAuth2 client not found
```

### 409x: Conflict Errors
```go
errcode.ResourceConflict  // E4091 - Resource already exists
errcode.DuplicateEntry    // E4092 - Duplicate entry
```

### 429x: Rate Limit Errors
```go
errcode.RateLimitExceeded  // E4291 - Too many requests
```

### 500x: Internal Server Errors
```go
errcode.InternalError         // E5001 - General internal error
errcode.MiddlewareError       // E5002 - Middleware error
errcode.ScopeValidationError  // E5003 - Scope validation error
errcode.ContextDataError      // E5004 - Context data error
errcode.DatabaseError         // E5010 - Database error
errcode.RedisError            // E5011 - Redis error
```

### 503x: Service Unavailable Errors
```go
errcode.ServiceUnavailable   // E5031 - Service temporarily unavailable
errcode.DatabaseUnavailable  // E5032 - Database unavailable
```

## 💻 Usage Examples

### Basic Usage with Response Package

```go
import (
    "system/pkg/util/errcode"
    "system/pkg/util/response"
)

// Error response
response.Error(
    c, 
    http.StatusUnauthorized, 
    errcode.AuthInvalidToken.String(),  // Convert to string
    "auth.invalid_token", 
    nil,
)

// AbortWithError
response.AbortWithError(
    c,
    http.StatusForbidden,
    errcode.AuthInsufficientScope.String(),
    "auth.insufficient_scope",
)
```

### In Middleware

```go
func RequireAuth(provider OAuth2Provider) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.AbortWithError(
                c, 
                http.StatusUnauthorized, 
                errcode.AuthMissingAuthHeader.String(),
                "auth.missing_authorization_header",
            )
            return
        }
        // ...
    }
}
```

### In Handlers

```go
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(
            c,
            http.StatusBadRequest,
            errcode.ValidationRequired.String(),
            "validation.failed",
            gin.H{"error": err.Error()},
        )
        return
    }
    
    // Check duplicate
    if userExists {
        response.Error(
            c,
            http.StatusConflict,
            errcode.DuplicateEntry.String(),
            "user.already_exists",
            nil,
        )
        return
    }
    
    // Success
    response.Success(c, http.StatusCreated, "user.created", userData, nil)
}
```

### Type Conversion

```go
// ErrorCode to string
code := errcode.AuthInvalidToken
codeStr := code.String()  // "E4011"

// Can be used directly in string context
fmt.Printf("Error code: %s", errcode.ValidationRequired)  // "Error code: E4001"
```

## 🔍 Finding Usage

Tìm tất cả nơi sử dụng một error code:

```go
// IDE: Click vào constant và "Find Usages"
errcode.AuthInvalidToken

// Or grep in terminal
grep -r "AuthInvalidToken" .
```

## ✅ Best Practices

1. **Always use constants, never hardcode strings:**
   ```go
   // ❌ Bad
   response.Error(c, 400, "E4001", "error", nil)
   
   // ✅ Good
   response.Error(c, 400, errcode.ValidationRequired.String(), "error", nil)
   ```

2. **Convert to string when needed:**
   ```go
   code := errcode.AuthInvalidToken
   codeStr := code.String()  // Explicit conversion
   ```

3. **Add new codes to appropriate category:**
   ```go
   // In pkg/util/errcode/codes.go
   const (
       AuthInvalidToken ErrorCode = "E4011"
       AuthNewErrorHere ErrorCode = "E4015"  // Add new code
   )
   ```

4. **Document each code with comment:**
   ```go
   const (
       ValidationRequired ErrorCode = "E4001" // Validation: field required
   )
   ```

## 📝 Adding New Error Codes

1. Mở file `pkg/util/errcode/codes.go`
2. Tìm category phù hợp (hoặc tạo mới)
3. Thêm constant với comment mô tả:

```go
// 401x: Unauthorized / Authentication Errors
const (
    AuthInvalidToken      ErrorCode = "E4011" // Invalid or expired token
    AuthNewFeatureError   ErrorCode = "E4016" // Your new error here
)
```

4. Update documentation nếu cần
5. Sử dụng trong code

## 🧪 Testing

```go
func TestErrorCodes(t *testing.T) {
    // Test string conversion
    assert.Equal(t, "E4011", errcode.AuthInvalidToken.String())
    
    // Test type safety
    var code errcode.ErrorCode = errcode.ValidationRequired
    assert.Equal(t, "E4001", code.String())
}
```

## 📚 Related Documentation

- [Standard Response Package](../response/README.md)
- [i18n Package](../../../internal/platform/i18n/README.md)
- [Middleware Usage Guide](../../../docs/MIDDLEWARE_USAGE_GUIDE.md)
