package audit

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

// fileObserver структура обработчика, пишет аудит в локальный файл.
type fileObserver struct {
	filePath string
	file     *os.File
}

// newFileObserver конструктор fileObserver.
func newFileObserver(filePath string) (*fileObserver, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("newFileObserver failed to open file: %s", filePath)
	}
	return &fileObserver{filePath, file}, nil
}

// notify логика сохранения аудита.
func (f *fileObserver) notify(message Message) {
	bytes, err := json.Marshal(message)
	if err != nil {
		logger.Logger.Errorf("fileObserver failed to marshal JSON: %s", err)
	}
	bytes = append(bytes, '\n')
	if _, err := f.file.Write(bytes); err != nil {
		logger.Logger.Errorf("fileObserver failed to write JSON: %s", err)
	}
}

// close закрытие файла
func (f *fileObserver) close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
