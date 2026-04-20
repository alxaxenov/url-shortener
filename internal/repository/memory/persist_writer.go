package memory

import (
	"errors"
	"fmt"
	"os"
)

type NewWriter struct{}

func (p *NewWriter) NewWriter(filepath string) (Writer, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("NewWriter open file error: %w", err)
	}
	return &writer{file: file}, nil
}

type writer struct {
	file *os.File
}

func (p *writer) Close() error {
	if p.file == nil {
		return errors.New("writer.Close file is nil")
	}
	p.file = nil
	return p.file.Close()
}

func (p *writer) Write(data []byte) (int, error) {
	if p.file == nil {
		return 0, errors.New("writer.Write file is nil")
	}
	return p.file.Write(data)
}
