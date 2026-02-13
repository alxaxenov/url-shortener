package db

import (
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	PingContext(context.Context) error
	Close() error
	Begin() (*sql.Tx, error)
}

type ConnectorInt interface {
	Open() (*sql.DB, error)
	Close() error
	Migrate(*sql.DB) error
	GetDB() DBTX
}

type Connector struct {
	DSN string
	DB  DBTX
}

func (c *Connector) GetDB() DBTX {
	return c.DB
}
