package db

import (
	"context"
	"fmt"
	"strings"
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
		return "", fmt.Errorf("SetValue failed to insert url: %w", err)
	}
	return inserted, nil
}

func (d *DBRepo) GetValue(ctx context.Context, short string) (string, error) {
	db := d.Connector.GetDB()
	row := db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url=$1", short)
	var originalURL string
	if err := row.Scan(&originalURL); err != nil {
		return "", fmt.Errorf("GetValue failed to fetch url: %w", err)
	}
	return originalURL, nil
}

func (d *DBRepo) SaveBatch(ctx context.Context, batches []service.UploadBatch) error {
	db := d.Connector.GetDB()
	createdAt := time.Now()
	valueStrings := make([]string, 0, len(batches))
	valueArgs := make([]interface{}, 0, len(batches)*2)
	for i, u := range batches {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, u.Short, u.Origin, createdAt)
	}
	query := fmt.Sprintf(
		"INSERT INTO urls (short_url, original_url, created_at) VALUES %s",
		strings.Join(valueStrings, ","),
	)
	_, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("SaveBatch failed to insert url: %w", err)
	}
	return nil
}

func NewDBRepo(c db_pack.ConnectorInt) service.ShortenerRepo {
	return &DBRepo{c}
}
