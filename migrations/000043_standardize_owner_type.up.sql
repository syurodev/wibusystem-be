-- Migration: Standardize owner_type values to 'user' or 'org'
-- Changes: 'tenant', 'group', 'organization' -> 'org'

-- Update existing records
UPDATE catalog.novels 
SET owner_type = 'org' 
WHERE owner_type IN ('tenant', 'group', 'organization');

-- Add check constraint to prevent invalid values
ALTER TABLE catalog.novels 
DROP CONSTRAINT IF EXISTS novels_owner_type_check;

ALTER TABLE catalog.novels 
ADD CONSTRAINT novels_owner_type_check 
CHECK (owner_type IN ('user', 'org'));
