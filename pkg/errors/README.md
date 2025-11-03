# Errors Package

Package `errors` chứa tất cả business logic errors được tổ chức theo domain.

## 📂 Structure

```
pkg/errors/
├── common.go      # Common errors (resource, validation, system)
├── auth.go        # Authentication, Authorization, Session, Consent errors
├── user.go        # User domain errors
├── oauth2.go      # OAuth2 domain errors
└── README.md      # This file
```

## 🎯 Nguyên tắc

### **1. Errors theo Domain**

Mỗi domain có file riêng để dễ quản lý và mở rộng:

```go
// user.go - User domain errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserInactive = errors.New("user is inactive")
)

// oauth2.go - OAuth2 domain errors
var (
    ErrClientNotFound = errors.New("oauth2 client not found")
    ErrInvalidClient = errors.New("invalid client credentials")
)
```

### **2. Business Logic Errors Only**

Package này chỉ chứa **business logic errors**, KHÔNG chứa technical errors:

```go
// ✅ Business logic errors (pkg/errors)
ErrUserNotFound
ErrInvalidCredentials
ErrSessionExpired

// ❌ Technical errors (infrastructure layer)
sql.ErrNoRows
redis.Nil
context.DeadlineExceeded
```

### **3. Service Layer Converts Technical → Business Errors**

```go
// Service converts technical errors to business errors
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        // Convert technical error to business error
        if errors.Is(err, sql.ErrNoRows) {
            return nil, pkgerrors.ErrUserNotFound  // Business error
        }
        return nil, pkgerrors.ErrDatabaseError  // Generic business error
    }
    return user, nil
}
```

## 📖 Usage Guide

### **Import Package**

```go
import (
    pkgerrors "system/pkg/errors"  // Alias to avoid conflict with std errors
)
```

### **In Service Layer**

Service throw business errors:

```go
func (s *OAuth2Service) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, pkgerrors.ErrInvalidCredentials  // Hide technical details
    }
    
    if !crypto.VerifyPassword(user.PasswordHash, password) {
        return nil, pkgerrors.ErrInvalidCredentials
    }
    
    if !user.Active {
        return nil, pkgerrors.ErrUserInactive
    }
    
    return user, nil
}
```

### **In Handler Layer**

Handler maps business errors to HTTP responses:

```go
import (
    "errors"
    pkgerrors "system/pkg/errors"
    "system/pkg/util/errcode"
    "system/pkg/util/response"
)

func (h *Handler) Login(c *gin.Context) {
    user, err := h.oauth2Service.AuthenticateUser(ctx, email, password)
    if err != nil {
        h.handleServiceError(c, err)
        return
    }
    
    response.Success(c, 200, "login.success", userData, nil)
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, pkgerrors.ErrInvalidCredentials):
        response.Error(c, 401, errcode.AuthInvalidCredentials.String(), "auth.invalid_credentials", nil)
    
    case errors.Is(err, pkgerrors.ErrUserNotFound):
        response.Error(c, 404, errcode.UserNotFound.String(), "user.not_found", nil)
    
    case errors.Is(err, pkgerrors.ErrUserInactive):
        response.Error(c, 403, errcode.AuthForbidden.String(), "user.inactive", nil)
    
    case errors.Is(err, pkgerrors.ErrSessionExpired):
        response.Error(c, 401, errcode.AuthInvalidToken.String(), "auth.session_expired", nil)
    
    default:
        // Log unexpected errors
        response.Error(c, 500, errcode.InternalError.String(), "general.error", nil)
    }
}
```

### **Error Checking**

Use `errors.Is()` for checking:

```go
if errors.Is(err, pkgerrors.ErrUserNotFound) {
    // Handle user not found
}

if errors.Is(err, pkgerrors.ErrInvalidCredentials) {
    // Handle invalid credentials
}
```

## 📋 Error Categories

### **common.go**
- Resource errors: `ErrResourceNotFound`, `ErrDuplicateEntry`, `ErrResourceConflict`
- Validation errors: `ErrInvalidInput`, `ErrValidationFailed`, `ErrInvalidFormat`
- System errors: `ErrInternalServer`, `ErrServiceUnavailable`, `ErrDatabaseError`, `ErrRedisError`

### **auth.go**
- Authentication: `ErrInvalidCredentials`, `ErrInvalidToken`, `ErrMissingAuthHeader`
- Authorization: `ErrForbidden`, `ErrInsufficientScope`, `ErrInsufficientPermission`
- Session: `ErrSessionExpired`, `ErrSessionNotFound`, `ErrSessionInvalid`
- Consent: `ErrConsentNotFound`, `ErrConsentRevoked`, `ErrConsentDenied`

### **user.go**
- User errors: `ErrUserNotFound`, `ErrUserAlreadyExists`, `ErrUserInactive`, `ErrUserBlocked`, `ErrEmailAlreadyTaken`, `ErrInvalidUserID`

### **oauth2.go**
- Client errors: `ErrClientNotFound`, `ErrInvalidClient`, `ErrClientInactive`, `ErrClientAlreadyExists`
- Authorization: `ErrAuthRequestNotFound`, `ErrAuthRequestExpired`, `ErrInvalidAuthRequest`
- Token errors: `ErrInvalidGrant`, `ErrInvalidScope`, `ErrUnsupportedGrantType`

## ✅ Best Practices

### **1. Always wrap with context**

```go
// ❌ Bad - No context
return nil, pkgerrors.ErrUserNotFound

// ✅ Good - With context
return nil, fmt.Errorf("failed to get user %s: %w", userID, pkgerrors.ErrUserNotFound)
```

### **2. Alias import to avoid conflict**

```go
import (
    "errors"  // Standard library
    pkgerrors "system/pkg/errors"  // This package
)
```

### **3. Service layer converts, Handler layer maps**

```
Repository → Technical Errors
     ↓
Service → Business Errors (pkg/errors)
     ↓
Handler → HTTP Responses (errcode + i18n)
```

### **4. Add new errors to appropriate file**

```go
// user.go - Add user-related errors here
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserSuspended = errors.New("user is suspended")  // New error
)
```

### **5. Document each error clearly**

```go
var (
    // ErrUserNotFound is returned when a user cannot be found in the database
    ErrUserNotFound = errors.New("user not found")
)
```

## 🧪 Testing

```go
func TestAuthenticateUser_InvalidPassword(t *testing.T) {
    user, err := service.AuthenticateUser(ctx, "test@example.com", "wrong")
    
    assert.Nil(t, user)
    assert.ErrorIs(t, err, pkgerrors.ErrInvalidCredentials)
}

func TestGetUser_NotFound(t *testing.T) {
    user, err := service.GetUser(ctx, uuid.New())
    
    assert.Nil(t, user)
    assert.ErrorIs(t, err, pkgerrors.ErrUserNotFound)
}
```

## 🆕 Adding New Domain Errors

Khi thêm domain mới (ví dụ: Product):

1. Tạo file mới: `pkg/errors/product.go`
2. Định nghĩa errors:
   ```go
   package errors
   
   import "errors"
   
   // Product domain errors
   var (
       ErrProductNotFound = errors.New("product not found")
       ErrProductOutOfStock = errors.New("product out of stock")
       ErrInvalidPrice = errors.New("invalid price")
   )
   ```
3. Update README.md với categories mới

## 📊 Error Flow Diagram

```
┌──────────────────────────────────────────────┐
│            HTTP Request                      │
└───────────────┬──────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────┐
│         Handler (Presentation)               │
│  - Maps business errors to HTTP responses    │
│  - Uses errcode + i18n                       │
└───────────────┬──────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────┐
│         Service (Business Logic)             │
│  - Throws pkg/errors (business errors)       │
│  - Converts technical → business errors      │
└───────────────┬──────────────────────────────┘
                │
                ▼
┌──────────────────────────────────────────────┐
│       Repository (Data Access)               │
│  - Throws technical errors (sql, redis)      │
│  - No business logic                         │
└──────────────────────────────────────────────┘
```

## 📚 Related Documentation

- [Error Codes Package](../util/errcode/README.md)
- [Response Package](../util/response/README.md)
- [Service Layer Guide](../../../docs/SERVICE_LAYER_GUIDE.md) (TBD)

## ❓ FAQ

**Q: Tại sao không để errors trong domain package?**  
A: Domain package nên chỉ chứa entities và interfaces. Errors là cross-cutting concern nên nằm ở pkg để reusable.

**Q: Khi nào dùng `fmt.Errorf` wrap?**  
A: Khi cần thêm context nhưng vẫn giữ error type để có thể dùng `errors.Is()`:
```go
return fmt.Errorf("user service: %w", pkgerrors.ErrUserNotFound)
```

**Q: Có nên tạo custom error type không?**  
A: Chỉ khi cần attach thêm data. Ví dụ:
```go
type ValidationError struct {
    Field string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on field %s: %s", e.Field, e.Message)
}
```

**Q: Handler có nên import pkg/errors không?**  
A: Có, handler cần import để check errors và map sang HTTP responses.
