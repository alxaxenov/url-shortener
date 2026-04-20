package handler

import (
	"net/http"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type IHandler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
	AddValueJSON(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	SaveBatch(w http.ResponseWriter, r *http.Request)
	UserURLs(w http.ResponseWriter, r *http.Request)
	DeleteURLs(w http.ResponseWriter, r *http.Request)
}

type IComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

const (
	timeoutDefault = 3 * time.Second
	timeoutBatch   = 5 * time.Second
)

func Serve(addr string, h IHandler, userMiddleware IComplexMiddleware) error {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)
	if userMiddleware != nil {
		r.Use(userMiddleware.Use)
	}

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
