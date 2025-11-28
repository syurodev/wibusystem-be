package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// PerformanceLogger provides methods for logging performance metrics
type PerformanceLogger struct {
	logger *zap.Logger
}

// NewPerformanceLogger creates a new PerformanceLogger instance
func NewPerformanceLogger(logger *zap.Logger) *PerformanceLogger {
	return &PerformanceLogger{
		logger: logger.Named("performance"),
	}
}

// OperationType định nghĩa loại operation được track
type OperationType string

const (
	OpTypeHTTP         OperationType = "http"
	OpTypeDatabase     OperationType = "database"
	OpTypeRedis        OperationType = "redis"
	OpTypeExternal     OperationType = "external_api"
	OpTypeOAuth2       OperationType = "oauth2"
	OpTypeAuthentication OperationType = "authentication"
)

// PerformanceMetric chứa thông tin về performance của một operation
type PerformanceMetric struct {
	OperationType OperationType `json:"operation_type"`
	Operation     string        `json:"operation"`
	Duration      time.Duration `json:"duration_ms"`
	Success       bool          `json:"success"`

	// HTTP specific
	Method        string        `json:"method,omitempty"`
	Path          string        `json:"path,omitempty"`
	StatusCode    int           `json:"status_code,omitempty"`

	// Database specific
	Query         string        `json:"query,omitempty"`
	QueryType     string        `json:"query_type,omitempty"` // SELECT, INSERT, UPDATE, DELETE
	RowsAffected  int64         `json:"rows_affected,omitempty"`

	// Cache specific
	CacheHit      *bool         `json:"cache_hit,omitempty"`

	// Error information
	Error         string        `json:"error,omitempty"`

	// Additional metadata
	Metadata      map[string]any `json:"metadata,omitempty"`

	Timestamp     time.Time     `json:"timestamp"`
}

// LogMetric logs a performance metric
func (pl *PerformanceLogger) LogMetric(ctx context.Context, metric *PerformanceMetric) {
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now().UTC()
	}

	fields := []zap.Field{
		zap.String("category", "performance"),
		zap.String("operation_type", string(metric.OperationType)),
		zap.String("operation", metric.Operation),
		zap.Int64("duration_ms", metric.Duration.Milliseconds()),
		zap.Bool("success", metric.Success),
		zap.Time("timestamp", metric.Timestamp),
	}

	// Add HTTP specific fields
	if metric.Method != "" {
		fields = append(fields, zap.String("method", metric.Method))
	}
	if metric.Path != "" {
		fields = append(fields, zap.String("path", metric.Path))
	}
	if metric.StatusCode > 0 {
		fields = append(fields, zap.Int("status_code", metric.StatusCode))
	}

	// Add database specific fields
	if metric.Query != "" {
		fields = append(fields, zap.String("query", metric.Query))
	}
	if metric.QueryType != "" {
		fields = append(fields, zap.String("query_type", metric.QueryType))
	}
	if metric.RowsAffected > 0 {
		fields = append(fields, zap.Int64("rows_affected", metric.RowsAffected))
	}

	// Add cache specific fields
	if metric.CacheHit != nil {
		fields = append(fields, zap.Bool("cache_hit", *metric.CacheHit))
	}

	// Add error if present
	if metric.Error != "" {
		fields = append(fields, zap.String("error", metric.Error))
	}

	// Add metadata
	if len(metric.Metadata) > 0 {
		fields = append(fields, zap.Any("metadata", metric.Metadata))
	}

	// Determine log level based on duration and success
	message := metric.Operation
	if !metric.Success {
		pl.logger.Error(message, fields...)
	} else if metric.Duration > 5*time.Second {
		// Slow operation warning
		pl.logger.Warn(message+" (slow)", fields...)
	} else {
		pl.logger.Info(message, fields...)
	}
}

// LogHTTPRequest logs HTTP request performance
func (pl *PerformanceLogger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, err error) {
	metric := &PerformanceMetric{
		OperationType: OpTypeHTTP,
		Operation:     "http_request",
		Method:        method,
		Path:          path,
		StatusCode:    statusCode,
		Duration:      duration,
		Success:       err == nil && statusCode < 500,
	}

	if err != nil {
		metric.Error = err.Error()
	}

	pl.LogMetric(ctx, metric)
}

// LogDatabaseQuery logs database query performance
func (pl *PerformanceLogger) LogDatabaseQuery(ctx context.Context, queryType, query string, duration time.Duration, rowsAffected int64, err error) {
	metric := &PerformanceMetric{
		OperationType: OpTypeDatabase,
		Operation:     "db_query",
		Query:         query,
		QueryType:     queryType,
		Duration:      duration,
		RowsAffected:  rowsAffected,
		Success:       err == nil,
	}

	if err != nil {
		metric.Error = err.Error()
	}

	pl.LogMetric(ctx, metric)
}

// LogCacheOperation logs cache operation performance
func (pl *PerformanceLogger) LogCacheOperation(ctx context.Context, operation string, key string, hit bool, duration time.Duration, err error) {
	metric := &PerformanceMetric{
		OperationType: OpTypeRedis,
		Operation:     operation,
		Duration:      duration,
		CacheHit:      &hit,
		Success:       err == nil,
		Metadata: map[string]any{
			"key": key,
		},
	}

	if err != nil {
		metric.Error = err.Error()
	}

	pl.LogMetric(ctx, metric)
}

// LogOAuth2Operation logs OAuth2 operation performance
func (pl *PerformanceLogger) LogOAuth2Operation(ctx context.Context, operation string, grantType string, duration time.Duration, success bool, err error) {
	metric := &PerformanceMetric{
		OperationType: OpTypeOAuth2,
		Operation:     operation,
		Duration:      duration,
		Success:       success,
		Metadata: map[string]any{
			"grant_type": grantType,
		},
	}

	if err != nil {
		metric.Error = err.Error()
	}

	pl.LogMetric(ctx, metric)
}

// PerformanceTimer helps measure operation duration
type PerformanceTimer struct {
	startTime time.Time
	logger    *PerformanceLogger
	metric    *PerformanceMetric
}

// NewTimer creates a new performance timer
func (pl *PerformanceLogger) NewTimer(operation string, opType OperationType) *PerformanceTimer {
	return &PerformanceTimer{
		startTime: time.Now(),
		logger:    pl,
		metric: &PerformanceMetric{
			OperationType: opType,
			Operation:     operation,
			Success:       true,
		},
	}
}

// WithMetadata adds metadata to the timer
func (pt *PerformanceTimer) WithMetadata(key string, value any) *PerformanceTimer {
	if pt.metric.Metadata == nil {
		pt.metric.Metadata = make(map[string]any)
	}
	pt.metric.Metadata[key] = value
	return pt
}

// WithHTTPInfo adds HTTP information to the timer
func (pt *PerformanceTimer) WithHTTPInfo(method, path string, statusCode int) *PerformanceTimer {
	pt.metric.Method = method
	pt.metric.Path = path
	pt.metric.StatusCode = statusCode
	return pt
}

// End stops the timer and logs the metric
func (pt *PerformanceTimer) End(ctx context.Context, err error) {
	pt.metric.Duration = time.Since(pt.startTime)
	if err != nil {
		pt.metric.Success = false
		pt.metric.Error = err.Error()
	}
	pt.logger.LogMetric(ctx, pt.metric)
}

// EndWithSuccess stops the timer with explicit success status
func (pt *PerformanceTimer) EndWithSuccess(ctx context.Context, success bool, err error) {
	pt.metric.Duration = time.Since(pt.startTime)
	pt.metric.Success = success
	if err != nil {
		pt.metric.Error = err.Error()
	}
	pt.logger.LogMetric(ctx, pt.metric)
}
