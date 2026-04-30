package config

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v11"
)

// basePathDefault дефолтный собственный путь сервиса.
var basePathDefault = "http://localhost:8080"

// Config структура конфига сервиса.
type Config struct {
	Addr             string `env:"SERVER_ADDRESS"`
	BasePath         string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DBDSN            string `env:"DATABASE_DSN"`
	AuthCookieSecret string `env:"AUTH_COOKIE_SECRET" envDefault:"secret_key"`
	AuditFile        string `env:"AUDIT_FILE"`
	AuditURL         string `env:"AUDIT_URL"`
	RunPPROF         bool   `env:"RUN_PPROF" envDefault:"false"`
}

// checkBasePath проверка наличия и валидности поля BasePath в Config.
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

// ParseConfig инициализация Config и парсинг флагов и переменных окружения.
// Переменные окружения имеют приоритет над флагами
func ParseConfig() (*Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.Addr, "a", ":8080", "server listen address")
	flag.StringVar(&cfg.BasePath, "b", basePathDefault, "base path")
	flag.StringVar(&cfg.FileStoragePath, "f", "file_storage.txt", "file storage path")
	flag.StringVar(&cfg.DBDSN, "d", "", "database connection string")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "audit url")
	flag.Parse()
	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("config parse error %w", err)
	}
	if err := cfg.checkBasePath(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
