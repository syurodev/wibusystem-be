# WebAuthn/Passkey Authentication

Hệ thống hiện đã hỗ trợ xác thực bằng Passkey (WebAuthn/FIDO2), cho phép người dùng đăng nhập mà không cần mật khẩu.

## 📋 Tổng Quan

WebAuthn (Web Authentication) là một tiêu chuẩn web cho phép người dùng xác thực bằng:
- **Passkeys**: Khóa mật mã được lưu trên thiết bị (iPhone, Android, máy tính)
- **Security Keys**: YubiKey, Google Titan Key, v.v.
- **Biometric**: Face ID, Touch ID, Windows Hello

## 🚀 Tính Năng

- ✅ Passwordless authentication (đăng nhập không cần mật khẩu)
- ✅ Phishing-resistant (chống phishing)
- ✅ Multi-device sync (sync giữa các thiết bị qua iCloud, Google Password Manager)
- ✅ Credential management (quản lý passkeys)
- ✅ Supports multiple credentials per user

## 🔧 Cấu Hình

### Environment Variables

Thêm các biến sau vào file `.env`:

```bash
# WebAuthn/Passkey Configuration
WEBAUTHN_RP_ID=localhost                                    # Domain của bạn (ví dụ: "example.com")
WEBAUTHN_RP_NAME=Wibutime                                   # Tên ứng dụng hiển thị khi đăng ký passkey
WEBAUTHN_RP_ORIGINS=http://localhost:8080,http://localhost:3000  # Danh sách origins được phép
WEBAUTHN_TIMEOUT=60000                                      # Timeout cho ceremony (milliseconds)
```

### Database Migration

Chạy migration để tạo bảng WebAuthn:

```bash
# Migration 000024 sẽ tạo:
# - webauthn_credentials: Lưu trữ passkeys
# - webauthn_sessions: Lưu trữ temporary sessions
# - Update users.password_hash thành optional (cho passwordless accounts)
```

## 📡 API Endpoints

### 1. Registration (Đăng Ký Passkey)

#### Begin Registration

**POST** `/api/v1/auth/passkey/register/begin`

Bắt đầu quá trình đăng ký passkey mới.

**Request Body:**
```json
{
  "email": "user@example.com",
  "full_name": "John Doe",
  "username": "johndoe" // optional
}
```

**Response:**
```json
{
  "challenge": "base64url_encoded_challenge",
  "rp": {
    "id": "localhost",
    "name": "Wibutime"
  },
  "user": {
    "id": "base64url_user_id",
    "name": "user@example.com",
    "displayName": "John Doe"
  },
  "pubKeyCredParams": [
    { "type": "public-key", "alg": -7 },
    { "type": "public-key", "alg": -257 }
  ],
  "timeout": 60000,
  "attestation": "none"
}
```

#### Finish Registration

**POST** `/api/v1/auth/passkey/register/finish`

Hoàn thành quá trình đăng ký passkey.

**Request Body:**
```json
{
  "id": "credential_id",
  "rawId": "base64url_raw_id",
  "type": "public-key",
  "response": {
    "clientDataJSON": "base64url_client_data",
    "attestationObject": "base64url_attestation",
    "transports": ["internal", "hybrid"]
  },
  "credential_name": "My iPhone 15 Pro" // optional
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "credential_id": "base64url_credential_id",
  "message": "Passkey registered successfully"
}
```

### 2. Authentication (Xác Thực)

#### Begin Authentication

**POST** `/api/v1/auth/passkey/authenticate/begin`

Bắt đầu quá trình xác thực bằng passkey.

**Request Body:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "challenge": "base64url_encoded_challenge",
  "timeout": 60000,
  "rpId": "localhost",
  "allowCredentials": [
    {
      "type": "public-key",
      "id": "base64url_credential_id",
      "transports": ["internal", "hybrid"]
    }
  ],
  "userVerification": "preferred"
}
```

#### Finish Authentication

**POST** `/api/v1/auth/passkey/authenticate/finish`

Hoàn thành quá trình xác thực.

**Request Body:**
```json
{
  "id": "credential_id",
  "rawId": "base64url_raw_id",
  "type": "public-key",
  "response": {
    "clientDataJSON": "base64url_client_data",
    "authenticatorData": "base64url_authenticator_data",
    "signature": "base64url_signature",
    "userHandle": "base64url_user_handle"
  }
}
```

**Response:**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "message": "Authentication successful"
}
```

### 3. Credential Management

#### List Credentials

**GET** `/api/v1/auth/passkey/credentials`

**Yêu cầu:** Bearer token authentication

Lấy danh sách tất cả passkeys của user.

**Response:**
```json
{
  "credentials": [
    {
      "id": "uuid",
      "credential_id": "base64url_credential_id",
      "credential_name": "My iPhone 15 Pro",
      "transports": ["internal", "hybrid"],
      "backup_eligible": true,
      "backup_state": true,
      "created_at": "2025-11-22T10:00:00Z",
      "last_used_at": "2025-11-22T14:30:00Z"
    }
  ]
}
```

#### Delete Credential

**DELETE** `/api/v1/auth/passkey/credentials`

**Yêu cầu:** Bearer token authentication

Xóa một passkey.

**Request Body:**
```json
{
  "credential_id": "base64url_credential_id"
}
```

**Response:**
```json
{
  "message": "Credential deleted successfully"
}
```

#### Update Credential Name

**PUT** `/api/v1/auth/passkey/credentials/name`

**Yêu cầu:** Bearer token authentication

Cập nhật tên của passkey.

**Request Body:**
```json
{
  "credential_id": "base64url_credential_id",
  "credential_name": "My MacBook Pro"
}
```

**Response:**
```json
{
  "message": "Credential name updated successfully"
}
```

## 🔐 Security Features

1. **Challenge-Response Authentication**: Mỗi ceremony sử dụng challenge ngẫu nhiên
2. **Origin Validation**: Chỉ chấp nhận requests từ origins được cấu hình
3. **Clone Detection**: Sign counter tracking để phát hiện credential bị nhân bản
4. **Session Timeout**: Sessions hết hạn sau 5 phút
5. **RPID Validation**: Đảm bảo passkey chỉ hoạt động trên domain chính xác

## 💡 Client Implementation Example

### Using SimpleWebAuthn (Browser)

```javascript
import { startRegistration, startAuthentication } from '@simplewebauthn/browser';

// Registration
async function registerPasskey(email, fullName) {
  // 1. Get options from server
  const beginResponse = await fetch('/api/v1/auth/passkey/register/begin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, full_name: fullName })
  });
  const options = await beginResponse.json();

  // 2. Start WebAuthn ceremony
  const credential = await startRegistration(options);

  // 3. Send credential to server
  const finishResponse = await fetch('/api/v1/auth/passkey/register/finish', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      ...credential,
      credential_name: 'My Device'
    })
  });

  return await finishResponse.json();
}

// Authentication
async function authenticateWithPasskey(email) {
  // 1. Get options from server
  const beginResponse = await fetch('/api/v1/auth/passkey/authenticate/begin', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  const options = await beginResponse.json();

  // 2. Start WebAuthn ceremony
  const credential = await startAuthentication(options);

  // 3. Send credential to server
  const finishResponse = await fetch('/api/v1/auth/passkey/authenticate/finish', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credential)
  });

  return await finishResponse.json();
}
```

## 🧪 Testing

### Local Development

1. **Localhost**: Passkeys hoạt động trên `localhost` (không cần HTTPS)
2. **Production**: Yêu cầu HTTPS và domain hợp lệ

### Browser Support

- ✅ Chrome/Edge 67+
- ✅ Safari 16+
- ✅ Firefox 122+
- ✅ Mobile: iOS 16+, Android 9+

### Testing Tools

- Chrome DevTools > Application > WebAuthn (Virtual Authenticators)
- https://webauthn.io/ - Test WebAuthn flows
- https://passkeys.dev/ - Passkey developer resources

## 🔄 Migration từ Password sang Passkey

Users có thể:
1. **Passwordless**: Tạo account chỉ với passkey (không có password)
2. **Hybrid**: Sử dụng cả password và passkey
3. **Upgrade**: Add passkey vào account hiện có (có password)

## 📚 References

- [WebAuthn Specification](https://www.w3.org/TR/webauthn-2/)
- [Passkeys.dev](https://passkeys.dev/)
- [SimpleWebAuthn Library](https://simplewebauthn.dev/)
- [go-webauthn/webauthn](https://github.com/go-webauthn/webauthn)

## ⚠️ Important Notes

1. **RPID** phải match với domain của bạn
   - Development: `localhost`
   - Production: `example.com` (không có subdomain)

2. **Origins** phải bao gồm tất cả URLs mà client có thể gọi API
   - Frontend URL
   - Backend URL (nếu khác nhau)

3. **Backup State**:
   - `backup_eligible=true`: Passkey có thể sync giữa devices (iCloud Keychain, Google Password Manager)
   - `backup_eligible=false`: Passkey chỉ tồn tại trên device đó (hardware security key)

4. **Production Deployment**:
   - Phải dùng HTTPS
   - RPID phải là domain hợp lệ
   - Không dùng IP address
