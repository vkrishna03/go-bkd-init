-- name: GetDeviceByID :one
SELECT * FROM devices WHERE id = $1;

-- name: GetDeviceByUserAndDeviceID :one
SELECT * FROM devices WHERE user_id = $1 AND device_id = $2;

-- name: ListUserDevices :many
SELECT * FROM devices WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListOnlineUserDevices :many
SELECT * FROM devices WHERE user_id = $1 AND is_online = TRUE ORDER BY last_seen DESC;

-- name: CreateDevice :one
INSERT INTO devices (user_id, device_id, device_name, device_type, has_camera, has_microphone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateDevice :one
UPDATE devices
SET device_name = COALESCE($2, device_name),
    has_camera = COALESCE($3, has_camera),
    has_microphone = COALESCE($4, has_microphone)
WHERE id = $1
RETURNING *;

-- name: UpdateDeviceOnlineStatus :exec
UPDATE devices
SET is_online = $2, last_seen = NOW()
WHERE id = $1;

-- name: UpdateDeviceLastSeen :exec
UPDATE devices
SET last_seen = NOW()
WHERE id = $1;

-- name: DeleteDevice :exec
DELETE FROM devices WHERE id = $1;

-- name: CountUserDevices :one
SELECT COUNT(*) FROM devices WHERE user_id = $1;
