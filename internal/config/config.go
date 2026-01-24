package config

import (
	"flag"
	"fmt"
	"log"
	"net/url"

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
	if cfg.BasePath == "" {
		flag.StringVar(&cfg.BasePath, "b", basePathDefault, "base path")
	}
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.checkBasePath(); err != nil {
		log.Fatal(err)
	}
	return &cfg
}
