package organization

import (
	"encoding/json"

	"github.com/gofrs/uuid/v5"
)

// CreateOrganizationRequest - Tạo org mới (user sẽ trở thành owner)
type CreateOrganizationRequest struct {
	Name        string          `json:"name" binding:"required,min=2,max=255"`
	Slug        string          `json:"slug" binding:"required,min=2,max=255"`
	Description json.RawMessage `json:"description"`
}

// UpdateOrganizationRequest - Cập nhật thông tin org
type UpdateOrganizationRequest struct {
	Name         *string          `json:"name" binding:"omitempty,min=2,max=255"`
	Description  json.RawMessage  `json:"description"`
	AvatarURL    *string          `json:"avatar_url"`
	IsRecruiting *bool            `json:"is_recruiting"`
}

// InviteMemberRequest - Invite member vào org (role mặc định = member)
type InviteMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// UpdateMemberRoleRequest - Thay đổi role của member (chỉ owner)
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// UpdateSettingsRequest - Cập nhật settings (chỉ owner)
type UpdateSettingsRequest struct {
	BypassInviteApproval *bool `json:"bypass_invite_approval"`
}

// ProcessInviteRequest - Duyệt/từ chối invite (owner/admin)
type ProcessInviteRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
}

// ReportOrganizationRequest - Report organization
type ReportOrganizationRequest struct {
	Reason      string  `json:"reason" binding:"required,oneof=spam harassment inappropriate_content copyright_violation fake_translations other"`
	Description *string `json:"description" binding:"omitempty,max=1000"`
}

// RespondReportRequest - Owner/Admin phản hồi report
type RespondReportRequest struct {
	Response string `json:"response" binding:"required,min=10,max=2000"`
}

// ListMembersRequest - Query params cho list members
type ListMembersRequest struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

// ListOrganizationsRequest - Query params cho list organizations
type ListOrganizationsRequest struct {
	Search       *string `form:"search"`
	Status       *string `form:"status" binding:"omitempty,oneof=active flagged suspended archived"`
	IsRecruiting *bool   `form:"is_recruiting"`
	SortBy       string  `form:"sort_by" binding:"omitempty,oneof=name member_count completed_translations created_at"`
	SortOrder    string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Limit        int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset       int     `form:"offset" binding:"omitempty,min=0"`
}
