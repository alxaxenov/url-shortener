package memory

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	repoModel "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
)

var timeFormat = time.RFC3339

type urlData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	UserID      int    `json:"user_id"`
	Active      bool   `json:"active"`
}

func (d urlData) isValid() bool {
	return d.ShortURL != "" && d.OriginalURL != ""
}

type persistInt interface {
	addData(string, string, time.Time, int, bool) error
	getData() ([]urlData, error)
}

type Value struct {
	original  string
	createdAt time.Time
	userID    int
	active    bool
}

type inMemoryRepo struct {
	urls      map[string]Value
	usersURLs map[int][]string
	maxUserID int
	persist   persistInt
}

func (r *inMemoryRepo) SetValue(ctx context.Context, k string, v string, userID int) (string, error) {
	createdAt := time.Now()
	r.urls[k] = Value{v, createdAt, userID, true}
	r.usersURLs[userID] = append(r.usersURLs[userID], k)
	if r.persist != nil {
		err := r.persist.addData(k, v, createdAt, userID, true)
		if err != nil {
			return "", fmt.Errorf("ошибка записи в файл: %w", err)
		}
	}
	return k, nil
}

func (r *inMemoryRepo) GetValue(ctx context.Context, k string) (string, bool, error) {
	if v, ok := r.urls[k]; ok {
		return v.original, v.active, nil
	}
	return "", false, errors.New("key not found")
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
		r.urls[v.ShortURL] = Value{v.OriginalURL, createdAt, v.UserID, v.Active}
		if !slices.Contains(r.usersURLs[v.UserID], v.ShortURL) {
			r.usersURLs[v.UserID] = append(r.usersURLs[v.UserID], v.ShortURL)
		}
		r.maxUserID = max(r.maxUserID, v.UserID)
	}
	return nil
}

func (r *inMemoryRepo) SaveBatch(ctx context.Context, batches []repoModel.UploadBatch, userID int) error {
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

func (r *inMemoryRepo) UserURLs(ctx context.Context, id int) ([]model.UserURLs, error) {
	data := make([]model.UserURLs, 0)
	shorts, ok := r.usersURLs[id]
	if !ok {
		return data, nil
	}
	for _, short := range shorts {
		origin, ok := r.urls[short]
		if !ok {
			logger.Logger.Infof("inMemoryRepo.UserURLs original not found id=%d short=%s", id, short)
			continue
		}
		if !origin.active {
			continue
		}
		data = append(data, model.UserURLs{Short: short, Origin: origin.original})
	}
	return data, nil
}

func (r *inMemoryRepo) DeleteURLs(ctx context.Context, deleteReq *model.DeleteRequest) (int, error) {
	affected := 0
	uniqueURLs := utils.UniqueSlice((*[]string)(&deleteReq.URLs))
	for _, shortURL := range uniqueURLs {
		cur, ok := r.urls[shortURL]
		if !ok || !cur.active || cur.userID != deleteReq.UserID {
			continue
		}
		r.urls[shortURL] = Value{original: cur.original, createdAt: cur.createdAt, active: false}
		err := r.persist.addData(shortURL, cur.original, cur.createdAt, deleteReq.UserID, false)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func NewInMemoryRepo(persist persistInt) (service.ShortenerRepo, error) {
	repo := &inMemoryRepo{make(map[string]Value), make(map[int][]string), 0, persist}
	err := repo.loadFromPersist()
	return repo, err
}
