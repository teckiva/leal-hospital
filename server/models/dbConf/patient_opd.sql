-- name: CreatePatientOPD :execresult
INSERT INTO patient_opd (
    patient_id, doctor_id, symptoms, prescription,
    medicines, future_suggestion, template_version
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetPatientOPDByID :one
SELECT * FROM patient_opd WHERE id = ?;

-- name: GetPatientOPDsByPatientID :many
SELECT * FROM patient_opd
WHERE patient_id = ?
ORDER BY created_on DESC;

-- name: GetLatestPatientOPD :one
SELECT * FROM patient_opd
WHERE patient_id = ?
ORDER BY created_on DESC
LIMIT 1;

-- name: GetPatientOPDsByDoctor :many
SELECT po.*, lp.name as patient_name, lp.mobile as patient_mobile
FROM patient_opd po
JOIN lael_patients lp ON po.patient_id = lp.id
WHERE po.doctor_id = ?
ORDER BY po.created_on DESC
LIMIT ? OFFSET ?;
