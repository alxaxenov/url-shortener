package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler/mocks"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortenerHandler_AddValue(t *testing.T) {
	type setupMock struct {
		service func(shortenerService *mocks.ShortenerService)
		reader  func(*mocks.Reader)
	}
	type args struct {
		contentType string
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
				service: func(shortenerService *mocks.ShortenerService) {},
				reader:  func(mockReader *mocks.Reader) {},
			},
			args: args{
				contentType: "application/json",
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
				service: func(shortenerService *mocks.ShortenerService) {},
				reader: func(mockReader *mocks.Reader) {
					mockReader.EXPECT().
						Read(mock.AnythingOfType("[]uint8")).
						Return(0, errors.New("reader test error")).
						Once()
				},
			},
			args: args{
				contentType: "text/plain",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "reader test error\n",
			},
		},
		{
			name: "parse body url error",
			mocks: setupMock{
				service: func(shortenerService *mocks.ShortenerService) {
					shortenerService.EXPECT().
						AddURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
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
			},
			args: args{
				contentType: "text/plain",
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
				service: func(shortenerService *mocks.ShortenerService) {
					shortenerService.EXPECT().
						AddURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
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
			},
			args: args{
				contentType: "text/plain",
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
				service: func(shortenerService *mocks.ShortenerService) {
					shortenerService.EXPECT().
						AddURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
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
			},
			args: args{
				contentType: "text/plain",
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
			mockService := mocks.NewShortenerService(t)
			tt.mocks.service(mockService)
			mockReader := mocks.NewReader(t)
			tt.mocks.reader(mockReader)

			h := &ShortenerHandler{
				Service: mockService,
			}
			r := httptest.NewRequest(http.MethodPost, "/", mockReader)
			r.Header.Set("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()
			h.AddValue(w, r)

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
		setupMock func(*mocks.ShortenerService)
	}{
		{
			name: "get url service error",
			setupMock: func(mockService *mocks.ShortenerService) {
				mockService.EXPECT().
					GetURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
					Return("", errors.New("service GetURL error")).
					Once()
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				location:   "",
			},
		},
		{
			name: "success",
			setupMock: func(mockService *mocks.ShortenerService) {
				mockService.EXPECT().
					GetURL(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string")).
					Return("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa", nil).
					Once()
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewShortenerService(t)
			tt.setupMock(mockService)
			h := &ShortenerHandler{
				Service: mockService,
			}

			r := httptest.NewRequest(http.MethodGet, "/abcdef", nil)
			w := httptest.NewRecorder()
			h.GetValue(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			mockService.AssertExpectations(t)
		})
	}
}
