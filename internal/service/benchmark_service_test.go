package service

import (
	"testing"

	"context"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service/mocks"
	"github.com/stretchr/testify/mock"
)

func BenchmarkGetShort(b *testing.B) {
	repoMock := mocks.NewIShortenerRepo(b)
	service := &ShortenerService{
		repo:          repoMock,
		basePath:      "",
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: nil,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.getShort()
	}
}

func BenchmarkSaveBatch(b *testing.B) {
	repoMock := mocks.NewIShortenerRepo(b)
	repoMock.EXPECT().SaveBatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	service := &ShortenerService{
		repo:          repoMock,
		basePath:      "http://base_path.com",
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: nil,
	}
	data := make([]model.BatchRequest, 10)
	for i := range data {
		data[i] = model.BatchRequest{OriginalURL: "https://example.com", CorrelationID: "id1"}
	}
	batch := model.LoadBatchRequest(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.SaveBatch(context.Background(), batch, 1)
	}
}

func BenchmarkAddURL(b *testing.B) {
	repoMock := mocks.NewIShortenerRepo(b)
	repoMock.EXPECT().SetValue(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("inserted", nil)
	service := &ShortenerService{
		repo:          repoMock,
		basePath:      "http://base_path.com",
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: nil,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.AddURL(context.Background(), "https://original.com", 1)

	}
}

func BenchmarkUserURLs(b *testing.B) {
	data := make([]model.UserURLs, 10)
	for i := range data {
		data[i] = model.UserURLs{"short", "http://original.com"}
	}
	repoMock := mocks.NewIShortenerRepo(b)
	repoMock.EXPECT().UserURLs(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID int) ([]model.UserURLs, error) {
			out := make([]model.UserURLs, len(data))
			copy(out, data)
			return out, nil
		})
	service := &ShortenerService{
		repo:          repoMock,
		basePath:      "http://base_path.com",
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: nil,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.UserURLs(context.Background(), 1)
	}
}

func BenchmarkGetURL(b *testing.B) {
	repoMock := mocks.NewIShortenerRepo(b)
	repoMock.EXPECT().GetValue(mock.Anything, mock.Anything).Return("http://origin.com", true, nil)
	service := &ShortenerService{
		repo:          repoMock,
		basePath:      "http://base_path.com",
		base62Chars:   "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
		DeleteMsgChan: nil,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.GetURL(context.Background(), "short")
	}
}
