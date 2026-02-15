package db

import (
	"context"
	"time"

	db_pack "github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

type DBRepo struct {
	Connector db_pack.ConnectorInt
}

func (d *DBRepo) SetValue(ctx context.Context, short string, origin string) (string, error) {
	db := d.Connector.GetDB()
	datetime := time.Now()
	row := db.QueryRowContext(
		ctx,
		"INSERT INTO urls (short_url, original_url, created_at) VALUES($1, $2, $3) ON CONFLICT (original_url) "+
			"DO UPDATE SET original_url = EXCLUDED.original_url RETURNING short_url", short, origin, datetime)
	var inserted string
	err := row.Scan(&inserted)
	if err != nil {
		return "", err
	}
	return inserted, nil
}

func (d *DBRepo) GetValue(ctx context.Context, short string) (string, error) {
	db := d.Connector.GetDB()
	row := db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url=$1", short)
	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		return "", err
	}
	return originalURL, nil
}

func (d *DBRepo) LoadBatch(ctx context.Context, batches []service.UploadBatch) error {
	db := d.Connector.GetDB()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urls (short_url, original_url, created_at) VALUES($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	createdAt := time.Now()
	for _, batch := range batches {
		_, err := stmt.ExecContext(ctx, batch.Short, batch.Origin, createdAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func NewDBRepo(c db_pack.ConnectorInt) service.ShortenerRepo {
	return &DBRepo{c}
}
