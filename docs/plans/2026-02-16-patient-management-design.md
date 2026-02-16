# Patient Management Module - Design Document

**Date**: 2026-02-16
**Module**: Patient Management (PRD Option 1)
**Author**: Claude (with user approval)

## Overview

This document describes the design for the Patient Management Module, which handles patient registration (new and revisit) for the Lael Hospital system. The module follows the KB-PAY architectural patterns established in the authentication module.

## Approach: Two-Stage Registration Flow

**Why this approach?**
- Clean separation of concerns between lookup and registration
- Matches the UI workflow (phone lookup → patient details display → revisit button)
- Business logic (data merge) stays on server
- Easy to test each stage independently
- Follows KB-PAY patterns cleanly

**Flow:**
1. **Stage 1**: Phone lookup endpoint returns all visit history if patient exists
2. **Stage 2**: Registration endpoint handles both new patients and revisits with server-side data merge

---

## API Contracts

### Endpoint 1: Patient Lookup

```
GET /api/patients/lookup?mobile={phone_number}
```

**Response (existing patient with visit history):**
```json
{
  "code": "SUCCESS",
  "msg": "Patient lookup successful",
  "model": {
    "mobile": "9876543210",
    "opdId": "LAELosaQGi2zM2d7wT8xYYuarRxptLu0Kq",
    "visits": [
      {
        "id": 123,
        "name": "Rajesh Kumar",
        "age": 45,
        "sex": "male",
        "address": {
          "locality": "Swami Vivekananda Nagar",
          "city": "Kushinagar",
          "state": "Uttar Pradesh",
          "pincode": "274401"
        },
        "visitNumber": 1,
        "createdOn": "2026-01-15T10:30:00Z"
      },
      {
        "id": 124,
        "name": "Rajesh Kumar",
        "age": 46,
        "sex": "male",
        "address": {
          "locality": "Swami Vivekananda Nagar",
          "city": "Kushinagar",
          "state": "Uttar Pradesh",
          "pincode": "274401"
        },
        "visitNumber": 2,
        "createdOn": "2026-02-10T14:20:00Z"
      }
    ],
    "latestVisit": {
      "visitNumber": 2,
      "age": 46,
      "name": "Rajesh Kumar",
      "sex": "male",
      "address": {
        "locality": "Swami Vivekananda Nagar",
        "city": "Kushinagar",
        "state": "Uttar Pradesh",
        "pincode": "274401"
      }
    }
  }
}
```

**Response (new patient - ERROR):**
```json
{
  "code": "499",
  "msg": "FAILED",
  "model": {
    "errorCode": "4000",
    "message": "Patient not found",
    "displayMessage": "Patient record not found."
  }
}
```

### Endpoint 2: Patient Registration

```
POST /api/patients/register
```

**Request Body:**
```json
{
  "mobile": "9876543210",
  "opdId": "LAELosaQGi2zM2d7wT8xYYuarRxptLu0Kq",  // null for new patients
  "name": "Rajesh Kumar",     // Required for new, optional for revisit
  "age": 45,                  // Required for new, optional for revisit
  "sex": "male",              // Required for new, optional for revisit
  "address": {
    "locality": "...",        // All optional, falls back to previous
    "city": "...",
    "state": "...",
    "pincode": "..."
  }
}
```

**Response:**
```json
{
  "code": "SUCCESS",
  "msg": "Patient registered successfully",
  "model": {
    "id": 124,
    "opdId": "LAELosaQGi2zM2d7wT8xYYuarRxptLu0Kq",
    "visitNumber": 3
  }
}
```

---

## OPD ID Generation

**Format**: `LAEL{base62_encoded_uuid7}`
**Example**: `LAELosaQGi2zM2d7wT8xYYuarRxptLu0Kq`

**Generation Method:**
- Generate UUID v7 (128-bit time-ordered UUID)
- Encode the raw bytes using base62 (alphanumeric: a-z, A-Z, 0-9)
- Prefix with "LAEL"

**Implementation:**
```go
import (
    "github.com/google/uuid"
    "github.com/jxskiss/base62"
)

func generateOPDID() string {
    uuid7 := uuid.NewV7()
    base62Str := base62.Encode(uuid7[:])
    return "LAEL" + base62Str
}
```

**Properties:**
- Length: ~30-36 characters
- Globally unique (UUID v7)
- Time-ordered (sortable)
- No collision check needed

---

## Data Merge Logic (Revisits)

**Server-side merge strategy:**
- Mobile and OPDId remain constant (never change)
- Visit number increments automatically
- For each optional field: use new value if provided (non-empty), else fallback to previous

**Implementation:**
```go
func (s *PatientSvc) mergePatientData(newData, previousData PatientData) PatientData {
    merged := PatientData{}

    // Fields that don't change
    merged.Mobile = previousData.Mobile
    merged.OPDId = previousData.OPDId
    merged.VisitNumber = previousData.VisitNumber + 1

    // Mergeable fields
    merged.Name = newData.Name
    if merged.Name == "" {
        merged.Name = previousData.Name
    }

    merged.Age = newData.Age
    if merged.Age == 0 {
        merged.Age = previousData.Age
    }

    merged.Sex = newData.Sex
    if merged.Sex == "" {
        merged.Sex = previousData.Sex
    }

    // Address fields: individual field-level merge
    merged.Address.Locality = newData.Address.Locality
    if merged.Address.Locality == "" {
        merged.Address.Locality = previousData.Address.Locality
    }

    merged.Address.City = newData.Address.City
    if merged.Address.City == "" {
        merged.Address.City = previousData.Address.City
    }

    merged.Address.State = newData.Address.State
    if merged.Address.State == "" {
        merged.Address.State = previousData.Address.State
    }

    merged.Address.Pincode = newData.Address.Pincode
    if merged.Address.Pincode == "" {
        merged.Address.Pincode = previousData.Address.Pincode
    }

    return merged
}
```

---

## Error Handling (KB-PAY Pattern)

### Service Layer Returns

**Custom error struct:**
```go
type ServiceError struct {
    Code string
    Type string
}

func HandleServiceErr(code string, errType string) *ServiceError {
    return &ServiceError{
        Code: code,
        Type: errType,
    }
}
```

**Service method signature:**
```go
func (s *PatientSvc) LookupPatient(ctx context.Context, mobile string) (*PatientLookupResponse, *ServiceError)
```

**Error return example:**
```go
if !isValidMobile(mobile) {
    logger.E(ctx, functionName, "Invalid mobile number format")
    return nil, HandleServiceErr("4003", "INTERNAL")
}

// Success
return response, &ServiceError{Code: status.OK}
```

### Handler Layer Processing

**Pattern:**
```go
import (
    "github.com/leal-hospital/server/utils/status"
    "github.com/leal-hospital/server/medierror"
)

func (h *PatientHandler) LookupPatient(c *gin.Context) {
    const functionName = "handler.patient.LookupPatient"

    registry := medierror.GetErrorRegistry(c)
    mobile := c.Query("mobile")

    resp, errCode := h.patientSvc.LookupPatient(c, mobile)

    if errCode.Code != status.OK {
        appErr := medierror.NewFromRegistry(medierror.ErrorCode(errCode.Code), nil)
        logger.E(c, functionName, "error_in_patient_lookup")
        c.JSON(200, appErr.ToResponse())
        return
    }

    response := utils.ResponseWithModel("SUCCESS", "Patient lookup successful", resp)
    c.JSON(200, response)
}
```

### Error Codes

Add to `server/config/errors.yaml`:

```yaml
"4003":
  message: Invalid mobile number format
  display_message: Please enter a valid 10-digit mobile number.
  decline_type: "BD"
  source: PATIENT_SERVICE

"4004":
  message: Missing required fields
  display_message: Please provide all required patient information.
  decline_type: "BD"
  source: PATIENT_SERVICE

"4005":
  message: Invalid OPD ID
  display_message: The OPD ID provided is invalid or does not match the patient.
  decline_type: "BD"
  source: PATIENT_SERVICE
```

**Existing codes:**
- `4000`: Patient not found
- `4001`: Invalid patient data
- `4002`: Patient already registered today
- `1005`: Database operation failed

---

## Validation Rules

### Lookup Endpoint
- Mobile: 10 digits, numeric only → Error code `4003`
- Database error → Error code `1005`
- Patient not found → Error code `4000`

### Register Endpoint

**New Patient:**
- Mobile, Name, Age, Sex are required → Error code `4004`
- Mobile: 10 digits, numeric → Error code `4003`
- Age: 1-150 range → Error code `4001`
- Sex: must be "male", "female", or "other" → Error code `4001`

**Revisit:**
- Mobile and OPDId are required → Error code `4004`
- Verify OPDId exists and matches mobile → Error code `4005`
- Other fields optional (merge with previous)

---

## Database & SQLC Queries

### Existing Schema

Table `lael_patients` already exists with:
- id, name, mobile, opd_id, age, sex
- address_locality, address_city, address_state, address_pincode
- visit_number, created_on, updated_on
- Indexes: mobile, opd_id, created_on

### Existing Queries

In `server/models/dbConf/patients.sql`:

```sql
-- name: CreatePatient :execresult
INSERT INTO lael_patients (
    name, mobile, opd_id, age, sex,
    address_locality, address_city, address_state, address_pincode,
    visit_number
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPatientByMobile :one
SELECT * FROM lael_patients
WHERE mobile = ?
ORDER BY created_on DESC;

-- name: GetPatientByOPDID :one
SELECT * FROM lael_patients WHERE opd_id = ?;
```

### New Queries Needed

Add to `server/models/dbConf/patients.sql`:

```sql
-- name: GetAllVisitsByMobile :many
SELECT * FROM lael_patients
WHERE mobile = ?
ORDER BY visit_number ASC;

-- name: GetLatestVisitByOPDID :one
SELECT * FROM lael_patients
WHERE opd_id = ?
ORDER BY visit_number DESC;
```

After adding queries, regenerate:
```bash
cd server/models/dbConf
sqlc generate
```

---

## Architecture: Three-Layer Pattern

### 1. Persistence Layer

**File**: `server/persistence/patient/driver.go`
```go
type PatientPersistenceDriver interface {
    WithConn(ctx context.Context, conn dbUtils.DBConnInterface) PatientPersistenceDriver
    GetAllVisitsByMobile(ctx context.Context, mobile string) ([]db.LaelPatient, error)
    GetLatestVisitByOPDID(ctx context.Context, opdId string) (*db.LaelPatient, error)
    CreatePatient(ctx context.Context, patient *domain.Patient) (int64, error)
}
```

**File**: `server/persistence/patient/patient.go`
```go
type PatientPersistence struct {
    Queries   *db.Queries
    DBManager dbUtils.DBManagerInterface
}

func NewPatientPersistence(queries *db.Queries, dbManager dbUtils.DBManagerInterface) PatientPersistenceDriver {
    return &PatientPersistence{
        Queries:   queries,
        DBManager: dbManager,
    }
}

func (p *PatientPersistence) WithConn(ctx context.Context, conn dbUtils.DBConnInterface) PatientPersistenceDriver {
    return &PatientPersistence{
        Queries:   db.New(conn.GetConn(ctx)),
        DBManager: p.DBManager,
    }
}
```

### 2. Service Layer

**File**: `server/services/patient/driver.go`
```go
type PatientSvcDriver interface {
    LookupPatient(ctx context.Context, mobile string) (*LookupResponse, *ServiceError)
    RegisterPatient(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *ServiceError)
}
```

**File**: `server/services/patient/types.go`
```go
type ServiceError struct {
    Code string
    Type string
}

func HandleServiceErr(code string, errType string) *ServiceError {
    return &ServiceError{Code: code, Type: errType}
}
```

**File**: `server/services/patient/patient.go`
```go
type PatientSvc struct {
    Persistence PatientPersistenceDriver
}

func NewPatientSvc(persistence PatientPersistenceDriver) PatientSvcDriver {
    return &PatientSvc{
        Persistence: persistence,
    }
}
```

### 3. Handler Layer

**File**: `server/handlers/patient/handler.go`
```go
type PatientHandler interface {
    LookupPatient() gin.HandlerFunc
    RegisterPatient() gin.HandlerFunc
}

type patientHandler struct {
    patientSvc PatientSvcDriver
}

func NewPatientHandler(svc PatientSvcDriver) PatientHandler {
    return &patientHandler{
        patientSvc: svc,
    }
}
```

---

## Dependency Injection Setup

Add to DI container:

```go
// Register Patient Persistence
container.RegisterFactory((*patientPersistence.PatientPersistenceDriver)(nil), func(c *di.Container) interface{} {
    queries := c.Resolve((*db.Queries)(nil)).(*db.Queries)
    dbManager := c.Resolve((*dbUtils.DBManagerInterface)(nil)).(dbUtils.DBManagerInterface)
    return patientPersistence.NewPatientPersistence(queries, dbManager)
})

// Register Patient Service
container.RegisterFactory((*patientService.PatientSvcDriver)(nil), func(c *di.Container) interface{} {
    persistence := c.Resolve((*patientPersistence.PatientPersistenceDriver)(nil)).(patientPersistence.PatientPersistenceDriver)
    return patientService.NewPatientSvc(persistence)
})

// Register Patient Handler
container.RegisterFactory((*patientHandler.PatientHandler)(nil), func(c *di.Container) interface{} {
    svc := c.Resolve((*patientService.PatientSvcDriver)(nil)).(patientService.PatientSvcDriver)
    return patientHandler.NewPatientHandler(svc)
})
```

---

## File Structure

```
server/
├── persistence/patient/
│   ├── driver.go          # Interface + constructor
│   └── patient.go         # Implementation with WithConn
├── services/patient/
│   ├── driver.go          # Interface + constructor
│   ├── patient.go         # Service implementation
│   └── types.go           # ServiceError struct + helper
├── handlers/patient/
│   ├── handler.go         # Handler implementation
│   └── dto/
│       ├── request.go     # Request DTOs
│       └── response.go    # Response DTOs
├── utils/status/
│   └── codes.go           # Status constants (OK = "200")
└── config/
    └── errors.yaml        # Add error codes 4003, 4004, 4005
```

---

## Implementation Checklist

- [ ] Add new SQLC queries to `patients.sql`
- [ ] Run `sqlc generate`
- [ ] Create `utils/status/codes.go` with status constants
- [ ] Create persistence layer (driver.go, patient.go)
- [ ] Create service layer (driver.go, types.go, patient.go)
- [ ] Create handler layer (handler.go, dto/)
- [ ] Add error codes to `errors.yaml`
- [ ] Register in DI container
- [ ] Add routes to router
- [ ] Test with Postman/curl
- [ ] Commit changes

---

## Testing Strategy

### Unit Tests
- Persistence layer: Test SQLC query execution
- Service layer: Test business logic (merge, validation, OPD ID generation)
- Handler layer: Test error handling and response formatting

### Integration Tests
- End-to-end: Phone lookup → patient display → revisit registration
- Test data merge scenarios
- Test error cases (invalid mobile, missing fields, etc.)

---

## Success Criteria

✅ Lookup returns all visit history for existing patients
✅ Lookup returns error for new patients
✅ Register creates new patient with UUID-based OPD ID
✅ Register creates revisit with incremented visit_number
✅ Data merge works correctly (new data overrides, falls back to previous)
✅ All validation rules enforced
✅ Error handling follows KB-PAY pattern
✅ All endpoints use status constants
✅ Code follows established architectural patterns
