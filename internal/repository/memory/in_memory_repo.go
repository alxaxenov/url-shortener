package memory

import (
	"errors"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository"
)

type inMemoryRepo struct {
	values map[string]string
}

func (r *inMemoryRepo) SetValue(k string, v string) error {
	r.values[k] = v
	return nil
}

func (r *inMemoryRepo) GetValue(k string) (string, error) {
	if v, ok := r.values[k]; ok {
		return v, nil
	}
	return "", errors.New("key not found")
}

func NewInMemoryRepo() repository.ShortenerRepo {
	return &inMemoryRepo{make(map[string]string)}
}
