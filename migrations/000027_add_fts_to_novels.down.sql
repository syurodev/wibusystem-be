DROP INDEX IF EXISTS catalog.idx_novels_title_trgm;
DROP INDEX IF EXISTS catalog.idx_novels_original_title_trgm;
-- We generally don't drop the extension in down migration as other tables might use it,
-- but for strict reversal:
-- DROP EXTENSION IF EXISTS pg_trgm;
