package memory

import (
	"errors"
	"os"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

type NewReader struct{}

func (c *NewReader) NewReader(filepath string) (ReaderInt, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		logger.Logger.Info("file does not exist")
		return nil, nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return &reader{file: file}, nil
}

type reader struct {
	file *os.File
}

func (c *reader) Close() error {
	if c.file == nil {
		return errors.New("reader.Close file is nil")
	}
	return c.file.Close()
}

func (c *reader) Read(p []byte) (n int, err error) {
	if c.file == nil {
		return 0, errors.New("reader.Read file is nil")
	}
	return c.file.Read(p)
}
