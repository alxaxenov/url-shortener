package pg

import (
	"context"
	"database/sql"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/migrations"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConnectorPG struct {
	db.Connector
}

func (c *ConnectorPG) Open() (*sql.DB, error) {
	dataBase, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return nil, err
	}
	c.DB = dataBase
	return dataBase, nil
}

func (c *ConnectorPG) Close() error {
	return c.DB.Close()
}

func (c *ConnectorPG) Migrate(dataBase *sql.DB) error {
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(dataBase, "."); err != nil {
		return err
	}
	return nil
}

func NewPGConnector(dsn string) *ConnectorPG {
	return &ConnectorPG{Connector: db.Connector{DSN: dsn}}
}

func ConnectAndSetup(ctx context.Context, connector *ConnectorPG) (*sql.DB, error) {
	dataBase, err := connector.Open()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := connector.DB.PingContext(ctx); err != nil {
		defer dataBase.Close()
		return nil, err
	}
	if err := connector.Migrate(dataBase); err != nil {
		defer dataBase.Close()
		return nil, err
	}
	return dataBase, nil
}
