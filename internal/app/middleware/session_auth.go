package middleware

import (
	"context"
	"net/http"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SessionValidator interface for validating user sessions
type SessionValidator interface {
	GetUserSession(ctx context.Context, sessionID string) (string, error)
}

// RequireSessionAuth là middleware xác thực user dựa trên session cookie.
// Middleware này được sử dụng cho các internal endpoints (HTMX, frontend) thay vì OAuth2 token.
func RequireSessionAuth(oauth2Service SessionValidator, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Lấy session_id từ cookie
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			logger.Warn("Session Auth Middleware: Missing session_id cookie")
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
			c.Abort()
			return
		}

		// 2. Validate session và lấy userID
		userID, err := oauth2Service.GetUserSession(c.Request.Context(), sessionID)
		if err != nil {
			logger.Error("Session Auth Middleware: Failed to get session", zap.Error(err))
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
			c.Abort()
			return
		}

		if userID == "" {
			logger.Warn("Session Auth Middleware: Session not found or expired")
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
			c.Abort()
			return
		}

		// 3. Set userID vào context
		c.Set("user_id", userID)
		
		logger.Info("Session Auth Middleware: User authenticated", zap.String("user_id", userID))

		c.Next()
	}
}

