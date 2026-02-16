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
