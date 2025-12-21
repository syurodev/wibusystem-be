-- Rollback: Recreate version triggers for novels and chapters
-- This restores automatic version increment on UPDATE

-- Recreate function if it was dropped
CREATE OR REPLACE FUNCTION catalog.increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate triggers
CREATE TRIGGER trg_novels_version
    BEFORE UPDATE ON catalog.novels
    FOR EACH ROW
    EXECUTE FUNCTION catalog.increment_version();

CREATE TRIGGER trg_chapters_version
    BEFORE UPDATE ON catalog.novel_chapters
    FOR EACH ROW
    EXECUTE FUNCTION catalog.increment_version();
