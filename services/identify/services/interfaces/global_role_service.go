package interfaces

import (
	"context"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
)

// GlobalRoleServiceInterface defines operations for managing global roles.
type GlobalRoleServiceInterface interface {
	ListRoles(ctx context.Context) ([]*d.GlobalRole, error)
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*d.GlobalRole, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}
