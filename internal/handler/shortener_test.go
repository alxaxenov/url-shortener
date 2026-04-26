package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing/synctest"

	"testing"

	"github.com/agiledragon/gomonkey/v2"
	DBMock "github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/mocks"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler/mocks"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/worker/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortenerHandler_AddValue(t *testing.T) {
	type setupMock struct {
		service func(IShortenerService *mocks.IShortenerService)
		reader  func(*mocks.Reader)
		audit   func(publisher *mocks.AuditPublisher)
	}
	type args struct {
		contentType string
		userID      any
	}
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name  string
		mocks setupMock
		args  args
		want  want
	}{
		{
			name: "wrong contect-type",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {},
				reader:  func(mockReader *mocks.Reader) {},
				audit:   func(publisher *mocks.AuditPublisher) {},
			},
			args: args{
				contentType: "application/json",
				userID:      nil,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "unexpected content-type\n",
			},
		},
		{
			name: "ReadALL error",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(0, errors.New("reader test error")).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {},
			},
			args: args{
				contentType: "text/plain",
				userID:      nil,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "reader test error\n",
			},
		},
		{
			name: "UserID error",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(len("not an url"), io.EOF).
						Run(func(p []byte) {
							copy(p, []byte("not an url"))
						}).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {},
			},
			args: args{
				contentType: "text/plain",
				userID:      "not an int",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "user id not found in context\n",
			},
		},
		{
			name: "already exist error",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {
					IShortenerService.EXPECT().
						AddURL(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("int")).
						Return("shortURL", service.NewAlreadyExists("short", "original")).
						Once()
				},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(len("not an url"), io.EOF).
						Run(func(p []byte) {
							copy(p, []byte("not an url"))
						}).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {
					publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			args: args{
				contentType: "text/plain",
				userID:      42,
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain",
				body:        "shortURL",
			},
		},
		{
			name: "parse body url error",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {
					IShortenerService.EXPECT().
						AddURL(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("int")).
						Return("", service.NewBadURL("incorrect URL", errors.New("parse error"))).
						Once()
				},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(len("not an url"), io.EOF).
						Run(func(p []byte) {
							copy(p, []byte("not an url"))
						}).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {
					publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			args: args{
				contentType: "text/plain",
				userID:      42,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "Некорректный URL [incorrect URL]\n",
			},
		},
		{
			name: "add url service error",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {
					IShortenerService.EXPECT().
						AddURL(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("int")).
						Return("", errors.New("service AddURL error")).
						Once()
				},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(len("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"), io.EOF).
						Run(func(p []byte) {
							copy(p, []byte("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"))
						}).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {
					publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
				},
			},
			args: args{
				contentType: "text/plain",
				userID:      42,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				body:        "Internal Server Error\n",
			},
		},
		{
			name: "created",
			mocks: setupMock{
				service: func(IShortenerService *mocks.IShortenerService) {
					IShortenerService.EXPECT().
						AddURL(mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"), mock.AnythingOfType("int")).
						Return("http://myservice.com/abcdef", nil).
						Once()
				},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(len("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"), io.EOF).
						Run(func(p []byte) {
							copy(p, []byte("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"))
						}).
						Once()
				},
				audit: func(publisher *mocks.AuditPublisher) {
					publisher.EXPECT().Publish(audit.Shorten, 42, "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa").Times(1)
				},
			},
			args: args{
				contentType: "text/plain",
				userID:      42,
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				body:        "http://myservice.com/abcdef",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mockService := mocks.NewIShortenerService(t)
				tt.mocks.service(mockService)
				mockReader := mocks.NewReader(t)
				tt.mocks.reader(mockReader)
				mockAudit := mocks.NewAuditPublisher(t)
				tt.mocks.audit(mockAudit)

				h := &ShortenerHandler{
					Service: mockService,
					audit:   mockAudit,
				}
				ctx := context.WithValue(context.Background(), utils.UserIDKey, tt.args.userID)
				r := httptest.NewRequest(http.MethodPost, "/", mockReader).WithContext(ctx)
				r.Header.Set("Content-Type", tt.args.contentType)
				w := httptest.NewRecorder()
				h.AddValue(w, r)
				synctest.Wait()

				res := w.Result()
				assert.Equal(t, tt.want.statusCode, res.StatusCode)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.body, string(resBody))
				mockService.AssertExpectations(t)
				mockReader.AssertExpectations(t)
			})
		})
	}
}

func TestShortenerHandler_GetValue(t *testing.T) {
	type want struct {
		location   string
		statusCode int
	}
	tests := []struct {
		name      string
		want      want
		setupMock func(*mocks.IShortenerService, *mocks.AuditPublisher)
	}{
		{
			name: "URL deleted error",
			setupMock: func(mockService *mocks.IShortenerService, publisher *mocks.AuditPublisher) {
				mockService.EXPECT().
					GetURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
					Return("", service.ErrURLDeleted).
					Once()
				publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
			},
			want: want{
				statusCode: http.StatusGone,
				location:   "",
			},
		},
		{
			name: "get url service error",
			setupMock: func(mockService *mocks.IShortenerService, publisher *mocks.AuditPublisher) {
				mockService.EXPECT().
					GetURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
					Return("", errors.New("service GetURL error")).
					Once()
				publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				location:   "",
			},
		},
		{
			name: "success",
			setupMock: func(mockService *mocks.IShortenerService, publisher *mocks.AuditPublisher) {
				mockService.EXPECT().
					GetURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
					Return("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa", nil).
					Once()
				publisher.EXPECT().Publish(audit.Follow, 0, "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa").Times(1)
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mockService := mocks.NewIShortenerService(t)
				mockPublisher := mocks.NewAuditPublisher(t)
				tt.setupMock(mockService, mockPublisher)
				h := &ShortenerHandler{
					Service: mockService,
					audit:   mockPublisher,
				}

				r := httptest.NewRequest(http.MethodGet, "/abcdef", nil)
				w := httptest.NewRecorder()
				h.GetValue(w, r)
				synctest.Wait()

				res := w.Result()
				defer res.Body.Close()
				assert.Equal(t, tt.want.statusCode, res.StatusCode)
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
				mockService.AssertExpectations(t)
			})
		})
	}
}

func TestShortenerHandler_Ping(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(dbtx *DBMock.DBTX)
		expStatus int
	}{
		{
			name: "ping error",
			setupMock: func(mockService *DBMock.DBTX) {
				mockService.EXPECT().PingContext(mock.Anything).Times(1).Return(errors.New("test error"))
			},
			expStatus: http.StatusInternalServerError,
		},
		{
			name: "ok",
			setupMock: func(mockService *DBMock.DBTX) {
				mockService.EXPECT().PingContext(mock.Anything).Times(1).Return(nil)
			},
			expStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDBTX := DBMock.NewDBTX(t)
			tt.setupMock(mockDBTX)

			h := &ShortenerHandler{DB: mockDBTX}

			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			h.Ping(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.expStatus, res.StatusCode)
			mockDBTX.AssertExpectations(t)
		})
	}
}

func TestShortenerHandler_DeleteURLs(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name      string
		setupMock func(smp *mocks.ISemaphore, srv *mocks.IShortenerService)
		body      io.Reader
		userID    any
		want      want
	}{
		{
			name:      "userID error",
			setupMock: func(smp *mocks.ISemaphore, srv *mocks.IShortenerService) {},
			body:      nil,
			userID:    "not int",
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "user id not found in context\n",
			},
		},
		{
			name:      "decoder error",
			setupMock: func(smp *mocks.ISemaphore, srv *mocks.IShortenerService) {},
			body:      strings.NewReader(`Decoder error`),
			userID:    42,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       "invalid character 'D' looking for beginning of value\n",
			},
		},
		{
			name: "ok",
			setupMock: func(smp *mocks.ISemaphore, srv *mocks.IShortenerService) {
				smp.EXPECT().Acquire().Times(0)
				smp.EXPECT().Release().Times(1)
				srv.EXPECT().AppendDelete(mock.AnythingOfType("int"), mock.Anything).Times(1)
			},
			body:   strings.NewReader(`["qwerty", "asdfg"]`),
			userID: 42,
			want: want{
				statusCode: http.StatusAccepted,
				body:       "Accepted",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mockSmp := mocks.NewISemaphore(t)
				mockSrv := mocks.NewIShortenerService(t)
				tt.setupMock(mockSmp, mockSrv)
				h := &ShortenerHandler{Service: mockSrv, deleteSemaphore: mockSmp}

				ctx := context.WithValue(context.Background(), utils.UserIDKey, tt.userID)
				r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", tt.body).WithContext(ctx)
				w := httptest.NewRecorder()
				h.DeleteURLs(w, r)
				synctest.Wait()

				assert.Equal(t, tt.want.statusCode, w.Code)
				assert.Equal(t, tt.want.body, w.Body.String())
				mockSmp.AssertExpectations(t)
			})
		})
	}
}

func TestShortenerHandler_UserURLs(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name       string
		setupMock  func(srv *mocks.IShortenerService)
		userID     any
		marshalErr bool
		want       want
	}{
		{
			name:      "userID error",
			setupMock: func(srv *mocks.IShortenerService) {},
			userID:    "not int",
			want:      want{http.StatusBadRequest, "user id not found in context\n"},
		},
		{
			name: "service error",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return([]model.UserURLs{}, errors.New("test error"))
			},
			userID: 42,
			want:   want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "no content",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return([]model.UserURLs{}, nil)
			},
			userID: 42,
			want:   want{http.StatusNoContent, "No Content\n"},
		},
		{
			name: "marshall error",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(
						[]model.UserURLs{{"qwerty", "asdfg"}},
						nil,
					)
			},
			userID:     42,
			marshalErr: true,
			want:       want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "ok",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().UserURLs(mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(
						[]model.UserURLs{{"qwerty", "asdfg"}, {"zxcv", "qazwsx"}},
						nil,
					)
			},
			userID: 42,
			want: want{
				http.StatusOK,
				`[{"short_url":"qwerty","original_url":"asdfg"},{"short_url":"zxcv","original_url":"qazwsx"}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mocks.NewIShortenerService(t)
			tt.setupMock(mockSrv)
			h := &ShortenerHandler{Service: mockSrv}

			ctx := context.WithValue(context.Background(), utils.UserIDKey, tt.userID)
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			if tt.marshalErr {
				patches := gomonkey.ApplyFunc(json.Marshal, func(any) ([]byte, error) {
					return nil, errors.New("forced marshaling error")
				})
				defer patches.Reset()
			}
			h.UserURLs(w, r)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.body, w.Body.String())
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestShortenerHandler_SaveBatch(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name       string
		setupMock  func(srv *mocks.IShortenerService)
		userID     any
		body       io.Reader
		marshalErr bool
		want       want
	}{
		{
			name:      "decode error",
			setupMock: func(srv *mocks.IShortenerService) {},
			userID:    nil,
			body:      strings.NewReader(`{}`),
			want:      want{http.StatusBadRequest, "json: cannot unmarshal object into Go value of type model.LoadBatchRequest\n"},
		},
		{
			name:      "userID error",
			setupMock: func(srv *mocks.IShortenerService) {},
			userID:    "not int",
			body:      strings.NewReader(`[]`),
			want:      want{http.StatusBadRequest, "user id not found in context\n"},
		},
		{
			name: "badURL error",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().SaveBatch(mock.Anything, mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(nil, service.NewBadURL("test_bad_url", nil))
			},
			userID: 42,
			body:   strings.NewReader(`[]`),
			want:   want{http.StatusBadRequest, "Некорректный URL [test_bad_url]\n"},
		},
		{
			name: "another error",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().SaveBatch(mock.Anything, mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(nil, errors.New("test error"))
			},
			userID: 42,
			body:   strings.NewReader(`[]`),
			want:   want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "marshal error",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().SaveBatch(mock.Anything, mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return([]model.BatchResponse{}, nil)
			},
			userID:     42,
			body:       strings.NewReader(`[]`),
			marshalErr: true,
			want:       want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "ok",
			setupMock: func(srv *mocks.IShortenerService) {
				srv.EXPECT().SaveBatch(mock.Anything, mock.Anything, mock.AnythingOfType("int")).Times(1).
					Return(
						[]model.BatchResponse{
							{"correlation_1", "short_1"},
							{"correlation_2", "short_2"},
						},
						nil,
					)
			},
			userID:     42,
			body:       strings.NewReader(`[]`),
			marshalErr: false,
			want: want{
				http.StatusCreated,
				`[{"correlation_id":"correlation_1","short_url":"short_1"},{"correlation_id":"correlation_2","short_url":"short_2"}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSrv := mocks.NewIShortenerService(t)
			tt.setupMock(mockSrv)
			h := &ShortenerHandler{Service: mockSrv}

			ctx := context.WithValue(context.Background(), utils.UserIDKey, tt.userID)
			r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", tt.body).WithContext(ctx)
			w := httptest.NewRecorder()
			if tt.marshalErr {
				patches := gomonkey.ApplyFunc(json.Marshal, func(any) ([]byte, error) {
					return nil, errors.New("forced marshaling error")
				})
				defer patches.Reset()
			}

			h.SaveBatch(w, r)

			assert.Equal(t, tt.want.statusCode, w.Code)
			assert.Equal(t, tt.want.body, w.Body.String())
			mockSrv.AssertExpectations(t)
		})
	}
}

func TestShortenerHandler_AddValueJSON(t *testing.T) {
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name       string
		setupMock  func(srv *mocks.IShortenerService, au *mocks.AuditPublisher)
		userID     any
		body       io.Reader
		marshalErr bool
		want       want
	}{
		{
			name:      "decode error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {},
			userID:    nil,
			body:      strings.NewReader(`[]`),
			want:      want{http.StatusBadRequest, "json: cannot unmarshal array into Go value of type model.AddURLRequest\n"},
		},
		{
			name:      "userID error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {},
			userID:    "not int",
			body:      strings.NewReader(`{}`),
			want:      want{http.StatusBadRequest, "user id not found in context\n"},
		},
		{
			name: "badURL error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {
				srv.EXPECT().AddURL(mock.Anything, "test_url_1", mock.AnythingOfType("int")).Times(1).
					Return("", service.NewBadURL("test_url_1", nil))
			},
			userID: 42,
			body:   strings.NewReader(`{"url": "test_url_1"}`),
			want:   want{http.StatusBadRequest, "Некорректный URL [test_url_1]\n"},
		},
		{
			name: "alreadyExist error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {
				srv.EXPECT().AddURL(mock.Anything, "test_url_2", mock.AnythingOfType("int")).Times(1).
					Return("short_exist", service.NewAlreadyExists("short_exist", "test_url_2"))
			},
			userID: 42,
			body:   strings.NewReader(`{"url": "test_url_2"}`),
			want:   want{http.StatusConflict, `{"result":"short_exist"}`},
		},
		{
			name: "another error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {
				srv.EXPECT().AddURL(mock.Anything, "test_url_3", mock.AnythingOfType("int")).Times(1).
					Return("", errors.New("forced error"))
			},
			userID: 42,
			body:   strings.NewReader(`{"url": "test_url_3"}`),
			want:   want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "marshal error",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {
				srv.EXPECT().AddURL(mock.Anything, "test_url_4", mock.AnythingOfType("int")).Times(1).
					Return("short_url", nil)
			},
			userID:     42,
			body:       strings.NewReader(`{"url": "test_url_4"}`),
			marshalErr: true,
			want:       want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
		{
			name: "ok",
			setupMock: func(srv *mocks.IShortenerService, au *mocks.AuditPublisher) {
				srv.EXPECT().AddURL(mock.Anything, "test_url_4", mock.AnythingOfType("int")).Times(1).
					Return("short_url", nil)
				au.EXPECT().Publish(audit.Shorten, 42, "test_url_4").Times(1)
			},
			userID: 42,
			body:   strings.NewReader(`{"url": "test_url_4"}`),
			want:   want{http.StatusCreated, `{"result":"short_url"}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				mockSrv := mocks.NewIShortenerService(t)
				mockAudit := mocks.NewAuditPublisher(t)
				tt.setupMock(mockSrv, mockAudit)
				h := &ShortenerHandler{Service: mockSrv, audit: mockAudit}

				ctx := context.WithValue(context.Background(), utils.UserIDKey, tt.userID)
				r := httptest.NewRequest(http.MethodPost, "/api/shorten", tt.body).WithContext(ctx)
				w := httptest.NewRecorder()
				if tt.marshalErr {
					patches := gomonkey.ApplyFunc(json.Marshal, func(any) ([]byte, error) {
						return nil, errors.New("forced marshaling error")
					})
					defer patches.Reset()
				}

				h.AddValueJSON(w, r)
				synctest.Wait()

				assert.Equal(t, tt.want.statusCode, w.Code)
				assert.Equal(t, tt.want.body, w.Body.String())
				mockSrv.AssertExpectations(t)
				mockAudit.AssertExpectations(t)
			})
		})
	}
}

func TestNewShortenerHandler(t *testing.T) {
	s := mocks.NewIShortenerService(t)
	d := DBMock.NewDBTX(t)
	a := mocks.NewAuditPublisher(t)

	handler := NewShortenerHandler(s, d, a)

	if handler == nil {
		t.Fatal("handler is nil")
	}
	_, ok := any(handler).(IHandler)
	if !ok {
		t.Error("does not implement IHandler")
	}
	if handler.Service != s {
		t.Error("Service not set correctly")
	}
	if handler.DB != d {
		t.Error("DB not set correctly")
	}
	if handler.audit != a {
		t.Error("audit not set correctly")
	}
	if handler.deleteSemaphore == nil {
		t.Error("deleteSemaphore is nil")
	}
	_, ok = any(handler.deleteSemaphore).(ISemaphore)
	if !ok {
		t.Error("does not implement ISemaphore")
	}
}
