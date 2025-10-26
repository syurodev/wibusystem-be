package oauth2

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho module OAuth2.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/userinfo", h.UserInfo)
}

// RegisterWellKnownRoutes đăng ký các routes cho /.well-known.
func (h *Handler) RegisterWellKnownRoutes(group *gin.RouterGroup) {
	group.GET("/openid-configuration", h.Discovery)
	group.GET("/jwks.json", h.JWKS)
}
