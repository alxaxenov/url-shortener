package main

import (
	"log"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/memory"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	flags := config.ParseFlags()

	inMemoryRepo := memory.NewInMemoryRepo()
	srv := service.NewShortenerService(inMemoryRepo, flags.BasePath)
	h := &handler.ShortenerHandler{Service: srv}

	return handler.Serve(flags.Addr, h)
}
