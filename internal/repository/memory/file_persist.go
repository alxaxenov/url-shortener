package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
)

//go:generate mockery --name FactoryWriterInt --with-expecter=true --inpackage --filename mock_factory_writer.go
type FactoryWriterInt interface {
	NewWriter(string) (WriterInt, error)
}

//go:generate mockery --name WriterInt --with-expecter=true --inpackage --filename mock_writer.go
type WriterInt interface {
	Close() error
	Write([]byte) (int, error)
}

//go:generate mockery --name FactoryReaderInt --with-expecter=true --inpackage --filename mock_factory_reader.go
type FactoryReaderInt interface {
	NewReader(string) (ReaderInt, error)
}

//go:generate mockery --name ReaderInt --with-expecter=true --inpackage --filename mock_reader.go
type ReaderInt interface {
	Close() error
	io.Reader
}

type filePersist struct {
	filePath      string
	writerFactory FactoryWriterInt
	readerFactory FactoryReaderInt
}

func (f *filePersist) addData(k string, v string, createdAt time.Time, userID int, active bool) error {
	producer, err := f.writerFactory.NewWriter(f.filePath)
	if err != nil {
		return err
	}
	defer producer.Close()

	createdAtString := createdAt.Format(timeFormat)
	bytes, err := json.Marshal(urlData{ShortURL: k, OriginalURL: v, CreatedAt: createdAtString, UserID: userID, Active: active})
	if err != nil {
		return fmt.Errorf("addData marshal url error: %w", err)
	}
	bytes = append(bytes, '\n')
	if _, err := producer.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (f *filePersist) getData() ([]urlData, error) {
	consumer, err := f.readerFactory.NewReader(f.filePath)
	if err != nil {
		return nil, err
	}
	if consumer == nil {
		return nil, nil
	}
	defer consumer.Close()

	scanner := bufio.NewScanner(consumer)
	var records []urlData
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var data urlData
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			logger.Logger.Infof("persist parse error %v", err)
			continue
		}
		if !data.isValid() {
			logger.Logger.Infof("persist parse data not valid %v", line)
			continue
		}
		records = append(records, data)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("persist scanner error: %w", err)
	}
	return records, nil
}

func NewFilePersist(path string) persistInt {
	return &filePersist{filePath: path, writerFactory: &NewWriter{}, readerFactory: &NewReader{}}
}
