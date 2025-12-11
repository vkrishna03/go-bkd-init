CREATE TABLE user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Limits
    max_devices INT DEFAULT 5,
    max_concurrent_streams INT DEFAULT 2,

    -- Usage tracking
    total_stream_minutes INT DEFAULT 0,
    total_streams_count INT DEFAULT 0,

    -- Preferences
    default_stream_quality stream_quality DEFAULT 'auto',
    default_stream_type stream_type DEFAULT 'both',

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
