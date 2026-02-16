-- name: CreateUser :execresult
INSERT INTO lael_users (
    name, mobile, email, designation, is_admin, is_approved, password_hash
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetUserByID :one
SELECT * FROM lael_users WHERE id = ?;

-- name: GetUserByMobile :one
SELECT * FROM lael_users WHERE mobile = ?;

-- name: GetUserByEmail :one
SELECT * FROM lael_users WHERE email = ?;

-- name: ListPendingStaffApprovals :many
SELECT * FROM lael_users
WHERE is_admin = FALSE AND is_approved = FALSE
ORDER BY created_on DESC;

-- name: ListStaffByStatus :many
SELECT * FROM lael_users
WHERE is_admin = FALSE AND status = ?
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: UpdateUserStatus :exec
UPDATE lael_users
SET status = ?, updated_on = NOW()
WHERE id = ?;

-- name: ApproveStaff :exec
UPDATE lael_users
SET is_approved = TRUE, approved_by = ?, updated_on = NOW()
WHERE id = ?;

-- name: UpdateUserPassword :exec
UPDATE lael_users
SET password_hash = ?, updated_on = NOW()
WHERE id = ?;

-- name: UpdateLastLogin :exec
UPDATE lael_users
SET last_login_at = NOW(), updated_on = NOW()
WHERE id = ?;
