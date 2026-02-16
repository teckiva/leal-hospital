package domain

// ============ Patient Domain Models ============

// PatientAddress represents a patient address
type PatientAddress struct {
	Locality string
	City     string
	State    string
	Pincode  string
}

// PatientVisit represents a single patient visit
type PatientVisit struct {
	ID          int64
	Name        string
	Age         int32
	Sex         string
	Address     PatientAddress
	VisitNumber int32
	CreatedOn   string
}

// ============ Request Domain Models ============

// PatientRegisterRequest represents patient registration request in domain layer
type PatientRegisterRequest struct {
	Mobile  string
	OpdId   *string // nil for new patients, set for revisits
	Name    string
	Age     int32
	Sex     string
	Address PatientAddress
}

// ============ Response Domain Models ============

// PatientLookupResponse represents patient lookup response in domain layer
type PatientLookupResponse struct {
	Mobile      string
	OpdId       string
	Visits      []PatientVisit
	LatestVisit PatientVisit
}

// PatientRegisterResponse represents patient registration response in domain layer
type PatientRegisterResponse struct {
	ID          int64
	OpdId       string
	VisitNumber int32
}
