package main

import (
	"log"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/memory"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
)

func main() {
	if err := logger.Initialize(); err != nil {
		log.Fatal(err)
	}
	if err := run(); err != nil {
		logger.Logger.Fatal(err)
	}
}

func run() error {
	cfg := config.ParseConfig()

	persistFile := memory.NewFilePersist(cfg.FileStoragePath)
	inMemoryRepo, err := memory.NewInMemoryRepo(persistFile)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	srv := service.NewShortenerService(inMemoryRepo, cfg.BasePath)
	h := &handler.ShortenerHandler{Service: srv}

	return handler.Serve(cfg.Addr, h)
}
