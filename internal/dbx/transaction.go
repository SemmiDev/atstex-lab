package dbx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// txCtxKey is an unexported type for context keys defined in this package.
type txCtxKey string

const txKey txCtxKey = "dbx_transaction"

// Executor is the subset of methods shared by *sqlx.DB and *sqlx.Tx.
// Repositories should depend on this interface so they work transparently
// with or without an active transaction.
type Executor interface {
	Rebind(query string) string
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	// We can add more as needed.
}

// TransactionManager manages database transactions. Services use
// WithTransaction to wrap multiple repository calls in a single transaction.
type TransactionManager interface {
	// WithTransaction executes fn inside a database transaction. If fn returns
	// an error the transaction is rolled back; otherwise it is committed.
	// The context passed to fn carries the active *sqlx.Tx so that
	// GetExecutor can extract it in the repository layer.
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// sqlxTxManager is the default TransactionManager backed by sqlx.
type sqlxTxManager struct {
	db *sqlx.DB
}

// NewTransactionManager creates a TransactionManager for the given sqlx.DB.
func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &sqlxTxManager{db: db}
}

// WithTransaction begins a transaction, injects it into the context, executes
// fn, and commits or rolls back depending on the outcome. Panics are caught,
// rolled back, and re-panicked.
func (m *sqlxTxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("dbx: failed to begin transaction: %w", err)
	}

	// Store the *sqlx.Tx in the context.
	txCtx := context.WithValue(ctx, txKey, tx)

	// Defer handles rollback on error/panic and commit on success.
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw after rollback
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("dbx: failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(txCtx)
	return err
}

// GetExecutor returns the active *sqlx.Tx from the context if present,
// otherwise falls back to the provided *sqlx.DB. This makes repositories
// work both inside and outside transactions without any code changes.
func GetExecutor(ctx context.Context, defaultDB *sqlx.DB) Executor {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return defaultDB
}
