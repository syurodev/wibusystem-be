CREATE TABLE view_events
(
    event_time DateTime DEFAULT now( ),
    media_type Enum8('novel' = 1, 'manga' = 2, 'anime' = 3),
    media_id   UUID,
    unit_id    UUID,
    user_id    Nullable(UUID),
    ip_address String,
    view_count UInt32   DEFAULT 1
)
    ENGINE = MergeTree PARTITION BY toYYYYMM( event_time ) ORDER BY ( media_type, media_id, event_time ) SETTINGS index_granularity = 8192;

CREATE MATERIALIZED VIEW clickhouse_dev.view_events_daily
            ( `event_date` Date,
              `media_type` Enum8('novel' = 1, 'manga' = 2, 'anime' = 3),
              `media_id` UUID,
              `total_views` UInt64,
              `unique_users` UInt64,
              `unique_ips` UInt64
                ) ENGINE = SummingMergeTree PARTITION BY toYYYYMM( event_date ) ORDER BY ( event_date, media_type, media_id ) SETTINGS index_granularity = 8192
AS
SELECT toDate( event_time ) AS event_date, media_type, media_id, sum( view_count ) AS total_views,
    uniqExact( user_id ) AS unique_users, uniqExact( ip_address ) AS unique_ips
FROM clickhouse_dev.view_events
GROUP BY event_date, media_type, media_id;

