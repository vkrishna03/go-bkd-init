-- name: GetStreamByID :one
SELECT * FROM streams WHERE id = $1;

-- name: ListUserStreams :many
SELECT * FROM streams WHERE user_id = $1 ORDER BY started_at DESC;

-- name: ListActiveUserStreams :many
SELECT * FROM streams
WHERE user_id = $1 AND status IN ('connecting', 'active', 'paused')
ORDER BY started_at DESC;

-- name: CreateStream :one
INSERT INTO streams (user_id, source_device_id, target_device_id, stream_type, quality)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateStreamStatus :exec
UPDATE streams
SET status = $2
WHERE id = $1;

-- name: UpdateStreamLatency :exec
UPDATE streams
SET latency_ms = $2
WHERE id = $1;

-- name: UpdateStreamConnectionType :exec
UPDATE streams
SET connection_type = $2
WHERE id = $1;

-- name: EndStream :exec
UPDATE streams
SET status = 'ended', ended_at = NOW()
WHERE id = $1;

-- name: CountActiveUserStreams :one
SELECT COUNT(*) FROM streams
WHERE user_id = $1 AND status IN ('connecting', 'active', 'paused');
