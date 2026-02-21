package handler

import (
	"net/http"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
	AddValueJSON(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	SaveBatch(w http.ResponseWriter, r *http.Request)
	UserURLs(w http.ResponseWriter, r *http.Request)
}

type ComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

func Serve(addr string, h Handler, userMiddleware ComplexMiddleware) error {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)
	if userMiddleware != nil {
		r.Use(userMiddleware.Use)
	}

	r.Post("/", timeoutHandler(h.AddValue, 3*time.Second, "").ServeHTTP)
	r.Post("/api/shorten", timeoutHandler(h.AddValueJSON, 3*time.Second, "").ServeHTTP)
	r.Post("/api/shorten/batch", timeoutHandler(h.SaveBatch, 5*time.Second, "").ServeHTTP)
	r.Get("/{id}", timeoutHandler(h.GetValue, 3*time.Second, "").ServeHTTP)
	r.Get("/ping", timeoutHandler(h.Ping, 3*time.Second, "").ServeHTTP)
	r.Get("/api/user/urls", timeoutHandler(h.UserURLs, 3*time.Second, "").ServeHTTP)

	logger.Logger.Info("Running server on", addr)
	return http.ListenAndServe(addr, r)
}
