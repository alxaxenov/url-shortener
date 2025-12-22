package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/memory"
)

const basePath = "http://localhost:8080"

type ShortenerServiceSt struct {
	repo repository.ShortenerRepo
}

func (s *ShortenerServiceSt) AddURL(u string) (string, error) {
	if exist, err := s.repo.GetValue(u); err == nil {
		return exist, nil
	}

	hashURL := s.getShort(u)

	joined, err := url.JoinPath(basePath, hashURL)
	if err != nil {
		return "", err
	}

	if err := s.repo.SetValue(hashURL, u); err != nil {
		return "", err
	}

	return joined, nil
}

func (s *ShortenerServiceSt) getShort(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:8]
}

func (s *ShortenerServiceSt) GetURL(short string) (string, error) {
	return s.repo.GetValue(short)
}

var ShortenerService ShortenerServiceSt = ShortenerServiceSt{memory.InMemoryRepo}
