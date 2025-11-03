# I18n (Internationalization) Package

Package này cung cấp hỗ trợ đa ngôn ngữ cho application sử dụng `go-i18n/v2`.

## Features

- ✅ Hỗ trợ đa ngôn ngữ (EN, VI)
- ✅ Embedded translation files vào binary
- ✅ Template data support (dynamic values)
- ✅ Fallback language (English)
- ✅ Domain-based organization
- ✅ Gin middleware integration

## Cấu trúc thư mục

```
internal/platform/i18n/
├── i18n.go              # Core implementation
├── locales/             # Translation files (embedded)
│   └── common/          # Common domain
│       ├── en.json      # English translations
│       └── vi.json      # Vietnamese translations
├── example_usage.go     # Usage examples
└── README.md            # This file
```

## Cấu trúc Translation Files

### Cấu trúc folder

```
locales/{domain}/{lang}.json
```

- `{domain}`: Tên domain (common, user, product, etc.)
- `{lang}`: Language code (en, vi)

### Ví dụ JSON file

**locales/common/en.json:**
```json
{
  "general.success": "Operation completed successfully.",
  "validation.required": "The field '{{.field}}' is required.",
  "rate_limit.exceeded": "Too many requests. Please try again in {{.seconds}} seconds."
}
```

**locales/common/vi.json:**
```json
{
  "general.success": "Thao tác được thực hiện thành công.",
  "validation.required": "Trường '{{.field}}' là bắt buộc.",
  "rate_limit.exceeded": "Quá nhiều yêu cầu. Vui lòng thử lại sau {{.seconds}} giây."
}
```

## Khởi tạo

I18n được khởi tạo trong `main.go`:

```go
import "system/internal/platform/i18n"

func main() {
    // Khởi tạo i18n
    if err := i18n.InitI18n(logger); err != nil {
        log.Fatal(err)
    }
}
```

## Sử dụng trong Gin Handler

### 1. Basic Usage

```go
func GetUser(c *gin.Context) {
    localizer := c.MustGet("localizer").(*i18n.Localizer)

    message := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "general.success",
    })

    c.JSON(200, gin.H{
        "message": message,
    })
}
```

### 2. Với Template Data

```go
func ValidateEmail(c *gin.Context) {
    localizer := c.MustGet("localizer").(*i18n.Localizer)

    message := localizer.MustLocalize(&i18n.LocalizeConfig{
        MessageID: "validation.required",
        TemplateData: map[string]any{
            "field": "email",
        },
    })

    c.JSON(400, gin.H{
        "error": message, // "The field 'email' is required."
    })
}
```

### 3. Với StandardResponse Helper

```go
import "system/pkg/util/response"

func CreateUser(c *gin.Context) {
    // Success response
    response.Success(c, 201, "user.created", userData, nil)

    // Error response
    response.Error(c, 400, "user.email_exists", nil)
}
```

## Language Detection

Language được detect từ HTTP header `Accept-Language`:

```bash
# English
curl -H "Accept-Language: en" http://localhost:8080/api/users

# Vietnamese
curl -H "Accept-Language: vi" http://localhost:8080/api/users

# Vietnamese with priority
curl -H "Accept-Language: vi-VN,vi;q=0.9,en;q=0.8" http://localhost:8080/api/users
```

## Tạo Localizer Manually

### Từ Accept-Language header

```go
acceptLang := c.GetHeader("Accept-Language")
localizer := i18n.GetLocalizerFromAcceptLanguage(acceptLang)
```

### Từ language tag

```go
import "golang.org/x/text/language"

// Vietnamese
localizerVi := i18n.GetLocalizer(language.Vietnamese)

// English
localizerEn := i18n.GetLocalizer(language.English)
```

## Message ID Naming Convention

Sử dụng dot notation: `{category}.{action}`

### Ví dụ:

- **General:** `general.success`, `general.error`
- **Validation:** `validation.required`, `validation.invalid_format`
- **Auth:** `auth.unauthorized`, `auth.forbidden`
- **Resource:** `resource.not_found`, `resource.conflict`
- **User:** `user.created`, `user.updated`, `user.deleted`
- **Product:** `product.created`, `product.out_of_stock`

## Template Data

Sử dụng `{{.VariableName}}` trong JSON:

### JSON:
```json
{
  "user.verification_sent": "Verification email has been sent to {{.email}}."
}
```

### Code:
```go
message := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID: "user.verification_sent",
    TemplateData: map[string]any{
        "email": "user@example.com",
    },
})
// Output: "Verification email has been sent to user@example.com."
```

## Thêm Domain Mới

### Bước 1: Tạo folder và files

```bash
mkdir -p internal/platform/i18n/locales/product
touch internal/platform/i18n/locales/product/en.json
touch internal/platform/i18n/locales/product/vi.json
```

### Bước 2: Thêm translations

**locales/product/en.json:**
```json
{
  "product.created": "Product created successfully.",
  "product.not_found": "Product not found."
}
```

**locales/product/vi.json:**
```json
{
  "product.created": "Tạo sản phẩm thành công.",
  "product.not_found": "Không tìm thấy sản phẩm."
}
```

### Bước 3: Update i18n.go

```go
// Thêm domain vào slice
domains := []string{"common", "product"}
```

### Bước 4: Rebuild

```bash
go build -o bin/server ./cmd/server
```

## Thêm Language Mới

### Bước 1: Tạo files cho language mới

```bash
# Thêm Japanese
touch internal/platform/i18n/locales/common/ja.json
```

### Bước 2: Thêm translations

**locales/common/ja.json:**
```json
{
  "general.success": "操作が正常に完了しました。"
}
```

### Bước 3: Update i18n.go

```go
// Thêm language vào slice
languages := []string{"en", "vi", "ja"}
```

## Error Handling

### MustLocalize (panic nếu fail)

```go
message := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID: "general.success",
})
```

### Localize (return error)

```go
message, err := localizer.Localize(&i18n.LocalizeConfig{
    MessageID: "some.key",
    DefaultMessage: &i18n.Message{
        ID:    "some.key",
        Other: "Default message",
    },
})
if err != nil {
    // Handle error hoặc sử dụng default message
}
```

## Current Message IDs

### Common Domain (locales/common/)

| Message ID | EN | VI |
|-----------|----|----|
| `general.success` | Operation completed successfully. | Thao tác được thực hiện thành công. |
| `general.error` | An internal error occurred. | Đã xảy ra lỗi nội bộ. |
| `validation.failed` | Validation failed. | Xác thực thất bại. |
| `validation.required` | The field '{{.field}}' is required. | Trường '{{.field}}' là bắt buộc. |
| `validation.invalid_format` | The field '{{.field}}' has an invalid format. | Trường '{{.field}}' có định dạng không hợp lệ. |
| `auth.unauthorized` | Authentication required. Please log in. | Cần xác thực. Vui lòng đăng nhập. |
| `auth.forbidden` | You do not have permission. | Bạn không có quyền truy cập. |
| `resource.not_found` | The requested resource was not found. | Không tìm thấy tài nguyên. |
| `resource.conflict` | The resource already exists. | Tài nguyên đã tồn tại. |
| `rate_limit.exceeded` | Too many requests. Try again in {{.seconds}} seconds. | Quá nhiều yêu cầu. Thử lại sau {{.seconds}} giây. |

## Best Practices

### 1. Luôn dùng i18n cho user-facing messages

```go
// ✅ ĐÚNG
response.Success(c, 200, "user.created", user, nil)

// ❌ SAI
c.JSON(200, gin.H{"message": "User created successfully"})
```

### 2. Sync message IDs giữa các languages

Đảm bảo en.json và vi.json có cùng message IDs:

```json
// ✅ ĐÚNG - Cả 2 files đều có key này
// en.json
{"user.created": "User created successfully."}

// vi.json
{"user.created": "Tạo người dùng thành công."}
```

### 3. Sử dụng template data cho dynamic values

```go
// ✅ ĐÚNG
message := localizer.MustLocalize(&i18n.LocalizeConfig{
    MessageID: "validation.required",
    TemplateData: map[string]any{"field": "email"},
})

// ❌ SAI - Hard-coded
message := "The field email is required."
```

### 4. Internal logs giữ tiếng Anh

```go
// Internal logs - keep English
logger.Error("Failed to connect to database", zap.Error(err))

// User-facing messages - use i18n
response.Error(c, 500, "general.error", nil)
```

### 5. Test với nhiều languages

```bash
# Test English
curl -H "Accept-Language: en" http://localhost:8080/api/test

# Test Vietnamese
curl -H "Accept-Language: vi" http://localhost:8080/api/test
```

## Troubleshooting

### Lỗi: "pattern locales/*/*.json: no matching files found"

**Nguyên nhân:** Translation files không tồn tại hoặc structure sai.

**Giải pháp:**
1. Kiểm tra structure: `internal/platform/i18n/locales/{domain}/{lang}.json`
2. Verify files tồn tại: `ls -la internal/platform/i18n/locales/common/`
3. Rebuild: `go build ./cmd/server`

### Message ID không được translate

**Nguyên nhân:** Message ID không tồn tại trong translation file.

**Giải pháp:**
1. Check message ID trong JSON file
2. Đảm bảo JSON format đúng
3. Sử dụng `DefaultMessage` để fallback:

```go
message, err := localizer.Localize(&i18n.LocalizeConfig{
    MessageID: "missing.key",
    DefaultMessage: &i18n.Message{
        ID:    "missing.key",
        Other: "Default text",
    },
})
```

### Template data không hoạt động

**Nguyên nhân:** Template syntax sai hoặc field name không match.

**Giải pháp:**

```json
// ✅ ĐÚNG
{"message": "Hello {{.Name}}"}
```

```go
// ✅ ĐÚNG - Field name match
TemplateData: map[string]any{"Name": "John"}
```

## Examples

Xem thêm examples chi tiết trong file `example_usage.go`.
