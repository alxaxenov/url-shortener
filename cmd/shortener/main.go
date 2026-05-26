package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"golang.org/x/sync/errgroup"
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

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit string = "N/A"

func main() {
	if err := logger.Initialize(); err != nil {
		log.Fatal(err)
	}

	logger.Logger.Infof("Build version: %s", buildVersion)
	logger.Logger.Infof("Build date: %s", buildDate)
	logger.Logger.Infof("Build commit: %s", buildCommit)

	if err := run(); err != nil {
		logger.Logger.Fatal(err)
	}
}

const (
	timeoutInner  = 10 * time.Second
	timeoutCommon = 30 * time.Second
)

func run() error {
	rootCtx, rootCancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer rootCancel()

	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(rootCtx, func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutCommon)
		defer cancel()

		<-ctx.Done()
		logger.Logger.Error("failed to shutdown gracefully")
		os.Exit(1)
	})

	cfg, err := config.ParseConfig()
	if err != nil {
		return err
	}

	if cfg.RunPPROF {
		pprofSRV := http.Server{Addr: ":6060"}
		g.Go(func() error {
			if err := pprofSRV.ListenAndServe(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return fmt.Errorf("start pprof server failed: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			defer logger.Logger.Info("closed pprof server")
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeoutInner)
			defer shutdownCancel()
			if err := pprofSRV.Shutdown(shutdownCtx); err != nil {
				logger.Logger.Info("failed to shutdown pprof server")
			}
			return nil
		})
		//go func() {
		//	logger.Logger.Info("pprof listening on :6060")
		//	if err := http.ListenAndServe(":6060", nil); err != nil {
		//		logger.Logger.Info("pprof error: %v", err)
		//	}
		//}()
	}

	var dbConn db.DBTX
	var repo service.IShortenerRepo
	if cfg.DBDSN != "" {
		connector := pg.NewPGConnector(cfg.DBDSN)
		dbConn, err = connector.Open(context.Background())
		if err != nil {
			return err
		}
		repo = repo_db.NewDBRepo(connector)
		g.Go(func() error {
			defer logger.Logger.Info("closed DB connection")
			<-ctx.Done()
			return dbConn.Close()
		})

	} else {
		persistFile := memory.NewFilePersist(cfg.FileStoragePath)
		repo, err = memory.NewInMemoryRepo(persistFile)
		if err != nil {
			return err
		}
	}

	srv := service.NewShortenerService(ctx, repo, cfg.BasePath, 3, service.NewHasher())
	g.Go(func() error {
		defer logger.Logger.Info("closed delete channel")
		<-ctx.Done()
		srv.CLoseDeleteChan()
		return nil
	})

	auditPudlisher, err := audit.NewPublisher(cfg.AuditFile, cfg.AuditURL)
	if err != nil {
		return err
	}
	g.Go(func() error {
		defer logger.Logger.Info("closed audit publisher")
		<-ctx.Done()
		auditPudlisher.Close()
		return nil
	})

	h := handler.NewShortenerHandler(srv, dbConn, auditPudlisher)
	userMiddleware := middleware.NewUserMiddleware(cfg.AuthCookieSecret, repo)

	server, err := handler.NewServer(cfg.Addr, h, userMiddleware, cfg.EnableHTTPS)
	if err != nil {
		return err
	}
	// Старт сервера
	g.Go(func() error {
		if err := server.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			return fmt.Errorf("start server failed: %w", err)
		}
		return nil
	})
	// Завершение работы сервера
	g.Go(func() error {
		defer logger.Logger.Info("server closed")
		<-rootCtx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeoutInner)
		defer shutdownCancel()
		if err := server.Stop(shutdownCtx); err != nil {
			logger.Logger.Info("shutdown server error: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
