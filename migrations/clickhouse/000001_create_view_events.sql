-- ClickHouse schema for view tracking analytics
-- This file creates tables for storing view events from novels, mangas, and anime

-- Drop existing tables if they exist (for development)
DROP TABLE IF EXISTS view_events_daily;
DROP TABLE IF EXISTS view_events;

-- Main events table for raw view tracking data
-- Uses MergeTree engine for efficient storage and querying
CREATE TABLE IF NOT EXISTS view_events (
    event_time DateTime DEFAULT now(),
    media_type Enum8('novel' = 1, 'manga' = 2, 'anime' = 3),
    media_id UUID,
    unit_id UUID, -- chapter_id or episode_id
    user_id Nullable(UUID),
    ip_address String,
    view_count UInt32 DEFAULT 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(event_time)
ORDER BY (media_type, media_id, event_time)
SETTINGS index_granularity = 8192;

-- Create indexes for common queries
CREATE INDEX idx_user_id ON view_events(user_id) TYPE bloom_filter GRANULARITY 4;
CREATE INDEX idx_media_id ON view_events(media_id) TYPE bloom_filter GRANULARITY 4;

-- Materialized view for daily aggregation
-- This pre-aggregates data by day for faster analytics queries
CREATE MATERIALIZED VIEW IF NOT EXISTS view_events_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, media_type, media_id)
AS SELECT
    toDate(event_time) AS event_date,
    media_type,
    media_id,
    sum(view_count) AS total_views,
    uniqExact(user_id) AS unique_users,
    uniqExact(ip_address) AS unique_ips
FROM view_events
GROUP BY event_date, media_type, media_id;

-- Verify tables created successfully
SELECT 'view_events table created' AS status;
SELECT 'view_events_daily materialized view created' AS status;
