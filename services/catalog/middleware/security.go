package middleware

import (
	"fmt"
	"mime"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	r "wibusystem/pkg/common/response"
)

// RateLimiter implements a lightweight in-memory rate limiter keyed by client IP.
type RateLimiter struct {
	requests map[string]*clientInfo
	mutex    sync.Mutex
	limit    int
	window   time.Duration
}

type clientInfo struct {
	timestamps []time.Time
}

// NewRateLimiter creates a new rate limiter instance.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()

	return rl
}

// RateLimit provides a Gin middleware enforcing the configured limits.
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := rl.clientKey(c)

		if !rl.allow(clientKey) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", rl.limit))
			c.Header("X-Rate-Limit-Window", rl.window.String())
			c.Header("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))

			c.JSON(http.StatusTooManyRequests, r.StandardResponse{
				Success: false,
				Message: "Rate limit exceeded",
				Data:    nil,
				Error: &r.ErrorDetail{
					Code:        "rate_limit_exceeded",
					Description: "Too many requests. Please slow down and try again shortly.",
				},
				Meta: map[string]any{"limit": rl.limit, "window": rl.window.String()},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) allow(clientKey string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	info, exists := rl.requests[clientKey]
	if !exists {
		info = &clientInfo{}
		rl.requests[clientKey] = info
	}

	filtered := info.timestamps[:0]
	for _, ts := range info.timestamps {
		if ts.After(windowStart) {
			filtered = append(filtered, ts)
		}
	}
	info.timestamps = filtered

	if len(info.timestamps) >= rl.limit {
		return false
	}

	info.timestamps = append(info.timestamps, now)
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		threshold := time.Now().Add(-rl.window * 2)
		for key, info := range rl.requests {
			if len(info.timestamps) == 0 || info.timestamps[len(info.timestamps)-1].Before(threshold) {
				delete(rl.requests, key)
			}
		}
		rl.mutex.Unlock()
	}
}

func (rl *RateLimiter) clientKey(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		ip = "unknown"
	}
	return ip
}

// SecurityHeaders adds basic security headers to responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// RequestLogger formats access logs similar to Identify service for consistency.
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s \"%s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
		)
	})
}

// RequestID injects an idempotent request identifier.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// APIVersionMiddleware annotates requests with an API version header.
func APIVersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("API-Version")
		if version == "" {
			version = "v1"
		}
		c.Set("api_version", version)
		c.Header("API-Version", version)
		c.Next()
	}
}

// ValidateContentType ensures POST/PUT/PATCH requests use supported media types.
func ValidateContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" {
				c.JSON(http.StatusBadRequest, r.StandardResponse{
					Success: false,
					Message: "Missing content type",
					Error: &r.ErrorDetail{
						Code:        "missing_content_type",
						Description: "Content-Type header is required",
					},
					Meta: map[string]any{},
				})
				c.Abort()
				return
			}

			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				c.JSON(http.StatusUnsupportedMediaType, r.StandardResponse{
					Success: false,
					Message: "Unsupported media type",
					Error: &r.ErrorDetail{
						Code:        "unsupported_media_type",
						Description: "Invalid Content-Type",
					},
					Meta: map[string]any{},
				})
				c.Abort()
				return
			}

			if mediaType != "application/json" && mediaType != "application/x-www-form-urlencoded" && mediaType != "multipart/form-data" {
				c.JSON(http.StatusUnsupportedMediaType, r.StandardResponse{
					Success: false,
					Message: "Unsupported media type",
					Error: &r.ErrorDetail{
						Code:        "unsupported_media_type",
						Description: "Content-Type must be application/json, application/x-www-form-urlencoded, or multipart/form-data",
					},
					Meta: map[string]any{},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ErrorHandler recovers from panics and returns a standardized response.
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, recovered any) {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: "Internal server error",
			Error: &r.ErrorDetail{
				Code:        "internal_server_error",
				Description: "An unexpected error occurred",
			},
			Meta: map[string]any{},
		})
		c.Abort()
	})
}
