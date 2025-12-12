package domain

// OrganizationStatus định nghĩa trạng thái của organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusFlagged   OrganizationStatus = "flagged"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

// OrganizationMemberRole định nghĩa role của thành viên trong organization
type OrganizationMemberRole string

const (
	OrganizationMemberRoleOwner  OrganizationMemberRole = "owner"
	OrganizationMemberRoleAdmin  OrganizationMemberRole = "admin"
	OrganizationMemberRoleMember OrganizationMemberRole = "member"
)
