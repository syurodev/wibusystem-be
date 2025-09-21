package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	r "wibusystem/pkg/common/response"
)

// setupPublicRoutes configures all public, non-versioned routes
// These include OAuth2 endpoints, OIDC discovery, health checks, and login pages
func setupPublicRoutes(router *gin.Engine, deps *Dependencies) {
	// Health check endpoint
	router.GET("/health", healthCheckHandler(deps))

	// Minimal built-in login page to support OAuth2 authorize flow
	router.GET("/login", loginPageHandler)

	// OAuth2 endpoints (public, with OAuth2-specific middleware)
	setupOAuth2Routes(router, deps)

	// OpenID Connect discovery endpoints
	setupOIDCDiscoveryRoutes(router, deps)

	// Dynamic Client Registration endpoints
	setupDCRRoutes(router, deps)
}

// healthCheckHandler returns the health status of the service
func healthCheckHandler(deps *Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		if err := deps.DBManager.Health(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, r.StandardResponse{
				Success: false,
				Message: "Unhealthy",
				Data:    nil,
				Error: &r.ErrorDetail{
					Code:        "database_error",
					Description: "database connection failed",
				},
				Meta: map[string]interface{}{},
			})
			return
		}

		c.JSON(http.StatusOK, r.StandardResponse{
			Success: true,
			Message: "Healthy",
			Data:    map[string]interface{}{"service": "identify-service"},
			Error:   nil,
			Meta:    map[string]interface{}{"timestamp": time.Now().UTC()},
		})
	}
}

// loginPageHandler serves a minimal login page for OAuth2 flows
func loginPageHandler(c *gin.Context) {
	redirect := c.Query("redirect_uri")
	if redirect == "" {
		redirect = "/oauth2/authorize"
	}
	html := `<!doctype html><html><head><meta charset="utf-8"><title>Sign in</title>
	<style>body{font-family:system-ui,sans-serif;padding:40px;max-width:480px;margin:auto}label{display:block;margin-top:12px}input{width:100%;padding:10px;margin-top:6px}button{margin-top:16px;padding:10px 14px}</style>
	</head><body>
	<h2>Sign in</h2>
	<form id="f">
	<label>Email<input type="email" name="email" required></label>
	<label>Password<input type="password" name="password" required></label>
	<button type="submit">Sign in</button>
	</form>
	<script>
	const f=document.getElementById('f');
	f.addEventListener('submit', async (e)=>{
	  e.preventDefault();
	  const data=Object.fromEntries(new FormData(f).entries());
	  const res=await fetch('/api/v1/auth/login',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});
	  if(res.ok){ window.location=decodeURIComponent('` + redirect + `'); } else { alert('Login failed'); }
	});
	</script>
	</body></html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// setupOAuth2Routes configures OAuth2/OIDC protocol endpoints
func setupOAuth2Routes(router *gin.Engine, deps *Dependencies) {
	oauth2Group := router.Group("/oauth2")
	oauth2Group.Use(deps.Middleware.SetupOAuth2Middleware()...)
	{
		oauth2Group.GET("/authorize", deps.Handlers.OAuth2.AuthorizeHandler)
		oauth2Group.POST("/authorize", deps.Handlers.OAuth2.AuthorizeHandler)
		oauth2Group.POST("/token", deps.Handlers.OAuth2.TokenHandler)
		oauth2Group.POST("/introspect", deps.Handlers.OAuth2.IntrospectHandler)
		oauth2Group.POST("/revoke", deps.Handlers.OAuth2.RevokeHandler)
	}

	// Consent endpoints (production-style)
	router.GET("/oauth2/consent", deps.Handlers.OAuth2.GetConsent)
	router.POST("/oauth2/consent", deps.Handlers.OAuth2.PostConsent)
}

// setupOIDCDiscoveryRoutes configures OpenID Connect discovery endpoints
func setupOIDCDiscoveryRoutes(router *gin.Engine, deps *Dependencies) {
	// Serve both underscore and standard dashed paths
	discoveryHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"issuer":                                deps.Config.OAuth2.Issuer,
			"authorization_endpoint":                deps.Config.OAuth2.Issuer + "/oauth2/authorize",
			"token_endpoint":                        deps.Config.OAuth2.Issuer + "/oauth2/token",
			"userinfo_endpoint":                     deps.Config.OAuth2.Issuer + "/api/v1/userinfo",
			"jwks_uri":                              deps.Config.OAuth2.Issuer + "/.well-known/jwks.json",
			"registration_endpoint":                 deps.Config.OAuth2.Issuer + "/register",
			"scopes_supported":                      []string{"openid", "profile", "email", "offline_access", "admin"},
			"response_types_supported":              []string{"code", "token", "id_token", "code token", "code id_token", "token id_token", "code token id_token"},
			"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
			"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce", "email", "email_verified", "name", "preferred_username"},
		})
	}
	router.GET("/.well-known/openid_configuration", discoveryHandler)
	router.GET("/.well-known/openid-configuration", discoveryHandler)

	// JWKS endpoint
	router.GET("/.well-known/jwks.json", deps.Handlers.OAuth2.JWKSHandler)
}

// setupDCRRoutes configures Dynamic Client Registration endpoints
func setupDCRRoutes(router *gin.Engine, deps *Dependencies) {
	router.POST("/register", deps.Handlers.OAuth2.RegisterClient)
	router.GET("/register/:client_id", deps.Handlers.OAuth2.GetRegisteredClient)
	router.PUT("/register/:client_id", deps.Handlers.OAuth2.UpdateRegisteredClient)
	router.DELETE("/register/:client_id", deps.Handlers.OAuth2.DeleteRegisteredClient)
}
