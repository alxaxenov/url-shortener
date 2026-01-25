package config

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/caarlos0/env/v11"
)

var basePathDefault = "http://localhost:8080"

type Config struct {
	Addr     string `env:"SERVER_ADDRESS"`
	BasePath string `env:"BASE_URL"`
}

func (c Config) checkBasePath() error {
	if c.BasePath == basePathDefault {
		return nil
	}
	data, err := url.ParseRequestURI(c.BasePath)
	if err != nil {
		return fmt.Errorf("base url parse error. %w", err)
	}
	if data.Scheme != "http" && data.Scheme != "https" {
		return fmt.Errorf("base url protocol missing: %v", c.BasePath)
	}
	return nil
}

func ParseConfig() *Config {
	cfg := Config{}
	flag.StringVar(&cfg.Addr, "a", ":8080", "server listen address")
	flag.StringVar(&cfg.BasePath, "b", basePathDefault, "base path")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		logger.Logger.Fatal(err)
	}
	if err := cfg.checkBasePath(); err != nil {
		logger.Logger.Fatal(err)
	}
	return &cfg
}
