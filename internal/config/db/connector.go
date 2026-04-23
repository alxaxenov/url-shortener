// Package db содержит сущности для описания взаимодействия с абстрактной базой данных.
package db

import (
	"context"
	"database/sql"
)

// DBTX интерфейс запросов к бд.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	PingContext(context.Context) error
	Close() error
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// IConnector интерфейс коннектора, содержит общие методы для взаимодействия с бд, кроме методов запросов.
type IConnector interface {
	Open(ctx context.Context) (*sql.DB, error)
	Close() error
	Migrate(*sql.DB) error
	GetDB() DBTX
}

// Connector структура коннектора.
type Connector struct {
	DSN string
	DB  DBTX
}

// GetDB получения сущности, содержащей методы запросов к бд.
func (c *Connector) GetDB() DBTX {
	return c.DB
}
