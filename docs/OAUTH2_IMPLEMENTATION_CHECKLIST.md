# Checklist Triển khai OAuth2 Server

Tài liệu này theo dõi tiến độ triển khai các tính năng cho OAuth2 Authorization Server.

---

## ✅ Đã hoàn thành

### Phase 0: Thiết lập Nền tảng & Cấu hình
- [x] **Fosite Provider**: Khởi tạo và cấu hình Fosite provider.
  - [x] Tắt grant type `ResourceOwnerPasswordCredentials` (ROPC) không an toàn.
  - [x] Cấu hình đúng các strategy và dependency cho Fosite v0.49.0.
- [x] **Storage Layer**: Thiết lập kiến trúc lưu trữ kết hợp (Hybrid).
  - [x] **SQLStore**: Dùng cho dữ liệu bền vững.
    - [x] Implement `fosite.ClientManager` (quản lý client).
    - [x] Implement các phương thức JTI (`ClientAssertionJWTValid`, `SetClientAssertionJWT`).
  - [x] **RedisStore**: Dùng cho dữ liệu tạm thời.
    - [x] Implement `AuthorizeCodeStorage` (authorization code).
    - [x] Implement `PKCERequestStorage` (PKCE sessions).
    - [x] Implement `AccessTokenStorage` (access token).
    - [x] Implement `OpenIDConnectRequestStorage` (OIDC sessions).
    - [x] Implement `TokenRevocationStorage` (danh sách token bị thu hồi).
  - [x] **HybridStore**: Kết hợp SQL và Redis.
    - [x] Implement chiến lược **Cache-Aside** cho `RefreshTokenStorage` (lưu ở SQL, cache ở Redis) để đảm bảo an toàn và hiệu năng.
- [x] **Repository Pattern**: Tách biệt logic truy vấn khỏi lớp storage.
  - [x] Tạo `OAuth2ClientRepository` để quản lý `oauth2_clients`.
  - [x] Tạo `OAuth2SessionRepository` để quản lý `oauth2_sessions`.
  - [x] Tái cấu trúc để sử dụng `pgx` thuần, loại bỏ `sqlx` và `lib/pq`.
- [x] **Cấu hình**: Thêm các cấu hình cần thiết cho OAuth2 (`Issuer`, `PrivateKey`, `HMACSecret`).
- [x] **Endpoints Cơ bản**:
  - [x] Implement OpenID Connect Discovery Endpoint (`/.well-known/openid-configuration`).
  - [x] Implement JWKS Endpoint (`/.well-known/jwks.json`).
- [x] **Refactoring & Code Quality**:
  - [x] Sử dụng `enum` (hằng số) cho các giá trị scope.

---

## ⏳ Cần thực hiện

### Phase 1: Luồng Ủy quyền Cốt lõi (Authorization Code Flow)
- [ ] **Authorization Endpoint (`/oauth2/auth`)**
  - [ ] Tạo handler `Authorize`.
  - [ ] Xử lý việc xác thực người dùng (cần tạo trang/logic đăng nhập).
  - [ ] Xử lý việc người dùng chấp thuận (cần tạo trang/logic consent).
  - [ ] Tích hợp với `provider.NewAuthorizeRequest` và `provider.NewAuthorizeResponse` của Fosite.
- [ ] **Token Endpoint (`/oauth2/token`)**
  - [ ] Tạo handler `Token`.
  - [ ] Xử lý grant type `authorization_code`.
  - [ ] Xử lý grant type `refresh_token`.
  - [ ] Xử lý grant type `client_credentials`.
  - [ ] Tích hợp với `provider.NewAccessRequest` và `provider.NewAccessResponse` của Fosite.

### Phase 2: Hoàn thiện các Endpoint còn lại
- [ ] **UserInfo Endpoint (`/oauth2/userinfo`)**
  - [ ] **Thay thế `MockStore`**: Sử dụng một `UserRepository` thật sự để lấy thông tin người dùng từ database thay vì dữ liệu giả.
- [ ] **Token Revocation Endpoint (`/oauth2/revoke`)**
  - [ ] Tạo handler `Revoke`.
  - [ ] Tích hợp với `provider.NewRevocationRequest`.
- [ ] **Token Introspection Endpoint (`/oauth2/introspect`)**
  - [ ] Tạo handler `Introspect`.
  - [ ] Endpoint này cần được bảo vệ, chỉ cho phép các client đã xác thực gọi.

### Phase 3: Nâng cao & Bảo mật
- [ ] **Tạo Authentication Middleware**: Viết một middleware để bảo vệ các resource server API, xác thực Bearer Token bằng cách gọi logic introspection.
- [ ] **(Tùy chọn)** Implement Dynamic Client Registration (`/oauth2/register`).
- [ ] **(Tùy chọn)** Implement Pushed Authorization Requests (PAR) (`/oauth2/par`).
