package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
)

type ShortenerRepo interface {
	SetValue(context.Context, string, string) (string, error)
	GetValue(context.Context, string) (string, error)
	SaveBatch(context.Context, []UploadBatch) error
}

type ShortenerService struct {
	repo        ShortenerRepo
	basePath    string
	base62Chars string
}

func (s *ShortenerService) AddURL(ctx context.Context, u string) (string, error) {
	if _, err := url.ParseRequestURI(u); err != nil {
		return "", NewBadURL(u, err)
	}

	hashURL, err := s.getShort()
	if err != nil {
		return "", err
	}

	inserted, err := s.repo.SetValue(ctx, hashURL, u)
	if err != nil {
		return "", err
	}
	if inserted == "" {
		return "", fmt.Errorf("AddURL inserted is empty [%s]", u)
	}

	joined, err := url.JoinPath(s.basePath, inserted)
	if err != nil {
		return "", fmt.Errorf("AddURL failed to join path %w", err)
	}
	var resultErr error
	if inserted != hashURL {
		resultErr = NewAlreadyExists(inserted, u)
	}
	return joined, resultErr
}

func (s *ShortenerService) getShort() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("getShort rand error: %w", err)
	}
	num := binary.BigEndian.Uint64(append([]byte{0, 0}, b...))
	var res = make([]byte, 0, 8)
	for num > 0 {
		res = append(res, s.base62Chars[num%62])
		num /= 62
	}
	return string(res), nil
}

func (s *ShortenerService) GetURL(ctx context.Context, short string) (string, error) {
	return s.repo.GetValue(ctx, short)
}

type UploadBatch struct {
	Short  string
	Origin string
}

func (s *ShortenerService) SaveBatch(ctx context.Context, batches model.LoadBatchRequest) ([]model.BatchResponse, error) {
	var resultBatches []model.BatchResponse
	var UploadBatches []UploadBatch
	for _, batch := range batches {
		if _, err := url.ParseRequestURI(batch.OriginalURL); err != nil {
			return nil, NewBadURL(batch.OriginalURL, err)
		}
		hashURL, err := s.getShort()
		if err != nil {
			return nil, err
		}
		joined, err := url.JoinPath(s.basePath, hashURL)
		if err != nil {
			return nil, fmt.Errorf("SaveBatch failed to join path %w", err)
		}
		UploadBatches = append(UploadBatches, UploadBatch{hashURL, batch.OriginalURL})
		resultBatches = append(resultBatches, model.BatchResponse{CorrelationID: batch.CorrelationID, ShortURL: joined})
	}
	if err := s.repo.SaveBatch(ctx, UploadBatches); err != nil {
		return nil, err
	}
	return resultBatches, nil
}

func NewShortenerService(repo ShortenerRepo, basePath string) *ShortenerService {
	return &ShortenerService{
		repo:        repo,
		basePath:    basePath,
		base62Chars: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
	}
}
