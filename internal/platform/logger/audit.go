package logger

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// AuditEventType định nghĩa loại audit event
type AuditEventType string

const (
	// Authentication Events
	EventLoginAttempt       AuditEventType = "auth.login.attempt"
	EventLoginSuccess       AuditEventType = "auth.login.success"
	EventLoginFailure       AuditEventType = "auth.login.failure"
	EventLogout             AuditEventType = "auth.logout"
	EventPasswordChange     AuditEventType = "auth.password.change"
	EventPasswordResetReq   AuditEventType = "auth.password.reset.request"
	EventPasswordReset      AuditEventType = "auth.password.reset"
	EventEmailVerification  AuditEventType = "auth.email.verification"
	EventAccountCreated     AuditEventType = "auth.account.created"
	EventAccountDeleted     AuditEventType = "auth.account.deleted"
	EventAccountLocked      AuditEventType = "auth.account.locked"
	EventAccountUnlocked    AuditEventType = "auth.account.unlocked"

	// OAuth2 Events
	EventOAuth2Authorize          AuditEventType = "oauth2.authorize"
	EventOAuth2TokenIssued        AuditEventType = "oauth2.token.issued"
	EventOAuth2TokenRefreshed     AuditEventType = "oauth2.token.refreshed"
	EventOAuth2TokenRevoked       AuditEventType = "oauth2.token.revoked"
	EventOAuth2TokenIntrospected  AuditEventType = "oauth2.token.introspected"
	EventOAuth2ConsentGranted     AuditEventType = "oauth2.consent.granted"
	EventOAuth2ConsentRevoked     AuditEventType = "oauth2.consent.revoked"
	EventOAuth2ConsentDenied      AuditEventType = "oauth2.consent.denied"

	// OAuth2 Client Management Events
	EventOAuth2ClientCreated      AuditEventType = "oauth2.client.created"
	EventOAuth2ClientUpdated      AuditEventType = "oauth2.client.updated"
	EventOAuth2ClientDeleted      AuditEventType = "oauth2.client.deleted"
	EventOAuth2ClientSecretRegen  AuditEventType = "oauth2.client.secret.regenerated"

	// Session Events
	EventSessionCreated     AuditEventType = "session.created"
	EventSessionExpired     AuditEventType = "session.expired"
	EventSessionInvalidated AuditEventType = "session.invalidated"

	// Security Events
	EventUnauthorizedAccess   AuditEventType = "security.unauthorized.access"
	EventSuspiciousActivity   AuditEventType = "security.suspicious.activity"
	EventRateLimitExceeded    AuditEventType = "security.rate_limit.exceeded"
	EventInvalidToken         AuditEventType = "security.invalid.token"
	EventCSRFDetected         AuditEventType = "security.csrf.detected"

	// Admin Events
	EventAdminAction          AuditEventType = "admin.action"
	EventConfigurationChange  AuditEventType = "admin.config.change"
)

// AuditEventStatus định nghĩa kết quả của audit event
type AuditEventStatus string

const (
	StatusSuccess AuditEventStatus = "success"
	StatusFailure AuditEventStatus = "failure"
	StatusPending AuditEventStatus = "pending"
	StatusDenied  AuditEventStatus = "denied"
)

// AuditEvent chứa thông tin về một audit event
type AuditEvent struct {
	// Core fields
	EventType   AuditEventType   `json:"event_type"`
	Status      AuditEventStatus `json:"status"`
	Timestamp   time.Time        `json:"timestamp"`

	// Actor information (who performed the action)
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	Username    string     `json:"username,omitempty"`
	ClientID    *uuid.UUID `json:"client_id,omitempty"`

	// Request context
	IPAddress   string     `json:"ip_address,omitempty"`
	UserAgent   string     `json:"user_agent,omitempty"`
	RequestID   string     `json:"request_id,omitempty"`

	// Target information (what was affected)
	TargetType  string     `json:"target_type,omitempty"`  // user, client, token, session
	TargetID    string     `json:"target_id,omitempty"`

	// OAuth2 specific
	GrantType   string     `json:"grant_type,omitempty"`
	Scopes      []string   `json:"scopes,omitempty"`
	TokenType   string     `json:"token_type,omitempty"`   // access_token, refresh_token, id_token

	// Additional metadata
	Message     string     `json:"message,omitempty"`
	ErrorCode   string     `json:"error_code,omitempty"`
	ErrorDetail string     `json:"error_detail,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`

	// Performance metrics
	Duration    int64      `json:"duration_ms,omitempty"`  // milliseconds
}

// AuditLogger provides methods for logging audit events
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger creates a new AuditLogger instance
func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger.Named("audit"),
	}
}

// LogEvent logs an audit event với structured fields
func (al *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	fields := []zap.Field{
		zap.String("event_type", string(event.EventType)),
		zap.String("status", string(event.Status)),
		zap.Time("timestamp", event.Timestamp),
		zap.String("category", "audit"),
	}

	// Add user information
	if event.UserID != nil {
		fields = append(fields, zap.String("user_id", event.UserID.String()))
	}
	if event.Username != "" {
		fields = append(fields, zap.String("username", event.Username))
	}
	if event.ClientID != nil {
		fields = append(fields, zap.String("client_id", event.ClientID.String()))
	}

	// Add request context
	if event.IPAddress != "" {
		fields = append(fields, zap.String("ip_address", event.IPAddress))
	}
	if event.UserAgent != "" {
		fields = append(fields, zap.String("user_agent", event.UserAgent))
	}
	if event.RequestID != "" {
		fields = append(fields, zap.String("request_id", event.RequestID))
	}

	// Add target information
	if event.TargetType != "" {
		fields = append(fields, zap.String("target_type", event.TargetType))
	}
	if event.TargetID != "" {
		fields = append(fields, zap.String("target_id", event.TargetID))
	}

	// Add OAuth2 specific fields
	if event.GrantType != "" {
		fields = append(fields, zap.String("grant_type", event.GrantType))
	}
	if len(event.Scopes) > 0 {
		fields = append(fields, zap.Strings("scopes", event.Scopes))
	}
	if event.TokenType != "" {
		fields = append(fields, zap.String("token_type", event.TokenType))
	}

	// Add error information
	if event.ErrorCode != "" {
		fields = append(fields, zap.String("error_code", event.ErrorCode))
	}
	if event.ErrorDetail != "" {
		fields = append(fields, zap.String("error_detail", event.ErrorDetail))
	}

	// Add performance metrics
	if event.Duration > 0 {
		fields = append(fields, zap.Int64("duration_ms", event.Duration))
	}

	// Add metadata
	if len(event.Metadata) > 0 {
		fields = append(fields, zap.Any("metadata", event.Metadata))
	}

	// Log với level phù hợp
	message := event.Message
	if message == "" {
		message = string(event.EventType)
	}

	switch event.Status {
	case StatusFailure:
		al.logger.Warn(message, fields...)
	case StatusDenied:
		al.logger.Warn(message, fields...)
	default:
		al.logger.Info(message, fields...)
	}
}

// Helper methods cho các event types phổ biến

// LogLoginAttempt logs a login attempt
func (al *AuditLogger) LogLoginAttempt(ctx context.Context, username, ipAddress, userAgent string, success bool, errorDetail string) {
	event := &AuditEvent{
		EventType: EventLoginAttempt,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if success {
		event.Status = StatusSuccess
		event.Message = "User login successful"
	} else {
		event.Status = StatusFailure
		event.Message = "User login failed"
		event.ErrorDetail = errorDetail
	}

	al.LogEvent(ctx, event)
}

// LogTokenIssued logs when an OAuth2 token is issued
func (al *AuditLogger) LogTokenIssued(ctx context.Context, userID, clientID *uuid.UUID, grantType, tokenType string, scopes []string, ipAddress string) {
	event := &AuditEvent{
		EventType: EventOAuth2TokenIssued,
		Status:    StatusSuccess,
		Message:   "OAuth2 token issued",
		UserID:    userID,
		ClientID:  clientID,
		GrantType: grantType,
		TokenType: tokenType,
		Scopes:    scopes,
		IPAddress: ipAddress,
	}

	al.LogEvent(ctx, event)
}

// LogTokenRevoked logs when an OAuth2 token is revoked
func (al *AuditLogger) LogTokenRevoked(ctx context.Context, userID, clientID *uuid.UUID, tokenType string, ipAddress string) {
	event := &AuditEvent{
		EventType: EventOAuth2TokenRevoked,
		Status:    StatusSuccess,
		Message:   "OAuth2 token revoked",
		UserID:    userID,
		ClientID:  clientID,
		TokenType: tokenType,
		IPAddress: ipAddress,
	}

	al.LogEvent(ctx, event)
}

// LogConsentGranted logs when user grants consent to a client
func (al *AuditLogger) LogConsentGranted(ctx context.Context, userID, clientID uuid.UUID, scopes []string, ipAddress, userAgent string) {
	event := &AuditEvent{
		EventType: EventOAuth2ConsentGranted,
		Status:    StatusSuccess,
		Message:   "User granted consent to client",
		UserID:    &userID,
		ClientID:  &clientID,
		Scopes:    scopes,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	al.LogEvent(ctx, event)
}

// LogConsentDenied logs when user denies consent to a client
func (al *AuditLogger) LogConsentDenied(ctx context.Context, userID, clientID uuid.UUID, ipAddress, userAgent string) {
	event := &AuditEvent{
		EventType: EventOAuth2ConsentDenied,
		Status:    StatusDenied,
		Message:   "User denied consent to client",
		UserID:    &userID,
		ClientID:  &clientID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	al.LogEvent(ctx, event)
}

// LogClientCreated logs OAuth2 client creation
func (al *AuditLogger) LogClientCreated(ctx context.Context, adminUserID *uuid.UUID, clientID uuid.UUID, clientName string, ipAddress string) {
	event := &AuditEvent{
		EventType:  EventOAuth2ClientCreated,
		Status:     StatusSuccess,
		Message:    "OAuth2 client created",
		UserID:     adminUserID,
		TargetType: "client",
		TargetID:   clientID.String(),
		IPAddress:  ipAddress,
		Metadata: map[string]any{
			"client_name": clientName,
		},
	}

	al.LogEvent(ctx, event)
}

// LogUnauthorizedAccess logs unauthorized access attempts
func (al *AuditLogger) LogUnauthorizedAccess(ctx context.Context, path, method, ipAddress, userAgent string, reason string) {
	event := &AuditEvent{
		EventType:   EventUnauthorizedAccess,
		Status:      StatusFailure,
		Message:     "Unauthorized access attempt",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		ErrorDetail: reason,
		Metadata: map[string]any{
			"path":   path,
			"method": method,
		},
	}

	al.LogEvent(ctx, event)
}

// LogRateLimitExceeded logs rate limit violations
func (al *AuditLogger) LogRateLimitExceeded(ctx context.Context, endpoint, ipAddress string, limit int) {
	event := &AuditEvent{
		EventType: EventRateLimitExceeded,
		Status:    StatusFailure,
		Message:   "Rate limit exceeded",
		IPAddress: ipAddress,
		Metadata: map[string]any{
			"endpoint": endpoint,
			"limit":    limit,
		},
	}

	al.LogEvent(ctx, event)
}
