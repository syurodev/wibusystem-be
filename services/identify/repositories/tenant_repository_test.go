package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	m "wibusystem/pkg/common/model"
)

// TestTenantRepository_Create tests tenant creation
func TestTenantRepository_Create(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		tenant      *m.Tenant
		expectError bool
	}{
		{
			name: "create tenant with all fields",
			tenant: &m.Tenant{
				Name:        "Test Company",
				Slug:        "test-company",
				Description: stringPtr("A test company"),
				Settings:    &m.TenantSettings{"theme": "dark"},
				Status:      m.TenantStatusActive,
			},
			expectError: false,
		},
		{
			name: "create tenant with minimal fields",
			tenant: &m.Tenant{
				Name:   "Minimal Company",
				Status: m.TenantStatusActive,
			},
			expectError: false,
		},
		{
			name: "create tenant with duplicate slug should fail",
			tenant: &m.Tenant{
				Name:   "Duplicate Slug Company",
				Slug:   "test-company", // Same as first test
				Status: m.TenantStatusActive,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.tenant)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tt.tenant.ID)
			assert.NotZero(t, tt.tenant.CreatedAt)
			assert.NotZero(t, tt.tenant.UpdatedAt)
		})
	}
}

// TestTenantRepository_GetByID tests tenant retrieval by ID
func TestTenantRepository_GetByID(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create a test tenant first
	tenant := &m.Tenant{
		Name:        "Get By ID Test",
		Slug:        "get-by-id-test",
		Description: stringPtr("Test description"),
		Settings:    &m.TenantSettings{"test": true},
		Status:      m.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Test getting the tenant
	retrieved, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)

	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)
	assert.Equal(t, tenant.Slug, retrieved.Slug)
	assert.Equal(t, *tenant.Description, *retrieved.Description)
	assert.Equal(t, tenant.Status, retrieved.Status)

	// Test settings
	if tenant.Settings != nil && retrieved.Settings != nil {
		assert.Equal(t, (*tenant.Settings)["test"], (*retrieved.Settings)["test"])
	}

	// Test non-existent tenant
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

// TestTenantRepository_GetBySlug tests tenant retrieval by slug
func TestTenantRepository_GetBySlug(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create a test tenant
	tenant := &m.Tenant{
		Name:   "Slug Test",
		Slug:   "slug-test-unique",
		Status: m.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Test getting by slug
	retrieved, err := repo.GetBySlug(ctx, tenant.Slug)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Slug, retrieved.Slug)

	// Test non-existent slug
	_, err = repo.GetBySlug(ctx, "non-existent-slug")
	assert.Error(t, err)
}

// TestTenantRepository_SlugExists tests slug existence check
func TestTenantRepository_SlugExists(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create a tenant with slug
	tenant := &m.Tenant{
		Name:   "Slug Exists Test",
		Slug:   "slug-exists-test",
		Status: m.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Test existing slug
	exists, err := repo.SlugExists(ctx, tenant.Slug)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test non-existing slug
	exists, err = repo.SlugExists(ctx, "non-existent-slug")
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestTenantRepository_Update tests tenant updates
func TestTenantRepository_Update(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create a test tenant
	tenant := &m.Tenant{
		Name:        "Update Test",
		Slug:        "update-test",
		Description: stringPtr("Original description"),
		Status:      m.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))
	originalUpdatedAt := tenant.UpdatedAt

	// Update the tenant
	tenant.Name = "Updated Name"
	tenant.Description = stringPtr("Updated description")
	tenant.Settings = &m.TenantSettings{"updated": true}

	err := repo.Update(ctx, tenant)
	require.NoError(t, err)

	// Verify updated_at changed
	assert.True(t, tenant.UpdatedAt.After(originalUpdatedAt))

	// Retrieve and verify changes
	retrieved, err := repo.GetByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "Updated description", *retrieved.Description)
}

// TestTenantRepository_List tests listing tenants with pagination
func TestTenantRepository_List(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create multiple test tenants
	for i := 0; i < 5; i++ {
		tenant := &m.Tenant{
			Name:   fmt.Sprintf("List Test %d", i),
			Status: m.TenantStatusActive,
		}
		require.NoError(t, repo.Create(ctx, tenant))
	}

	// Test pagination
	tenants, total, err := repo.List(ctx, 3, 0) // First 3 tenants
	require.NoError(t, err)
	assert.LessOrEqual(t, len(tenants), 3)
	assert.GreaterOrEqual(t, total, int64(5))

	// Test second page
	tenants2, total2, err := repo.List(ctx, 3, 3) // Next 3 tenants
	require.NoError(t, err)
	assert.Equal(t, total, total2) // Total should be same
}

// TestTenantRepository_Delete tests tenant deletion
func TestTenantRepository_Delete(t *testing.T) {
	repo := getTenantTestRepo(t)
	ctx := context.Background()

	// Create a test tenant
	tenant := &m.Tenant{
		Name:   "Delete Test",
		Status: m.TenantStatusActive,
	}
	require.NoError(t, repo.Create(ctx, tenant))

	// Delete the tenant
	err := repo.Delete(ctx, tenant.ID)
	require.NoError(t, err)

	// Verify tenant is deleted
	_, err = repo.GetByID(ctx, tenant.ID)
	assert.Error(t, err)

	// Test deleting non-existent tenant
	err = repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// getTenantTestRepo returns a tenant repository for testing
// This would need to be implemented based on your test setup
func getTenantTestRepo(t *testing.T) TenantRepository {
	// This is a placeholder - you would need to implement this
	// based on your test database setup
	t.Skip("Test database setup not implemented yet")
	return nil
}
