package memory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

var timeFormat = time.RFC3339

type urlData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

func (d urlData) isValid() bool {
	return d.ShortURL != "" && d.OriginalURL != ""
}

type persistInt interface {
	addData(string, string, time.Time) error
	getData() ([]urlData, error)
}

type Value struct {
	original  string
	createdAt time.Time
}

type inMemoryRepo struct {
	values  map[string]Value
	persist persistInt
}

func (r *inMemoryRepo) SetValue(ctx context.Context, k string, v string) (string, error) {
	createdAt := time.Now()
	r.values[k] = Value{v, createdAt}
	if r.persist != nil {
		err := r.persist.addData(k, v, createdAt)
		if err != nil {
			return "", fmt.Errorf("ошибка записи в файл: %w", err)
		}
	}
	return k, nil
}

func (r *inMemoryRepo) GetValue(ctx context.Context, k string) (string, error) {
	if v, ok := r.values[k]; ok {
		return v.original, nil
	}
	return "", errors.New("key not found")
}

func (r *inMemoryRepo) loadFromPersist() error {
	if r.persist == nil {
		logger.Logger.Info("persist repo is nil")
		return nil
	}
	data, err := r.persist.getData()
	if err != nil {
		return err
	}
	if data == nil {
		logger.Logger.Info("data from persist is nil")
		return nil
	}
	for _, v := range data {
		createdAt, err := time.Parse(timeFormat, v.CreatedAt)
		if err != nil {
			logger.Logger.Info("failed to parse created at time %s", v.CreatedAt)
			continue
		}
		r.values[v.ShortURL] = Value{v.OriginalURL, createdAt}
	}
	return nil
}

func (r *inMemoryRepo) LoadBatch(ctx context.Context, batches []service.UploadBatch) error {
	for _, batch := range batches {
		_, err := r.SetValue(ctx, batch.Short, batch.Origin)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewInMemoryRepo(persist persistInt) (service.ShortenerRepo, error) {
	repo := &inMemoryRepo{make(map[string]Value), persist}
	err := repo.loadFromPersist()
	return repo, err
}
