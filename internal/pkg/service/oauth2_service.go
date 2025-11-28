package service

import (
	"context"
	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/crypto"
	"system/pkg/util/random"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OAuth2Service chứa business logic cho OAuth2 operations.
type OAuth2Service struct {
	userRepo             domain.UserRepository
	sessionRepo          domain.SessionRepository
	authRequestRepo      domain.AuthRequestRepository
	consentRepo          domain.ConsentRepository
	oauth2SessionRepo    domain.OAuth2SessionRepository
	clientRepo           domain.OAuth2ClientRepository
}

// NewOAuth2Service tạo instance mới của OAuth2Service.
func NewOAuth2Service(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	authRequestRepo domain.AuthRequestRepository,
	consentRepo domain.ConsentRepository,
	oauth2SessionRepo domain.OAuth2SessionRepository,
	clientRepo domain.OAuth2ClientRepository,
) *OAuth2Service {
	return &OAuth2Service{
		userRepo:          userRepo,
		sessionRepo:       sessionRepo,
		authRequestRepo:   authRequestRepo,
		consentRepo:       consentRepo,
		oauth2SessionRepo: oauth2SessionRepo,
		clientRepo:        clientRepo,
	}
}

// AuthenticateUser xác thực user với email và password.
// Trả về user nếu thành công, business error nếu thất bại.
func (s *OAuth2Service) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, error) {
	// Get user từ database
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Convert technical error to business error
		return nil, pkgerrors.ErrInvalidCredentials
	}

	// Verify password với bcrypt
	if !crypto.VerifyPassword(user.PasswordHash, password) {
		return nil, pkgerrors.ErrInvalidCredentials
	}

	return user, nil
}

// CreateUserSession tạo session mới cho user.
// Trả về sessionID.
func (s *OAuth2Service) CreateUserSession(ctx context.Context, userID uuid.UUID, ttl time.Duration) (string, error) {
	// Generate secure random session ID
	sessionID, err := random.GenerateSessionID()
	if err != nil {
		return "", err
	}

	// Store session in Redis
	err = s.sessionRepo.CreateSession(ctx, sessionID, userID.String(), ttl)
	if err != nil {
		return "", err
	}

	// Update last login time (async, không cần chờ)
	go func() {
		_ = s.userRepo.UpdateLastLogin(context.Background(), userID)
	}()

	return sessionID, nil
}

// GetUserSession lấy userID từ sessionID.
func (s *OAuth2Service) GetUserSession(ctx context.Context, sessionID string) (string, error) {
	return s.sessionRepo.GetSession(ctx, sessionID)
}

// DeleteUserSession xóa session.
func (s *OAuth2Service) DeleteUserSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

// GetUserByID lấy thông tin user theo ID.
func (s *OAuth2Service) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// CheckUserConsent kiểm tra xem user đã consent cho client chưa.
func (s *OAuth2Service) CheckUserConsent(ctx context.Context, userID, clientID uuid.UUID) (bool, error) {
	consent, err := s.consentRepo.GetActiveConsent(ctx, userID, clientID)
	if err != nil || consent == nil {
		return false, err
	}

	return !consent.Revoked, nil
}

// CreateUserConsent tạo consent mới cho user.
func (s *OAuth2Service) CreateUserConsent(ctx context.Context, userID, clientID uuid.UUID, scopes []string) error {
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
func (s *OAuth2Service) RevokeUserConsent(ctx context.Context, userID, clientID uuid.UUID) error {
	return s.consentRepo.RevokeConsent(ctx, userID, clientID)
}

// GetUserConsents lấy danh sách consents của user.
func (s *OAuth2Service) GetUserConsents(ctx context.Context, userID uuid.UUID, includeRevoked bool) ([]*domain.OAuth2Consent, error) {
	return s.consentRepo.GetUserConsents(ctx, userID, includeRevoked)
}

// LogoutUser thực hiện logout hoàn toàn: xóa session và optionally revoke tokens.
func (s *OAuth2Service) LogoutUser(ctx context.Context, sessionID string, revokeTokens bool) error {
	// Get userID từ session trước khi xóa
	userID, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		// Session không tồn tại hoặc đã expired - không coi là lỗi
		return nil
	}

	// Xóa session khỏi Redis
	if err := s.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
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
func (s *OAuth2Service) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	return s.oauth2SessionRepo.RevokeAllUserSessions(ctx, userID.String())
}

// GetClientInfo lấy thông tin OAuth2 client theo ID.
func (s *OAuth2Service) GetClientInfo(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error) {
	return s.clientRepo.GetByID(ctx, clientID)
}

// GetUserByEmail lấy thông tin user theo email.
func (s *OAuth2Service) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}
