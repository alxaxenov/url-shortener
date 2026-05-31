package config

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"

	"github.com/caarlos0/env/v11"
)

// basePathDefault дефолтный собственный путь сервиса.
var basePathDefault = "http://localhost:8080"

// Config структура конфига сервиса.
type Config struct {
	Addr             string `env:"SERVER_ADDRESS" json:"server_address"`
	GAddr            string `env:"SERVER_GRPC_ADDRESS"`
	BasePath         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DBDSN            string `env:"DATABASE_DSN" json:"database_dsn"`
	AuthCookieSecret string `env:"AUTH_COOKIE_SECRET"`
	AuditFile        string `env:"AUDIT_FILE"`
	AuditURL         string `env:"AUDIT_URL"`
	RunPPROF         bool   `env:"RUN_PPROF"`
	EnableHTTPS      bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigPath       string `env:"CONFIG"`
	TrustedSubnet    string `env:"TRUSTED_SUBNET"`
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

func configDefault() *Config {
	return &Config{
		Addr:             ":8080",
		GAddr:            ":8081",
		BasePath:         basePathDefault,
		FileStoragePath:  "file_storage.txt",
		DBDSN:            "",
		AuthCookieSecret: "secret_key",
		AuditFile:        "",
		AuditURL:         "",
		RunPPROF:         false,
		EnableHTTPS:      false,
		TrustedSubnet:    "",
	}
}

// ParseConfig инициализация Config и парсинг флагов и переменных окружения.
// Переменные окружения имеют приоритет над флагами
func ParseConfig() (*Config, error) {
	cfg, err := getFromFile()
	if err != nil || cfg == nil {
		return nil, err
	}
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "server listen address")
	flag.StringVar(&cfg.BasePath, "b", cfg.BasePath, "base path")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.DBDSN, "d", cfg.DBDSN, "database connection string")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "audit url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable https")
	flag.StringVar(&cfg.ConfigPath, "c", "", "config file path")
	flag.StringVar(&cfg.ConfigPath, "config", "", "config file path")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "trusted subnet ip addr")
	flag.Parse()
	err = env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("config parse error %w", err)
	}
	if err := cfg.checkBasePath(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func getFromFile() (*Config, error) {
	var cPath string
	defaultConfig := configDefault()

	fs := flag.NewFlagSet("config_file", flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(io.Discard)
	fs.StringVar(&cPath, "c", "", "")
	fs.StringVar(&cPath, "config", "", "")
	_ = fs.Parse(os.Args[1:])
	if cPath == "" {
		cPath = os.Getenv("CONFIG")
	}
	if cPath == "" {
		return defaultConfig, nil
	}
	defaultConfig.ConfigPath = cPath

	file, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file %s not found", cPath)
		}
		return nil, fmt.Errorf("config file %s open error. %w", cPath, err)
	}
	defer file.Close()
	var data []byte
	re := regexp.MustCompile(" // .+$")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		result := re.ReplaceAllString(line, "")
		data = append(data, []byte(result)...)
	}
	err = scanner.Err()
	if err != nil {
		return nil, fmt.Errorf("config file %s read error. %w", cPath, err)
	}

	err = json.Unmarshal(data, defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("config file %s unmarshal error. %w", cPath, err)
	}
	return defaultConfig, nil
}
