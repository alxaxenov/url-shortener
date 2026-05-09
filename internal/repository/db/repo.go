// Package db содержит реализацию репозитория для взаимодействия с бд Postgresql
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	db_pack "github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	repoModel "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
)

// DBRepo структура репозитория.
type DBRepo struct {
	Connector db_pack.IConnector
}

// NewDBRepo конструктор DBRepo.
func NewDBRepo(c db_pack.IConnector) *DBRepo {
	return &DBRepo{c}
}

// SetValue сохранение нового URL. Если такой оригинальный URL уже есть в базе, возвращается существующий хэш.
func (d *DBRepo) SetValue(ctx context.Context, short string, origin string, userID int) (string, error) {
	db := d.Connector.GetDB()
	datetime := time.Now()
	row := db.QueryRowContext(
		ctx,
		"INSERT INTO urls (short_url, original_url, created_at, user_id, active) VALUES($1, $2, $3, $4, $5) ON CONFLICT (original_url) "+
			"DO UPDATE SET original_url = EXCLUDED.original_url RETURNING short_url", short, origin, datetime, userID, true)
	var inserted string
	err := row.Scan(&inserted)
	if err != nil {
		return "", fmt.Errorf("SetValue failed to insert url: %w", err)
	}
	return inserted, nil
}

// GetValue получение оригинального URL по хэшу короткого.
func (d *DBRepo) GetValue(ctx context.Context, short string) (string, bool, error) {
	db := d.Connector.GetDB()
	row := db.QueryRowContext(ctx, "SELECT original_url, active FROM urls WHERE short_url=$1", short)
	var originalURL string
	var active bool
	if err := row.Scan(&originalURL, &active); err != nil {
		return "", false, fmt.Errorf("GetValue failed to fetch url: %w", err)
	}
	return originalURL, active, nil
}

// SaveBatch сохранение батча новых URL.
func (d *DBRepo) SaveBatch(ctx context.Context, batches []repoModel.UploadBatch, userID int) error {
	if len(batches) == 0 {
		return errors.New("SaveBatch no batches to upload")
	}
	db := d.Connector.GetDB()
	createdAt := time.Now()
	valueStrings := make([]string, 0, len(batches))
	valueArgs := make([]interface{}, 0, len(batches)*2)
	for i, u := range batches {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))
		valueArgs = append(valueArgs, u.Short, u.Origin, createdAt, userID, true)
	}
	query := fmt.Sprintf(
		"INSERT INTO urls (short_url, original_url, created_at, user_id, active) VALUES %s",
		strings.Join(valueStrings, ","),
	)
	_, err := db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("SaveBatch failed to insert url: %w", err)
	}
	return nil
}

// CreateUser создание нового пользователя в таблице users.
func (d *DBRepo) CreateUser(ctx context.Context) (int, error) {
	db := d.Connector.GetDB()
	datetime := time.Now()
	row := db.QueryRowContext(ctx, "INSERT INTO users (created_at) VALUES($1) RETURNING id", datetime)
	var addedID int
	if err := row.Scan(&addedID); err != nil {
		return 0, fmt.Errorf("CreateUser failed to create: %w", err)
	}
	return addedID, nil
}

// UserURLs получение всех активных URL пользователя.
func (d *DBRepo) UserURLs(ctx context.Context, id int) ([]model.UserURLs, error) {
	db := d.Connector.GetDB()
	rows, err := db.QueryContext(ctx, "SELECT short_url, original_url FROM urls WHERE user_id = $1 and active = true", id)
	if err != nil {
		return nil, fmt.Errorf("UserURLs failed to fetch urls: %w", err)
	}
	defer rows.Close()
	URLs := make([]model.UserURLs, 0)
	for rows.Next() {
		var URL model.UserURLs
		err = rows.Scan(&URL.Short, &URL.Origin)
		if err != nil {
			return nil, fmt.Errorf("UserURLs failed to scan url: %w", err)
		}
		URLs = append(URLs, URL)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("UserURLs rows.Err() failed: %w", err)
	}
	return URLs, nil
}

// DeleteURLs архивация записей. Архивируются только записи, которые были добавлены текущим пользователем.
func (d *DBRepo) DeleteURLs(ctx context.Context, deleteReq *model.DeleteRequest) (int, error) {
	db := d.Connector.GetDB()
	query := "UPDATE urls SET active = false WHERE user_id = $1 AND short_url = ANY($2) AND active = true"

	uniqueURLs := utils.UniqueSlice((*[]string)(&deleteReq.URLs))
	result, err := db.ExecContext(ctx, query, deleteReq.UserID, uniqueURLs)

	if err != nil {
		return 0, fmt.Errorf("DeleteURLs error %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Logger.Error("DeleteURLs failed to fetch affected rows", "error", err)
	}

	return int(rowsAffected), nil
}
