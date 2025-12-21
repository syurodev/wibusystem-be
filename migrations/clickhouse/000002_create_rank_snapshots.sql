-- ClickHouse schema for rank snapshots
-- Stores rankings for genres, creators, orgs, and media at specific points in time

-- Enum values for period_type
-- 'week' = 1
-- 'month' = 2
-- 'year' = 3

-- Enum values for entity_type
-- 'genre' = 1
-- 'creator' = 2
-- 'org' = 3
-- 'novel' = 4
-- 'manga' = 5
-- 'anime' = 6

CREATE TABLE IF NOT EXISTS rank_snapshots (
    snapshot_date Date,
    period_type Enum8('week' = 1, 'month' = 2, 'year' = 3),
    entity_type Enum8('genre' = 1, 'creator' = 2, 'org' = 3, 'novel' = 4, 'manga' = 5, 'anime' = 6),
    entity_id UUID,
    rank UInt32,
    total_views UInt64,
    unique_users UInt64
) ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(snapshot_date)
ORDER BY (period_type, entity_type, snapshot_date, rank)
TTL snapshot_date + INTERVAL 13 MONTH;
