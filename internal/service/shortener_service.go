package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	repoModel "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/model"
)

// IShortenerRepo интерфейс логики работы с базой данных.
//
//go:generate mockery --name IShortenerRepo --with-expecter=true --filename mock_shortener_repo.go
type IShortenerRepo interface {
	SetValue(context.Context, string, string, int) (string, error)
	GetValue(context.Context, string) (string, bool, error)
	SaveBatch(context.Context, []repoModel.UploadBatch, int) error
	CreateUser(context.Context) (int, error)
	UserURLs(context.Context, int) ([]model.UserURLs, error)
	DeleteURLs(context.Context, *model.DeleteRequest) (int, error)
}

// ShortenerService структура слоя сервиса.
type ShortenerService struct {
	repo          IShortenerRepo
	basePath      string
	base62Chars   string
	DeleteMsgChan chan model.DeleteRequest
}

// NewShortenerService конструктор ShortenerService.
func NewShortenerService(repo IShortenerRepo, basePath string, deleteWorkers int) *ShortenerService {
	instance := &ShortenerService{
		repo:          repo,
		basePath:      basePath,
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: make(chan model.DeleteRequest, 1024),
	}
	for range deleteWorkers {
		go instance.deleteWorker()
	}
	return instance
}

// AddURL сохранение нового URL.
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

// GetURL получение оригинального URL.
func (s *ShortenerService) GetURL(ctx context.Context, short string) (string, error) {
	v, active, err := s.repo.GetValue(ctx, short)
	if err != nil {
		return "", err
	}
	if !active {
		return "", ErrURLDeleted
	}
	return v, nil
}

// SaveBatch сохранение нескольких новых URL батчем.
func (s *ShortenerService) SaveBatch(ctx context.Context, batches model.LoadBatchRequest, userID int) ([]model.BatchResponse, error) {
	resultBatches := make([]model.BatchResponse, 0, len(batches))
	UploadBatches := make([]repoModel.UploadBatch, 0, len(batches))
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
		UploadBatches = append(UploadBatches, repoModel.UploadBatch{hashURL, batch.OriginalURL})
		resultBatches = append(resultBatches, model.BatchResponse{CorrelationID: batch.CorrelationID, ShortURL: joined})
	}
	if err := s.repo.SaveBatch(ctx, UploadBatches, userID); err != nil {
		return nil, err
	}
	return resultBatches, nil
}

// UserURLs получение всех сохраненных URL пользователя. Возвращаются только активные записи.
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

// CLoseDeleteChan метод для закрытия очереди на архивацию URL.
func (s *ShortenerService) CLoseDeleteChan() {
	close(s.DeleteMsgChan)
}

// AppendDelete добавление запроса в очередь на архивацию.
func (s *ShortenerService) AppendDelete(userID int, URLs model.DeleteURLs) {
	s.DeleteMsgChan <- model.DeleteRequest{UserID: userID, URLs: URLs}
}

// getShort генерация случайного хэша для короткого URL.
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

// deleteWorker логика воркера, обрабатывающего запросы на архивацию.
func (s *ShortenerService) deleteWorker() {
	for msg := range s.DeleteMsgChan {
		if len(msg.URLs) == 0 {
			logger.Logger.Infof("Delete URLs is empty user %d", msg.UserID)
			continue
		}
		affected, err := s.repo.DeleteURLs(context.Background(), &msg)
		if err != nil {
			logger.Logger.Error("DeleteURLs error", "error", err)
		} else {
			logger.Logger.Infof("DeleteURLs affected %d rows", affected)
		}
	}
}
