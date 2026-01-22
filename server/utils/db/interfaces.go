package db

import (
	"context"
	"database/sql"
)

// DBManagerInterface manages database connections
type DBManagerInterface interface {
	Initialize() error
	GetWriteDBConn(ctx context.Context) (DBConnInterface, error)
	GetReadDBConn(ctx context.Context) (DBConnInterface, error)
	Close() error
}

// DBConnInterface wraps a database connection
type DBConnInterface interface {
	GetConn(ctx context.Context) *sql.Conn
	Close(ctx context.Context) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (TxnInterface, error)
}

// TxnInterface wraps a database transaction
type TxnInterface interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	GetTx() *sql.Tx
}
