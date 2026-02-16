-- name: CreatePatient :execresult
INSERT INTO lael_patients (
    name, mobile, opd_id, age, sex,
    address_locality, address_city, address_state, address_pincode,
    visit_number
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPatientByID :one
SELECT * FROM lael_patients WHERE id = ?;

-- name: GetPatientByMobile :one
SELECT * FROM lael_patients
WHERE mobile = ?
ORDER BY created_on DESC
LIMIT 1;

-- name: GetPatientByOPDID :one
SELECT * FROM lael_patients WHERE opd_id = ?;

-- name: SearchPatientsByMobile :many
SELECT * FROM lael_patients
WHERE mobile LIKE ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: SearchPatientsByName :many
SELECT * FROM lael_patients
WHERE name LIKE ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: GetTodayPatients :many
SELECT * FROM lael_patients
WHERE DATE(created_on) = CURDATE()
ORDER BY created_on DESC;

-- name: GetPatientsByDateRange :many
SELECT * FROM lael_patients
WHERE created_on BETWEEN ? AND ?
ORDER BY created_on DESC
LIMIT ? OFFSET ?;

-- name: UpdatePatientVisitNumber :exec
UPDATE lael_patients
SET visit_number = visit_number + 1, updated_on = NOW()
WHERE mobile = ?;

-- name: GetAllVisitsByMobile :many
SELECT * FROM lael_patients
WHERE mobile = ?
ORDER BY visit_number ASC;

-- name: GetLatestVisitByOPDID :one
SELECT * FROM lael_patients
WHERE opd_id = ?
ORDER BY visit_number DESC;
