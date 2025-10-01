-- Add missing columns to tenants table
ALTER TABLE tenants
ADD COLUMN IF NOT EXISTS slug VARCHAR(50),
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active' NOT NULL,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL;

-- Create unique index on slug for faster lookups and uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug_unique ON tenants(slug) WHERE slug IS NOT NULL;

-- Create index on status for filtering
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- Add check constraint for status (with conditional logic to avoid duplicate)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'check_tenant_status'
        AND table_name = 'tenants'
    ) THEN
        ALTER TABLE tenants ADD CONSTRAINT check_tenant_status
        CHECK (status IN ('active', 'suspended', 'inactive'));
    END IF;
END $$;

-- Create trigger to update updated_at on tenant updates
CREATE OR REPLACE FUNCTION update_tenants_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_tenants_updated_at ON tenants;
CREATE TRIGGER trigger_update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_tenants_updated_at();
