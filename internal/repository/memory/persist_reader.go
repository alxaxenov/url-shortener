package memory

import (
	"errors"
	"fmt"
	"os"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

// NewReader структура, чтобы иметь возможность мокировать фабрику.
type NewReader struct{}

// NewReader конструктор, открывающий файл на чтение. Возвращает ссылку на reader.
func (c *NewReader) NewReader(filepath string) (Reader, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		logger.Logger.Info("file does not exist")
		return nil, nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("NewReader open file error: %w", err)
	}
	return &reader{file: file}, nil
}

// reader структура чтения из файла.
type reader struct {
	file *os.File
}

// Close завершения работы с файлом.
func (c *reader) Close() error {
	if c.file == nil {
		return errors.New("reader.Close file is nil")
	}
	return c.file.Close()
}

// Read чтение из файла.
func (c *reader) Read(p []byte) (n int, err error) {
	if c.file == nil {
		return 0, errors.New("reader.Read file is nil")
	}
	return c.file.Read(p)
}
