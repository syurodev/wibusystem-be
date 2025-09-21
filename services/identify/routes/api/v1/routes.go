// Package v1 contains all API version 1 routes and handlers.
// This package encapsulates the v1 API implementation, making it easy to
// maintain multiple API versions in parallel and migrate between them.
package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/pkg/database/factory"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/config"
	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
)

// SetupV1Routes configures all API v1 endpoints and their middleware
func SetupV1Routes(
	router *gin.Engine,
	dbManager *factory.DatabaseManager,
	cfg *config.Config,
	translator *i18n.Translator,
	h *handlers.Handlers,
	m *middleware.Manager,
) {
	// API v1 group
	v1 := router.Group("/api/v1")
	v1.Use(m.SetupAPIMiddleware()...)

	// Setup auth routes (public and protected)
	setupAuthRoutes(v1, h, m)

	// OpenID Connect UserInfo endpoint
	v1.GET("/userinfo", m.Auth.RequireAuth(), m.Auth.RequireScope("openid"), h.OAuth2.UserInfoHandler)

	// Setup protected resource routes
	setupUserRoutes(v1, h, m)
	setupTenantRoutes(v1, h, m)

	// Setup admin routes (on main router, not v1 group)
	setupAdminRoutes(router, h, m)
}
