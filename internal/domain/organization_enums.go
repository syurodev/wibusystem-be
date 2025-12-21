package domain

// OrganizationStatus định nghĩa trạng thái của organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusFlagged   OrganizationStatus = "flagged"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

// IsValid kiểm tra xem status có hợp lệ không
func (s OrganizationStatus) IsValid() bool {
	switch s {
	case OrganizationStatusActive, OrganizationStatusFlagged, OrganizationStatusSuspended, OrganizationStatusArchived:
		return true
	default:
		return false
	}
}

// OrganizationMemberRole định nghĩa role của thành viên trong organization
type OrganizationMemberRole string

const (
	OrganizationMemberRoleOwner  OrganizationMemberRole = "owner"
	OrganizationMemberRoleAdmin  OrganizationMemberRole = "admin"
	OrganizationMemberRoleMember OrganizationMemberRole = "member"
)

// IsValid kiểm tra xem role có hợp lệ không
func (r OrganizationMemberRole) IsValid() bool {
	switch r {
	case OrganizationMemberRoleOwner, OrganizationMemberRoleAdmin, OrganizationMemberRoleMember:
		return true
	default:
		return false
	}
}
