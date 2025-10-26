
## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- `openssl` command-line tool
- `golang-migrate` (see `migrations/README.md`)

### Running the Application

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd system
    ```

2.  **Generate RSA Keys for OAuth2:**
    The OAuth2/OIDC service requires an RSA key pair for signing tokens.

    ```bash
    # Create a directory for keys
    mkdir -p configs/keys

    # Generate 2048-bit RSA private key
    openssl genrsa -out configs/keys/private_key.pem 2048

    # Extract public key (for reference)
    openssl rsa -in configs/keys/private_key.pem -pubout -out configs/keys/public_key.pem
    ```

3.  **Configure Environment:**
    Create a `.env` file and set the following variables for OAuth2:
    ```env
    # .env
    OAUTH2_ISSUER=http://localhost:8080
    OAUTH2_PRIVATE_KEY_PATH=configs/keys/private_key.pem
    OAUTH2_KEY_ID=a-unique-key-id
    ```

4.  **Start Services:**
    Start the PostgreSQL and Redis databases using Docker.
    ```bash
    make docker-up
    ```

5.  **Run Database Migrations:**
    Apply all database migrations.
    ```bash
    make migrate-up
    ```

6.  **Run the Application:**
    ```bash
    go run ./cmd/server/main.go
    ```
    The server will be running on `http://localhost:8080`.

---

## 📁 Project Structure
```
/system
├── cmd/
│   └── server/
│       └── main.go                  // ENTRY POINT: Khởi tạo hệ thống (Config, Logger, DB, Router) và chạy server.
|
├── configs/                         // Chứa cấu hình và logic tải cấu hình (Viper/Godotenv).
│   └── config.go
|
├── internal/                        // MÃ LÕI NỘI BỘ (Chỉ import được trong module 'system').
│   ├── app/                         // Lớp Ứng dụng/Trình bày (Presentation Layer)
│   │   ├── handler/
│   │   │   ├── v1/                  // API Version 1
│   │   │   │   ├── user/            // Chia theo Domain (package 'user')
│   │   │   │   │   ├── handler.go   // Chứa struct UserHandler và các phương thức HTTP (Get, Create, Update)
│   │   │   │   │   └── dto.go       // Chứa UserRequest/UserResponse DTOs cho V1
│   │   │   │   └── product/         // Chia theo Domain (package 'product')
│   │   │   │       ├── handler.go
│   │   │   │       └── dto.go
│   │   │   └── v2/                  // API Version 2 (Nếu cần)
│   │   │       └── user/
│   │   │           ├── handler.go
│   │   │           └── dto.go
│   │   └── middleware/              // Các Middleware tùy chỉnh (package 'middleware')
│   │       └── auth.go
│   ├── domain/                      // Lớp Domain (Cốt lõi nghiệp vụ - Entity và Interface)
│   │   ├── user.go                  // Struct User (Entity) và Interface UserService, UserRepository
│   │   └── product.go
│   ├── pkg/
│   │   ├── service/                 // Lớp Dịch vụ (Business Logic - package 'service')
│   │   │   ├── user.go              // Triển khai logic nghiệp vụ (gọi Repository, thực thi rules).
│   │   │   └── product.go
│   │   └── repository/              // Lớp Truy cập Dữ liệu (package 'repository')
│   │       ├── user_repo.go         // Triển khai User Repository (dùng pgx).
│   │       └── product_repo.go
│   └── platform/                    // Lớp Cơ sở Hạ tầng (Infrastructure)
│       ├── database/
│       │   └── postgres.go          // Logic kết nối/khởi tạo pgx pool (package 'database')
│       └── logger/
│           └── zap.go               // Cấu hình và khởi tạo Zap logger (package 'logger')
|
├── migrations/                      // Các tệp migration SQL (thư mục số nhiều)
│   ├── 000001_create_users_table.up.sql
│   └── 000001_create_users_table.down.sql
|
├── pkg/                             // MÃ DÙNG CHUNG (Có thể import bởi các module/dịch vụ khác).
│   └── utils/                       // Các hàm tiện ích chung (thư mục số nhiều: 'utils')
│       └── validation.go
|
├── .env                             // Tệp chứa các biến môi trường
├── go.mod
├── go.sum
└── Dockerfile
```