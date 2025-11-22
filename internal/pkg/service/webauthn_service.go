package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/configs"
	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// WebAuthnService xử lý WebAuthn/Passkey authentication logic.
type WebAuthnService struct {
	webAuthn       *webauthn.WebAuthn
	userRepo       domain.UserRepository
	credentialRepo domain.WebAuthnCredentialRepository
	sessionRepo    domain.WebAuthnSessionRepository
	config         *configs.WebAuthnConfig
}

// NewWebAuthnService tạo instance mới của WebAuthnService.
func NewWebAuthnService(
	cfg *configs.WebAuthnConfig,
	userRepo domain.UserRepository,
	credentialRepo domain.WebAuthnCredentialRepository,
	sessionRepo domain.WebAuthnSessionRepository,
) (*WebAuthnService, error) {
	// Khởi tạo WebAuthn instance
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.RPName,
		RPID:          cfg.RPID,
		RPOrigins:     cfg.RPOrigins,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
				TimeoutUVD: time.Duration(cfg.Timeout) * time.Millisecond,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(cfg.Timeout) * time.Millisecond,
				TimeoutUVD: time.Duration(cfg.Timeout) * time.Millisecond,
			},
		},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn instance: %w", err)
	}

	return &WebAuthnService{
		webAuthn:       w,
		userRepo:       userRepo,
		credentialRepo: credentialRepo,
		sessionRepo:    sessionRepo,
		config:         cfg,
	}, nil
}

// BeginRegistration bắt đầu quá trình đăng ký passkey cho user mới hoặc hiện tại.
func (s *WebAuthnService) BeginRegistration(ctx context.Context, email, fullName string) (*protocol.CredentialCreation, error) {
	// Kiểm tra xem user đã tồn tại chưa
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Nếu user chưa tồn tại, tạo user mới (passwordless)
	if user == nil {
		user = &domain.User{
			ID:            uuid.Must(uuid.NewV7()),
			Email:         email,
			EmailVerified: false,
			PasswordHash:  "", // Passwordless account
			FullName:      &fullName,
			Status:        "active",
			Settings: map[string]any{
				"theme":                 "system",
				"language":              "en",
				"notifications_enabled": true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Lấy credentials hiện có của user để exclude
	existingCreds, err := s.credentialRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing credentials: %w", err)
	}

	// Tạo WebAuthn user wrapper
	webAuthnUser := &webAuthnUser{
		id:          user.ID.Bytes(),
		name:        user.Email,
		displayName: getDisplayName(user),
		credentials: convertToWebAuthnCredentials(existingCreds),
	}

	// Bắt đầu registration ceremony
	options, session, err := s.webAuthn.BeginRegistration(webAuthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Lưu session vào database
	challengeB64 := base64.RawURLEncoding.EncodeToString(session.Challenge)
	webAuthnSession := &domain.WebAuthnSession{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      &user.ID,
		Challenge:   challengeB64,
		SessionType: domain.SessionTypeRegistration,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	if err := s.sessionRepo.Create(ctx, webAuthnSession); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return options, nil
}

// FinishRegistration hoàn thành quá trình đăng ký passkey.
func (s *WebAuthnService) FinishRegistration(ctx context.Context, email string, credentialName *string, response *protocol.ParsedCredentialCreationData) (*domain.User, *domain.WebAuthnCredential, error) {
	// Lấy user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, pkgerrors.ErrUserNotFound
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Lấy session từ challenge
	challengeB64 := base64.RawURLEncoding.EncodeToString(response.Response.CollectedClientData.Challenge)
	session, err := s.sessionRepo.GetByChallenge(ctx, challengeB64)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, pkgerrors.ErrInvalidOrExpiredSession
		}
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Verify session belongs to user
	if session.UserID == nil || *session.UserID != user.ID {
		return nil, nil, pkgerrors.ErrInvalidSession
	}

	// Lấy credentials hiện có
	existingCreds, err := s.credentialRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get existing credentials: %w", err)
	}

	// Tạo WebAuthn user wrapper
	webAuthnUser := &webAuthnUser{
		id:          user.ID.Bytes(),
		name:        user.Email,
		displayName: getDisplayName(user),
		credentials: convertToWebAuthnCredentials(existingCreds),
	}

	// Tạo session data từ database session
	sessionData := &webauthn.SessionData{
		Challenge:            response.Response.CollectedClientData.Challenge,
		UserID:               user.ID.Bytes(),
		UserVerification:     protocol.VerificationPreferred,
		AllowedCredentialIDs: [][]byte{},
	}

	// Verify registration
	credential, err := s.webAuthn.CreateCredential(webAuthnUser, *sessionData, response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify registration: %w", err)
	}

	// Lưu credential vào database
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	webAuthnCred := &domain.WebAuthnCredential{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          user.ID,
		CredentialID:    credentialID,
		PublicKey:       credential.PublicKey,
		AttestationType: domain.AttestationType(credential.AttestationType),
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       int32(credential.Authenticator.SignCount),
		Transports:      convertTransports(credential.Transport),
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		CredentialName:  credentialName,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.credentialRepo.Create(ctx, webAuthnCred); err != nil {
		return nil, nil, fmt.Errorf("failed to save credential: %w", err)
	}

	// Xóa session
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		// Log but don't fail
	}

	return user, webAuthnCred, nil
}

// BeginAuthentication bắt đầu quá trình xác thực bằng passkey.
func (s *WebAuthnService) BeginAuthentication(ctx context.Context, email string) (*protocol.CredentialAssertion, error) {
	// Lấy user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Lấy credentials của user
	credentials, err := s.credentialRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	if len(credentials) == 0 {
		return nil, pkgerrors.ErrNoPasskeyRegistered
	}

	// Tạo WebAuthn user wrapper
	webAuthnUser := &webAuthnUser{
		id:          user.ID.Bytes(),
		name:        user.Email,
		displayName: getDisplayName(user),
		credentials: convertToWebAuthnCredentials(credentials),
	}

	// Bắt đầu authentication ceremony
	options, session, err := s.webAuthn.BeginLogin(webAuthnUser)
	if err != nil {
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
	}

	// Lưu session vào database
	challengeB64 := base64.RawURLEncoding.EncodeToString(session.Challenge)
	webAuthnSession := &domain.WebAuthnSession{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      &user.ID,
		Challenge:   challengeB64,
		SessionType: domain.SessionTypeAuthentication,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	if err := s.sessionRepo.Create(ctx, webAuthnSession); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return options, nil
}

// FinishAuthentication hoàn thành quá trình xác thực bằng passkey.
func (s *WebAuthnService) FinishAuthentication(ctx context.Context, email string, response *protocol.ParsedCredentialAssertionData) (*domain.User, *domain.WebAuthnCredential, error) {
	// Lấy user
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, pkgerrors.ErrUserNotFound
		}
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Lấy session từ challenge
	challengeB64 := base64.RawURLEncoding.EncodeToString(response.Response.CollectedClientData.Challenge)
	session, err := s.sessionRepo.GetByChallenge(ctx, challengeB64)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, pkgerrors.ErrInvalidOrExpiredSession
		}
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Verify session belongs to user
	if session.UserID == nil || *session.UserID != user.ID {
		return nil, nil, pkgerrors.ErrInvalidSession
	}

	// Lấy credentials của user
	credentials, err := s.credentialRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Tạo WebAuthn user wrapper
	webAuthnUser := &webAuthnUser{
		id:          user.ID.Bytes(),
		name:        user.Email,
		displayName: getDisplayName(user),
		credentials: convertToWebAuthnCredentials(credentials),
	}

	// Tạo session data từ database session
	sessionData := &webauthn.SessionData{
		Challenge:            response.Response.CollectedClientData.Challenge,
		UserID:               user.ID.Bytes(),
		UserVerification:     protocol.VerificationPreferred,
		AllowedCredentialIDs: getAllowedCredentialIDs(credentials),
	}

	// Verify authentication
	credential, err := s.webAuthn.ValidateLogin(webAuthnUser, *sessionData, response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify authentication: %w", err)
	}

	// Cập nhật sign count
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	if err := s.credentialRepo.UpdateSignCount(ctx, credentialID, int32(credential.Authenticator.SignCount)); err != nil {
		// Log but don't fail
	}

	// Lấy credential từ database
	dbCredential, err := s.credentialRepo.GetByCredentialID(ctx, credentialID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Xóa session
	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		// Log but don't fail
	}

	// Cập nhật last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log but don't fail
	}

	return user, dbCredential, nil
}

// GetUserCredentials lấy danh sách credentials của user.
func (s *WebAuthnService) GetUserCredentials(ctx context.Context, userID uuid.UUID) ([]*domain.WebAuthnCredential, error) {
	return s.credentialRepo.GetByUserID(ctx, userID)
}

// DeleteCredential xóa một credential.
func (s *WebAuthnService) DeleteCredential(ctx context.Context, userID uuid.UUID, credentialID string) error {
	// Verify credential belongs to user
	credential, err := s.credentialRepo.GetByCredentialID(ctx, credentialID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrCredentialNotFound
		}
		return fmt.Errorf("failed to get credential: %w", err)
	}

	if credential.UserID != userID {
		return pkgerrors.ErrUnauthorized
	}

	return s.credentialRepo.Delete(ctx, credential.ID)
}

// UpdateCredentialName cập nhật tên của credential.
func (s *WebAuthnService) UpdateCredentialName(ctx context.Context, userID uuid.UUID, credentialID string, name string) error {
	// Verify credential belongs to user
	credential, err := s.credentialRepo.GetByCredentialID(ctx, credentialID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrCredentialNotFound
		}
		return fmt.Errorf("failed to get credential: %w", err)
	}

	if credential.UserID != userID {
		return pkgerrors.ErrUnauthorized
	}

	credential.CredentialName = &name
	credential.UpdatedAt = time.Now()

	return s.credentialRepo.Update(ctx, credential)
}

// ==================== Helper Types & Functions ====================

// webAuthnUser implements webauthn.User interface
type webAuthnUser struct {
	id          []byte
	name        string
	displayName string
	credentials []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte {
	return u.id
}

func (u *webAuthnUser) WebAuthnName() string {
	return u.name
}

func (u *webAuthnUser) WebAuthnDisplayName() string {
	return u.displayName
}

func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (u *webAuthnUser) WebAuthnIcon() string {
	return ""
}

// getDisplayName returns user's display name
func getDisplayName(user *domain.User) string {
	if user.FullName != nil && *user.FullName != "" {
		return *user.FullName
	}
	return user.Email
}

// convertToWebAuthnCredentials converts domain credentials to webauthn credentials
func convertToWebAuthnCredentials(creds []*domain.WebAuthnCredential) []webauthn.Credential {
	result := make([]webauthn.Credential, len(creds))
	for i, cred := range creds {
		credID, _ := base64.RawURLEncoding.DecodeString(cred.CredentialID)
		result[i] = webauthn.Credential{
			ID:              credID,
			PublicKey:       cred.PublicKey,
			AttestationType: string(cred.AttestationType),
			Transport:       convertToProtocolTransports(cred.Transports),
			Flags: webauthn.CredentialFlags{
				BackupEligible: cred.BackupEligible,
				BackupState:    cred.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:    cred.AAGUID,
				SignCount: uint32(cred.SignCount),
			},
		}
	}
	return result
}

// convertTransports converts protocol transports to string slice
func convertTransports(transports []protocol.AuthenticatorTransport) []string {
	result := make([]string, len(transports))
	for i, t := range transports {
		result[i] = string(t)
	}
	return result
}

// convertToProtocolTransports converts string slice to protocol transports
func convertToProtocolTransports(transports []string) []protocol.AuthenticatorTransport {
	result := make([]protocol.AuthenticatorTransport, len(transports))
	for i, t := range transports {
		result[i] = protocol.AuthenticatorTransport(t)
	}
	return result
}

// getAllowedCredentialIDs gets allowed credential IDs for authentication
func getAllowedCredentialIDs(creds []*domain.WebAuthnCredential) [][]byte {
	result := make([][]byte, len(creds))
	for i, cred := range creds {
		credID, _ := base64.RawURLEncoding.DecodeString(cred.CredentialID)
		result[i] = credID
	}
	return result
}
