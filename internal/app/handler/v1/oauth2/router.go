package oauth2

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho module OAuth2.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	// Core OAuth2 endpoints
	group.GET("/auth", h.Authorize)         // Authorization endpoint
	group.POST("/token", h.Token)           // Token endpoint
	group.GET("/userinfo", h.UserInfo)      // UserInfo endpoint
	group.POST("/revoke", h.Revoke)         // Token revocation endpoint (RFC 7009)
	group.POST("/introspect", h.Introspect) // Token introspection endpoint (RFC 7662)
	group.GET("/logout", h.Logout)          // RP-Initiated Logout endpoint (OpenID Connect)
	group.POST("/logout", h.Logout)         // Support both GET and POST for logout

	// Login flow
	group.GET("/login", h.LoginPage)    // Login page
	group.POST("/login", h.LoginSubmit) // Login form submit

	// Consent flow
	group.GET("/consent", h.ConsentPage)    // Consent page
	group.POST("/consent", h.ConsentSubmit) // Consent form submit
}

// RegisterWellKnownRoutes đăng ký các routes cho /.well-known.
func (h *Handler) RegisterWellKnownRoutes(group *gin.RouterGroup) {
	group.GET("/openid-configuration", h.Discovery)
	group.GET("/jwks.json", h.JWKS)
}
