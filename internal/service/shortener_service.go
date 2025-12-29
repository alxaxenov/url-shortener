package service

import (
	"crypto/rand"
	"encoding/binary"
	"net/url"
)

type ShortenerRepo interface {
	SetValue(string, string) error
	GetValue(string) (string, error)
}

type ShortenerService struct {
	repo        ShortenerRepo
	basePath    string
	base62Chars string
}

func (s *ShortenerService) AddURL(u string) (string, error) {
	hashURL, err := s.getShort()
	if err != nil {
		return "", err
	}

	joined, err := url.JoinPath(s.basePath, hashURL)
	if err != nil {
		return "", err
	}

	if err := s.repo.SetValue(hashURL, u); err != nil {
		return "", err
	}

	return joined, nil
}

func (s *ShortenerService) getShort() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	num := binary.BigEndian.Uint64(append([]byte{0, 0}, b...))
	var res = make([]byte, 0, 8)
	for num > 0 {
		res = append(res, s.base62Chars[num%62])
		num /= 62
	}
	return string(res), nil
}

func (s *ShortenerService) GetURL(short string) (string, error) {
	return s.repo.GetValue(short)
}

func NewShortenerService(repo ShortenerRepo, basePath string) *ShortenerService {
	return &ShortenerService{
		repo:        repo,
		basePath:    basePath,
		base62Chars: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
	}
}
