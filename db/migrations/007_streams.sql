CREATE TABLE streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    target_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    stream_type stream_type NOT NULL,
    status stream_status DEFAULT 'connecting',
    connection_type connection_type,
    quality stream_quality DEFAULT 'auto',
    latency_ms INT,
    started_at TIMESTAMP DEFAULT NOW(),
    ended_at TIMESTAMP
);

CREATE INDEX idx_streams_user_id ON streams(user_id);
CREATE INDEX idx_streams_source_device_id ON streams(source_device_id);
CREATE INDEX idx_streams_status ON streams(status);
CREATE INDEX idx_streams_started_at ON streams(started_at);
