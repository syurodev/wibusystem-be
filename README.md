# WibuSystem - Go Monorepo

Đây là monorepo Go sử dụng Go Workspaces để quản lý nhiều microservices và shared modules.

## Cấu trúc project

```
.
├── go.mod              # Root module
├── go.work             # Workspace definition
├── services/           # Microservices
│   └── identify/
│       ├── go.mod      # Auth service module
│       └── main.go     # Example auth service
└── pkg/                # Shared modules
    ├── common/
    │   ├── go.mod      # Common shared module
    │   └── types.go    # Shared types
    └── database/
        ├── go.mod      # Database shared module
        └── connection.go # DB interfaces
```

## Go Workspace Commands

### Quản lý workspace
```bash
# Đồng bộ dependencies cho tất cả modules
go work sync

# Thêm module mới vào workspace
go work use ./path/to/module

# Xóa module khỏi workspace
go work edit -dropuse ./path/to/module

# Xem thông tin workspace
go work edit -json
```

### Build và test
```bash
# Build service cụ thể
go build ./services/identify

# Build tất cả services
go build ./services/...

# Chạy service
go run ./services/identify

# Test tất cả modules
go test ./...

# Test service cụ thể
go test ./services/identify/...
```

### Làm việc với dependencies
```bash
# Thêm dependency cho service cụ thể
cd services/identify
go get github.com/gin-gonic/gin

# Thêm dependency cho shared module
cd pkg/database
go get gorm.io/gorm

# Clean up modules
go mod tidy
```

## Thêm service mới

1. Tạo thư mục service:
```bash
mkdir services/new-service
```

2. Tạo go.mod:
```bash
cd services/new-service
go mod init wibusystem/services/new-service
```

3. Thêm dependencies cho shared modules trong go.mod:
```go
require (
    wibusystem/pkg/common v0.0.0
    wibusystem/pkg/database v0.0.0
)

replace (
    wibusystem/pkg/common => ../../pkg/common
    wibusystem/pkg/database => ../../pkg/database
)
```

4. Thêm vào workspace:
```bash
go work use ./services/new-service
```

## Thêm shared module mới

1. Tạo thư mục:
```bash
mkdir shared/new-module
```

2. Tạo go.mod:
```bash
cd shared/new-module
go mod init wibusystem/pkg/new-module
```

3. Thêm vào workspace:
```bash
go work use ./pkg/new-module
```

4. Sử dụng trong services bằng cách thêm vào go.mod của service:
```go
require wibusystem/pkg/new-module v0.0.0
replace wibusystem/pkg/new-module => ../../pkg/new-module
```

## Lưu ý

- Workspace cho phép import trực tiếp giữa các modules mà không cần publish
- Sử dụng `replace` directives để point đến local paths
- Chạy `go work sync` sau khi thay đổi dependencies
- Mỗi service và shared module có go.mod riêng để độc lập về dependencies
