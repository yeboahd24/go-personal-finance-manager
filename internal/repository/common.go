package repository

import (
	"context"
	"database/sql"
)

// QueryExecutor interface for both *sql.DB and *sql.Tx
type QueryExecutor interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
