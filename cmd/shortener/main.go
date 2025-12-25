package main

import (
	"fmt"
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	config.ParseFlags()
	h := &handler.ShortenerHandler{Service: service.ShortenerService}

	r := chi.NewRouter()
	r.Post("/", h.AddValue)
	r.Get("/{id}", h.GetValue)

	fmt.Println("Running server on", config.Flags.Addr)
	err := http.ListenAndServe(config.Flags.Addr, r)
	if err != nil {
		panic(err)
	}
}
