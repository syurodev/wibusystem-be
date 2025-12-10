package auth

import (
	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all auth API routes (v1/auth)
func (h *Handler) RegisterAPIRoutes(group *gin.RouterGroup, sessionAuth gin.HandlerFunc) {
	authGroup := group.Group("/auth")
	
	// Public auth endpoints
	authGroup.POST("/register", h.Register)
	authGroup.GET("/verify-email", h.VerifyEmail)
	authGroup.POST("/verify-email", h.VerifyEmail)
	authGroup.POST("/forgot-password", h.ForgotPassword)
	authGroup.POST("/reset-password", h.ResetPassword)

	// WebAuthn Registration (requires auth)
	authGroup.POST("/passkey/register/begin", sessionAuth, h.PasskeyRegisterBegin)
	authGroup.POST("/passkey/register/finish", sessionAuth, h.PasskeyRegisterFinish)

	// WebAuthn Authentication (public)
	authGroup.POST("/passkey/authenticate/begin", h.PasskeyAuthenticateBegin)
	authGroup.POST("/passkey/authenticate/finish", h.PasskeyAuthenticateFinish)

	// WebAuthn Credential management (requires auth)
	authGroup.GET("/passkey/credentials", sessionAuth, h.PasskeyListCredentials)
	authGroup.DELETE("/passkey/credentials", sessionAuth, h.PasskeyDeleteCredential)
	authGroup.PUT("/passkey/credentials/name", sessionAuth, h.PasskeyUpdateCredentialName)
}

// RegisterPageRoutes registers auth HTML pages and account routes
func (h *Handler) RegisterPageRoutes(router *gin.RouterGroup, sessionAuth gin.HandlerFunc) {
	// Passkey HTML pages
	authHTML := router.Group("/auth")
	authHTML.GET("/passkey/setup", h.PasskeySetupPage)
	authHTML.GET("/passkey/register", h.PasskeyRegisterPage)
	authHTML.GET("/passkey/manage", h.PasskeyManagePage)

	// Account management (requires session auth)
	account := router.Group("/account", sessionAuth)
	account.GET("/passkeys", h.PasskeyListFragment)
	account.DELETE("/passkeys/:id", h.PasskeyDeleteFragment)
	account.PUT("/passkeys/:id/name", h.PasskeyUpdateNameFragment)
}

// RegisterOAuth2Pages registers auth pages mounted under OAuth2 prefix
func (h *Handler) RegisterOAuth2Pages(group *gin.RouterGroup) {
	group.GET("/register", h.RegisterPage)
	group.GET("/verify-email", h.VerifyEmailPage)
	group.GET("/forgot-password", h.ForgotPasswordPage)
	group.GET("/reset-password", h.ResetPasswordPage)
}
