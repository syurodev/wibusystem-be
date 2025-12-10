package oauth2

import (
	"context"
	"strings"
	"system/internal/domain"
	"system/internal/modules/user"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/crypto"
	"system/pkg/util/stringutil"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OAuth2Service chứa business logic cho OAuth2 operations.
type oauth2ServiceImpl struct {
	userService          user.UserService // Use UserService instead of UserRepo and SessionRepo
	authRequestRepo      domain.AuthRequestRepository
	consentRepo          domain.ConsentRepository
	oauth2SessionRepo    domain.OAuth2SessionRepository
	clientRepo           domain.OAuth2ClientRepository
}

// NewOAuth2Service tạo instance mới của OAuth2Service.
func NewService(
	userService user.UserService,
	authRequestRepo domain.AuthRequestRepository,
	consentRepo domain.ConsentRepository,
	oauth2SessionRepo domain.OAuth2SessionRepository,
	clientRepo domain.OAuth2ClientRepository,
) *oauth2ServiceImpl {
	return &oauth2ServiceImpl{
		userService:       userService,
		authRequestRepo:   authRequestRepo,
		consentRepo:       consentRepo,
		oauth2SessionRepo: oauth2SessionRepo,
		clientRepo:        clientRepo,
	}
}

// AuthenticateUser xác thực user với email/username và password.
// Nếu identifier chứa '@' sẽ tìm theo email, ngược lại tìm theo username.
// Trả về user nếu thành công, business error nếu thất bại.
func (s *oauth2ServiceImpl) AuthenticateUser(ctx context.Context, identifier, password string) (*domain.User, error) {
	var user *domain.User
	var err error

	// Detect email vs username
	if strings.Contains(identifier, "@") {
		// Tìm theo email
		user, err = s.userService.GetUserByEmail(ctx, identifier)
	} else {
		// Tìm theo username
		user, err = s.userService.GetUserByUsername(ctx, identifier)
	}

	if err != nil {
		// Convert technical error to business error
		return nil, pkgerrors.Unauthorized(I18nInvalidCredentials, "invalid credentials")
	}

	// Verify password với bcrypt
	if !crypto.VerifyPassword(user.PasswordHash, password) {
		return nil, pkgerrors.Unauthorized(I18nInvalidCredentials, "invalid credentials")
	}

	return user, nil
}

// CreateUserSession tạo session mới cho user.
// Trả về sessionID.
func (s *oauth2ServiceImpl) CreateUserSession(ctx context.Context, userID uuid.UUID, ttl time.Duration, userAgent, ip string) (string, error) {
	// Generate secure random session ID
	sessionID, err := stringutil.GenerateSessionID()
	if err != nil {
		return "", err
	}

	session := &domain.UserSession{
		SessionID:  sessionID,
		UserID:     userID.String(),
		UserAgent:  userAgent,
		IP:         ip,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		LastActive: time.Now(),
	}

	// Store session in Redis
	err = s.userService.CreateSession(ctx, session, ttl)
	if err != nil {
		return "", err
	}

	// Update last login time (async, không cần chờ)
	go func() {
		_ = s.userService.UpdateLastLogin(context.Background(), userID)
	}()

	return sessionID, nil
}

// GetUserSession lấy userID từ sessionID.
func (s *oauth2ServiceImpl) GetUserSession(ctx context.Context, sessionID string) (string, error) {
	session, err := s.userService.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	return session.UserID, nil
}

// DeleteUserSession xóa session.
func (s *oauth2ServiceImpl) DeleteUserSession(ctx context.Context, sessionID string) error {
	// Need userID to delete session via UserService? 
	// UserService.DeleteSession(ctx, userIDStr, sessionID)
	// But OAuth2Service previously just used sessionRepo.DeleteSession(ctx, sessionID)
	// Let's use UserService.GetSession first to get userID
	
	session, err := s.userService.GetSession(ctx, sessionID)
	if err != nil {
		return nil // Session already gone
	}
	
	return s.userService.DeleteSession(ctx, session.UserID, sessionID)
}

// GetUserByID lấy thông tin user theo ID.
func (s *oauth2ServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userService.GetUserByID(ctx, userID)
}

// CheckUserConsent kiểm tra xem user đã consent cho client chưa.
func (s *oauth2ServiceImpl) CheckUserConsent(ctx context.Context, userID, clientID uuid.UUID) (bool, error) {
	consent, err := s.consentRepo.GetActiveConsent(ctx, userID, clientID)
	if err != nil || consent == nil {
		return false, err
	}

	return !consent.Revoked, nil
}

// CreateUserConsent tạo consent mới cho user.
func (s *oauth2ServiceImpl) CreateUserConsent(ctx context.Context, userID, clientID uuid.UUID, scopes []string) error {
	consent := &domain.OAuth2Consent{
		ID:            uuid.Must(uuid.NewV4()),
		UserID:        userID,
		ClientID:      clientID,
		GrantedScopes: scopes,
		Revoked:       false,
		GrantedAt:     time.Now(),
		ConsentMethod: domain.ConsentMethodExplicit,
	}

	return s.consentRepo.CreateConsent(ctx, consent)
}

// RevokeUserConsent thu hồi consent.
func (s *oauth2ServiceImpl) RevokeUserConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	return s.consentRepo.RevokeConsent(ctx, userID, clientID)
}

// GetUserConsents lấy danh sách consents của user.
func (s *oauth2ServiceImpl) GetUserConsents(ctx context.Context, userID uuid.UUID, includeRevoked bool) ([]*domain.OAuth2Consent, error) {
	return s.consentRepo.GetUserConsents(ctx, userID, includeRevoked)
}

// LogoutUser thực hiện logout hoàn toàn: xóa session và optionally revoke tokens.
func (s *oauth2ServiceImpl) LogoutUser(ctx context.Context, sessionID string, revokeTokens bool) error {
	// Get userID từ session trước khi xóa
	userID, err := s.GetUserSession(ctx, sessionID)
	if err != nil {
		// Session không tồn tại hoặc đã expired - không coi là lỗi
		return nil
	}

	// Xóa session khỏi Redis using UserService
	// Need to check if user still exists? getUserSession returns string
	if err := s.DeleteUserSession(ctx, sessionID); err != nil {
		return err
	}

	// Revoke tất cả OAuth2 tokens nếu được yêu cầu
	if revokeTokens && userID != "" {
		if err := s.oauth2SessionRepo.RevokeAllUserSessions(ctx, userID); err != nil {
			// Log error nhưng không fail logout
			return err
		}
	}

	return nil
}

// RevokeUserTokens revokes tất cả OAuth2 tokens của một user.
func (s *oauth2ServiceImpl) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.oauth2SessionRepo.RevokeAllUserSessions(ctx, userID.String())
}

// GetClientInfo lấy thông tin OAuth2 client theo ID.
func (s *oauth2ServiceImpl) GetClientInfo(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error) {
	return s.clientRepo.GetByID(ctx, clientID)
}

// GetUserByEmail lấy thông tin user theo email.
func (s *oauth2ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userService.GetUserByEmail(ctx, email)
}

// GetUserByIdentifier lấy thông tin user theo email hoặc username.
// Nếu identifier chứa '@' sẽ tìm theo email, ngược lại tìm theo username.
func (s *oauth2ServiceImpl) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	if strings.Contains(identifier, "@") {
		return s.userService.GetUserByEmail(ctx, identifier)
	}
	return s.userService.GetUserByUsername(ctx, identifier)
}

// GetGlobalPermissions lấy danh sách global permissions của user.
func (s *oauth2ServiceImpl) GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.userService.GetGlobalPermissions(ctx, userID)
}

// GetOrganizationPermissions lấy danh sách permissions của user trong organization.
func (s *oauth2ServiceImpl) GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	return s.userService.GetOrganizationPermissions(ctx, userID, organizationID)
}

// GetGlobalRoles lấy danh sách global roles của user.
func (s *oauth2ServiceImpl) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.userService.GetGlobalRoles(ctx, userID)
}

// GetOrganizationRoles lấy danh sách roles của user trong organization.
func (s *oauth2ServiceImpl) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	return s.userService.GetOrganizationRoles(ctx, userID, organizationID)
}
