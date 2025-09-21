package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // low, medium, high, critical
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityEventType constants
const (
	EventTypeLogin              = "login"
	EventTypeLoginFailed        = "login_failed"
	EventTypeLogout             = "logout"
	EventTypePasswordChange     = "password_change"
	EventTypeAccountLocked      = "account_locked"
	EventTypeUserCreated        = "user_created"
	EventTypeUserDeleted        = "user_deleted"
	EventTypeTenantCreated      = "tenant_created"
	EventTypeTenantDeleted      = "tenant_deleted"
	EventTypePermissionChange   = "permission_change"
	EventTypeSuspiciousActivity = "suspicious_activity"
	EventTypeTokenIssued        = "token_issued"
	EventTypeTokenRevoked       = "token_revoked"
	EventTypeAPIAccess          = "api_access"
	EventTypeRateLimitHit       = "rate_limit_hit"
)

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// SecurityService handles security monitoring and audit logging
type SecurityService struct {
	events        []SecurityEvent
	mutex         sync.RWMutex
	loginAttempts map[string][]time.Time
	attemptsMutex sync.RWMutex
	maxAttempts   int
	lockoutWindow time.Duration
	cleanupTicker *time.Ticker
}

// NewSecurityService creates a new security service
func NewSecurityService(maxAttempts int, lockoutWindow time.Duration) *SecurityService {
	ss := &SecurityService{
		events:        make([]SecurityEvent, 0),
		loginAttempts: make(map[string][]time.Time),
		maxAttempts:   maxAttempts,
		lockoutWindow: lockoutWindow,
		cleanupTicker: time.NewTicker(time.Hour), // Cleanup every hour
	}

	// Start cleanup goroutine
	go ss.cleanupOldAttempts()

	return ss
}

// LogSecurityEvent logs a security event
func (ss *SecurityService) LogSecurityEvent(event SecurityEvent) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// Generate ID if not provided
	if event.ID == "" {
		event.ID = ss.generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Add to events slice
	ss.events = append(ss.events, event)

	// Log to system logger based on severity
	ss.logToSystem(event)

	// Handle specific event types
	ss.handleSpecificEvent(event)
}

// LogLoginAttempt logs a login attempt and checks for suspicious activity
func (ss *SecurityService) LogLoginAttempt(identifier, ipAddress, userAgent string, success bool, userID *uuid.UUID) bool {
	ss.attemptsMutex.Lock()
	defer ss.attemptsMutex.Unlock()

	key := fmt.Sprintf("%s:%s", identifier, ipAddress)
	now := time.Now()

	// Get existing attempts for this identifier/IP combination
	attempts, exists := ss.loginAttempts[key]
	if !exists {
		attempts = make([]time.Time, 0)
	}

	// Remove old attempts (outside lockout window)
	validAttempts := make([]time.Time, 0)
	cutoff := now.Add(-ss.lockoutWindow)
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			validAttempts = append(validAttempts, attempt)
		}
	}

	// Add current attempt if failed
	if !success {
		validAttempts = append(validAttempts, now)
		ss.loginAttempts[key] = validAttempts

		// Log failed login event
		severity := SeverityMedium
		if len(validAttempts) >= ss.maxAttempts {
			severity = SeverityHigh
		}

		ss.LogSecurityEvent(SecurityEvent{
			Type:        EventTypeLoginFailed,
			UserID:      userID,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Description: fmt.Sprintf("Failed login attempt for %s", identifier),
			Severity:    severity,
			Metadata: map[string]interface{}{
				"identifier":     identifier,
				"attempt_count":  len(validAttempts),
				"max_attempts":   ss.maxAttempts,
				"lockout_window": ss.lockoutWindow.String(),
			},
		})

		// Check if account should be locked
		if len(validAttempts) >= ss.maxAttempts {
			ss.LogSecurityEvent(SecurityEvent{
				Type:        EventTypeAccountLocked,
				UserID:      userID,
				IPAddress:   ipAddress,
				UserAgent:   userAgent,
				Description: fmt.Sprintf("Account locked due to too many failed attempts: %s", identifier),
				Severity:    SeverityCritical,
				Metadata: map[string]interface{}{
					"identifier":    identifier,
					"attempt_count": len(validAttempts),
					"lockout_until": now.Add(ss.lockoutWindow),
				},
			})
			return false // Account locked
		}
	} else {
		// Successful login - clear failed attempts
		delete(ss.loginAttempts, key)

		ss.LogSecurityEvent(SecurityEvent{
			Type:        EventTypeLogin,
			UserID:      userID,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Description: fmt.Sprintf("Successful login for %s", identifier),
			Severity:    SeverityLow,
			Metadata: map[string]interface{}{
				"identifier": identifier,
			},
		})
	}

	return true // Not locked
}

// IsAccountLocked checks if an account is currently locked
func (ss *SecurityService) IsAccountLocked(identifier, ipAddress string) bool {
	ss.attemptsMutex.RLock()
	defer ss.attemptsMutex.RUnlock()

	key := fmt.Sprintf("%s:%s", identifier, ipAddress)
	attempts, exists := ss.loginAttempts[key]
	if !exists {
		return false
	}

	// Check if we have enough recent failed attempts
	now := time.Now()
	cutoff := now.Add(-ss.lockoutWindow)
	recentAttempts := 0

	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			recentAttempts++
		}
	}

	return recentAttempts >= ss.maxAttempts
}

// DetectSuspiciousActivity analyzes patterns for suspicious behavior
func (ss *SecurityService) DetectSuspiciousActivity(ipAddress, userAgent string, userID *uuid.UUID) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	// Count events from this IP in the last hour
	ipEvents := 0
	for _, event := range ss.events {
		if event.IPAddress == ipAddress && event.Timestamp.After(oneHourAgo) {
			ipEvents++
		}
	}

	// Alert if too many events from single IP
	if ipEvents > 100 {
		ss.LogSecurityEvent(SecurityEvent{
			Type:        EventTypeSuspiciousActivity,
			UserID:      userID,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Description: fmt.Sprintf("High activity from IP: %d events in 1 hour", ipEvents),
			Severity:    SeverityHigh,
			Metadata: map[string]interface{}{
				"event_count": ipEvents,
				"time_window": "1 hour",
				"threshold":   100,
			},
		})
	}

	// Check for suspicious user agent patterns
	if ss.isSuspiciousUserAgent(userAgent) {
		ss.LogSecurityEvent(SecurityEvent{
			Type:        EventTypeSuspiciousActivity,
			UserID:      userID,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Description: "Suspicious user agent detected",
			Severity:    SeverityMedium,
			Metadata: map[string]interface{}{
				"user_agent": userAgent,
				"reason":     "suspicious_pattern",
			},
		})
	}
}

// GetSecurityEvents returns security events with optional filtering
func (ss *SecurityService) GetSecurityEvents(ctx context.Context, filters map[string]interface{}, limit int) []SecurityEvent {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	var filtered []SecurityEvent

	for _, event := range ss.events {
		if ss.matchesFilters(event, filters) {
			filtered = append(filtered, event)
			if len(filtered) >= limit && limit > 0 {
				break
			}
		}
	}

	return filtered
}

// GetSecurityStats returns security statistics
func (ss *SecurityService) GetSecurityStats(ctx context.Context) map[string]interface{} {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)
	oneDayAgo := now.Add(-24 * time.Hour)

	stats := map[string]interface{}{
		"total_events":     len(ss.events),
		"events_last_hour": 0,
		"events_last_24h":  0,
		"event_types":      make(map[string]int),
		"severity_counts":  make(map[string]int),
		"active_lockouts":  len(ss.loginAttempts),
	}

	eventTypes := stats["event_types"].(map[string]int)
	severityCounts := stats["severity_counts"].(map[string]int)

	for _, event := range ss.events {
		eventTypes[event.Type]++
		severityCounts[event.Severity]++

		if event.Timestamp.After(oneHourAgo) {
			stats["events_last_hour"] = stats["events_last_hour"].(int) + 1
		}

		if event.Timestamp.After(oneDayAgo) {
			stats["events_last_24h"] = stats["events_last_24h"].(int) + 1
		}
	}

	return stats
}

// Helper methods

func (ss *SecurityService) generateEventID() string {
	return fmt.Sprintf("sec_%d_%d", time.Now().UnixNano(), len(ss.events))
}

func (ss *SecurityService) logToSystem(event SecurityEvent) {
	switch event.Severity {
	case SeverityCritical, SeverityHigh:
		log.Printf("[SECURITY-%s] %s: %s (IP: %s)",
			event.Severity, event.Type, event.Description, event.IPAddress)
	case SeverityMedium:
		log.Printf("[SECURITY-%s] %s: %s",
			event.Severity, event.Type, event.Description)
	default:
		// Low severity events are logged at debug level
		log.Printf("[SECURITY-DEBUG] %s: %s", event.Type, event.Description)
	}
}

func (ss *SecurityService) handleSpecificEvent(event SecurityEvent) {
	switch event.Type {
	case EventTypeAccountLocked:
		// Could send email notification, slack alert, etc.
		log.Printf("ALERT: Account locked - %s", event.Description)

	case EventTypeSuspiciousActivity:
		// Could trigger additional monitoring, IP blocking, etc.
		log.Printf("ALERT: Suspicious activity - %s", event.Description)

	case EventTypeLoginFailed:
		// Could trigger progressive delays, CAPTCHA, etc.
		if metadata, ok := event.Metadata["attempt_count"].(int); ok && metadata >= ss.maxAttempts-1 {
			log.Printf("WARNING: Account near lockout - %s", event.Description)
		}
	}
}

func (ss *SecurityService) cleanupOldAttempts() {
	for range ss.cleanupTicker.C {
		ss.attemptsMutex.Lock()
		now := time.Now()
		cutoff := now.Add(-ss.lockoutWindow * 2) // Keep data for 2x lockout window

		for key, attempts := range ss.loginAttempts {
			validAttempts := make([]time.Time, 0)
			for _, attempt := range attempts {
				if attempt.After(cutoff) {
					validAttempts = append(validAttempts, attempt)
				}
			}

			if len(validAttempts) == 0 {
				delete(ss.loginAttempts, key)
			} else {
				ss.loginAttempts[key] = validAttempts
			}
		}
		ss.attemptsMutex.Unlock()

		// Also cleanup old events (keep only last 10000)
		ss.mutex.Lock()
		if len(ss.events) > 10000 {
			ss.events = ss.events[len(ss.events)-10000:]
		}
		ss.mutex.Unlock()
	}
}

func (ss *SecurityService) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap", "nmap", "nikto", "dirb", "gobuster",
		"python-requests", "curl", "wget", "bot", "crawler",
		"scanner", "exploit", "hack", "injection",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}

	return false
}

func (ss *SecurityService) matchesFilters(event SecurityEvent, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "type":
			if event.Type != value.(string) {
				return false
			}
		case "severity":
			if event.Severity != value.(string) {
				return false
			}
		case "user_id":
			if event.UserID == nil || *event.UserID != value.(uuid.UUID) {
				return false
			}
		case "ip_address":
			if event.IPAddress != value.(string) {
				return false
			}
		case "since":
			if event.Timestamp.Before(value.(time.Time)) {
				return false
			}
		}
	}
	return true
}

// Close cleanup resources
func (ss *SecurityService) Close() {
	if ss.cleanupTicker != nil {
		ss.cleanupTicker.Stop()
	}
}

// IPWhitelist manages allowed IP addresses
type IPWhitelist struct {
	allowedIPs []net.IPNet
	mutex      sync.RWMutex
}

// NewIPWhitelist creates a new IP whitelist
func NewIPWhitelist(cidrs []string) (*IPWhitelist, error) {
	whitelist := &IPWhitelist{
		allowedIPs: make([]net.IPNet, 0, len(cidrs)),
	}

	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %s: %w", cidr, err)
		}
		whitelist.allowedIPs = append(whitelist.allowedIPs, *network)
	}

	return whitelist, nil
}

// IsAllowed checks if an IP address is in the whitelist
func (ipw *IPWhitelist) IsAllowed(ip string) bool {
	ipw.mutex.RLock()
	defer ipw.mutex.RUnlock()

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	for _, network := range ipw.allowedIPs {
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// AddIP adds an IP or CIDR to the whitelist
func (ipw *IPWhitelist) AddIP(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %s: %w", cidr, err)
	}

	ipw.mutex.Lock()
	defer ipw.mutex.Unlock()

	ipw.allowedIPs = append(ipw.allowedIPs, *network)
	return nil
}
