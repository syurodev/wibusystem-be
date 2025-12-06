-- Migration 000002: Add Analytics Metadata
-- Description: Adds denormalized columns for comprehensive analytics and creates materialized views

-- 1. Add new columns to view_events table
ALTER TABLE view_events
    -- Content Metadata
    ADD COLUMN IF NOT EXISTS author_ids Array(UUID),
    ADD COLUMN IF NOT EXISTS genre_ids Array(UUID),
    ADD COLUMN IF NOT EXISTS artist_ids Array(UUID),
    ADD COLUMN IF NOT EXISTS tag_ids Array(UUID),
    ADD COLUMN IF NOT EXISTS group_id Nullable(UUID),
    ADD COLUMN IF NOT EXISTS owner_id Nullable(UUID),
    ADD COLUMN IF NOT EXISTS studio_id Nullable(UUID),
    ADD COLUMN IF NOT EXISTS original_language LowCardinality(String),

    -- User Context
    ADD COLUMN IF NOT EXISTS platform LowCardinality(String), -- 'web', 'android', 'ios'
    ADD COLUMN IF NOT EXISTS os LowCardinality(String),       -- 'windows', 'macos', 'android', 'ios'
    ADD COLUMN IF NOT EXISTS browser LowCardinality(String),  -- 'chrome', 'safari', 'firefox'
    ADD COLUMN IF NOT EXISTS country_code LowCardinality(String), -- ISO 3166-1 alpha-2
    ADD COLUMN IF NOT EXISTS city String,
    ADD COLUMN IF NOT EXISTS referrer String,

    -- User Status
    ADD COLUMN IF NOT EXISTS is_premium Boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS user_role LowCardinality(String) DEFAULT 'guest';

-- 2. Create Materialized Views for Aggregation

-- 2.1 View Stats by Author
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_author
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, author_id)
AS SELECT
    toDate(event_time) AS event_date,
    author_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
ARRAY JOIN author_ids AS author_id
GROUP BY event_date, author_id;

-- 2.2 View Stats by Genre
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_genre
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, genre_id)
AS SELECT
    toDate(event_time) AS event_date,
    genre_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
ARRAY JOIN genre_ids AS genre_id
GROUP BY event_date, genre_id;

-- 2.3 View Stats by Artist
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_artist
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, artist_id)
AS SELECT
    toDate(event_time) AS event_date,
    artist_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
ARRAY JOIN artist_ids AS artist_id
GROUP BY event_date, artist_id;

-- 2.4 View Stats by Group (Translation Group / Organization)
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_group
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, group_id)
AS SELECT
    toDate(event_time) AS event_date,
    assumeNotNull(group_id) AS group_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
WHERE group_id IS NOT NULL
GROUP BY event_date, group_id;

-- 2.5 View Stats by Studio (For Anime)
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_studio
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, studio_id)
AS SELECT
    toDate(event_time) AS event_date,
    assumeNotNull(studio_id) AS studio_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
WHERE studio_id IS NOT NULL
GROUP BY event_date, studio_id;

-- 2.6 View Stats by Tag (Future Proofing)
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_tag
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, tag_id)
AS SELECT
    toDate(event_time) AS event_date,
    tag_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
ARRAY JOIN tag_ids AS tag_id
GROUP BY event_date, tag_id;

-- 2.7 View Stats by Geography
CREATE MATERIALIZED VIEW IF NOT EXISTS view_stats_by_geo
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, country_code)
AS SELECT
    toDate(event_time) AS event_date,
    country_code,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users
FROM view_events
GROUP BY event_date, country_code;
