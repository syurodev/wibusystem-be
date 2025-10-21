// Package domain contains the core business entities and logic for the Identity module.
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the status of a tenant.
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusTrial     TenantStatus = "trial"
)

// Valid checks if the TenantStatus is valid.
func (s TenantStatus) Valid() bool {
	switch s {
	case TenantStatusActive, TenantStatusInactive, TenantStatusSuspended, TenantStatusTrial:
		return true
	default:
		return false
	}
}

// String returns the string representation of TenantStatus.
func (s TenantStatus) String() string {
	return string(s)
}

// Scan implements the sql.Scanner interface.
func (s *TenantStatus) Scan(value any) error {
	if value == nil {
		*s = TenantStatusTrial
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan TenantStatus: %v", value)
	}

	*s = TenantStatus(str)
	if !s.Valid() {
		return fmt.Errorf("invalid TenantStatus: %s", str)
	}

	return nil
}

// Value implements the driver.Valuer interface.
func (s TenantStatus) Value() (driver.Value, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid TenantStatus: %s", s)
	}
	return string(s), nil
}

// MemberRole represents the role of a user in a tenant.
type MemberRole string

const (
	MemberRoleOwner  MemberRole = "owner"
	MemberRoleAdmin  MemberRole = "admin"
	MemberRoleEditor MemberRole = "editor"
	MemberRoleMember MemberRole = "member"
	MemberRoleViewer MemberRole = "viewer"
)

// Valid checks if the MemberRole is valid.
func (r MemberRole) Valid() bool {
	switch r {
	case MemberRoleOwner, MemberRoleAdmin, MemberRoleEditor, MemberRoleMember, MemberRoleViewer:
		return true
	default:
		return false
	}
}

// String returns the string representation of MemberRole.
func (r MemberRole) String() string {
	return string(r)
}

// CanManageMembers returns true if the role can manage members.
func (r MemberRole) CanManageMembers() bool {
	return r == MemberRoleOwner || r == MemberRoleAdmin
}

// CanManageContent returns true if the role can manage content.
func (r MemberRole) CanManageContent() bool {
	return r == MemberRoleOwner || r == MemberRoleAdmin || r == MemberRoleEditor
}

// CanViewContent returns true if the role can view content.
func (r MemberRole) CanViewContent() bool {
	return true // All roles can view
}

// Tenant represents an organization/tenant in the multi-tenant system.
type Tenant struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description *string        `json:"description,omitempty"`
	LogoURL     *string        `json:"logo_url,omitempty"`
	Status      TenantStatus   `json:"status"`
	OwnerID     uuid.UUID      `json:"owner_id"`
	Settings    TenantSettings `json:"settings"`
	Metadata    TenantMetadata `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty"`
}

// TenantSettings represents tenant configuration settings.
type TenantSettings struct {
	MaxMembers            int    `json:"max_members,omitempty"`
	AllowPublicContent    bool   `json:"allow_public_content"`
	RequireApproval       bool   `json:"require_approval"`
	DefaultContentPrivacy string `json:"default_content_privacy,omitempty"`
}

// TenantMetadata represents arbitrary metadata for a tenant.
type TenantMetadata map[string]any

// Scan implements the sql.Scanner interface for TenantSettings.
func (s *TenantSettings) Scan(value any) error {
	if value == nil {
		*s = TenantSettings{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TenantSettings: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface for TenantSettings.
func (s TenantSettings) Value() (driver.Value, error) {
	if s == (TenantSettings{}) {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for TenantMetadata.
func (m *TenantMetadata) Scan(value any) error {
	if value == nil {
		*m = TenantMetadata{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TenantMetadata: expected []byte, got %T", value)
	}

	return json.Unmarshal(bytes, m)
}

// Value implements the driver.Valuer interface for TenantMetadata.
func (m TenantMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return json.Marshal(map[string]any{})
	}
	return json.Marshal(m)
}

// TenantMember represents a user's membership in a tenant.
type TenantMember struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Role        MemberRole `json:"role"`
	Permissions []string   `json:"permissions"`
	InvitedBy   *uuid.UUID `json:"invited_by,omitempty"`
	JoinedAt    time.Time  `json:"joined_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Slug validation regex (lowercase alphanumeric and hyphens, 3-100 chars)
var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,98}[a-z0-9])?$`)

// Domain errors for Tenant
var (
	ErrInvalidTenantID     = errors.New("invalid tenant ID")
	ErrInvalidOwnerID      = errors.New("invalid owner ID")
	ErrTenantNameEmpty     = errors.New("tenant name cannot be empty")
	ErrTenantNameTooLong   = errors.New("tenant name exceeds maximum length (255 characters)")
	ErrInvalidSlug         = errors.New("invalid slug format (must be lowercase alphanumeric with hyphens, 3-100 chars)")
	ErrSlugReserved        = errors.New("slug is reserved")
	ErrInvalidTenantStatus = errors.New("invalid tenant status")
	ErrTenantDeleted       = errors.New("tenant has been deleted")
	ErrInvalidMemberRole   = errors.New("invalid member role")
	ErrCannotRemoveOwner   = errors.New("cannot remove tenant owner")
	ErrMaxMembersReached   = errors.New("maximum number of members reached")
)

// Reserved slugs that cannot be used for tenants
var reservedSlugs = map[string]bool{
	"admin":     true,
	"api":       true,
	"www":       true,
	"app":       true,
	"dashboard": true,
	"settings":  true,
	"account":   true,
	"profile":   true,
	"system":    true,
	"public":    true,
	"private":   true,
	"support":   true,
	"help":      true,
	"docs":      true,
	"blog":      true,
	"about":     true,
	"contact":   true,
	"terms":     true,
	"privacy":   true,
}

// NewTenant creates a new Tenant with the given name, slug, and owner.
func NewTenant(name, slug string, ownerID uuid.UUID) (*Tenant, error) {
	// Validate inputs
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTenantNameEmpty
	}
	if len(name) > 255 {
		return nil, ErrTenantNameTooLong
	}

	if ownerID == uuid.Nil {
		return nil, ErrInvalidOwnerID
	}

	// Validate and normalize slug
	slug = strings.ToLower(strings.TrimSpace(slug))
	if err := ValidateSlug(slug); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tenant := &Tenant{
		ID:      uuid.New(),
		Name:    name,
		Slug:    slug,
		Status:  TenantStatusTrial,
		OwnerID: ownerID,
		Settings: TenantSettings{
			AllowPublicContent: true,
			RequireApproval:    false,
		},
		Metadata:  TenantMetadata{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return tenant, nil
}

// ValidateSlug validates a tenant slug.
func ValidateSlug(slug string) error {
	slug = strings.ToLower(strings.TrimSpace(slug))

	if !slugRegex.MatchString(slug) {
		return ErrInvalidSlug
	}

	if reservedSlugs[slug] {
		return ErrSlugReserved
	}

	return nil
}

// UpdateName updates the tenant's name.
func (t *Tenant) UpdateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrTenantNameEmpty
	}
	if len(name) > 255 {
		return ErrTenantNameTooLong
	}

	t.Name = name
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateDescription updates the tenant's description.
func (t *Tenant) UpdateDescription(description string) {
	description = strings.TrimSpace(description)
	if description == "" {
		t.Description = nil
	} else {
		t.Description = &description
	}
	t.UpdatedAt = time.Now().UTC()
}

// UpdateLogoURL updates the tenant's logo URL.
func (t *Tenant) UpdateLogoURL(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		t.LogoURL = nil
	} else {
		t.LogoURL = &url
	}
	t.UpdatedAt = time.Now().UTC()
}

// Activate activates the tenant.
func (t *Tenant) Activate() error {
	if t.IsDeleted() {
		return ErrTenantDeleted
	}
	t.Status = TenantStatusActive
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Deactivate deactivates the tenant.
func (t *Tenant) Deactivate() error {
	if t.IsDeleted() {
		return ErrTenantDeleted
	}
	t.Status = TenantStatusInactive
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Suspend suspends the tenant.
func (t *Tenant) Suspend() error {
	if t.IsDeleted() {
		return ErrTenantDeleted
	}
	t.Status = TenantStatusSuspended
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Delete soft-deletes the tenant.
func (t *Tenant) Delete() {
	now := time.Now().UTC()
	t.DeletedAt = &now
	t.Status = TenantStatusInactive
	t.UpdatedAt = now
}

// IsDeleted returns true if the tenant is soft-deleted.
func (t *Tenant) IsDeleted() bool {
	return t.DeletedAt != nil
}

// IsActive returns true if the tenant is active and not deleted.
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive && !t.IsDeleted()
}

// UpdateSettings updates the tenant's settings.
func (t *Tenant) UpdateSettings(settings TenantSettings) {
	t.Settings = settings
	t.UpdatedAt = time.Now().UTC()
}

// SetMetadata sets a metadata value.
func (t *Tenant) SetMetadata(key string, value any) {
	if t.Metadata == nil {
		t.Metadata = TenantMetadata{}
	}
	t.Metadata[key] = value
	t.UpdatedAt = time.Now().UTC()
}

// GetMetadata gets a metadata value.
func (t *Tenant) GetMetadata(key string) (any, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	value, ok := t.Metadata[key]
	return value, ok
}

// Validate validates the tenant entity.
func (t *Tenant) Validate() error {
	if t.ID == uuid.Nil {
		return ErrInvalidTenantID
	}

	if t.OwnerID == uuid.Nil {
		return ErrInvalidOwnerID
	}

	if strings.TrimSpace(t.Name) == "" {
		return ErrTenantNameEmpty
	}

	if len(t.Name) > 255 {
		return ErrTenantNameTooLong
	}

	if err := ValidateSlug(t.Slug); err != nil {
		return err
	}

	if !t.Status.Valid() {
		return ErrInvalidTenantStatus
	}

	return nil
}

// Clone creates a deep copy of the Tenant.
func (t *Tenant) Clone() *Tenant {
	clone := *t

	if t.Description != nil {
		desc := *t.Description
		clone.Description = &desc
	}

	if t.LogoURL != nil {
		url := *t.LogoURL
		clone.LogoURL = &url
	}

	if t.DeletedAt != nil {
		timestamp := *t.DeletedAt
		clone.DeletedAt = &timestamp
	}

	// Deep copy settings and metadata
	clone.Settings = t.Settings
	clone.Metadata = make(TenantMetadata, len(t.Metadata))
	for k, v := range t.Metadata {
		clone.Metadata[k] = v
	}

	return &clone
}

// NewTenantMember creates a new TenantMember.
func NewTenantMember(tenantID, userID uuid.UUID, role MemberRole, invitedBy *uuid.UUID) (*TenantMember, error) {
	if tenantID == uuid.Nil {
		return nil, ErrInvalidTenantID
	}

	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	if !role.Valid() {
		return nil, ErrInvalidMemberRole
	}

	now := time.Now().UTC()
	member := &TenantMember{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      userID,
		Role:        role,
		Permissions: []string{},
		InvitedBy:   invitedBy,
		JoinedAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return member, nil
}

// UpdateRole updates the member's role.
func (m *TenantMember) UpdateRole(role MemberRole) error {
	if !role.Valid() {
		return ErrInvalidMemberRole
	}

	if m.Role == MemberRoleOwner {
		return ErrCannotRemoveOwner
	}

	m.Role = role
	m.UpdatedAt = time.Now().UTC()
	return nil
}

// AddPermission adds a permission to the member.
func (m *TenantMember) AddPermission(permission string) {
	for _, p := range m.Permissions {
		if p == permission {
			return // Already has permission
		}
	}
	m.Permissions = append(m.Permissions, permission)
	m.UpdatedAt = time.Now().UTC()
}

// RemovePermission removes a permission from the member.
func (m *TenantMember) RemovePermission(permission string) {
	for i, p := range m.Permissions {
		if p == permission {
			m.Permissions = append(m.Permissions[:i], m.Permissions[i+1:]...)
			m.UpdatedAt = time.Now().UTC()
			return
		}
	}
}

// HasPermission checks if the member has a specific permission.
func (m *TenantMember) HasPermission(permission string) bool {
	// Owners and admins have all permissions
	if m.Role == MemberRoleOwner || m.Role == MemberRoleAdmin {
		return true
	}

	for _, p := range m.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsOwner returns true if the member is the owner.
func (m *TenantMember) IsOwner() bool {
	return m.Role == MemberRoleOwner
}

// CanManageMembers returns true if the member can manage other members.
func (m *TenantMember) CanManageMembers() bool {
	return m.Role.CanManageMembers()
}

// CanManageContent returns true if the member can manage content.
func (m *TenantMember) CanManageContent() bool {
	return m.Role.CanManageContent()
}
