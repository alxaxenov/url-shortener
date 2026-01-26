package repository

import (
	"encoding/json"
	"os"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

type fileData []urlData

type urlData struct {
	Uuid        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func LoadFromFile(repo ShortenerRepo, path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		logger.Logger.Info("file does not exist")
		return nil
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	data := fileData{}
	if err := decoder.Decode(&data); err != nil {
		return err
	}
	for _, url := range data {
		repo.SetValue(url.ShortURL, url.OriginalURL)
	}
	return nil
}
