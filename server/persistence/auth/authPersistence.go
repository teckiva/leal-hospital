package auth

import (
	"context"
	"database/sql"

	"github.com/leal-hospital/server/logger"
	"github.com/leal-hospital/server/models/db"
	utilsdb "github.com/leal-hospital/server/utils/db"
)

// AuthPersistence defines the interface for authentication persistence operations
type AuthPersistence interface {
	WithConn(ctx context.Context, conn utilsdb.DBConnInterface) AuthPersistence
	GetUserByEmail(ctx context.Context, email string) (*db.LaelUser, error)
	GetUserByID(ctx context.Context, id int64) (*db.LaelUser, error)
	CreateUser(ctx context.Context, arg db.CreateUserParams) (int64, error)
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
}

// authPersistence implements the AuthPersistence interface
type authPersistence struct {
	Querier *db.Queries
}

// NewAuthPersistence creates a new instance of AuthPersistence
func NewAuthPersistence(dbConn *sql.DB) AuthPersistence {
	querier := db.New(dbConn)
	return &authPersistence{
		Querier: querier,
	}
}

// WithConn wraps the persistence with a new connection for transaction scope
func (p *authPersistence) WithConn(ctx context.Context, conn utilsdb.DBConnInterface) AuthPersistence {
	c := conn.GetConn(ctx)
	querier := db.New(c)

	return &authPersistence{
		Querier: querier,
	}
}

// GetUserByEmail retrieves a user by email address
func (p *authPersistence) GetUserByEmail(ctx context.Context, email string) (*db.LaelUser, error) {
	const functionName = "persistence.auth.authPersistence.GetUserByEmail"

	user, err := p.Querier.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info(functionName, "User not found:", email)
			return nil, err
		}
		logger.Error(functionName, "Failed to fetch user by email:", err)
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (p *authPersistence) GetUserByID(ctx context.Context, id int64) (*db.LaelUser, error) {
	const functionName = "persistence.auth.authPersistence.GetUserByID"

	user, err := p.Querier.GetUserByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info(functionName, "User not found:", id)
			return nil, err
		}
		logger.Error(functionName, "Failed to fetch user by ID:", err)
		return nil, err
	}

	return &user, nil
}

// CreateUser creates a new user in the database
func (p *authPersistence) CreateUser(ctx context.Context, arg db.CreateUserParams) (int64, error) {
	const functionName = "persistence.auth.authPersistence.CreateUser"

	result, err := p.Querier.CreateUser(ctx, arg)
	if err != nil {
		logger.Error(functionName, "Failed to create user:", err)
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		logger.Error(functionName, "Failed to get last insert ID:", err)
		return 0, err
	}

	return userID, nil
}

// UpdateUserPassword updates a user's password
func (p *authPersistence) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	const functionName = "persistence.auth.authPersistence.UpdateUserPassword"

	err := p.Querier.UpdateUserPassword(ctx, arg)
	if err != nil {
		logger.Error(functionName, "Failed to update user password:", err)
		return err
	}

	return nil
}
