package organization

import "encoding/json"

// OrganizationResponse - Response cơ bản cho org
type OrganizationResponse struct {
	ID                    string           `json:"id"`
	Name                  string           `json:"name"`
	Slug                  string           `json:"slug"`
	Status                string           `json:"status"` // active, flagged, suspended, archived
	Description           *json.RawMessage `json:"description,omitempty"`
	AvatarURL             *string          `json:"avatar_url,omitempty"`
	IsRecruiting          bool             `json:"is_recruiting"`
	MemberCount           int              `json:"member_count"`
	CompletedTranslations int              `json:"completed_translations"`
	CreatedAt             string           `json:"created_at"`

	// Rank Comparison (Optional)
	CurrentRank  *int `json:"current_rank,omitempty"`
	PreviousRank *int `json:"previous_rank,omitempty"`
	RankChange   *int `json:"rank_change,omitempty"`
}

// OrganizationDetailResponse - Response chi tiết
type OrganizationDetailResponse struct {
	OrganizationResponse
	Settings       *OrganizationSettingsResponse `json:"settings,omitempty"`        // Chỉ owner/admin thấy
	MyRole         *string                       `json:"my_role,omitempty"`         // Role của current user
	CanInvite      bool                          `json:"can_invite"`                // User có thể invite không
	PendingInvites int                           `json:"pending_invites,omitempty"` // Số invite chờ duyệt (owner/admin)
	ReportCount    int                           `json:"report_count,omitempty"`    // Chỉ moderator thấy
	HasReported    bool                          `json:"has_reported"`              // User đã report chưa
}

// OrganizationSettingsResponse - Settings của org
type OrganizationSettingsResponse struct {
	BypassInviteApproval bool `json:"bypass_invite_approval"`
}

// MemberResponse - Response cho member
type MemberResponse struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Username    *string `json:"username,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Role        string  `json:"role"` // owner/admin/member
	JoinedAt    *string `json:"joined_at,omitempty"`
	Status      string  `json:"status"`
}

// PendingInviteResponse - Response cho pending invite
type PendingInviteResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Username    *string `json:"username,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	InvitedBy   string  `json:"invited_by"`
	InviterName string  `json:"inviter_name"`
	ExpiresAt   string  `json:"expires_at"`
	CreatedAt   string  `json:"created_at"`
}

// ReportResponse - Response cho report (cho org owner/admin)
type ReportResponse struct {
	ID             string  `json:"id"`
	ReporterID     string  `json:"reporter_id"`
	ReporterName   string  `json:"reporter_name"`
	Reason         string  `json:"reason"`
	Description    *string `json:"description,omitempty"`
	OrgResponse    *string `json:"org_response,omitempty"`
	OrgRespondedBy *string `json:"org_responded_by,omitempty"`
	OrgRespondedAt *string `json:"org_responded_at,omitempty"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
}

// ListOrganizationsResponse - Response cho list organizations
type ListOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Total         int64                  `json:"total"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

// ListMembersResponse - Response cho list members
type ListMembersResponse struct {
	Members []MemberResponse `json:"members"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// ListPendingInvitesResponse - Response cho list pending invites
type ListPendingInvitesResponse struct {
	Invites []PendingInviteResponse `json:"invites"`
	Total   int64                   `json:"total"`
}

// ListReportsResponse - Response cho list reports
type ListReportsResponse struct {
	Reports []ReportResponse `json:"reports"`
	Total   int64            `json:"total"`
}

// MyOrganizationsResponse - Response cho danh sách org của user
type MyOrganizationsResponse struct {
	Owned  *OrganizationResponse  `json:"owned,omitempty"` // Org mà user là owner
	Member []OrganizationResponse `json:"member"`          // Các org mà user là member
}
