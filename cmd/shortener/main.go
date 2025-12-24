package main

import (
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	h := &handler.ShortenerHandler{Service: service.ShortenerService}

	r := chi.NewRouter()
	r.Post("/", h.AddValue)
	r.Get("/{id}", h.GetValue)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
