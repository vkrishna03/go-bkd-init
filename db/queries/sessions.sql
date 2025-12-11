-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE refresh_token = $1 AND expires_at > NOW();

-- name: CreateSession :one
INSERT INTO sessions (user_id, device_id, refresh_token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions WHERE refresh_token = $1;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();
