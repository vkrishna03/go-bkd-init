-- name: GetUserSettings :one
SELECT * FROM user_settings WHERE user_id = $1;

-- name: CreateUserSettings :one
INSERT INTO user_settings (user_id)
VALUES ($1)
RETURNING *;

-- name: UpdateUserSettings :one
UPDATE user_settings
SET max_devices = COALESCE($2, max_devices),
    max_concurrent_streams = COALESCE($3, max_concurrent_streams),
    default_stream_quality = COALESCE($4, default_stream_quality),
    default_stream_type = COALESCE($5, default_stream_type),
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: IncrementStreamCount :exec
UPDATE user_settings
SET total_streams_count = total_streams_count + 1,
    updated_at = NOW()
WHERE user_id = $1;

-- name: IncrementStreamMinutes :exec
UPDATE user_settings
SET total_stream_minutes = total_stream_minutes + $2,
    updated_at = NOW()
WHERE user_id = $1;
