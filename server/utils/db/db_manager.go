package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/leal-hospital/server/config"
)

const (
	defaultWriteMaxOpenConns = 25
	defaultWriteMaxIdleConns = 10
	defaultReadMaxOpenConns  = 50
	defaultReadMaxIdleConns  = 15
	defaultConnMaxLifetime   = 30 // minutes
)

var (
	instance DBManagerInterface
	once     sync.Once
)

// DBManager implements DBManagerInterface
type DBManager struct {
	writeDB *sql.DB
	readDB  *sql.DB
	config  *config.DBConfig
	mu      sync.RWMutex
}

// DBConn implements DBConnInterface
type DBConn struct {
	conn *sql.Conn
}

// Txn implements TxnInterface
type Txn struct {
	tx *sql.Tx
}

// GetDBManager returns singleton instance
func GetDBManager() DBManagerInterface {
	once.Do(func() {
		instance = &DBManager{}
	})
	return instance
}

// NewDBManager creates a new DBManager with config
func NewDBManager(cfg *config.DBConfig) DBManagerInterface {
	return &DBManager{
		config: cfg,
	}
}

// Initialize sets up database connections
func (dm *DBManager) Initialize() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.config == nil {
		return fmt.Errorf("database configuration is nil")
	}

	// Build DSN from config
	dsn := dm.config.GetDSN()
	if dsn == "" {
		return fmt.Errorf("database DSN is empty")
	}

	// Open write connection
	writeDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open write database: %w", err)
	}

	// Configure write connection pool
	writeDB.SetMaxOpenConns(defaultWriteMaxOpenConns)
	writeDB.SetMaxIdleConns(defaultWriteMaxIdleConns)

	// Ping to verify connection
	if err := writeDB.PingContext(context.Background()); err != nil {
		return fmt.Errorf("failed to ping write database: %w", err)
	}

	dm.writeDB = writeDB
	dm.readDB = writeDB // Use same connection for reads initially

	return nil
}

// GetWriteDBConn returns a write database connection
func (dm *DBManager) GetWriteDBConn(ctx context.Context) (DBConnInterface, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.writeDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	conn, err := dm.writeDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get write connection: %w", err)
	}

	return &DBConn{conn: conn}, nil
}

// GetReadDBConn returns a read database connection
func (dm *DBManager) GetReadDBConn(ctx context.Context) (DBConnInterface, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.readDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	conn, err := dm.readDB.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get read connection: %w", err)
	}

	return &DBConn{conn: conn}, nil
}

// Close closes all database connections
func (dm *DBManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	var err error
	if dm.writeDB != nil {
		err = dm.writeDB.Close()
	}
	if dm.readDB != nil && dm.readDB != dm.writeDB {
		if closeErr := dm.readDB.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// GetConn returns the underlying connection
func (dc *DBConn) GetConn(ctx context.Context) *sql.Conn {
	return dc.conn
}

// Close closes the connection
func (dc *DBConn) Close(ctx context.Context) error {
	return dc.conn.Close()
}

// BeginTx starts a transaction
func (dc *DBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (TxnInterface, error) {
	tx, err := dc.conn.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Txn{tx: tx}, nil
}

// Commit commits the transaction
func (t *Txn) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Txn) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// GetTx returns the underlying transaction
func (t *Txn) GetTx() *sql.Tx {
	return t.tx
}
