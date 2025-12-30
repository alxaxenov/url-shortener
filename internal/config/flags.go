package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Flags struct {
	Addr     string
	BasePath string
}

func newFlags() *Flags {
	return &Flags{
		Addr:     ":8080",
		BasePath: "http://localhost:8080",
	}
}

func ParseFlags() *Flags {
	flags := newFlags()
	flag.StringVar(&flags.Addr, "a", ":8080", "server listen address")
	flag.Func(
		"b",
		"base path for shortened urls (default \"http://localhost:8080\")",
		func(basePath string) error {
			data, err := url.ParseRequestURI(basePath)
			if err != nil {
				return fmt.Errorf("url parse error: %w", err)
			}
			if data.Scheme != "http" && data.Scheme != "https" {
				return fmt.Errorf("url protocol missing: %v", basePath)
			}
			flags.BasePath = basePath
			return nil
		})
	flag.Parse()
	return flags
}
