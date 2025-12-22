package shortener_service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/shortener_repo"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/shortener_repo/in_memory"
)

const basePath = "http://localhost:8080"

type shortenerService struct {
	repo shortener_repo.ShortenerRepo
}

func (s *shortenerService) AddURL(u string) (string, error) {
	if exist, err := s.repo.GetValue(u); err == nil {
		return exist, nil
	}

	hash_url := s.getShort(u)

	joined, err := url.JoinPath(basePath, hash_url)
	if err != nil {
		return "", err
	}

	if err := s.repo.SetValue(hash_url, u); err != nil {
		return "", err
	}

	return joined, nil
}

func (s *shortenerService) getShort(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:8]
}

func (s *shortenerService) GetURL(short string) (string, error) {
	return s.repo.GetValue(short)
}

var ShortenerService shortenerService = shortenerService{in_memory.InMemoryRepo}
