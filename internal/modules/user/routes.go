package user

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	// User settings routes
	// Assuming the group passed here is already /users or similar, 
	// OR we are passed the v1 group and we create /user group.
	// Let's follow the pattern seen in router.go:
	// genreGroup := apiV1.Group("/genres") -> genreHandler.RegisterRoutes(genreGroup)
	// So we should expect a group that is specific to user or we define it.
	
	// However, usually /user/settings implies singular "user" (current user).
	// So in router.go we should probably create a group "/user".
	
	rg.GET("/settings", h.GetSettings)
	rg.PATCH("/settings", h.UpdateSettings)

	// Account management
	rg.GET("/profile", h.GetProfile)
	rg.PUT("/profile", h.UpdateProfile)
	rg.PUT("/security/password", h.ChangePassword)
	rg.GET("/sessions", h.GetSessions)
	rg.DELETE("/sessions/:id", h.DeleteSession)
}
