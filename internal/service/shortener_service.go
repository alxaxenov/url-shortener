package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/url"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository"
)

type ShortenerRepo interface {
	SetValue(context.Context, string, string, int) (string, error)
	GetValue(context.Context, string) (string, bool, error)
	SaveBatch(context.Context, []UploadBatch, int) error
	CreateUser(context.Context) (int, error)
	UserURLs(context.Context, int) ([]model.UserURLs, error)
	DeleteURLs(ctx context.Context, deleteMap *repository.DeleteMap) (int, error)
}

type ShortenerService struct {
	repo          ShortenerRepo
	basePath      string
	base62Chars   string
	DeleteMsgChan chan model.DeleteRequest
}

func (s *ShortenerService) AddURL(ctx context.Context, u string, userID int) (string, error) {
	if _, err := url.ParseRequestURI(u); err != nil {
		return "", NewBadURL(u, err)
	}

	hashURL, err := s.getShort()
	if err != nil {
		return "", err
	}

	inserted, err := s.repo.SetValue(ctx, hashURL, u, userID)
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

func (s *ShortenerService) GetURL(ctx context.Context, short string) (string, bool, error) {
	return s.repo.GetValue(ctx, short)
}

type UploadBatch struct {
	Short  string
	Origin string
}

func (s *ShortenerService) SaveBatch(ctx context.Context, batches model.LoadBatchRequest, userID int) ([]model.BatchResponse, error) {
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
	if err := s.repo.SaveBatch(ctx, UploadBatches, userID); err != nil {
		return nil, err
	}
	return resultBatches, nil
}

func (s *ShortenerService) UserURLs(ctx context.Context, userID int) ([]model.UserURLs, error) {
	data, err := s.repo.UserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range data {
		data[i].Short, err = url.JoinPath(s.basePath, data[i].Short)
		if err != nil {
			return nil, fmt.Errorf("UserURLs failed to join path for %s %w", data[i].Short, err)
		}
	}
	return data, nil
}

func (s *ShortenerService) deleteURLs() {
	ticker := time.NewTicker(10 * time.Second)
	var messages []model.DeleteRequest

	for {
		select {
		case msg := <-s.DeleteMsgChan:
			messages = append(messages, msg)
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}
			deleteMap := repository.DeleteMap{}
			for _, msg := range messages {
				deleteMap[msg.UserID] = append(deleteMap[msg.UserID], msg.URLs...)
			}
			affected, err := s.repo.DeleteURLs(context.Background(), &deleteMap)

			if err != nil {
				logger.Logger.Error("DeleteURLs error", "error", err)
			} else {
				logger.Logger.Infof("DeleteURLs affected %d rows", affected)
			}

			messages = nil
		}
	}
}

func (s *ShortenerService) AppendDelete(userID int, URLs model.DeleteURLs) {
	s.DeleteMsgChan <- model.DeleteRequest{UserID: userID, URLs: URLs}
}

func NewShortenerService(repo ShortenerRepo, basePath string) *ShortenerService {
	instance := &ShortenerService{
		repo:          repo,
		basePath:      basePath,
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: make(chan model.DeleteRequest, 1024),
	}
	go instance.deleteURLs()
	return instance
}
