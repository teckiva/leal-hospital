-- name: CreateSession :execresult
INSERT INTO lael_sessions (
    user_id, session_token, session_expiry, device_info, ip_address
) VALUES (?, ?, ?, ?, ?);

-- name: GetSessionByToken :one
SELECT * FROM lael_sessions WHERE session_token = ?;

-- name: UpdateSessionActivity :exec
UPDATE lael_sessions
SET last_active_at = NOW()
WHERE session_token = ?;

-- name: DeleteSession :exec
DELETE FROM lael_sessions WHERE session_token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM lael_sessions WHERE session_expiry < NOW();

-- name: GetUserSessions :many
SELECT * FROM lael_sessions
WHERE user_id = ?
ORDER BY last_active_at DESC;
