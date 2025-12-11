CREATE TYPE device_type AS ENUM ('phone', 'tablet', 'desktop');
CREATE TYPE stream_type AS ENUM ('video', 'audio', 'both');
CREATE TYPE stream_status AS ENUM ('connecting', 'active', 'paused', 'ended', 'failed');
CREATE TYPE connection_type AS ENUM ('p2p', 'relay');
CREATE TYPE stream_quality AS ENUM ('low', 'medium', 'high', 'auto');
