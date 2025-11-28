package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"system/configs"
	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// WebAuthnService defines the interface for WebAuthn operations
type WebAuthnService interface {
	// Registration ceremony
	BeginRegistration(ctx context.Context, userID uuid.UUID, clientOrigin string) (*protocol.CredentialCreation, error)
	FinishRegistration(ctx context.Context, userID uuid.UUID, response *protocol.ParsedCredentialCreationData, credentialName *string, clientOrigin string) (*domain.WebAuthnCredential, error)

	// Authentication ceremony
	BeginAuthentication(ctx context.Context, userID uuid.UUID) (*protocol.CredentialAssertion, error)
	FinishAuthentication(ctx context.Context, userID uuid.UUID, response *protocol.ParsedCredentialAssertionData) (*domain.User, error)

	// Credential management
	ListUserCredentials(ctx context.Context, userID uuid.UUID) ([]*domain.WebAuthnCredential, error)
	DeleteCredential(ctx context.Context, userID uuid.UUID, credentialID uuid.UUID) error
	UpdateCredentialName(ctx context.Context, userID uuid.UUID, credentialID uuid.UUID, name string) error
	HasPasskeys(ctx context.Context, userID uuid.UUID) (bool, error)
}

// webauthnService implements WebAuthnService
type webauthnService struct {
	config         configs.WebAuthnConfig
	credentialRepo domain.WebAuthnCredentialRepository
	sessionRepo    domain.WebAuthnSessionRepository
	userRepo       domain.UserRepository
	sessionTimeout time.Duration
	logger         *zap.Logger
}

// NewWebAuthnService creates a new instance of webauthnService
func NewWebAuthnService(
	config configs.WebAuthnConfig,
	credentialRepo domain.WebAuthnCredentialRepository,
	sessionRepo domain.WebAuthnSessionRepository,
	userRepo domain.UserRepository,
	logger *zap.Logger,
) (WebAuthnService, error) {
	svc := &webauthnService{
		config:         config,
		credentialRepo: credentialRepo,
		sessionRepo:    sessionRepo,
		userRepo:       userRepo,
		sessionTimeout: 5 * time.Minute,
		logger:         logger,
	}

	// Validate config by creating a test instance using the centralized helper
	// This ensures we catch config errors at startup
	_, err := svc.initWebAuthn("")
	if err != nil {
		return nil, fmt.Errorf("failed to validate WebAuthn config: %w", err)
	}

	return svc, nil
}

// =================================================================================
// CENTRALIZED CONFIGURATION HELPER (CORE FIX)
// =================================================================================

// generateWebAuthnConfig creates a consistent configuration object
// This is the Single Source of Truth for WebAuthn config
func (s *webauthnService) generateWebAuthnConfig(clientOrigin string) *webauthn.Config {
	// 1. Copy base origins
	origins := make([]string, len(s.config.RPOrigins))
	copy(origins, s.config.RPOrigins)

	// 2. Dynamically add clientOrigin if provided and unique
	if clientOrigin != "" {
		found := false
		for _, o := range origins {
			if o == clientOrigin {
				found = true
				break
			}
		}
		if !found {
			origins = append(origins, clientOrigin)
		}
	}

	wconfig := &webauthn.Config{
		RPDisplayName: s.config.RPName,
		RPID:          s.config.RPID,
		RPOrigins:     origins,
		Debug: 		   true,
		// CRITICAL FIX: Always allow "none" attestation for TouchID/FaceID/Windows Hello
		AttestationPreference: protocol.PreferNoAttestation,
		Timeouts: webauthn.TimeoutsConfig{
			Login: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(s.config.Timeout) * time.Millisecond,
				TimeoutUVD: time.Duration(s.config.Timeout) * time.Millisecond,
			},
			Registration: webauthn.TimeoutConfig{
				Enforce:    true,
				Timeout:    time.Duration(s.config.Timeout) * time.Millisecond,
				TimeoutUVD: time.Duration(s.config.Timeout) * time.Millisecond,
			},
		},
	}

	fmt.Printf("\n====== DEBUG WEBAUTHN CONFIG ======\n")
    fmt.Printf("RPID: '%s'\n", wconfig.RPID)
    fmt.Printf("Origins: %v\n", wconfig.RPOrigins)
    fmt.Printf("AttestationPreference: '%s'\n", wconfig.AttestationPreference)
    fmt.Printf("===================================\n\n")

	// 3. Return the unified config
	return wconfig
}

// initWebAuthn creates a new WebAuthn instance using the centralized config
func (s *webauthnService) initWebAuthn(clientOrigin string) (*webauthn.WebAuthn, error) {
	wconfig := s.generateWebAuthnConfig(clientOrigin)
	return webauthn.New(wconfig)
}

// =================================================================================
// SERVICE METHODS
// =================================================================================

func (s *webauthnService) BeginRegistration(ctx context.Context, userID uuid.UUID, clientOrigin string) (*protocol.CredentialCreation, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	existingCreds, err := s.credentialRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing credentials: %w", err)
	}

	webauthnCreds := convertToWebAuthnCredentials(existingCreds)
	webauthnUser := &webauthnUser{user: user, credentials: webauthnCreds}

	var exclusions []protocol.CredentialDescriptor
	for _, cred := range webauthnCreds {
		exclusions = append(exclusions, protocol.CredentialDescriptor{
			Type:            protocol.PublicKeyCredentialType,
			CredentialID:    cred.ID,
			Transport:       cred.Transport,
			AttestationType: cred.AttestationType,
		})
	}

	// Use helper to init WebAuthn
	webAuthn, err := s.initWebAuthn(clientOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	options, sessionData, err := webAuthn.BeginRegistration(
		webauthnUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: (*bool)(nil),
			ResidentKey:        protocol.ResidentKeyRequirementPreferred,
			UserVerification:   protocol.VerificationPreferred,
		}),
		// CRITICAL FIX: Explicitly prefer no attestation to match config
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
		webauthn.WithExclusions(exclusions),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	session := &domain.WebAuthnSession{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      &userID,
		Challenge:   string(sessionData.Challenge),
		SessionType: domain.SessionTypeRegistration,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(s.sessionTimeout),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return options, nil
}

func (s *webauthnService) FinishRegistration(
	ctx context.Context,
	userID uuid.UUID,
	response *protocol.ParsedCredentialCreationData,
	credentialName *string,
	clientOrigin string,
) (*domain.WebAuthnCredential, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	challengeStr := string(response.Response.CollectedClientData.Challenge)
	session, err := s.sessionRepo.GetByChallenge(ctx, challengeStr)
	if err != nil {
		return nil, pkgerrors.ErrInvalidOrExpiredSession
	}

	if session.UserID == nil || *session.UserID != userID {
		return nil, pkgerrors.ErrInvalidSession
	}
	if session.SessionType != domain.SessionTypeRegistration {
		return nil, pkgerrors.ErrInvalidSession
	}

	existingCreds, err := s.credentialRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing credentials: %w", err)
	}

	webauthnUser := &webauthnUser{user: user, credentials: convertToWebAuthnCredentials(existingCreds)}

	sessionData := webauthn.SessionData{
		Challenge:        session.Challenge,
		UserID:           user.ID.Bytes(),
		UserVerification: protocol.VerificationPreferred,
	}

	// Use helper to init WebAuthn (Ensures config is identical to BeginRegistration)
	webAuthn, err := s.initWebAuthn(clientOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	// Debug log
	s.logger.Info("Attestation details",
		zap.String("format", response.Response.AttestationObject.Format),
		zap.Int("attStmt_len", len(response.Response.AttestationObject.AttStatement)))

	credential, err := webAuthn.CreateCredential(webauthnUser, sessionData, response)
	if err != nil {
		s.logger.Error("WebAuthn CreateCredential failed",
			zap.Error(err),
			zap.String("attestation_format", response.Response.AttestationObject.Format))
		return nil, pkgerrors.ErrCredentialVerificationFailed
	}

	credIDStr := base64.RawURLEncoding.EncodeToString(credential.ID)
	if _, err := s.credentialRepo.GetByCredentialID(ctx, credIDStr); err == nil {
		return nil, pkgerrors.ErrCredentialAlreadyExists
	}

	transports := make([]string, len(credential.Transport))
	for i, t := range credential.Transport {
		transports[i] = string(t)
	}

	domainCred := &domain.WebAuthnCredential{
		ID:              uuid.Must(uuid.NewV7()),
		UserID:          userID,
		CredentialID:    credIDStr,
		PublicKey:       credential.PublicKey,
		AttestationType: domain.AttestationType(credential.AttestationType),
		AAGUID:          credential.Authenticator.AAGUID,
		SignCount:       int32(credential.Authenticator.SignCount),
		Transports:      transports,
		BackupEligible:  credential.Flags.BackupEligible,
		BackupState:     credential.Flags.BackupState,
		CredentialName:  credentialName,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.credentialRepo.Create(ctx, domainCred); err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	_ = s.sessionRepo.Delete(ctx, session.ID)

	return domainCred, nil
}

func (s *webauthnService) BeginAuthentication(ctx context.Context, userID uuid.UUID) (*protocol.CredentialAssertion, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	existingCreds, err := s.credentialRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	if len(existingCreds) == 0 {
		return nil, pkgerrors.ErrNoPasskeyRegistered
	}

	webauthnCreds := convertToWebAuthnCredentials(existingCreds)
	webauthnUser := &webauthnUser{user: user, credentials: webauthnCreds}

	// Init WebAuthn with empty origin (default) for login
	webAuthn, err := s.initWebAuthn("")
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	options, sessionData, err := webAuthn.BeginLogin(
		webauthnUser,
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
	}

	session := &domain.WebAuthnSession{
		ID:          uuid.Must(uuid.NewV7()),
		UserID:      &userID,
		Challenge:   sessionData.Challenge,
		SessionType: domain.SessionTypeAuthentication,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(s.sessionTimeout),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return options, nil
}

func (s *webauthnService) FinishAuthentication(
	ctx context.Context,
	userID uuid.UUID,
	response *protocol.ParsedCredentialAssertionData,
) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	challengeStr := string(response.Response.CollectedClientData.Challenge)
	session, err := s.sessionRepo.GetByChallenge(ctx, challengeStr)
	if err != nil {
		return nil, pkgerrors.ErrInvalidOrExpiredSession
	}

	if session.UserID == nil || *session.UserID != userID {
		return nil, pkgerrors.ErrInvalidSession
	}
	if session.SessionType != domain.SessionTypeAuthentication {
		return nil, pkgerrors.ErrInvalidSession
	}

	existingCreds, err := s.credentialRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	webauthnUser := &webauthnUser{user: user, credentials: convertToWebAuthnCredentials(existingCreds)}

	sessionData := webauthn.SessionData{
		Challenge:        session.Challenge,
		UserID:           user.ID.Bytes(),
		UserVerification: protocol.VerificationPreferred,
	}

	webAuthn, err := s.initWebAuthn("")
	if err != nil {
		return nil, fmt.Errorf("failed to create WebAuthn instance: %w", err)
	}

	credential, err := webAuthn.ValidateLogin(webauthnUser, sessionData, response)
	if err != nil {
		return nil, pkgerrors.ErrAuthenticationFailed
	}

	credIDStr := base64.RawURLEncoding.EncodeToString(credential.ID)
	// Ignore error on stats update, don't block login
	_ = s.credentialRepo.UpdateSignCount(ctx, credIDStr, int32(credential.Authenticator.SignCount))
	_ = s.userRepo.UpdateLastLogin(ctx, userID)
	_ = s.sessionRepo.Delete(ctx, session.ID)

	return user, nil
}

func (s *webauthnService) ListUserCredentials(ctx context.Context, userID uuid.UUID) ([]*domain.WebAuthnCredential, error) {
	return s.credentialRepo.GetByUserID(ctx, userID)
}

func (s *webauthnService) DeleteCredential(ctx context.Context, userID uuid.UUID, credentialID uuid.UUID) error {
	credential, err := s.credentialRepo.GetByID(ctx, credentialID)
	if err != nil {
		return pkgerrors.ErrCredentialNotFound
	}
	if credential.UserID != userID {
		return pkgerrors.ErrCredentialNotFound
	}
	return s.credentialRepo.Delete(ctx, credentialID)
}

func (s *webauthnService) UpdateCredentialName(ctx context.Context, userID uuid.UUID, credentialID uuid.UUID, name string) error {
	credential, err := s.credentialRepo.GetByID(ctx, credentialID)
	if err != nil {
		return pkgerrors.ErrCredentialNotFound
	}
	if credential.UserID != userID {
		return pkgerrors.ErrCredentialNotFound
	}
	credential.CredentialName = &name
	credential.UpdatedAt = time.Now()
	return s.credentialRepo.Update(ctx, credential)
}

func (s *webauthnService) HasPasskeys(ctx context.Context, userID uuid.UUID) (bool, error) {
	credentials, err := s.credentialRepo.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return len(credentials) > 0, nil
}

// =================================================================================
// HELPER TYPES & FUNCTIONS
// =================================================================================

type webauthnUser struct {
	user        *domain.User
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte { return u.user.ID.Bytes() }
func (u *webauthnUser) WebAuthnName() string { return u.user.Email }
func (u *webauthnUser) WebAuthnDisplayName() string {
	if u.user.FullName != nil {
		return *u.user.FullName
	}
	return u.user.Email
}
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }
func (u *webauthnUser) WebAuthnIcon() string {
	if u.user.AvatarURL != nil {
		return *u.user.AvatarURL
	}
	return ""
}

func convertToWebAuthnCredentials(credentials []*domain.WebAuthnCredential) []webauthn.Credential {
	result := make([]webauthn.Credential, len(credentials))
	for i, cred := range credentials {
		credIDBytes, _ := base64.RawURLEncoding.DecodeString(cred.CredentialID)
		transports := make([]protocol.AuthenticatorTransport, len(cred.Transports))
		for j, t := range cred.Transports {
			transports[j] = protocol.AuthenticatorTransport(t)
		}
		result[i] = webauthn.Credential{
			ID:              credIDBytes,
			PublicKey:       cred.PublicKey,
			AttestationType: string(cred.AttestationType),
			Transport:       transports,
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