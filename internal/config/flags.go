package config

import (
	"errors"
	"flag"
	"net/url"
)

type FlagsSt struct {
	Addr     string
	BasePath string
}

func newFlags() *FlagsSt {
	return &FlagsSt{
		Addr:     ":8080",
		BasePath: "http://localhost:8080",
	}
}

var Flags = newFlags()

func ParseFlags() {
	flag.StringVar(&Flags.Addr, "a", ":8080", "server listen address")
	flag.Func(
		"b",
		"base path for shortened urls (default \"http://localhost:8080\")",
		func(flagValue string) error {
			if flagValue == "" {
				Flags.Addr = ":8080"
				return nil
			}
			data, err := url.ParseRequestURI(flagValue)
			if err != nil {
				return err
			}
			if data.Scheme != "http" && data.Scheme != "https" {
				return errors.New("url protocol missing")
			}
			Flags.BasePath = flagValue
			return nil
		})
	flag.Parse()
}
