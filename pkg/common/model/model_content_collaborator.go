package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Collaborator status constants
const (
	CollaboratorStatusPending  = "PENDING"  // Invitation sent, not accepted
	CollaboratorStatusActive   = "ACTIVE"   // Actively collaborating
	CollaboratorStatusInactive = "INACTIVE" // Temporarily inactive
	CollaboratorStatusRemoved  = "REMOVED"  // Removed from collaboration
)

// Collaborator permission constants
const (
	PermissionRead           = "READ"            // Can view content
	PermissionEdit           = "EDIT"            // Can edit content
	PermissionPublish        = "PUBLISH"         // Can publish/unpublish content
	PermissionDelete         = "DELETE"          // Can delete content
	PermissionManageChapters = "MANAGE_CHAPTERS" // Can add/remove chapters
	PermissionManagePricing  = "MANAGE_PRICING"  // Can update pricing
	PermissionViewAnalytics  = "VIEW_ANALYTICS"  // Can view analytics
	PermissionManageCollab   = "MANAGE_COLLAB"   // Can manage other collaborators
)

// RevenueNotes represents additional revenue sharing terms as JSONB
type RevenueNotes map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (rn RevenueNotes) Value() (driver.Value, error) {
	if rn == nil {
		return nil, nil
	}
	return json.Marshal(rn)
}

// Scan implements sql.Scanner interface for database retrieval
func (rn *RevenueNotes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into RevenueNotes", value)
	}

	return json.Unmarshal(bytes, rn)
}

// CustomPermissions represents custom permission overrides as JSONB
type CustomPermissions map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (cp CustomPermissions) Value() (driver.Value, error) {
	if cp == nil {
		return nil, nil
	}
	return json.Marshal(cp)
}

// Scan implements sql.Scanner interface for database retrieval
func (cp *CustomPermissions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into CustomPermissions", value)
	}

	return json.Unmarshal(bytes, cp)
}

// ContentCollaborator represents collaborative content creation
type ContentCollaborator struct {
	ID uuid.UUID `json:"id" db:"id"`

	// Content identification
	ContentType string    `json:"content_type" db:"content_type"` // NOVEL, VOLUME, CHAPTER
	ContentID   uuid.UUID `json:"content_id" db:"content_id"`     // ID of the content

	// Collaborator identification
	CollaboratorID   uuid.UUID `json:"collaborator_id" db:"collaborator_id"`     // User ID of collaborator
	CollaboratorType string    `json:"collaborator_type" db:"collaborator_type"` // 'user' or 'tenant'

	// Collaboration details
	Role        *string        `json:"role,omitempty" db:"role"`               // Optional role label (e.g., 'Co-Author', 'Editor')
	Permissions pq.StringArray `json:"permissions" db:"permissions"`           // Array of permissions
	Status      string         `json:"status" db:"status"`                     // Collaboration status

	// Revenue sharing
	RevenueSharePercent *float64      `json:"revenue_share_percent,omitempty" db:"revenue_share_percent"` // Percentage of revenue (0.00 to 100.00)
	RevenueNotes        *RevenueNotes `json:"revenue_notes,omitempty" db:"revenue_notes"`                 // Additional revenue terms

	// Workflow tracking
	InvitedByUserID  uuid.UUID  `json:"invited_by_user_id" db:"invited_by_user_id"`   // User who sent invitation
	AcceptedAt       *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`       // When invitation was accepted
	RemovedByUserID  *uuid.UUID `json:"removed_by_user_id,omitempty" db:"removed_by_user_id"` // User who removed collaborator
	RemovedAt        *time.Time `json:"removed_at,omitempty" db:"removed_at"`         // When collaborator was removed

	// Metadata
	CollaborationNotes *string            `json:"collaboration_notes,omitempty" db:"collaboration_notes"` // Notes about collaboration
	CustomPermissions  *CustomPermissions `json:"custom_permissions,omitempty" db:"custom_permissions"`   // Custom permission overrides

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// HasPermission checks if collaborator has a specific permission
func (cc *ContentCollaborator) HasPermission(permission string) bool {
	if cc.Status != CollaboratorStatusActive {
		return false
	}
	for _, p := range cc.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if collaborator has any of the specified permissions
func (cc *ContentCollaborator) HasAnyPermission(permissions ...string) bool {
	if cc.Status != CollaboratorStatusActive {
		return false
	}
	for _, required := range permissions {
		for _, p := range cc.Permissions {
			if p == required {
				return true
			}
		}
	}
	return false
}

// HasAllPermissions checks if collaborator has all specified permissions
func (cc *ContentCollaborator) HasAllPermissions(permissions ...string) bool {
	if cc.Status != CollaboratorStatusActive {
		return false
	}
	for _, required := range permissions {
		found := false
		for _, p := range cc.Permissions {
			if p == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// IsActive returns true if collaboration is active
func (cc *ContentCollaborator) IsActive() bool {
	return cc.Status == CollaboratorStatusActive
}

// IsPending returns true if collaboration invitation is pending
func (cc *ContentCollaborator) IsPending() bool {
	return cc.Status == CollaboratorStatusPending
}

// CanAccept checks if invitation can be accepted
func (cc *ContentCollaborator) CanAccept() bool {
	return cc.Status == CollaboratorStatusPending
}

// CanRemove checks if collaborator can be removed
func (cc *ContentCollaborator) CanRemove() bool {
	return cc.Status == CollaboratorStatusActive || cc.Status == CollaboratorStatusInactive
}
