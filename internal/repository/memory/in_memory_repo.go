package memory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

var timeFormat = time.RFC3339

type urlData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	UserID      int    `json:"user_id"`
}

func (d urlData) isValid() bool {
	return d.ShortURL != "" && d.OriginalURL != ""
}

type persistInt interface {
	addData(string, string, time.Time, int) error
	getData() ([]urlData, error)
}

type Value struct {
	original  string
	createdAt time.Time
	userID    int
}

type inMemoryRepo struct {
	urls      map[string]Value
	usersURLs map[int][]string
	maxUserID int
	persist   persistInt
}

func (r *inMemoryRepo) SetValue(ctx context.Context, k string, v string, userID int) (string, error) {
	createdAt := time.Now()
	r.urls[k] = Value{v, createdAt, userID}
	r.usersURLs[userID] = append(r.usersURLs[userID], k)
	if r.persist != nil {
		err := r.persist.addData(k, v, createdAt, userID)
		if err != nil {
			return "", fmt.Errorf("ошибка записи в файл: %w", err)
		}
	}
	return k, nil
}

func (r *inMemoryRepo) GetValue(ctx context.Context, k string) (string, error) {
	if v, ok := r.urls[k]; ok {
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
		r.urls[v.ShortURL] = Value{v.OriginalURL, createdAt, v.UserID}
		r.usersURLs[v.UserID] = append(r.usersURLs[v.UserID], v.ShortURL)
		r.maxUserID = max(r.maxUserID, v.UserID)
	}
	return nil
}

func (r *inMemoryRepo) SaveBatch(ctx context.Context, batches []service.UploadBatch, userID int) error {
	for _, batch := range batches {
		_, err := r.SetValue(ctx, batch.Short, batch.Origin, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *inMemoryRepo) CreateUser(ctx context.Context) (int, error) {
	r.maxUserID = r.maxUserID + 1
	return r.maxUserID, nil
}

func (d *inMemoryRepo) UserURLs(ctx context.Context, id int) ([]model.UserURLs, error) {
	data := make([]model.UserURLs, 0)
	shorts, ok := d.usersURLs[id]
	if !ok {
		return data, nil
	}
	for _, short := range shorts {
		origin, ok := d.urls[short]
		if !ok {
			logger.Logger.Infof("inMemoryRepo.UserURLs original not found id=%d short=%s", id, short)
			continue
		}
		data = append(data, model.UserURLs{Short: short, Origin: origin.original})
	}
	return data, nil
}

func NewInMemoryRepo(persist persistInt) (service.ShortenerRepo, error) {
	repo := &inMemoryRepo{make(map[string]Value), make(map[int][]string), 0, persist}
	err := repo.loadFromPersist()
	return repo, err
}
