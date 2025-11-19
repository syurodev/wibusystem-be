package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"system/pkg/util/errcode"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"go.uber.org/zap"
)

// OAuth2Provider interface để inject vào middleware
type OAuth2Provider interface {
	IntrospectToken(ctx context.Context, token string, tokenUse fosite.TokenUse, session fosite.Session, scopes ...string) (fosite.TokenUse, fosite.AccessRequester, error)
}

// RequireAuth là middleware xác thực Bearer Token sử dụng OAuth2 introspection.
// Middleware này bảo vệ các protected resources, yêu cầu valid access token.
//
// Usage:
//
//	protectedRoutes := router.Group("/api")
//	protectedRoutes.Use(middleware.RequireAuth(oauth2Provider))
//	protectedRoutes.Use(middleware.RequireAuth(oauth2Provider, logger))
//	protectedRoutes.GET("/profile", profileHandler)
func RequireAuth(provider OAuth2Provider, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy Authorization header
		authHeader := c.GetHeader("Authorization")
		
		// DEBUG LOGGING
		logger.Info("Auth Middleware: Checking Authorization header", 
			zap.String("path", c.Request.URL.Path),
			zap.String("auth_header_len", string(len(authHeader))),
			zap.Bool("has_auth_header", authHeader != ""),
		)

		if authHeader == "" {
			logger.Warn("Auth Middleware: Missing Authorization header")
			response.AbortWithError(c, http.StatusUnauthorized, errcode.AuthMissingAuthHeader.String(), "auth.missing_authorization_header")
			return
		}

		// Parse Bearer token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			logger.Warn("Auth Middleware: Invalid Authorization format (missing Bearer prefix)")
			// Header không có "Bearer " prefix
			response.AbortWithError(c, http.StatusUnauthorized, errcode.AuthInvalidAuthorizationFormat.String(), "auth.invalid_authorization_format")
			return
		}

		logger.Info("Auth Middleware: Token found, introspecting...", zap.String("token_prefix", token[:min(5, len(token))]))

		// Introspect token để xác thực
		session := &fosite.DefaultSession{}
		tokenUse, ar, err := provider.IntrospectToken(c.Request.Context(), token, fosite.AccessToken, session)
		if err != nil {
			logger.Error("Auth Middleware: Token introspection failed", zap.Error(err))
			response.AbortWithError(c, http.StatusUnauthorized, errcode.AuthInvalidToken.String(), "auth.invalid_token")
			return
		}

		logger.Info("Auth Middleware: Token valid", 
			zap.String("subject", ar.GetSession().GetSubject()),
			zap.String("client_id", ar.GetClient().GetID()),
		)

		// Token valid - lưu thông tin vào context để handlers sử dụng
		c.Set("oauth2_token_use", string(tokenUse))
		c.Set("oauth2_client_id", ar.GetClient().GetID())
		c.Set("oauth2_subject", ar.GetSession().GetSubject())
		c.Set("oauth2_scopes", ar.GetGrantedScopes())
		c.Set("oauth2_access_request", ar)

		c.Next()
	}
}

// RequireScope là middleware kiểm tra token có chứa scope cần thiết không.
// Phải sử dụng sau RequireAuth middleware.
//
// Usage:
//
//	protectedRoutes.Use(middleware.RequireAuth(oauth2Provider))
//	protectedRoutes.Use(middleware.RequireScope("admin"))
//	protectedRoutes.GET("/admin/users", adminUsersHandler)
func RequireScope(requiredScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy scopes từ context (được set bởi RequireAuth)
		scopesValue, exists := c.Get("oauth2_scopes")
		if !exists {
			response.AbortWithError(c, http.StatusInternalServerError, errcode.MiddlewareError.String(), "auth.middleware_error")
			return
		}

		scopes, ok := scopesValue.([]string)
		if !ok {
			response.AbortWithError(c, http.StatusInternalServerError, errcode.ContextDataError.String(), "auth.middleware_error")
			return
		}

		// Kiểm tra required scope có trong granted scopes không
		hasScope := slices.Contains(scopes, requiredScope)

		if !hasScope {
			response.AbortWithError(c, http.StatusForbidden, errcode.AuthInsufficientScope.String(), "auth.insufficient_scope")
			return
		}

		c.Next()
	}
}

// RequireAnyScope là middleware kiểm tra token có chứa ít nhất một trong các scopes cần thiết.
// Phải sử dụng sau RequireAuth middleware.
//
// Usage:
//
//	protectedRoutes.Use(middleware.RequireAuth(oauth2Provider))
//	protectedRoutes.Use(middleware.RequireAnyScope("read", "write", "admin"))
//	protectedRoutes.GET("/api/data", dataHandler)
func RequireAnyScope(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy scopes từ context (được set bởi RequireAuth)
		scopesValue, exists := c.Get("oauth2_scopes")
		if !exists {
			response.AbortWithError(c, http.StatusInternalServerError, errcode.ScopeValidationError.String(), "auth.middleware_error")
			return
		}

		scopes, ok := scopesValue.([]string)
		if !ok {
			response.AbortWithError(c, http.StatusInternalServerError, errcode.ContextDataError.String(), "auth.middleware_error")
			return
		}

		// Kiểm tra có ít nhất một required scope trong granted scopes không
		hasAnyScope := false
		for _, requiredScope := range requiredScopes {
			for _, scope := range scopes {
				if scope == requiredScope {
					hasAnyScope = true
					break
				}
			}
			if hasAnyScope {
				break
			}
		}

		if !hasAnyScope {
			response.AbortWithError(c, http.StatusForbidden, errcode.AuthInsufficientAnyScope.String(), "auth.insufficient_scope")
			return
		}

		c.Next()
	}
}

// GetUserID helper function để lấy user ID từ context trong handlers
func GetUserID(c *gin.Context) (string, bool) {
	subject, exists := c.Get("oauth2_subject")
	if !exists {
		return "", false
	}

	userID, ok := subject.(string)
	return userID, ok
}

// GetClientID helper function để lấy client ID từ context trong handlers
func GetClientID(c *gin.Context) (string, bool) {
	clientID, exists := c.Get("oauth2_client_id")
	if !exists {
		return "", false
	}

	id, ok := clientID.(string)
	return id, ok
}

// GetScopes helper function để lấy granted scopes từ context trong handlers
func GetScopes(c *gin.Context) ([]string, bool) {
	scopes, exists := c.Get("oauth2_scopes")
	if !exists {
		return nil, false
	}

	scopeList, ok := scopes.([]string)
	return scopeList, ok
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
