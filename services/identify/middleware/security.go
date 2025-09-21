package middleware

import (
    "fmt"
    "net/http"
    "sync"
    "time"
    "mime"

    r "wibusystem/pkg/common/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string]*ClientInfo
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// ClientInfo tracks request information for a client
type ClientInfo struct {
	requests  []time.Time
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int, windowDuration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*ClientInfo),
		limit:    requestsPerMinute,
		window:   windowDuration,
	}

	// Clean up old entries periodically
	go rl.cleanup()

	return rl
}

// RateLimit middleware that limits requests per client
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := getClientIdentifier(c)

		if !rl.allowRequest(clientID) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", rl.limit))
			c.Header("X-Rate-Limit-Window", rl.window.String())
			c.Header("Retry-After", "60")

            c.JSON(http.StatusTooManyRequests, r.StandardResponse{Success: false, Message: "Rate limit exceeded", Data: nil, Error: &r.ErrorDetail{Code: "rate_limit_exceeded", Description: "Too many requests. Please try again later."}, Meta: map[string]interface{}{"limit": rl.limit, "window": rl.window.String()}})
            c.Abort()
            return
        }

		c.Next()
	}
}

// allowRequest checks if a request should be allowed
func (rl *RateLimiter) allowRequest(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	clientInfo, exists := rl.requests[clientID]
	if !exists {
		clientInfo = &ClientInfo{
			requests:  make([]time.Time, 0),
			lastReset: now,
		}
		rl.requests[clientID] = clientInfo
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range clientInfo.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	clientInfo.requests = validRequests

	// Check if limit is exceeded
	if len(clientInfo.requests) >= rl.limit {
		return false
	}

	// Add current request
	clientInfo.requests = append(clientInfo.requests, now)
	return true
}

// cleanup removes old client info entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2) // Keep entries for 2x window duration

		for clientID, clientInfo := range rl.requests {
			if clientInfo.lastReset.Before(cutoff) && len(clientInfo.requests) == 0 {
				delete(rl.requests, clientID)
			}
		}
		rl.mutex.Unlock()
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(c *gin.Context) string {
	// Try to get client ID from OAuth2 context first
	if userInfo, exists := GetUserFromContext(c); exists {
		return fmt.Sprintf("user:%s", userInfo.Subject)
	}

	// Fall back to IP address
	clientIP := c.ClientIP()

	// Include User-Agent for better granularity (optional)
	userAgent := c.GetHeader("User-Agent")
	if userAgent != "" {
		return fmt.Sprintf("ip:%s:ua:%s", clientIP, userAgent[:min(50, len(userAgent))])
	}

	return fmt.Sprintf("ip:%s", clientIP)
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (adjust as needed)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		// Strict Transport Security (only in production with HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RequestLogger middleware logs requests with security information
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// APIVersionMiddleware adds API versioning support
func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("API-Version")
		if version == "" {
			version = "v1" // default version
		}
		c.Set("api_version", version)
		c.Header("API-Version", version)
		c.Next()
	}
}

// ValidateContentType middleware validates content type for POST/PUT requests
func ValidateContentType() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
            contentType := c.GetHeader("Content-Type")
            if contentType == "" {
                c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Missing content type", Data: nil, Error: &r.ErrorDetail{Code: "missing_content_type", Description: "Content-Type header is required"}, Meta: map[string]interface{}{}})
                c.Abort()
                return
            }

            // Allow JSON and form data (ignore parameters like charset)
            mediaType, _, err := mime.ParseMediaType(contentType)
            if err != nil {
                c.JSON(http.StatusUnsupportedMediaType, r.StandardResponse{Success: false, Message: "Unsupported media type", Data: nil, Error: &r.ErrorDetail{Code: "unsupported_media_type", Description: "Invalid Content-Type"}, Meta: map[string]interface{}{}})
                c.Abort()
                return
            }
            if mediaType != "application/json" &&
                mediaType != "application/x-www-form-urlencoded" &&
                mediaType != "multipart/form-data" {
                c.JSON(http.StatusUnsupportedMediaType, r.StandardResponse{Success: false, Message: "Unsupported media type", Data: nil, Error: &r.ErrorDetail{Code: "unsupported_media_type", Description: "Content-Type must be application/json or application/x-www-form-urlencoded"}, Meta: map[string]interface{}{}})
                c.Abort()
                return
            }
        }

		c.Next()
	}
}

// IPWhitelist middleware restricts access to specific IP addresses (for admin endpoints)
func IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	allowedIPMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedIPMap[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

        if len(allowedIPMap) > 0 && !allowedIPMap[clientIP] {
            c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "Access denied from this IP address"}, Meta: map[string]interface{}{"ip": clientIP}})
            c.Abort()
            return
        }

		c.Next()
	}
}

// ErrorHandler middleware handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        if err, ok := recovered.(string); ok {
            c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Internal server error", Data: nil, Error: &r.ErrorDetail{Code: "internal_server_error", Description: "An unexpected error occurred"}, Meta: map[string]interface{}{}})
            // Log the actual error (in production, use proper logging)
            fmt.Printf("Panic recovered: %s\n", err)
        }
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
