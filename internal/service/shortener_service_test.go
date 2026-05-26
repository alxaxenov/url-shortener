package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	repoModel "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortenerService_GetURL(t *testing.T) {
	type want struct {
		url string
		err error
	}
	tests := []struct {
		name      string
		setupMock func(repo *mocks.IShortenerRepo)
		want      want
	}{
		{
			name: "repo error",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().GetValue(mock.Anything, mock.AnythingOfType("string")).Times(1).
					Return("", false, errors.New("repo_error"))
			},
			want: want{"", errors.New("repo_error")},
		},
		{
			name: "not active",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().GetValue(mock.Anything, mock.AnythingOfType("string")).Times(1).
					Return("", false, nil)
			},
			want: want{"", ErrURLDeleted},
		},
		{
			name: "success",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().GetValue(mock.Anything, mock.AnythingOfType("string")).Times(1).
					Return("original_url", true, nil)
			},
			want: want{"original_url", nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewIShortenerRepo(t)
			tt.setupMock(repoMock)
			s := &ShortenerService{repo: repoMock}

			got, err := s.GetURL(context.Background(), "test_short")
			assert.Equal(t, tt.want.url, got)
			assert.Equal(t, tt.want.err, err)
			repoMock.AssertExpectations(t)
		})
	}
}

func TestShortenerService_UserURLs(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		setupMock func(repo *mocks.IShortenerRepo)
		want      []model.UserURLs
		expErr    func(*testing.T, error)
	}{
		{
			name: "repo error",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(nil, errors.New("repo_error"))
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "repo_error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "empty data",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return([]model.UserURLs{}, nil)
			},
			want: []model.UserURLs{},
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "join error",
			basePath: "incorrect_base_path:port",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return([]model.UserURLs{{"short", "origin"}}, nil)
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "UserURLs failed to join path for"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:     "ok",
			basePath: "http://test_base.com",
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(
						[]model.UserURLs{
							{"short_1", "origin_1"},
							{"short_2", "origin_2"},
						},
						nil,
					)
			},
			want: []model.UserURLs{
				{"http://test_base.com/short_1", "origin_1"},
				{"http://test_base.com/short_2", "origin_2"},
			},
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewIShortenerRepo(t)
			tt.setupMock(repoMock)
			s := &ShortenerService{repo: repoMock, basePath: tt.basePath}

			got, err := s.UserURLs(context.Background(), 42)
			tt.expErr(t, err)
			assert.Equalf(t, tt.want, got, "UserURLs - %s, unexpected result", tt.name)
			repoMock.AssertExpectations(t)
		})
	}
}

func TestShortenerService_AddURL(t *testing.T) {
	tests := []struct {
		name      string
		originURL string
		basePath  string
		setupMock func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher)
		want      string
		expErr    func(*testing.T, error)
	}{
		{
			name:      "parse new error",
			originURL: "not_url",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {},
			want:      "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "Некорректный URL [not_url]"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "GetShort error",
			originURL: "http://origin_url.com",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("", errors.New("get_short_error"))
			},
			want: "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "get_short_error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "SetValue error",
			originURL: "http://origin_url.com",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("hash", nil)
				repo.EXPECT().SetValue(mock.Anything, "hash", "http://origin_url.com", mock.AnythingOfType("int")).
					Times(1).Return("", errors.New("set_value_error"))
			},
			want: "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "set_value_error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "inserted empty",
			originURL: "http://origin_url.com",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("hash", nil)
				repo.EXPECT().SetValue(mock.Anything, "hash", "http://origin_url.com", mock.AnythingOfType("int")).
					Times(1).Return("", nil)
			},
			want: "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "AddURL inserted is empty"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "join empty",
			originURL: "http://origin_url.com",
			basePath:  "not_url:port",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("hash", nil)
				repo.EXPECT().SetValue(mock.Anything, "hash", "http://origin_url.com", mock.AnythingOfType("int")).
					Times(1).Return("hash", nil)
			},
			want: "",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "AddURL failed to join path"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "already exist",
			originURL: "http://origin_url.com",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("new_hash", nil)
				repo.EXPECT().SetValue(mock.Anything, "new_hash", "http://origin_url.com", mock.AnythingOfType("int")).
					Times(1).Return("old_hash", nil)
			},
			want: "old_hash",
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "URL уже существует [http://origin_url.com]"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name:      "ok",
			originURL: "http://origin_url.com",
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("hash", nil)
				repo.EXPECT().SetValue(mock.Anything, "hash", "http://origin_url.com", mock.AnythingOfType("int")).
					Times(1).Return("hash", nil)
			},
			want: "hash",
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewIShortenerRepo(t)
			hasherMock := mocks.NewIhasher(t)
			tt.setupMock(repoMock, hasherMock)
			s := &ShortenerService{
				repo:     repoMock,
				hasher:   hasherMock,
				basePath: tt.basePath,
			}
			got, err := s.AddURL(context.Background(), tt.originURL, 42)
			tt.expErr(t, err)
			assert.Equalf(t, tt.want, got, "AddURL - %s, unexpected result", tt.name)
			repoMock.AssertExpectations(t)
			hasherMock.AssertExpectations(t)
		})
	}
}

func TestShortenerService_deleteWorker(t *testing.T) {
	tests := []struct {
		name      string
		urls      model.DeleteURLs
		setupMock func(repo *mocks.IShortenerRepo)
	}{
		{
			name:      "empty",
			urls:      model.DeleteURLs{},
			setupMock: func(repo *mocks.IShortenerRepo) {},
		},
		{
			name: "delete error",
			urls: model.DeleteURLs{"one"},
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().DeleteURLs(context.Background(), &model.DeleteRequest{42, model.DeleteURLs{"one"}}).
					Times(1).Return(0, errors.New("delete_error"))
			},
		},
		{
			name: "ok",
			urls: model.DeleteURLs{"one", "two"},
			setupMock: func(repo *mocks.IShortenerRepo) {
				repo.EXPECT().DeleteURLs(context.Background(), &model.DeleteRequest{42, model.DeleteURLs{"one", "two"}}).
					Times(1).Return(2, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				repoMock := mocks.NewIShortenerRepo(t)
				tt.setupMock(repoMock)
				s := NewShortenerService(context.Background(), repoMock, "", 1, nil)

				s.AppendDelete(42, tt.urls)
				s.CLoseDeleteChan()
				synctest.Wait()
				repoMock.AssertExpectations(t)
			})

		})
	}
}

func TestShortenerService_SaveBatch(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		batches   model.LoadBatchRequest
		setupMock func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher)
		want      []model.BatchResponse
		expErr    func(*testing.T, error)
	}{
		{
			name:      "empty batches",
			basePath:  "",
			batches:   model.LoadBatchRequest{},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {},
			expErr:    func(t *testing.T, err error) { assert.NoError(t, err) },
			want:      []model.BatchResponse{},
		},
		{
			name:      "parse error",
			basePath:  "",
			batches:   model.LoadBatchRequest{{"correlation", "not_url"}},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "Некорректный URL "
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
			want: nil,
		},
		{
			name:     "get short error",
			basePath: "",
			batches:  model.LoadBatchRequest{{"correlation", "http://url.com"}},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("", errors.New("get_short_error"))
			},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "get_short_error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
			want: nil,
		},
		{
			name:     "join error",
			basePath: "not_url:port",
			batches:  model.LoadBatchRequest{{"correlation", "http://url.com"}},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("QWERTY", nil)
			},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "SaveBatch failed to join path"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
			want: nil,
		},
		{
			name:     "SaveBatch error",
			basePath: "http://base_path.com",
			batches:  model.LoadBatchRequest{{"correlation", "http://url.com"}},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("QWERTY", nil)
				repo.EXPECT().SaveBatch(
					mock.Anything,
					[]repoModel.UploadBatch{{"QWERTY", "http://url.com"}},
					mock.AnythingOfType("int"),
				).Times(1).Return(errors.New("save_batch_error"))
			},
			expErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				expectedPrefix := "save_batch_error"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
			want: nil,
		},
		{
			name:     "ok",
			basePath: "http://base_path.com",
			batches: model.LoadBatchRequest{
				{"correlation_1", "http://url_1.com"},
				{"correlation_2", "http://url_2.com"},
			},
			setupMock: func(repo *mocks.IShortenerRepo, hash *mocks.Ihasher) {
				hash.EXPECT().GetShort().Times(1).Return("QWERTY", nil)
				hash.EXPECT().GetShort().Times(1).Return("ASDFG", nil)
				repo.EXPECT().SaveBatch(
					mock.Anything,
					[]repoModel.UploadBatch{{"QWERTY", "http://url_1.com"}, {"ASDFG", "http://url_2.com"}},
					mock.AnythingOfType("int"),
				).Times(1).Return(nil)
			},
			expErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			want: []model.BatchResponse{
				{"correlation_1", "http://base_path.com/QWERTY"},
				{"correlation_2", "http://base_path.com/ASDFG"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := mocks.NewIShortenerRepo(t)
			hasherMock := mocks.NewIhasher(t)
			tt.setupMock(repoMock, hasherMock)
			s := &ShortenerService{
				repo:     repoMock,
				hasher:   hasherMock,
				basePath: tt.basePath,
			}

			got, err := s.SaveBatch(context.Background(), tt.batches, 42)
			tt.expErr(t, err)
			assert.Equalf(t, tt.want, got, "AddURL - %s, unexpected result", tt.name)
			repoMock.AssertExpectations(t)
			hasherMock.AssertExpectations(t)
		})
	}
}
