package interfaces

import (
	"context"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
)

// TenantRoleServiceInterface defines operations for configuring tenant roles.
type TenantRoleServiceInterface interface {
	ListPermissions(ctx context.Context) ([]*d.Permission, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*d.TenantRole, error)
	CreateRole(ctx context.Context, tenantID uuid.UUID, payload d.TenantRolePayload) (*d.TenantRole, error)
	UpdateRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, payload d.TenantRolePayload) (*d.TenantRole, error)
	DeleteRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error
	AssignRoleToUser(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) error
}
