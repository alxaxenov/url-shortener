package main

import (
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler/middlewares"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

func main() {
	h := &handler.ShortenerHandler{Service: &service.ShortenerService}

	http.Handle("/", middlewares.MethodVerification(http.MethodPost)(http.HandlerFunc(h.AddValue)))
	http.Handle("/{id}", middlewares.MethodVerification(http.MethodGet)(http.HandlerFunc(h.GetValue)))
	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
