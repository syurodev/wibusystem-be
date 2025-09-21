// Package repositories contains data access interfaces and Postgres-backed
// implementations for users, credentials, tenants, and memberships.
package repositories

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repositories aggregates all repository interfaces for convenient injection.
type Repositories struct {
	User       UserRepository
	Credential CredentialRepository
	Tenant     TenantRepository
	Membership MembershipRepository
}

// NewRepositories creates and returns all repositories using the given
// pgx connection pool.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:       NewUserRepository(pool),
		Credential: NewCredentialRepository(pool),
		Tenant:     NewTenantRepository(pool),
		Membership: NewMembershipRepository(pool),
	}
}
