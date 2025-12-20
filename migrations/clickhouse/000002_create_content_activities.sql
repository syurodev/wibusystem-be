CREATE TABLE IF NOT EXISTS content_activities (
    event_time DateTime DEFAULT now(),
    
    -- Action Type: 1=create, 2=publish, 3=delete
    action_type Enum8('create' = 1, 'publish' = 2, 'delete' = 3),
    
    -- Media Type: 1=novel, 2=manga, 3=anime
    media_type Enum8('novel' = 1, 'manga' = 2, 'anime' = 3),
    media_id UUID,
    
    -- Target Type: 1=media, 2=volume, 3=chapter, 4=episode
    target_type Enum8('media' = 1, 'volume' = 2, 'chapter' = 3, 'episode' = 4),
    target_id UUID,
    
    -- Actor
    user_id UUID,
    org_id Nullable(UUID),
    
    -- KPI Weight (e.g., word count)
    weight Int64 DEFAULT 1
) ENGINE = MergeTree 
PARTITION BY toYYYYMM(event_time) 
ORDER BY (event_time, media_type, user_id)
SETTINGS index_granularity = 8192;
