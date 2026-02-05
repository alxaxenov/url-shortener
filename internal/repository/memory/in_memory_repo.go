package memory

import (
	"errors"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository"
)

type urlData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (d urlData) isValid() bool {
	return d.ShortURL != "" && d.OriginalURL != ""
}

type persistInt interface {
	addData(string, string) error
	getData() ([]urlData, error)
}

type inMemoryRepo struct {
	values  map[string]string
	persist persistInt
}

func (r *inMemoryRepo) SetValue(k string, v string) error {
	r.values[k] = v
	if r.persist == nil {
		return nil
	}
	return r.persist.addData(k, v)
}

func (r *inMemoryRepo) GetValue(k string) (string, error) {
	if v, ok := r.values[k]; ok {
		return v, nil
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
		r.values[v.ShortURL] = v.OriginalURL
	}
	return nil
}

func NewInMemoryRepo(persist persistInt) (repository.ShortenerRepo, error) {
	repo := &inMemoryRepo{make(map[string]string), persist}
	err := repo.loadFromPersist()
	return repo, err
}
