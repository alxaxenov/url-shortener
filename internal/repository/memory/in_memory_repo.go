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
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
)

// timeFormat формат хранения временных меток.
var timeFormat = time.RFC3339

type (
	// urlData структура хранения записи в файле.
	urlData struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		CreatedAt   string `json:"created_at"`
		UserID      int    `json:"user_id"`
		Active      bool   `json:"active"`
	}

	// Value структура хранения записи в памяти.
	Value struct {
		original  string
		createdAt time.Time
		userID    int
		active    bool
	}
)

// isValid проверка валидности записи из файла. критерий - хэш и оригинальный URL не пустые строки.
func (d urlData) isValid() bool {
	return d.ShortURL != "" && d.OriginalURL != ""
}

// Ipersist интерфейс взаимодействия с файловой системой
type Ipersist interface {
	addData(string, string, time.Time, int, bool) error
	getData() ([]urlData, error)
}

// inMemoryRepo структура репозитория в памяти.
// Записи хранятся в мапе urls map[хэш]Value.
// Так же ведется маппинг добавления записей по пользователю usersURLs map[user id][]хэш.
type inMemoryRepo struct {
	urls      map[string]Value
	usersURLs map[int][]string
	maxUserID int
	persist   Ipersist
}

// NewInMemoryRepo конструктор inMemoryRepo.
func NewInMemoryRepo(persist Ipersist) (*inMemoryRepo, error) {
	repo := &inMemoryRepo{make(map[string]Value), make(map[int][]string), 0, persist}
	err := repo.loadFromPersist()
	return repo, err
}

// SetValue сохранение нового URL. Опционально возможна запись в файл.
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

// GetValue получение оригинального URL по хэшу короткого.
func (r *inMemoryRepo) GetValue(ctx context.Context, k string) (string, bool, error) {
	if v, ok := r.urls[k]; ok {
		return v.original, v.active, nil
	}
	return "", false, errors.New("key not found")
}

// loadFromPersist загрузка данных из файла в память.
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

// SaveBatch сохранение батча новых URL.
func (r *inMemoryRepo) SaveBatch(ctx context.Context, batches []repoModel.UploadBatch, userID int) error {
	for _, batch := range batches {
		_, err := r.SetValue(ctx, batch.Short, batch.Origin, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateUser создание пользователя.
func (r *inMemoryRepo) CreateUser(ctx context.Context) (int, error) {
	r.maxUserID = r.maxUserID + 1
	return r.maxUserID, nil
}

// UserURLs получение всех активных URL пользователя.
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

// DeleteURLs архивация записей. Архивируются только записи, которые были добавлены текущим пользователем.
func (r *inMemoryRepo) DeleteURLs(ctx context.Context, deleteReq *model.DeleteRequest) (int, error) {
	affected := 0
	uniqueURLs := utils.UniqueSlice((*[]string)(&deleteReq.URLs))
	for _, shortURL := range uniqueURLs {
		cur, ok := r.urls[shortURL]
		if !ok || !cur.active || cur.userID != deleteReq.UserID {
			continue
		}
		r.urls[shortURL] = Value{original: cur.original, createdAt: cur.createdAt, active: false}
		if r.persist != nil {
			err := r.persist.addData(shortURL, cur.original, cur.createdAt, deleteReq.UserID, false)
			if err != nil {
				return affected, err
			}
		}
		affected++
	}
	return affected, nil
}

// URLsAndUsersCount Подсчет количества URL и пользователей в базе.
func (r *inMemoryRepo) URLsAndUsersCount(ctx context.Context) (int, int, error) {
	return len(r.urls), len(r.usersURLs), nil
}
