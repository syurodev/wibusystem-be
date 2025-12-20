package organization

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho organization
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	const identifierPath = "/:identifier"

	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(requireAuth)
	{
		// User's organizations - MUST BE BEFORE public routes using wildcard
		protected.GET("/me", h.GetMyOrganizations) // GET /api/v1/organizations/me

		// Organization CRUD
		protected.POST("", h.CreateOrganization)               // POST /api/v1/organizations
		protected.PUT(identifierPath, h.UpdateOrganization)    // PUT /api/v1/organizations/:identifier
		protected.DELETE(identifierPath, h.DeleteOrganization) // DELETE /api/v1/organizations/:identifier

		// Settings (owner only)
		protected.PUT(identifierPath+"/settings", h.UpdateSettings) // PUT /api/v1/organizations/:identifier/settings

		// Membership
		protected.POST(identifierPath+"/members", h.InviteMember)                  // POST /api/v1/organizations/:identifier/members
		protected.DELETE(identifierPath+"/members/:user_id", h.RemoveMember)       // DELETE /api/v1/organizations/:identifier/members/:user_id
		protected.PUT(identifierPath+"/members/:user_id/role", h.UpdateMemberRole) // PUT /api/v1/organizations/:identifier/members/:user_id/role
		protected.DELETE(identifierPath+"/leave", h.LeaveOrganization)             // DELETE /api/v1/organizations/:identifier/leave

		// Pending Invites
		protected.GET(identifierPath+"/invites", h.ListPendingInvites)        // GET /api/v1/organizations/:identifier/invites
		protected.POST(identifierPath+"/invites/:id", h.ProcessPendingInvite) // POST /api/v1/organizations/:identifier/invites/:id

		// Reports
		protected.POST(identifierPath+"/report", h.ReportOrganization)           // POST /api/v1/organizations/:identifier/report
		protected.GET(identifierPath+"/reports", h.ListReports)                  // GET /api/v1/organizations/:identifier/reports
		protected.POST(identifierPath+"/reports/:id/respond", h.RespondToReport) // POST /api/v1/organizations/:identifier/reports/:id/respond
	}

	// Public routes (no auth required for reading)
	// These must be registered AFTER specific routes like /me if they use wildcards at the same level
	// But since /me is in protected group and public is in router group, we need to be careful about middleware order
	// However, if we mount protected group on router, the order matters for the underlying engine
	
	router.GET("", h.ListOrganizations)              // GET /api/v1/organizations
	router.GET("/top", h.GetTopOrgsByViews)          // GET /api/v1/organizations/top
	router.GET(identifierPath, h.GetOrganization)    // GET /api/v1/organizations/:identifier
	router.GET(identifierPath+"/members", h.ListMembers) // GET /api/v1/organizations/:identifier/members
}

