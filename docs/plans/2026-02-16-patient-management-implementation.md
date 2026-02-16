# Patient Management Module Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement patient registration API with lookup and register endpoints following KB-PAY patterns.

**Architecture:** Three-layer architecture (persistence → service → handler) with two-stage registration flow. Phone lookup returns visit history; register endpoint handles new patients and revisits with server-side data merge.

**Tech Stack:** Go, Gin, SQLC, MySQL, UUID v7, Base62 encoding

---

## Task 1: Add SQLC Queries

**Files:**
- Modify: `server/models/dbConf/patients.sql`

**Step 1: Add new SQLC queries**

Add these queries to `server/models/dbConf/patients.sql` after existing queries:

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

**Step 2: Regenerate SQLC code**

Run:
```bash
cd server/models/dbConf
sqlc generate
```

Expected: No errors, generates `GetAllVisitsByMobile` and `GetLatestVisitByOPDID` functions in `server/models/db/patients.sql.go`

**Step 3: Verify generated code**

Run:
```bash
grep -n "GetAllVisitsByMobile\|GetLatestVisitByOPDID" server/models/db/patients.sql.go
```

Expected: Shows function definitions for both queries

---

## Task 2: Create Status Constants

**Files:**
- Create: `server/utils/status/codes.go`

**Step 1: Create status constants file**

```go
package status

const (
	OK = "200"
)
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./utils/status
```

Expected: SUCCESS, no errors

---

## Task 3: Add Error Codes to Registry

**Files:**
- Modify: `server/config/errors.yaml`

**Step 1: Add new error codes**

Add after existing patient service errors (4000-4002):

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

**Step 2: Verify YAML is valid**

Run:
```bash
cd server
go run cmd/main.go --validate-errors
```

Expected: No YAML parsing errors (or manually check syntax)

---

## Task 4: Install Required Dependencies

**Files:**
- Modify: `server/go.mod`

**Step 1: Add UUID and Base62 libraries**

Run:
```bash
cd server
go get github.com/google/uuid
go get github.com/jxskiss/base62
```

Expected: Dependencies added to go.mod and go.sum

**Step 2: Verify imports work**

Run:
```bash
go mod tidy
go mod verify
```

Expected: No errors, all modules verified

---

## Task 5: Create Persistence Layer - Driver Interface

**Files:**
- Create: `server/persistence/patient/driver.go`

**Step 1: Create persistence driver interface**

```go
package patient

import (
	"context"

	"github.com/leal-hospital/server/models/db"
	dbUtils "github.com/leal-hospital/server/utils/db"
)

// PatientPersistenceDriver defines the interface for patient data operations
type PatientPersistenceDriver interface {
	WithConn(ctx context.Context, conn dbUtils.DBConnInterface) PatientPersistenceDriver
	GetAllVisitsByMobile(ctx context.Context, mobile string) ([]db.LaelPatient, error)
	GetLatestVisitByOPDID(ctx context.Context, opdId string) (*db.LaelPatient, error)
	CreatePatient(ctx context.Context, params db.CreatePatientParams) (int64, error)
}

// PatientPersistence implements PatientPersistenceDriver
type PatientPersistence struct {
	Queries   *db.Queries
	DBManager dbUtils.DBManagerInterface
}

// NewPatientPersistence creates a new patient persistence instance
func NewPatientPersistence(queries *db.Queries, dbManager dbUtils.DBManagerInterface) PatientPersistenceDriver {
	return &PatientPersistence{
		Queries:   queries,
		DBManager: dbManager,
	}
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./persistence/patient
```

Expected: SUCCESS

---

## Task 6: Create Persistence Layer - Implementation

**Files:**
- Create: `server/persistence/patient/patient.go`

**Step 1: Implement persistence methods**

```go
package patient

import (
	"context"

	"github.com/leal-hospital/server/models/db"
	dbUtils "github.com/leal-hospital/server/utils/db"
)

// WithConn returns a new persistence instance with the given database connection
func (p *PatientPersistence) WithConn(ctx context.Context, conn dbUtils.DBConnInterface) PatientPersistenceDriver {
	return &PatientPersistence{
		Queries:   db.New(conn.GetConn(ctx)),
		DBManager: p.DBManager,
	}
}

// GetAllVisitsByMobile retrieves all visits for a patient by mobile number
func (p *PatientPersistence) GetAllVisitsByMobile(ctx context.Context, mobile string) ([]db.LaelPatient, error) {
	return p.Queries.GetAllVisitsByMobile(ctx, mobile)
}

// GetLatestVisitByOPDID retrieves the latest visit for a patient by OPD ID
func (p *PatientPersistence) GetLatestVisitByOPDID(ctx context.Context, opdId string) (*db.LaelPatient, error) {
	patient, err := p.Queries.GetLatestVisitByOPDID(ctx, opdId)
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

// CreatePatient creates a new patient record
func (p *PatientPersistence) CreatePatient(ctx context.Context, params db.CreatePatientParams) (int64, error) {
	result, err := p.Queries.CreatePatient(ctx, params)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./persistence/patient
```

Expected: SUCCESS

---

## Task 7: Create Service Layer - Types and Error Handling

**Files:**
- Create: `server/services/patient/types.go`

**Step 1: Create service types and error struct**

```go
package patient

// ServiceError represents a service layer error
type ServiceError struct {
	Code string
	Type string
}

// HandleServiceErr creates a new ServiceError
func HandleServiceErr(code string, errType string) *ServiceError {
	return &ServiceError{
		Code: code,
		Type: errType,
	}
}

// Address represents a patient address
type Address struct {
	Locality string `json:"locality,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	Pincode  string `json:"pincode,omitempty"`
}

// Visit represents a single patient visit
type Visit struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Age         int32   `json:"age"`
	Sex         string  `json:"sex"`
	Address     Address `json:"address"`
	VisitNumber int32   `json:"visitNumber"`
	CreatedOn   string  `json:"createdOn"`
}

// LatestVisit represents the most recent visit data
type LatestVisit struct {
	VisitNumber int32   `json:"visitNumber"`
	Age         int32   `json:"age"`
	Name        string  `json:"name"`
	Sex         string  `json:"sex"`
	Address     Address `json:"address"`
}

// LookupResponse is the response for patient lookup
type LookupResponse struct {
	Mobile      string      `json:"mobile"`
	OpdId       string      `json:"opdId"`
	Visits      []Visit     `json:"visits"`
	LatestVisit LatestVisit `json:"latestVisit"`
}

// RegisterRequest is the request for patient registration
type RegisterRequest struct {
	Mobile  string   `json:"mobile"`
	OpdId   *string  `json:"opdId"`
	Name    string   `json:"name"`
	Age     int32    `json:"age"`
	Sex     string   `json:"sex"`
	Address Address  `json:"address"`
}

// RegisterResponse is the response for patient registration
type RegisterResponse struct {
	ID          int64  `json:"id"`
	OpdId       string `json:"opdId"`
	VisitNumber int32  `json:"visitNumber"`
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 8: Create Service Layer - Driver Interface

**Files:**
- Create: `server/services/patient/driver.go`

**Step 1: Create service driver interface**

```go
package patient

import (
	"context"

	patientPersistence "github.com/leal-hospital/server/persistence/patient"
)

// PatientSvcDriver defines the interface for patient service operations
type PatientSvcDriver interface {
	LookupPatient(ctx context.Context, mobile string) (*LookupResponse, *ServiceError)
	RegisterPatient(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *ServiceError)
}

// PatientSvc implements PatientSvcDriver
type PatientSvc struct {
	Persistence patientPersistence.PatientPersistenceDriver
}

// NewPatientSvc creates a new patient service instance
func NewPatientSvc(persistence patientPersistence.PatientPersistenceDriver) PatientSvcDriver {
	return &PatientSvc{
		Persistence: persistence,
	}
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 9: Create Service Layer - Validation Helpers

**Files:**
- Create: `server/services/patient/validation.go`

**Step 1: Create validation helper functions**

```go
package patient

import (
	"regexp"
)

// isValidMobile validates if mobile number is 10 digits numeric
func isValidMobile(mobile string) bool {
	if len(mobile) != 10 {
		return false
	}
	matched, _ := regexp.MatchString(`^[0-9]{10}$`, mobile)
	return matched
}

// isValidAge validates if age is between 1 and 150
func isValidAge(age int32) bool {
	return age >= 1 && age <= 150
}

// isValidSex validates if sex is one of the allowed values
func isValidSex(sex string) bool {
	return sex == "male" || sex == "female" || sex == "other"
}

// validateNewPatient validates required fields for new patient registration
func validateNewPatient(req *RegisterRequest) bool {
	if req.Mobile == "" || req.Name == "" || req.Age == 0 || req.Sex == "" {
		return false
	}
	return true
}

// validateRevisitPatient validates required fields for revisit registration
func validateRevisitPatient(req *RegisterRequest) bool {
	if req.Mobile == "" || req.OpdId == nil || *req.OpdId == "" {
		return false
	}
	return true
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 10: Create Service Layer - OPD ID Generator

**Files:**
- Create: `server/services/patient/opdid.go`

**Step 1: Create OPD ID generation function**

```go
package patient

import (
	"github.com/google/uuid"
	"github.com/jxskiss/base62"
)

// generateOPDID generates a unique OPD ID with LAEL prefix
func generateOPDID() string {
	uuid7 := uuid.NewV7()
	base62Str := base62.EncodeToString(uuid7[:])
	return "LAEL" + base62Str
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 11: Create Service Layer - Data Merge Logic

**Files:**
- Create: `server/services/patient/merge.go`

**Step 1: Create data merge helper**

```go
package patient

import (
	"github.com/leal-hospital/server/models/db"
)

// mergePatientData merges new patient data with previous visit data
func mergePatientData(newData *RegisterRequest, previousData *db.LaelPatient) db.CreatePatientParams {
	merged := db.CreatePatientParams{
		Mobile:  previousData.Mobile,
		OpdID:   previousData.OpdID,
		VisitNumber: previousData.VisitNumber + 1,
	}

	// Merge Name
	if newData.Name != "" {
		merged.Name = newData.Name
	} else {
		merged.Name = previousData.Name
	}

	// Merge Age
	if newData.Age > 0 {
		merged.Age = newData.Age
	} else {
		merged.Age = previousData.Age
	}

	// Merge Sex
	if newData.Sex != "" {
		merged.Sex = db.LaelPatientsSex(newData.Sex)
	} else {
		merged.Sex = previousData.Sex
	}

	// Merge Address fields
	if newData.Address.Locality != "" {
		merged.AddressLocality.String = newData.Address.Locality
		merged.AddressLocality.Valid = true
	} else {
		merged.AddressLocality = previousData.AddressLocality
	}

	if newData.Address.City != "" {
		merged.AddressCity.String = newData.Address.City
		merged.AddressCity.Valid = true
	} else {
		merged.AddressCity = previousData.AddressCity
	}

	if newData.Address.State != "" {
		merged.AddressState.String = newData.Address.State
		merged.AddressState.Valid = true
	} else {
		merged.AddressState = previousData.AddressState
	}

	if newData.Address.Pincode != "" {
		merged.AddressPincode.String = newData.Address.Pincode
		merged.AddressPincode.Valid = true
	} else {
		merged.AddressPincode = previousData.AddressPincode
	}

	return merged
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 12: Create Service Layer - LookupPatient Implementation

**Files:**
- Create: `server/services/patient/patient.go`

**Step 1: Implement LookupPatient method**

```go
package patient

import (
	"context"
	"database/sql"

	"github.com/leal-hospital/server/utils/status"
)

// LookupPatient looks up a patient by mobile number and returns all visit history
func (s *PatientSvc) LookupPatient(ctx context.Context, mobile string) (*LookupResponse, *ServiceError) {
	const functionName = "services.patient.PatientSvc.LookupPatient"

	// Validate mobile number format
	if !isValidMobile(mobile) {
		return nil, HandleServiceErr("4003", "INTERNAL")
	}

	// Get all visits for this mobile number
	visits, err := s.Persistence.GetAllVisitsByMobile(ctx, mobile)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, HandleServiceErr("4000", "INTERNAL")
		}
		return nil, HandleServiceErr("1005", "INTERNAL")
	}

	// No visits found
	if len(visits) == 0 {
		return nil, HandleServiceErr("4000", "INTERNAL")
	}

	// Build response
	response := &LookupResponse{
		Mobile: mobile,
		OpdId:  visits[0].OpdID,
		Visits: make([]Visit, len(visits)),
	}

	// Convert database models to response models
	for i, visit := range visits {
		response.Visits[i] = Visit{
			ID:          visit.ID,
			Name:        visit.Name,
			Age:         visit.Age,
			Sex:         string(visit.Sex),
			VisitNumber: visit.VisitNumber,
			CreatedOn:   visit.CreatedOn.Format("2006-01-02T15:04:05Z"),
			Address: Address{
				Locality: visit.AddressLocality.String,
				City:     visit.AddressCity.String,
				State:    visit.AddressState.String,
				Pincode:  visit.AddressPincode.String,
			},
		}
	}

	// Set latest visit (last in array since ordered by visit_number ASC)
	latestVisit := visits[len(visits)-1]
	response.LatestVisit = LatestVisit{
		VisitNumber: latestVisit.VisitNumber,
		Age:         latestVisit.Age,
		Name:        latestVisit.Name,
		Sex:         string(latestVisit.Sex),
		Address: Address{
			Locality: latestVisit.AddressLocality.String,
			City:     latestVisit.AddressCity.String,
			State:    latestVisit.AddressState.String,
			Pincode:  latestVisit.AddressPincode.String,
		},
	}

	return response, &ServiceError{Code: status.OK}
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 13: Create Service Layer - RegisterPatient Implementation

**Files:**
- Modify: `server/services/patient/patient.go`

**Step 1: Add RegisterPatient method**

Add this method to `patient.go`:

```go
// RegisterPatient registers a new patient or creates a revisit
func (s *PatientSvc) RegisterPatient(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *ServiceError) {
	const functionName = "services.patient.PatientSvc.RegisterPatient"

	// Validate mobile number format
	if !isValidMobile(req.Mobile) {
		return nil, HandleServiceErr("4003", "INTERNAL")
	}

	var params db.CreatePatientParams

	// Check if this is a new patient or revisit
	if req.OpdId == nil || *req.OpdId == "" {
		// New patient registration
		if !validateNewPatient(req) {
			return nil, HandleServiceErr("4004", "INTERNAL")
		}

		// Validate age range
		if !isValidAge(req.Age) {
			return nil, HandleServiceErr("4001", "INTERNAL")
		}

		// Validate sex
		if !isValidSex(req.Sex) {
			return nil, HandleServiceErr("4001", "INTERNAL")
		}

		// Generate new OPD ID
		opdId := generateOPDID()

		// Build params for new patient
		params = db.CreatePatientParams{
			Name:        req.Name,
			Mobile:      req.Mobile,
			OpdID:       opdId,
			Age:         req.Age,
			Sex:         db.LaelPatientsSex(req.Sex),
			VisitNumber: 1,
		}

		// Set address fields
		if req.Address.Locality != "" {
			params.AddressLocality.String = req.Address.Locality
			params.AddressLocality.Valid = true
		}
		if req.Address.City != "" {
			params.AddressCity.String = req.Address.City
			params.AddressCity.Valid = true
		}
		if req.Address.State != "" {
			params.AddressState.String = req.Address.State
			params.AddressState.Valid = true
		}
		if req.Address.Pincode != "" {
			params.AddressPincode.String = req.Address.Pincode
			params.AddressPincode.Valid = true
		}

	} else {
		// Revisit registration
		if !validateRevisitPatient(req) {
			return nil, HandleServiceErr("4004", "INTERNAL")
		}

		// Get latest visit by OPD ID
		previousVisit, err := s.Persistence.GetLatestVisitByOPDID(ctx, *req.OpdId)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, HandleServiceErr("4005", "INTERNAL")
			}
			return nil, HandleServiceErr("1005", "INTERNAL")
		}

		// Verify OPD ID matches mobile number
		if previousVisit.Mobile != req.Mobile {
			return nil, HandleServiceErr("4005", "INTERNAL")
		}

		// Merge data with previous visit
		params = mergePatientData(req, previousVisit)
	}

	// Create patient record
	id, err := s.Persistence.CreatePatient(ctx, params)
	if err != nil {
		return nil, HandleServiceErr("1005", "INTERNAL")
	}

	// Build response
	response := &RegisterResponse{
		ID:          id,
		OpdId:       params.OpdID,
		VisitNumber: params.VisitNumber,
	}

	return response, &ServiceError{Code: status.OK}
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./services/patient
```

Expected: SUCCESS

---

## Task 14: Create Handler Layer - DTOs

**Files:**
- Create: `server/handlers/patient/dto/request.go`
- Create: `server/handlers/patient/dto/response.go`

**Step 1: Create request DTOs**

File: `server/handlers/patient/dto/request.go`

```go
package dto

import (
	"github.com/leal-hospital/server/services/patient"
)

// Address represents a patient address in request
type Address struct {
	Locality string `json:"locality,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	Pincode  string `json:"pincode,omitempty"`
}

// RegisterRequest is the HTTP request for patient registration
type RegisterRequest struct {
	Mobile  string   `json:"mobile" binding:"required"`
	OpdId   *string  `json:"opdId"`
	Name    string   `json:"name"`
	Age     int32    `json:"age"`
	Sex     string   `json:"sex"`
	Address Address  `json:"address"`
}

// ToServiceRequest converts DTO to service layer request
func (r *RegisterRequest) ToServiceRequest() *patient.RegisterRequest {
	return &patient.RegisterRequest{
		Mobile: r.Mobile,
		OpdId:  r.OpdId,
		Name:   r.Name,
		Age:    r.Age,
		Sex:    r.Sex,
		Address: patient.Address{
			Locality: r.Address.Locality,
			City:     r.Address.City,
			State:    r.Address.State,
			Pincode:  r.Address.Pincode,
		},
	}
}
```

**Step 2: Create response DTOs**

File: `server/handlers/patient/dto/response.go`

```go
package dto

import (
	"github.com/leal-hospital/server/services/patient"
)

// AddressResponse represents an address in response
type AddressResponse struct {
	Locality string `json:"locality,omitempty"`
	City     string `json:"city,omitempty"`
	State    string `json:"state,omitempty"`
	Pincode  string `json:"pincode,omitempty"`
}

// VisitResponse represents a visit in response
type VisitResponse struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Age         int32           `json:"age"`
	Sex         string          `json:"sex"`
	Address     AddressResponse `json:"address"`
	VisitNumber int32           `json:"visitNumber"`
	CreatedOn   string          `json:"createdOn"`
}

// LatestVisitResponse represents the latest visit in response
type LatestVisitResponse struct {
	VisitNumber int32           `json:"visitNumber"`
	Age         int32           `json:"age"`
	Name        string          `json:"name"`
	Sex         string          `json:"sex"`
	Address     AddressResponse `json:"address"`
}

// LookupResponse is the HTTP response for patient lookup
type LookupResponse struct {
	Mobile      string              `json:"mobile"`
	OpdId       string              `json:"opdId"`
	Visits      []VisitResponse     `json:"visits"`
	LatestVisit LatestVisitResponse `json:"latestVisit"`
}

// RegisterResponse is the HTTP response for patient registration
type RegisterResponse struct {
	ID          int64  `json:"id"`
	OpdId       string `json:"opdId"`
	VisitNumber int32  `json:"visitNumber"`
}

// FromServiceLookupResponse converts service response to DTO
func FromServiceLookupResponse(svcResp *patient.LookupResponse) *LookupResponse {
	visits := make([]VisitResponse, len(svcResp.Visits))
	for i, visit := range svcResp.Visits {
		visits[i] = VisitResponse{
			ID:          visit.ID,
			Name:        visit.Name,
			Age:         visit.Age,
			Sex:         visit.Sex,
			VisitNumber: visit.VisitNumber,
			CreatedOn:   visit.CreatedOn,
			Address: AddressResponse{
				Locality: visit.Address.Locality,
				City:     visit.Address.City,
				State:    visit.Address.State,
				Pincode:  visit.Address.Pincode,
			},
		}
	}

	return &LookupResponse{
		Mobile: svcResp.Mobile,
		OpdId:  svcResp.OpdId,
		Visits: visits,
		LatestVisit: LatestVisitResponse{
			VisitNumber: svcResp.LatestVisit.VisitNumber,
			Age:         svcResp.LatestVisit.Age,
			Name:        svcResp.LatestVisit.Name,
			Sex:         svcResp.LatestVisit.Sex,
			Address: AddressResponse{
				Locality: svcResp.LatestVisit.Address.Locality,
				City:     svcResp.LatestVisit.Address.City,
				State:    svcResp.LatestVisit.Address.State,
				Pincode:  svcResp.LatestVisit.Address.Pincode,
			},
		},
	}
}

// FromServiceRegisterResponse converts service response to DTO
func FromServiceRegisterResponse(svcResp *patient.RegisterResponse) *RegisterResponse {
	return &RegisterResponse{
		ID:          svcResp.ID,
		OpdId:       svcResp.OpdId,
		VisitNumber: svcResp.VisitNumber,
	}
}
```

**Step 3: Verify files compile**

Run:
```bash
cd server
go build ./handlers/patient/dto
```

Expected: SUCCESS

---

## Task 15: Create Handler Layer - Handler Implementation

**Files:**
- Create: `server/handlers/patient/handler.go`

**Step 1: Create handler interface and implementation**

```go
package patient

import (
	"github.com/gin-gonic/gin"

	"github.com/leal-hospital/server/handlers/patient/dto"
	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/medierror"
	"github.com/leal-hospital/server/services/patient"
	"github.com/leal-hospital/server/utils"
	"github.com/leal-hospital/server/utils/status"
)

// PatientHandler defines the interface for patient handling operations
type PatientHandler interface {
	LookupPatient() gin.HandlerFunc
	RegisterPatient() gin.HandlerFunc
}

// patientHandler implements the PatientHandler interface
type patientHandler struct {
	patientSvc patient.PatientSvcDriver
}

// NewPatientHandler creates a new instance of PatientHandler
func NewPatientHandler(svc patient.PatientSvcDriver) PatientHandler {
	return &patientHandler{
		patientSvc: svc,
	}
}

// LookupPatient handles GET /api/patients/lookup
func (h *patientHandler) LookupPatient() gin.HandlerFunc {
	const functionName = "handler.patient.LookupPatient"

	return func(c *gin.Context) {
		registry := medierror.GetErrorRegistry(c)
		mobile := c.Query("mobile")

		// Call service layer
		resp, errCode := h.patientSvc.LookupPatient(c, mobile)

		// Handle errors
		if errCode.Code != status.OK {
			appErr := medierror.NewFromRegistry(medierror.ErrorCode(errCode.Code), nil)
			logger.E(c, functionName, "error_in_patient_lookup")
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Success response
		response := utils.ResponseWithModel("SUCCESS", "Patient lookup successful", dto.FromServiceLookupResponse(resp))
		c.JSON(200, response)
	}
}

// RegisterPatient handles POST /api/patients/register
func (h *patientHandler) RegisterPatient() gin.HandlerFunc {
	const functionName = "handler.patient.RegisterPatient"

	return func(c *gin.Context) {
		registry := medierror.GetErrorRegistry(c)

		// Bind request body
		var req dto.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.E(c, functionName, "invalid request body")
			c.JSON(200, medierror.ErrBadRequestWithMsg("Invalid request body", "Please provide valid patient information", err).ToResponse())
			return
		}

		// Call service layer
		resp, errCode := h.patientSvc.RegisterPatient(c, req.ToServiceRequest())

		// Handle errors
		if errCode.Code != status.OK {
			appErr := medierror.NewFromRegistry(medierror.ErrorCode(errCode.Code), nil)
			logger.E(c, functionName, "error_in_patient_registration")
			c.JSON(200, appErr.ToResponse())
			return
		}

		// Success response
		response := utils.ResponseWithModel("SUCCESS", "Patient registered successfully", dto.FromServiceRegisterResponse(resp))
		c.JSON(200, response)
	}
}
```

**Step 2: Verify file compiles**

Run:
```bash
cd server
go build ./handlers/patient
```

Expected: SUCCESS

---

## Task 16: Register Components in DI Container

**Files:**
- Modify: `server/di/container.go` (or wherever DI registration happens)

**Step 1: Add patient module registrations**

Find the DI container setup and add these registrations:

```go
import (
	patientPersistence "github.com/leal-hospital/server/persistence/patient"
	patientService "github.com/leal-hospital/server/services/patient"
	patientHandler "github.com/leal-hospital/server/handlers/patient"
)

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

**Step 2: Verify application compiles**

Run:
```bash
cd server
go build ./cmd/main.go
```

Expected: SUCCESS

---

## Task 17: Register Routes

**Files:**
- Modify: `server/routes/routes.go` (or wherever routes are registered)

**Step 1: Add patient routes**

Find the route registration and add:

```go
import (
	patientHandler "github.com/leal-hospital/server/handlers/patient"
)

// Resolve patient handler from DI container
patientH := container.Resolve((*patientHandler.PatientHandler)(nil)).(patientHandler.PatientHandler)

// Register patient routes
apiRoutes := router.Group("/api")
{
	patientRoutes := apiRoutes.Group("/patients")
	{
		patientRoutes.GET("/lookup", patientH.LookupPatient())
		patientRoutes.POST("/register", patientH.RegisterPatient())
	}
}
```

**Step 2: Verify application compiles and runs**

Run:
```bash
cd server
go build ./cmd/main.go
./main
```

Expected: Server starts without errors

---

## Task 18: Manual Testing

**Step 1: Test patient lookup for new patient (should error)**

Run:
```bash
curl -X GET "http://localhost:8080/api/patients/lookup?mobile=9999999999" -v
```

Expected:
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

**Step 2: Test new patient registration**

Run:
```bash
curl -X POST "http://localhost:8080/api/patients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9999999999",
    "name": "Test Patient",
    "age": 30,
    "sex": "male",
    "address": {
      "locality": "Test Locality",
      "city": "Test City",
      "state": "Test State",
      "pincode": "123456"
    }
  }'
```

Expected:
```json
{
  "code": "SUCCESS",
  "msg": "Patient registered successfully",
  "model": {
    "id": 1,
    "opdId": "LAELxxxxxxxxxxxxx",
    "visitNumber": 1
  }
}
```

**Step 3: Test patient lookup for existing patient**

Run:
```bash
curl -X GET "http://localhost:8080/api/patients/lookup?mobile=9999999999" -v
```

Expected: Returns patient with visit history

**Step 4: Test revisit registration**

Run:
```bash
curl -X POST "http://localhost:8080/api/patients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9999999999",
    "opdId": "LAELxxxxxxxxxxxxx",
    "age": 31
  }'
```

Expected:
```json
{
  "code": "SUCCESS",
  "msg": "Patient registered successfully",
  "model": {
    "id": 2,
    "opdId": "LAELxxxxxxxxxxxxx",
    "visitNumber": 2
  }
}
```

**Step 5: Verify data merge worked**

Run lookup again and verify visit 2 has age=31, name from visit 1, and address from visit 1.

---

## Task 19: Error Case Testing

**Step 1: Test invalid mobile format**

Run:
```bash
curl -X GET "http://localhost:8080/api/patients/lookup?mobile=123" -v
```

Expected: Error code 4003 (Invalid mobile number format)

**Step 2: Test missing required fields (new patient)**

Run:
```bash
curl -X POST "http://localhost:8080/api/patients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "8888888888"
  }'
```

Expected: Error code 4004 (Missing required fields)

**Step 3: Test invalid OPD ID (revisit)**

Run:
```bash
curl -X POST "http://localhost:8080/api/patients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "9999999999",
    "opdId": "LAELinvalidid"
  }'
```

Expected: Error code 4005 (Invalid OPD ID)

**Step 4: Test OPD ID mismatch**

Run:
```bash
curl -X POST "http://localhost:8080/api/patients/register" \
  -H "Content-Type: application/json" \
  -d '{
    "mobile": "7777777777",
    "opdId": "LAELxxxxxxxxxxxxx"
  }'
```

Expected: Error code 4005 (Invalid OPD ID - mobile doesn't match)

---

## Success Criteria

✅ All files compile without errors
✅ Server starts successfully
✅ Lookup returns error for new patient (4000)
✅ New patient registration creates patient with UUID-based OPD ID
✅ Lookup returns all visit history for existing patient
✅ Revisit registration increments visit_number
✅ Data merge works correctly (new data overrides, previous data fallback)
✅ Invalid mobile returns error 4003
✅ Missing required fields returns error 4004
✅ Invalid OPD ID returns error 4005
✅ All responses follow KB-PAY error pattern

---

## Notes

- Each task is designed to be completed in 2-5 minutes
- Compile checks after each file creation ensure no breaking changes
- Manual testing validates end-to-end functionality
- Error case testing ensures proper validation
