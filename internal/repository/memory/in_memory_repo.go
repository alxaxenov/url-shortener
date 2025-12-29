package memory

import "errors"

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

func NewInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{make(map[string]string)}
}
