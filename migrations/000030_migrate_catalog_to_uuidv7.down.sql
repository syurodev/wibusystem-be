-- Migration: Rollback remaining tables from UUID v7 to UUID v4
-- Description: Revert `identify.organizations` and `catalog` schema tables from uuidv7() to uuidv7()
-- Author: System
-- Created: 2025-12-06

-- Schema: identify
ALTER TABLE identify.organizations ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog
ALTER TABLE catalog.novels ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_volumes ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_chapters ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog (history)
ALTER TABLE catalog.novel_history ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_volume_histories ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_chapter_histories ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog (ownership & reports)
ALTER TABLE catalog.ownership_transfers ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.exclusive_translation_reports ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog (translations & contributions)
ALTER TABLE catalog.novel_synopsis_translations ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.synopsis_translation_contributions ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_chapter_translations ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.translation_contributions ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.translation_history ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog (supporting tables)
ALTER TABLE catalog.genres ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_genres ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.authors ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_authors ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.artists ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_artists ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.translators ALTER COLUMN id SET DEFAULT uuidv7();
ALTER TABLE catalog.novel_translators ALTER COLUMN id SET DEFAULT uuidv7();

-- Schema: catalog (assignments)
ALTER TABLE catalog.novel_organization_assignments ALTER COLUMN id SET DEFAULT uuidv7();
