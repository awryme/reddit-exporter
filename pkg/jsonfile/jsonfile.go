package jsonfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrFileNotFound = fmt.Errorf("json file not found")

func Read[T any](filename string) (T, error) {
	var value T
	file, err := os.Open(filename)
	if errors.Is(err, os.ErrNotExist) {
		return value, ErrFileNotFound
	}
	if err != nil {
		return value, fmt.Errorf("open json file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&value)
	if err != nil {
		return value, fmt.Errorf("decode json file: %w", err)
	}
	return value, nil
}

func Write[T any](filename string, data T) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdir for json file: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create json file: %w", err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(data)
	if err != nil {
		return fmt.Errorf("encode json file: %w", err)
	}
	return nil
}
