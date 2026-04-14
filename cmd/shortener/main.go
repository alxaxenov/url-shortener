package main

import (
	"context"
	"log"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/config"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/config/db/pg"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	repo_db "github.com/alxaxenov/url-shortener/tree/v2/internal/repository/db"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/repository/memory"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/worker/audit"
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
	cfg, err := config.ParseConfig()
	if err != nil {
		return err
	}

	var dbConn db.DBTX
	var repo service.ShortenerRepo
	if cfg.DBDSN != "" {
		connector := pg.NewPGConnector(cfg.DBDSN)
		dbConn, err = connector.Open(context.Background())
		if err != nil {
			return err
		}
		defer dbConn.Close()
		repo = repo_db.NewDBRepo(connector)

	} else {
		persistFile := memory.NewFilePersist(cfg.FileStoragePath)
		repo, err = memory.NewInMemoryRepo(persistFile)
		if err != nil {
			return err
		}
	}

	srv := service.NewShortenerService(repo, cfg.BasePath, 3)
	defer srv.CLoseDeleteChan()
	auditPudlisher := audit.NewPublisher(cfg.AuditFile, cfg.AuditURL)
	h := handler.NewShortenerHandler(srv, dbConn, auditPudlisher)
	userMiddleware := middleware.NewUserMiddleware(cfg.AuthCookieSecret, repo)

	return handler.Serve(cfg.Addr, h, userMiddleware)
}
