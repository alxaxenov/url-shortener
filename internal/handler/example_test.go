package handler_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"context"

	handlerPack "github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/memory"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
	"github.com/go-chi/chi/v5"
)

var (
	memoryRepo, _ = memory.NewInMemoryRepo(nil)
	srv           = service.NewShortenerService(context.Background(), memoryRepo, "http://test_path.com", 1, service.NewHasher())
	handler       = handlerPack.NewShortenerHandler(srv, nil, nil)
	ctx           = context.WithValue(context.Background(), utils.UserIDKey, 42)
	shorts        []string
)

func getShort(s string) string {
	segments := strings.Split(s, "/")
	last := segments[len(segments)-1]
	shorts = append(shorts, last)
	return last
}

func ExampleShortenerHandler_AddValue() {
	r := httptest.NewRequest("POST", "/", strings.NewReader("http://long_url.com")).WithContext(ctx)
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	handler.AddValue(w, r)
	getShort(w.Body.String())
	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleShortenerHandler_AddValueJSON() {
	r := httptest.NewRequest(
		"POST",
		"/api/shorten",
		strings.NewReader(`{"url": "http://another_long_url.com"}`),
	).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.AddValueJSON(w, r)
	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleShortenerHandler_GetValue() {
	r := httptest.NewRequest("GET", "/"+shorts[0], nil).WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", shorts[0])
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetValue(w, r)
	fmt.Println(w.Code)
	// Output:
	// 307
}

func ExampleShortenerHandler_SaveBatch() {
	body := `[{"correlation_id": "correlation_1", "original_url": "http://original_1.com"}, {"correlation_id": "correlation_2", "original_url": "http://original_2.com"}]`
	r := httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(body)).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.SaveBatch(w, r)
	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleShortenerHandler_UserURLs() {
	r := httptest.NewRequest("GET", "/api/user/urls", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.UserURLs(w, r)
	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleShortenerHandler_DeleteURLs() {
	body := fmt.Sprintf("[\"%s\"]", shorts[0])
	r := httptest.NewRequest("DELETE", "/api/user/urls", strings.NewReader(body)).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.DeleteURLs(w, r)
	fmt.Println(w.Code)
	// Output:
	// 202
}
