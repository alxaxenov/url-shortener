package memory

import (
	"errors"
	"fmt"
	"os"
)

// NewWriter структура, чтобы иметь возможность мокировать фабрику.
type NewWriter struct{}

// NewWriter конструктор, открывающий файл на добавление записи, создает файл, если его не существует.
// Возвращает ссылку на writer.
func (p *NewWriter) NewWriter(filepath string) (Writer, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("NewWriter open file error: %w", err)
	}
	return &writer{file: file}, nil
}

// writer структура записи в файл.
type writer struct {
	file *os.File
}

// Close завершения работы с файлом.
func (p *writer) Close() error {
	if p.file == nil {
		return errors.New("writer.Close file is nil")
	}
	p.file = nil
	return p.file.Close()
}

// Write запись в файл.
func (p *writer) Write(data []byte) (int, error) {
	if p.file == nil {
		return 0, errors.New("writer.Write file is nil")
	}
	return p.file.Write(data)
}
