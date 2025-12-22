package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

type addURLFunc func(string) (string, error)
type getURLFunc func(string) (string, error)

type mockShortenerService struct {
	addURLFunc addURLFunc
	getURLFunc getURLFunc
}

func (s *mockShortenerService) AddURL(url string) (string, error) {
	if s.addURLFunc != nil {
		return s.addURLFunc(url)
	}
	return "", nil
}

func (s *mockShortenerService) GetURL(url string) (string, error) {
	if s.getURLFunc != nil {
		return s.getURLFunc(url)
	}
	return "", nil
}

func newMockShortenerService(
	addURL func(string) (string, error), getURL func(string) (string, error),
) service.ShortenerServiceInt {
	return &mockShortenerService{addURLFunc: addURL, getURLFunc: getURL}
}

type mockReader struct {
	errorText string
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	return 0, errors.New(r.errorText)
}

func TestShortenerHandler_AddValue(t *testing.T) {
	type fields struct {
		Service service.ShortenerServiceInt
	}
	type args struct {
		body        io.Reader
		contentType string
	}
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "wrong contect-type",
			fields: fields{
				Service: newMockShortenerService(nil, nil),
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
			fields: fields{
				Service: newMockShortenerService(nil, nil),
			},
			args: args{
				body:        &mockReader{errorText: "reader test error"},
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
			fields: fields{
				Service: newMockShortenerService(nil, nil),
			},
			args: args{
				body:        strings.NewReader("not an url"),
				contentType: "text/plain",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "invalid body URL\n",
			},
		},
		{
			name: "add url service error",
			fields: fields{
				Service: newMockShortenerService(
					func(string) (string, error) { return "", errors.New("service AddURL error") }, nil,
				),
			},
			args: args{
				body:        strings.NewReader("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"),
				contentType: "text/plain",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "service AddURL error\n",
			},
		},
		{
			name: "created",
			fields: fields{
				Service: newMockShortenerService(
					func(string) (string, error) { return "http://myservice.com/abcdef", nil }, nil,
				),
			},
			args: args{
				body:        strings.NewReader("http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa"),
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
			h := &ShortenerHandler{
				Service: tt.fields.Service,
			}
			r := httptest.NewRequest(http.MethodPost, "/", tt.args.body)
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
		})
	}
}

func TestShortenerHandler_GetValue(t *testing.T) {
	type fields struct {
		Service service.ShortenerServiceInt
	}
	type want struct {
		location   string
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "get url service error",
			fields: fields{
				Service: newMockShortenerService(
					nil, func(string) (string, error) { return "", errors.New("service GetURL error") },
				),
			},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
		{
			name: "success",
			fields: fields{
				Service: newMockShortenerService(
					nil, func(string) (string, error) { return "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa", nil },
				),
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://ptmnjp.ru/xqbm8n/ekei4oj2yxfa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ShortenerHandler{
				Service: tt.fields.Service,
			}
			r := httptest.NewRequest(http.MethodGet, "/abcdef", nil)
			w := httptest.NewRecorder()
			h.GetValue(w, r)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}
