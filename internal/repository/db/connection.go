package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func GetDBConnection(dsn string, ctx context.Context) (*sql.DB, error) {
	if dsn == "" {
		logger.Logger.Info("bd dsn is empty")
		return nil, nil
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	password, _ := u.User.Password()
	dsnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		u.Hostname(), u.Port(), u.User.Username(), password, u.Path[1:])

	db, err := sql.Open("pgx", dsnStr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
