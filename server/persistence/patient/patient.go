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
