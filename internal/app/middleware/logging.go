package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/platform/logger"
)

// LoggingMiddleware creates a middleware for structured HTTP request logging
func LoggingMiddleware(appLogger *zap.Logger, perfLogger *logger.PerformanceLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.Must(uuid.NewV7()).String()
		c.Set("request_id", requestID)

		// Get client information
		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Store in context
		ctx := c.Request.Context()
		ctx = logger.WithRequestID(ctx, requestID)
		ctx = logger.WithIPAddress(ctx, ipAddress)
		ctx = logger.WithUserAgent(ctx, userAgent)
		c.Request = c.Request.WithContext(ctx)

		// Start timer
		startTime := time.Now()

		// Create logger with context
		ctxLogger := logger.FromContext(ctx, appLogger)

		// Log request start
		ctxLogger.Info("HTTP request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("user_agent", userAgent),
			zap.String("category", "http"),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Log request completion
		ctxLogger.Info("HTTP request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", statusCode),
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.Int("response_size", c.Writer.Size()),
			zap.String("category", "http"),
		)

		// Log performance metric
		var err error
		if len(c.Errors) > 0 {
			err = c.Errors.Last().Err
		}
		perfLogger.LogHTTPRequest(ctx, c.Request.Method, c.Request.URL.Path, statusCode, duration, err)

		// Log slow requests as warning
		if duration > 2*time.Second {
			ctxLogger.Warn("Slow HTTP request detected",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int64("duration_ms", duration.Milliseconds()),
				zap.String("category", "performance"),
			)
		}

		// Log errors
		if statusCode >= 500 {
			// Convert errors to string slice
			var errorMsgs []string
			for _, e := range c.Errors {
				errorMsgs = append(errorMsgs, e.Error())
			}

			ctxLogger.Error("HTTP request failed with server error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status_code", statusCode),
				zap.Strings("errors", errorMsgs),
				zap.String("category", "error"),
			)
		} else if statusCode >= 400 {
			// Convert errors to string slice
			var errorMsgs []string
			for _, e := range c.Errors {
				errorMsgs = append(errorMsgs, e.Error())
			}

			ctxLogger.Warn("HTTP request failed with client error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status_code", statusCode),
				zap.Strings("errors", errorMsgs),
				zap.String("category", "error"),
			)
		}
	}
}

// RecoveryMiddleware creates a middleware for panic recovery with structured logging
func RecoveryMiddleware(appLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				requestID := logger.GetRequestID(ctx)

				ctxLogger := logger.FromContext(ctx, appLogger)
				ctxLogger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("category", "panic"),
				)

				// Return 500 Internal Server Error
				c.JSON(500, gin.H{
					"error":      "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
