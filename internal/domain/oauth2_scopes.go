package domain

// Scope định nghĩa một kiểu tùy chỉnh cho OAuth2 scopes để tăng tính an toàn về kiểu.
type Scope string

// Các hằng số cho các OAuth2/OIDC scopes đã biết.
const (
	// ScopeOfflineAccess là scope tiêu chuẩn cho phép cấp refresh token.
	ScopeOfflineAccess Scope = "offline_access"
	// ScopeOffline là một alias phổ biến cho offline_access.
	ScopeOffline Scope = "offline"

	// ScopeOpenID là scope bắt buộc cho tất cả các luồng OpenID Connect.
	ScopeOpenID Scope = "openid"
	// ScopeProfile yêu cầu quyền truy cập vào các thông tin profile mặc định của người dùng.
	ScopeProfile Scope = "profile"
	// ScopeEmail yêu cầu quyền truy cập vào địa chỉ email của người dùng.
	ScopeEmail Scope = "email"

	// ScopeInternal yêu cầu quyền truy cập vào roles và permissions (chỉ dành cho internal clients).
	ScopeInternal Scope = "internal"
)
