package handler

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
}

func Serve(addr string, h Handler) error {
	r := chi.NewRouter()
	r.Post("/", h.AddValue)
	r.Get("/{id}", h.GetValue)

	log.Println("Running server on", addr)
	return http.ListenAndServe(addr, r)
}
