package service

import (
	"fmt"
)

type BadURL struct {
	URL string
	Err error
}

func (e *BadURL) Error() string {
	return fmt.Sprintf("Некорректный URL [%s]", e.URL)
}

func (e *BadURL) Unwrap() error {
	return e.Err
}

func NewBadURL(u string, err error) *BadURL {
	return &BadURL{URL: u, Err: err}
}

type AlreadyExists struct {
	ShortURL  string
	OriginURL string
}

func (e *AlreadyExists) Error() string {
	return fmt.Sprintf("URL уже существует [%s]", e.OriginURL)
}

func NewAlreadyExists(shortURL, originURL string) *AlreadyExists {
	return &AlreadyExists{ShortURL: shortURL, OriginURL: originURL}
}
