-- name: CreateOTP :execresult
INSERT INTO lael_otp (
    mobile, email, otp, expiry, otp_type, retry_count
) VALUES (?, ?, ?, ?, ?, 0);

-- name: GetLatestOTP :one
SELECT * FROM lael_otp
WHERE mobile = ? AND otp_type = ? AND is_validated = FALSE
ORDER BY created_on DESC
LIMIT 1;

-- name: GetLatestOTPByEmail :one
SELECT * FROM lael_otp
WHERE email = ? AND otp_type = ? AND is_validated = FALSE
ORDER BY created_on DESC
LIMIT 1;

-- name: ValidateOTP :exec
UPDATE lael_otp
SET is_validated = TRUE, updated_on = NOW()
WHERE id = ?;

-- name: IncrementRetryCount :exec
UPDATE lael_otp
SET retry_count = retry_count + 1, updated_on = NOW()
WHERE id = ?;

-- name: DeleteExpiredOTP :exec
DELETE FROM lael_otp
WHERE expiry < NOW();
