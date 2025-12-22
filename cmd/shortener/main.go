package main

import (
	"net/http"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler/middlewares"
)

func main() {
	http.Handle("/", middlewares.MethodVerification(http.MethodPost)(http.HandlerFunc(handler.AddValue)))
	http.Handle("/{id}", middlewares.MethodVerification(http.MethodGet)(http.HandlerFunc(handler.GetValue)))
	err := http.ListenAndServe(`:8080`, nil)
	if err != nil {
		panic(err)
	}
}
