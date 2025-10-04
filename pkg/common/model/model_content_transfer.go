package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Transfer status constants
const (
	TransferStatusPending   = "PENDING"   // Waiting for approval
	TransferStatusApproved  = "APPROVED"  // Approved, ready to execute
	TransferStatusCompleted = "COMPLETED" // Transfer completed successfully
	TransferStatusRejected  = "REJECTED"  // Transfer rejected
	TransferStatusCancelled = "CANCELLED" // Transfer cancelled by initiator
	TransferStatusExpired   = "EXPIRED"   // Transfer request expired
)

// Content entity type constants (for ownership transfer - different from ContentType enum)
const (
	ContentEntityNovel   = "NOVEL"   // Refers to a novel entity
	ContentEntityVolume  = "VOLUME"  // Refers to a volume entity
	ContentEntityChapter = "CHAPTER" // Refers to a chapter entity
)

// Owner type constants
const (
	OwnerTypeUser   = "user"
	OwnerTypeTenant = "tenant"
)

// TransferConditions represents additional transfer conditions as JSONB
type TransferConditions map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (tc TransferConditions) Value() (driver.Value, error) {
	if tc == nil {
		return nil, nil
	}
	return json.Marshal(tc)
}

// Scan implements sql.Scanner interface for database retrieval
func (tc *TransferConditions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TransferConditions", value)
	}

	return json.Unmarshal(bytes, tc)
}

// TransferNotes represents additional transfer metadata as JSONB
type TransferNotes map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (tn TransferNotes) Value() (driver.Value, error) {
	if tn == nil {
		return nil, nil
	}
	return json.Marshal(tn)
}

// Scan implements sql.Scanner interface for database retrieval
func (tn *TransferNotes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TransferNotes", value)
	}

	return json.Unmarshal(bytes, tn)
}

// ContentTransfer represents ownership transfer requests
type ContentTransfer struct {
	ID uuid.UUID `json:"id" db:"id"`

	// Content identification
	ContentType string    `json:"content_type" db:"content_type"` // NOVEL, VOLUME, CHAPTER
	ContentID   uuid.UUID `json:"content_id" db:"content_id"`     // ID of the content being transferred

	// Transfer parties
	FromOwnerID   uuid.UUID `json:"from_owner_id" db:"from_owner_id"`     // Current owner (user_id or tenant_id)
	FromOwnerType string    `json:"from_owner_type" db:"from_owner_type"` // 'user' or 'tenant'
	ToOwnerID     uuid.UUID `json:"to_owner_id" db:"to_owner_id"`         // New owner (user_id or tenant_id)
	ToOwnerType   string    `json:"to_owner_type" db:"to_owner_type"`     // 'user' or 'tenant'

	// Transfer details
	NewOwnershipType string `json:"new_ownership_type" db:"new_ownership_type"` // Target ownership type (PERSONAL, TENANT, COLLABORATIVE)
	NewAccessLevel   string `json:"new_access_level" db:"new_access_level"`     // Target access level (PRIVATE, TENANT_ONLY, PUBLIC)
	Status           string `json:"status" db:"status"`                         // Transfer status

	// Workflow tracking
	InitiatedByUserID  uuid.UUID  `json:"initiated_by_user_id" db:"initiated_by_user_id"`    // User who initiated
	ApprovedByUserID   *uuid.UUID `json:"approved_by_user_id,omitempty" db:"approved_by_user_id"`   // User who approved
	CompletedByUserID  *uuid.UUID `json:"completed_by_user_id,omitempty" db:"completed_by_user_id"` // User who executed transfer

	// Additional metadata
	TransferReason *string             `json:"transfer_reason,omitempty" db:"transfer_reason"` // Optional reason
	TransferNotes  *TransferNotes      `json:"transfer_notes,omitempty" db:"transfer_notes"`   // Additional metadata
	Conditions     *TransferConditions `json:"conditions,omitempty" db:"conditions"`           // Transfer conditions (e.g., royalty split)

	// Timestamps
	InitiatedAt time.Time  `json:"initiated_at" db:"initiated_at"`
	ApprovedAt  *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"` // Transfer request expiration
}

// IsActive returns true if transfer is in active state (PENDING or APPROVED)
func (ct *ContentTransfer) IsActive() bool {
	return ct.Status == TransferStatusPending || ct.Status == TransferStatusApproved
}

// CanApprove checks if transfer can be approved
func (ct *ContentTransfer) CanApprove() bool {
	return ct.Status == TransferStatusPending
}

// CanComplete checks if transfer can be completed
func (ct *ContentTransfer) CanComplete() bool {
	return ct.Status == TransferStatusApproved
}

// CanCancel checks if transfer can be cancelled
func (ct *ContentTransfer) CanCancel() bool {
	return ct.Status == TransferStatusPending || ct.Status == TransferStatusApproved
}

// IsExpired checks if transfer has expired
func (ct *ContentTransfer) IsExpired() bool {
	if ct.ExpiresAt == nil {
		return false
	}
	return ct.ExpiresAt.Before(time.Now()) && ct.Status == TransferStatusPending
}
