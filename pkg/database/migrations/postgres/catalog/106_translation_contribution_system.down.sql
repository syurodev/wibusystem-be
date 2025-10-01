-- Drop translation contribution system

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_translation_contributions_timestamp ON translation_contributions;
DROP TRIGGER IF EXISTS trigger_update_translation_vote_counts ON translation_votes;

-- Drop functions
DROP FUNCTION IF EXISTS update_translation_contributions_timestamp();
DROP FUNCTION IF EXISTS update_translation_vote_counts();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS translation_votes;
DROP TABLE IF EXISTS translation_contributions;