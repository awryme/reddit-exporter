package bookstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type BasicFS struct {
	dir string
}

func NewBasicFS(dir string) (*BasicFS, error) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("make store dir: %w", err)
	}
	return &BasicFS{dir: dir}, nil
}

func (store *BasicFS) SaveBook(id, title, format string, data io.Reader) error {
	filename := title + "." + format
	filename = strings.ReplaceAll(filename, "/", "_")
	fullname := filepath.Join(store.dir, filename)

	file, err := os.Create(fullname)
	if err != nil {
		return fmt.Errorf("create file for book '%s': %w", filename, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("copy book to file: %w", err)
	}
	return nil
}
