package handler

import (
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Handler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
	AddValueJSON(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
}

func Serve(addr string, h Handler) error {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)

	r.Post("/", h.AddValue)
	r.Post("/api/shorten", h.AddValueJSON)
	r.Get("/{id}", h.GetValue)
	r.Get("/ping", h.Ping)

	logger.Logger.Info("Running server on", addr)
	return http.ListenAndServe(addr, r)
}
