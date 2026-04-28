package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

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

// @Title UrlShortener API
// @Description Сервис для сокращение url.
// @Version 1.0

// @Host localhost:8080
// @schemes http

// @SecurityDefinitions.apikey CookieAuth
// @in header
// @Name AUTH_SHORTEN_KEY
// @description Авторизационный ключ, передаваемый в Cookie.

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

	if cfg.RunPPROF {
		go func() {
			log.Println("pprof listening on :6060")
			if err := http.ListenAndServe(":6060", nil); err != nil {
				log.Printf("pprof error: %v", err)
			}
		}()
	}

	var dbConn db.DBTX
	var repo service.IShortenerRepo
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

	srv := service.NewShortenerService(repo, cfg.BasePath, 3, service.NewHasher())
	defer srv.CLoseDeleteChan()
	auditPudlisher, err := audit.NewPublisher(cfg.AuditFile, cfg.AuditURL)
	if err != nil {
		return err
	}
	defer auditPudlisher.Close()
	h := handler.NewShortenerHandler(srv, dbConn, auditPudlisher)
	userMiddleware := middleware.NewUserMiddleware(cfg.AuthCookieSecret, repo)

	return handler.Serve(cfg.Addr, h, userMiddleware)
}
