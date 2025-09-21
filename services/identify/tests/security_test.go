package tests

import (
	"testing"
	"time"

	"wibusystem/services/identify/services"
)

func TestSecurityService_LoginAttempts(t *testing.T) {
	// Create security service with 3 max attempts and 5-minute lockout
	ss := services.NewSecurityService(3, 5*time.Minute)
	defer ss.Close()

	identifier := "test@example.com"
	ipAddress := "192.168.1.1"
	userAgent := "Test Browser"

	// Test successful login
	allowed := ss.LogLoginAttempt(identifier, ipAddress, userAgent, true, nil)
	if !allowed {
		t.Error("Expected successful login to be allowed")
	}

	// Test failed login attempts
	for i := 0; i < 2; i++ {
		allowed = ss.LogLoginAttempt(identifier, ipAddress, userAgent, false, nil)
		if !allowed {
			t.Errorf("Expected failed login attempt %d to be allowed", i+1)
		}
	}

	// Check if account is not yet locked
	if ss.IsAccountLocked(identifier, ipAddress) {
		t.Error("Account should not be locked yet")
	}

	// Third failed attempt should trigger lockout
	allowed = ss.LogLoginAttempt(identifier, ipAddress, userAgent, false, nil)
	if allowed {
		t.Error("Expected third failed login attempt to trigger lockout")
	}

	// Check if account is now locked
	if !ss.IsAccountLocked(identifier, ipAddress) {
		t.Error("Account should be locked after max attempts")
	}

	// Test that successful login clears attempts
	ss2 := services.NewSecurityService(3, 5*time.Minute)
	defer ss2.Close()

	// One failed attempt
	ss2.LogLoginAttempt(identifier, ipAddress, userAgent, false, nil)

	// Successful login should clear attempts
	ss2.LogLoginAttempt(identifier, ipAddress, userAgent, true, nil)

	// Account should not be locked
	if ss2.IsAccountLocked(identifier, ipAddress) {
		t.Error("Account should not be locked after successful login")
	}
}

func TestSecurityService_EventLogging(t *testing.T) {
	ss := services.NewSecurityService(5, 15*time.Minute)
	defer ss.Close()

	// Test logging different types of events
	testEvents := []services.SecurityEvent{
		{
			Type:        services.EventTypeLogin,
			IPAddress:   "192.168.1.1",
			UserAgent:   "Test Browser",
			Description: "User logged in",
			Severity:    services.SeverityLow,
		},
		{
			Type:        services.EventTypeLoginFailed,
			IPAddress:   "192.168.1.2",
			UserAgent:   "Test Browser",
			Description: "Failed login attempt",
			Severity:    services.SeverityMedium,
		},
		{
			Type:        services.EventTypeSuspiciousActivity,
			IPAddress:   "192.168.1.3",
			UserAgent:   "Suspicious Bot",
			Description: "Suspicious activity detected",
			Severity:    services.SeverityHigh,
		},
	}

	// Log all test events
	for _, event := range testEvents {
		ss.LogSecurityEvent(event)
	}

	// Retrieve events with no filters
	events := ss.GetSecurityEvents(nil, nil, 0)
	if len(events) != len(testEvents) {
		t.Errorf("Expected %d events, got %d", len(testEvents), len(events))
	}

	// Test filtering by severity
	highSeverityFilter := map[string]interface{}{
		"severity": services.SeverityHigh,
	}
	highEvents := ss.GetSecurityEvents(nil, highSeverityFilter, 0)
	if len(highEvents) != 1 {
		t.Errorf("Expected 1 high severity event, got %d", len(highEvents))
	}

	// Test filtering by type
	loginEventFilter := map[string]interface{}{
		"type": services.EventTypeLogin,
	}
	loginEvents := ss.GetSecurityEvents(nil, loginEventFilter, 0)
	if len(loginEvents) != 1 {
		t.Errorf("Expected 1 login event, got %d", len(loginEvents))
	}

	// Test limiting results
	limitedEvents := ss.GetSecurityEvents(nil, nil, 2)
	if len(limitedEvents) != 2 {
		t.Errorf("Expected 2 events with limit, got %d", len(limitedEvents))
	}
}

func TestSecurityService_SuspiciousActivity(t *testing.T) {
	ss := services.NewSecurityService(5, 15*time.Minute)
	defer ss.Close()

	// Generate many events from same IP to trigger suspicious activity detection
	ipAddress := "192.168.1.100"
	userAgent := "Normal Browser"

	// Log 101 events to trigger threshold (>100)
	for i := 0; i < 101; i++ {
		ss.LogSecurityEvent(services.SecurityEvent{
			Type:        services.EventTypeAPIAccess,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Description: "API access",
			Severity:    services.SeverityLow,
		})
	}

	// Trigger suspicious activity detection
	ss.DetectSuspiciousActivity(ipAddress, userAgent, nil)

	// Check for suspicious activity event
	suspiciousFilter := map[string]interface{}{
		"type": services.EventTypeSuspiciousActivity,
	}
	suspiciousEvents := ss.GetSecurityEvents(nil, suspiciousFilter, 0)
	if len(suspiciousEvents) == 0 {
		t.Error("Expected suspicious activity event to be logged")
	}
}

func TestSecurityService_Stats(t *testing.T) {
	ss := services.NewSecurityService(5, 15*time.Minute)
	defer ss.Close()

	// Log some test events
	events := []services.SecurityEvent{
		{Type: services.EventTypeLogin, Severity: services.SeverityLow, IPAddress: "1.1.1.1"},
		{Type: services.EventTypeLogin, Severity: services.SeverityLow, IPAddress: "1.1.1.1"},
		{Type: services.EventTypeLoginFailed, Severity: services.SeverityMedium, IPAddress: "1.1.1.1"},
		{Type: services.EventTypeSuspiciousActivity, Severity: services.SeverityHigh, IPAddress: "1.1.1.1"},
	}

	for _, event := range events {
		ss.LogSecurityEvent(event)
	}

	// Get stats
	stats := ss.GetSecurityStats(nil)

	// Check total events
	if stats["total_events"].(int) != len(events) {
		t.Errorf("Expected %d total events, got %d", len(events), stats["total_events"])
	}

	// Check event types
	eventTypes := stats["event_types"].(map[string]int)
	if eventTypes[services.EventTypeLogin] != 2 {
		t.Errorf("Expected 2 login events, got %d", eventTypes[services.EventTypeLogin])
	}
	if eventTypes[services.EventTypeLoginFailed] != 1 {
		t.Errorf("Expected 1 failed login event, got %d", eventTypes[services.EventTypeLoginFailed])
	}

	// Check severity counts
	severityCounts := stats["severity_counts"].(map[string]int)
	if severityCounts[services.SeverityLow] != 2 {
		t.Errorf("Expected 2 low severity events, got %d", severityCounts[services.SeverityLow])
	}
	if severityCounts[services.SeverityMedium] != 1 {
		t.Errorf("Expected 1 medium severity event, got %d", severityCounts[services.SeverityMedium])
	}
	if severityCounts[services.SeverityHigh] != 1 {
		t.Errorf("Expected 1 high severity event, got %d", severityCounts[services.SeverityHigh])
	}
}

func TestIPWhitelist(t *testing.T) {
	// Test creating whitelist
	cidrs := []string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"127.0.0.1/32",
	}

	whitelist, err := services.NewIPWhitelist(cidrs)
	if err != nil {
		t.Fatalf("Failed to create IP whitelist: %v", err)
	}

	// Test allowed IPs
	allowedIPs := []string{
		"192.168.1.1",
		"192.168.1.100",
		"10.0.0.1",
		"10.255.255.255",
		"127.0.0.1",
	}

	for _, ip := range allowedIPs {
		if !whitelist.IsAllowed(ip) {
			t.Errorf("IP %s should be allowed", ip)
		}
	}

	// Test disallowed IPs
	disallowedIPs := []string{
		"192.168.2.1",
		"172.16.0.1",
		"8.8.8.8",
		"127.0.0.2",
	}

	for _, ip := range disallowedIPs {
		if whitelist.IsAllowed(ip) {
			t.Errorf("IP %s should not be allowed", ip)
		}
	}

	// Test invalid IP
	if whitelist.IsAllowed("invalid-ip") {
		t.Error("Invalid IP should not be allowed")
	}

	// Test adding new IP
	err = whitelist.AddIP("203.0.113.0/24")
	if err != nil {
		t.Errorf("Failed to add IP to whitelist: %v", err)
	}

	// Test newly added IP
	if !whitelist.IsAllowed("203.0.113.1") {
		t.Error("Newly added IP should be allowed")
	}

	// Test adding invalid CIDR
	err = whitelist.AddIP("invalid-cidr")
	if err == nil {
		t.Error("Adding invalid CIDR should return error")
	}
}
