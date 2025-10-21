// Package domain contains the core business entities and logic for the Identity module.
package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Session represents a user session.
type Session struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	TokenHash      string    `json:"-"` // Never expose token hash
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	Revoked        bool      `json:"revoked"`
}

// Session errors
var (
	ErrInvalidSessionID        = errors.New("invalid session ID")
	ErrInvalidUserIDForSession = errors.New("invalid user ID for session")
	ErrSessionExpired          = errors.New("session has expired")
	ErrSessionRevoked          = errors.New("session has been revoked")
	ErrInvalidSessionToken     = errors.New("invalid session token")
	ErrTokenGenerationFailed   = errors.New("failed to generate session token")
)

const (
	// SessionTokenLength is the length of the session token in bytes
	SessionTokenLength = 32

	// DefaultSessionDuration is the default session duration
	DefaultSessionDuration = 24 * time.Hour

	// ExtendedSessionDuration is for "remember me" sessions
	ExtendedSessionDuration = 30 * 24 * time.Hour
)

// NewSession creates a new session for a user.
// It generates a secure random token and returns both the session and the plain token.
// The plain token should be sent to the client and never stored.
func NewSession(userID uuid.UUID, duration time.Duration, ipAddress, userAgent *string) (*Session, string, error) {
	if userID == uuid.Nil {
		return nil, "", ErrInvalidUserIDForSession
	}

	if duration <= 0 {
		duration = DefaultSessionDuration
	}

	// Generate secure random token
	token, tokenHash, err := generateSessionToken()
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	session := &Session{
		ID:             uuid.New(),
		UserID:         userID,
		TokenHash:      tokenHash,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ExpiresAt:      now.Add(duration),
		CreatedAt:      now,
		LastAccessedAt: now,
		Revoked:        false,
	}

	return session, token, nil
}

// generateSessionToken generates a secure random session token.
// Returns the plain token (to send to client) and the bcrypt hash (to store in DB).
func generateSessionToken() (plainToken string, tokenHash string, err error) {
	// Generate random bytes
	tokenBytes := make([]byte, SessionTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTokenGenerationFailed, err)
	}

	// Encode to base64 for the plain token
	plainToken = base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token with bcrypt for storage
	hash, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTokenGenerationFailed, err)
	}

	tokenHash = string(hash)
	return plainToken, tokenHash, nil
}

// VerifyToken verifies a plain token against the stored hash.
func (s *Session) VerifyToken(plainToken string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(s.TokenHash), []byte(plainToken))
	return err == nil
}

// IsValid checks if the session is valid (not expired and not revoked).
func (s *Session) IsValid() bool {
	now := time.Now().UTC()
	return !s.Revoked && s.ExpiresAt.After(now)
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// Revoke revokes the session.
func (s *Session) Revoke() {
	s.Revoked = true
}

// Extend extends the session expiration time.
func (s *Session) Extend(duration time.Duration) error {
	if s.Revoked {
		return ErrSessionRevoked
	}

	if duration <= 0 {
		duration = DefaultSessionDuration
	}

	s.ExpiresAt = time.Now().UTC().Add(duration)
	return nil
}

// UpdateLastAccessed updates the last accessed timestamp.
// This should be called on each authenticated request.
func (s *Session) UpdateLastAccessed() {
	s.LastAccessedAt = time.Now().UTC()
}

// UpdateIPAddress updates the IP address associated with the session.
func (s *Session) UpdateIPAddress(ipAddress string) {
	ipAddress = normalizeIPAddress(ipAddress)
	if ipAddress == "" {
		s.IPAddress = nil
	} else {
		s.IPAddress = &ipAddress
	}
}

// UpdateUserAgent updates the user agent associated with the session.
func (s *Session) UpdateUserAgent(userAgent string) {
	if userAgent == "" {
		s.UserAgent = nil
	} else {
		s.UserAgent = &userAgent
	}
}

// Validate validates the session entity.
func (s *Session) Validate() error {
	if s.ID == uuid.Nil {
		return ErrInvalidSessionID
	}

	if s.UserID == uuid.Nil {
		return ErrInvalidUserIDForSession
	}

	if s.TokenHash == "" {
		return ErrInvalidSessionToken
	}

	if s.ExpiresAt.Before(s.CreatedAt) {
		return errors.New("expiration time cannot be before creation time")
	}

	return nil
}

// RemainingTime returns the remaining time until the session expires.
func (s *Session) RemainingTime() time.Duration {
	remaining := time.Until(s.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ShouldRefresh checks if the session should be refreshed.
// Returns true if less than 25% of the session duration remains.
func (s *Session) ShouldRefresh() bool {
	if s.Revoked {
		return false
	}

	totalDuration := s.ExpiresAt.Sub(s.CreatedAt)
	remaining := s.RemainingTime()

	// Refresh if less than 25% of time remaining
	return remaining < (totalDuration / 4)
}

// Clone creates a deep copy of the Session.
func (s *Session) Clone() *Session {
	clone := *s

	if s.IPAddress != nil {
		ip := *s.IPAddress
		clone.IPAddress = &ip
	}

	if s.UserAgent != nil {
		ua := *s.UserAgent
		clone.UserAgent = &ua
	}

	return &clone
}

// normalizeIPAddress normalizes an IP address string.
func normalizeIPAddress(ip string) string {
	if ip == "" {
		return ""
	}

	// Try to parse as IP
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip // Return as-is if not a valid IP
	}

	return parsed.String()
}

// SessionMetadata contains additional information about a session for display purposes.
type SessionMetadata struct {
	ID             uuid.UUID `json:"id"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	Browser        string    `json:"browser,omitempty"`
	OS             string    `json:"os,omitempty"`
	Device         string    `json:"device,omitempty"`
}

// ToMetadata converts a Session to SessionMetadata for API responses.
func (s *Session) ToMetadata() SessionMetadata {
	metadata := SessionMetadata{
		ID:             s.ID,
		IPAddress:      s.IPAddress,
		UserAgent:      s.UserAgent,
		CreatedAt:      s.CreatedAt,
		LastAccessedAt: s.LastAccessedAt,
		ExpiresAt:      s.ExpiresAt,
		IsActive:       s.IsValid(),
	}

	// Parse user agent if available
	if s.UserAgent != nil {
		metadata.Browser, metadata.OS, metadata.Device = parseUserAgent(*s.UserAgent)
	}

	return metadata
}

// parseUserAgent extracts browser, OS, and device information from a user agent string.
// This is a simplified implementation - in production, use a proper UA parser library.
func parseUserAgent(ua string) (browser, os, device string) {
	// This is a very basic implementation
	// In production, use a library like github.com/mssola/user_agent

	// Detect OS
	if contains(ua, "Windows") {
		os = "Windows"
	} else if contains(ua, "Mac OS X") || contains(ua, "macOS") {
		os = "macOS"
	} else if contains(ua, "Linux") {
		os = "Linux"
	} else if contains(ua, "Android") {
		os = "Android"
		device = "Mobile"
	} else if contains(ua, "iOS") || contains(ua, "iPhone") || contains(ua, "iPad") {
		os = "iOS"
		if contains(ua, "iPad") {
			device = "Tablet"
		} else {
			device = "Mobile"
		}
	}

	// Detect browser
	if contains(ua, "Chrome") && !contains(ua, "Edg") {
		browser = "Chrome"
	} else if contains(ua, "Firefox") {
		browser = "Firefox"
	} else if contains(ua, "Safari") && !contains(ua, "Chrome") {
		browser = "Safari"
	} else if contains(ua, "Edg") {
		browser = "Edge"
	} else if contains(ua, "Opera") || contains(ua, "OPR") {
		browser = "Opera"
	}

	// Default device if not set
	if device == "" {
		device = "Desktop"
	}

	return browser, os, device
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
