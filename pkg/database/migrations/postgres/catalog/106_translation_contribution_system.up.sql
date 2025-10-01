-- Translation Contribution System
-- Enables community-driven translations with review workflow

-- Translation Contribution Requests
-- Stores user-submitted translations for community review and approval
CREATE TABLE translation_contributions (
    id UUID PRIMARY KEY DEFAULT uuidv7(), -- Unique identifier for the contribution

    -- Content reference
    reference_type VARCHAR(20) NOT NULL CHECK (reference_type IN ('novel', 'novel_chapter')), -- Type of content being translated
    reference_id UUID NOT NULL, -- ID of the content (novel or chapter)

    -- Translation content
    title TEXT NOT NULL, -- Translated title
    content JSONB NOT NULL, -- Translated content from Plate editor

    -- Language settings
    source_language VARCHAR(10) NOT NULL, -- Original language code (e.g., 'en', 'vi')
    target_language VARCHAR(10) NOT NULL, -- Target language code for translation
    is_machine_translation BOOLEAN DEFAULT false, -- Flag for machine-generated translations

    -- Contributor information
    user_id UUID NOT NULL, -- User ID from Identity service
    tenant_id UUID, -- Tenant ID if user belongs to a tenant

    -- Review workflow
    status VARCHAR(20) NOT NULL DEFAULT 'pending' -- Current status in review workflow
        CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason TEXT, -- Detailed reason for rejection
    reviewer_id UUID, -- ID of the user who reviewed this contribution

    -- Soft delete
    is_deleted BOOLEAN DEFAULT false, -- Soft delete flag

    -- Community engagement (cached counts)
    upvotes INTEGER DEFAULT 0, -- Number of positive votes
    downvotes INTEGER DEFAULT 0, -- Number of negative votes

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- When contribution was submitted
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Last modification time
);

-- Individual Vote Tracking
-- Tracks individual user votes on translation contributions
CREATE TABLE translation_votes (
    id UUID PRIMARY KEY DEFAULT uuidv7(), -- Unique identifier for the vote
    contribution_id UUID NOT NULL REFERENCES translation_contributions(id) ON DELETE CASCADE, -- Contribution being voted on
    user_id UUID NOT NULL, -- ID of the user casting the vote

    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('upvote', 'downvote')), -- Type of vote
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- When vote was cast

    UNIQUE(contribution_id, user_id) -- One vote per user per contribution
);

-- Indexes for performance
CREATE INDEX idx_translation_contributions_reference ON translation_contributions(reference_type, reference_id);
CREATE INDEX idx_translation_contributions_user ON translation_contributions(user_id);
CREATE INDEX idx_translation_contributions_tenant ON translation_contributions(tenant_id);
CREATE INDEX idx_translation_contributions_status ON translation_contributions(status);
CREATE INDEX idx_translation_contributions_languages ON translation_contributions(source_language, target_language);
CREATE INDEX idx_translation_contributions_created ON translation_contributions(created_at DESC);

CREATE INDEX idx_translation_votes_contribution ON translation_votes(contribution_id);
CREATE INDEX idx_translation_votes_user ON translation_votes(user_id);

-- Function to update vote counts
CREATE OR REPLACE FUNCTION update_translation_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    -- Update vote counts in translation_contributions
    UPDATE translation_contributions
    SET
        upvotes = (
            SELECT COUNT(*)
            FROM translation_votes
            WHERE contribution_id = COALESCE(NEW.contribution_id, OLD.contribution_id)
            AND vote_type = 'upvote'
        ),
        downvotes = (
            SELECT COUNT(*)
            FROM translation_votes
            WHERE contribution_id = COALESCE(NEW.contribution_id, OLD.contribution_id)
            AND vote_type = 'downvote'
        ),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = COALESCE(NEW.contribution_id, OLD.contribution_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update vote counts
CREATE TRIGGER trigger_update_translation_vote_counts
    AFTER INSERT OR UPDATE OR DELETE ON translation_votes
    FOR EACH ROW EXECUTE FUNCTION update_translation_vote_counts();

-- Function to update timestamp on translation_contributions
CREATE OR REPLACE FUNCTION update_translation_contributions_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update timestamp
CREATE TRIGGER trigger_update_translation_contributions_timestamp
    BEFORE UPDATE ON translation_contributions
    FOR EACH ROW EXECUTE FUNCTION update_translation_contributions_timestamp();