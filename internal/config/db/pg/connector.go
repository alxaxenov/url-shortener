package pg

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/migrations"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConnectorPG struct {
	db.Connector
}

func (c *ConnectorPG) Open(ctx context.Context) (*sql.DB, error) {
	dataBase, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	c.DB = dataBase
	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := c.DB.PingContext(newCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if err := c.Migrate(dataBase); err != nil {
		defer dataBase.Close()
		return nil, err
	}
	return dataBase, nil
}

func (c *ConnectorPG) Close() error {
	return c.DB.Close()
}

func (c *ConnectorPG) Migrate(dataBase *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration SetDialect error: %w", err)
	}
	if err := goose.Up(dataBase, "."); err != nil {
		return fmt.Errorf("migration Up error: %w", err)
	}
	return nil
}

func NewPGConnector(dsn string) *ConnectorPG {
	return &ConnectorPG{Connector: db.Connector{DSN: dsn}}
}
