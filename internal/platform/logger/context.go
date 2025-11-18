package logger

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// Context keys for logger
type contextKey string

const (
	contextKeyRequestID contextKey = "request_id"
	contextKeyUserID    contextKey = "user_id"
	contextKeyClientID  contextKey = "client_id"
	contextKeyIPAddress contextKey = "ip_address"
	contextKeyUserAgent contextKey = "user_agent"
)

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(contextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) *uuid.UUID {
	if userID, ok := ctx.Value(contextKeyUserID).(uuid.UUID); ok {
		return &userID
	}
	return nil
}

// WithClientID adds client ID to context
func WithClientID(ctx context.Context, clientID uuid.UUID) context.Context {
	return context.WithValue(ctx, contextKeyClientID, clientID)
}

// GetClientID retrieves client ID from context
func GetClientID(ctx context.Context) *uuid.UUID {
	if clientID, ok := ctx.Value(contextKeyClientID).(uuid.UUID); ok {
		return &clientID
	}
	return nil
}

// WithIPAddress adds IP address to context
func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, contextKeyIPAddress, ipAddress)
}

// GetIPAddress retrieves IP address from context
func GetIPAddress(ctx context.Context) string {
	if ipAddress, ok := ctx.Value(contextKeyIPAddress).(string); ok {
		return ipAddress
	}
	return ""
}

// WithUserAgent adds user agent to context
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, contextKeyUserAgent, userAgent)
}

// GetUserAgent retrieves user agent from context
func GetUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(contextKeyUserAgent).(string); ok {
		return userAgent
	}
	return ""
}

// FromContext creates a logger with context fields
func FromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	fields := []zap.Field{}

	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	if userID := GetUserID(ctx); userID != nil {
		fields = append(fields, zap.String("user_id", userID.String()))
	}
	if clientID := GetClientID(ctx); clientID != nil {
		fields = append(fields, zap.String("client_id", clientID.String()))
	}
	if ipAddress := GetIPAddress(ctx); ipAddress != "" {
		fields = append(fields, zap.String("ip_address", ipAddress))
	}

	if len(fields) > 0 {
		return logger.With(fields...)
	}
	return logger
}

// LoggerWithContext is a helper type that combines logger and context
type LoggerWithContext struct {
	logger *zap.Logger
	ctx    context.Context
}

// NewLoggerWithContext creates a new LoggerWithContext
func NewLoggerWithContext(ctx context.Context, logger *zap.Logger) *LoggerWithContext {
	return &LoggerWithContext{
		logger: FromContext(ctx, logger),
		ctx:    ctx,
	}
}

// Debug logs a debug message with context
func (l *LoggerWithContext) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message with context
func (l *LoggerWithContext) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message with context
func (l *LoggerWithContext) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message with context
func (l *LoggerWithContext) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message with context
func (l *LoggerWithContext) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// With adds fields to the logger
func (l *LoggerWithContext) With(fields ...zap.Field) *LoggerWithContext {
	return &LoggerWithContext{
		logger: l.logger.With(fields...),
		ctx:    l.ctx,
	}
}
