package handler

import (
	"net/http"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	"github.com/go-chi/chi/v5"

	_ "github.com/alxaxenov/url-shortener/tree/v2/docs"
	"github.com/swaggo/http-swagger"
)

// IHandler интерфейс http обработчиков.
type IHandler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
	AddValueJSON(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	SaveBatch(w http.ResponseWriter, r *http.Request)
	UserURLs(w http.ResponseWriter, r *http.Request)
	DeleteURLs(w http.ResponseWriter, r *http.Request)
}

// IComplexMiddleware интерфейс сложной middleware, для запуска которой необходимо вызвать метод.
type IComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

// Настройки времени таймаута хендлеров.
const (
	timeoutDefault = 3 * time.Second
	timeoutBatch   = 5 * time.Second
)

// Serve подключение хендлеров и запуск роутера.
func Serve(addr string, h IHandler, userMiddleware IComplexMiddleware) error {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)
	if userMiddleware != nil {
		r.Use(userMiddleware.Use)
	}

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Post("/", timeoutHandler(h.AddValue, timeoutDefault, ""))
	r.Get("/{id}", timeoutHandler(h.GetValue, timeoutDefault, ""))
	r.Get("/ping", timeoutHandler(h.Ping, timeoutDefault, ""))

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", timeoutHandler(h.AddValueJSON, timeoutDefault, ""))
		r.Post("/shorten/batch", timeoutHandler(h.SaveBatch, timeoutBatch, ""))

		r.Get("/user/urls", timeoutHandler(h.UserURLs, timeoutDefault, ""))
		r.Delete("/user/urls", timeoutHandler(h.DeleteURLs, timeoutDefault, ""))
	})

	logger.Logger.Info("Running server on", addr)
	return http.ListenAndServe(addr, r)
}
