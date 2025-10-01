-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_tenants_updated_at ON tenants;
DROP FUNCTION IF EXISTS update_tenants_updated_at();

-- Drop constraints
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS check_tenant_status;

-- Drop indexes
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_slug;

-- Drop columns (in reverse order)
ALTER TABLE tenants
DROP COLUMN IF EXISTS updated_at,
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS settings,
DROP COLUMN IF EXISTS description,
DROP COLUMN IF EXISTS slug;
