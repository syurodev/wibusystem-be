package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GinZap là một Gin middleware để ghi log các request HTTP bằng Zap.
func GinZap(logger *zap.Logger, isProd bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Xử lý request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			// Ghi log các lỗi đã xảy ra trong context của Gin
			for _, e := range c.Errors.Errors() {
				logger.Error(e)
			}
		} else {
			// Chỉ ghi log các request thành công ở môi trường dev để giảm nhiễu
			if !isProd {
				logger.Info(path,
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.String("path", path),
					zap.String("query", query),
					zap.String("ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					zap.Duration("latency", latency),
				)
			}
		}
	}
}

// Logger là biến toàn cục để sử dụng Zap Logger sau khi khởi tạo.
var Logger *zap.Logger

// InitLogger khởi tạo cấu hình Zap Logger.
// isProduction quyết định cấu hình sẽ là dev (dễ đọc) hay prod (JSON).
func InitLogger(isProduction bool) (*zap.Logger, error) {
	var cfg zap.Config

	if isProduction {
		// Cấu hình cho Production: Output JSON, level INFO
		cfg = zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		// Cấu hình cho Development: Output console, level DEBUG
		cfg = zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		// Đặt lại encoder để dễ đọc hơn trong console
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Tùy chỉnh định dạng thời gian
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		return nil, err
	}

	// Quan trọng: Thay thế logger mặc định của Go bằng Zap
	zap.ReplaceGlobals(Logger)

	return Logger, nil
}

// SyncLogger gọi Sync() để đảm bảo tất cả buffered logs được ghi ra (nên gọi ở cuối main).
func SyncLogger() {
	// Chỉ ghi log nếu Logger đã được khởi tạo
	if Logger != nil {
		// Log bất kỳ lỗi nào trong quá trình đồng bộ
		if err := Logger.Sync(); err != nil && err.Error() != "/dev/stderr: file already closed" {
			// Loại bỏ lỗi thường thấy khi đóng stderr/stdout
			// (Thường là lỗi an toàn, không cần thiết phải panic)
			_, _ = os.Stderr.WriteString("Error syncing Zap logger: " + err.Error() + "\n")
		}
	}
}
