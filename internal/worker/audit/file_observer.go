package audit

import (
	"encoding/json"
	"os"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

type fileObserver struct {
	filePath string
}

func newFileObserver(filePath string) *fileObserver {
	return &fileObserver{filePath}
}

func (f *fileObserver) notify(message Message) {
	bytes, err := json.Marshal(message)
	if err != nil {
		logger.Logger.Errorf("fileObserver failed to marshal JSON: %s", err)
	}
	bytes = append(bytes, '\n')
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Logger.Errorf("fileObserver open file error: %s", err)
	}
	defer file.Close()
	if _, err := file.Write(bytes); err != nil {
		logger.Logger.Errorf("fileObserver failed to write JSON: %s", err)
	}
}
