package main

import (
	"context"
	"log"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/pg"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	repo_db "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/db"
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

	var dbConn db.DBTX
	var repo service.ShortenerRepo
	var err error
	if cfg.DBDSN != "" {
		connector := pg.NewPGConnector(cfg.DBDSN)
		dbConn, err = pg.ConnectAndSetup(context.Background(), connector)
		if err != nil {
			logger.Logger.Fatal(err)
		}
		defer dbConn.Close()
		repo = repo_db.NewDBRepo(connector)

	} else {
		persistFile := memory.NewFilePersist(cfg.FileStoragePath)
		repo, err = memory.NewInMemoryRepo(persistFile)
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}

	srv := service.NewShortenerService(repo, cfg.BasePath)
	h := &handler.ShortenerHandler{Service: srv, DB: dbConn}

	return handler.Serve(cfg.Addr, h)
}
