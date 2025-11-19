package bookstore

import (
	"bytes"
	"fmt"
	"io"
)

type BookStore interface {
	SaveBook(id, title, format string, data io.Reader) error
}

type MultiStore struct {
	stores map[string]BookStore
}

func NewMultiStore(stores map[string]BookStore) *MultiStore {
	return &MultiStore{stores}
}

func (ms *MultiStore) SaveBook(id, title, format string, data io.Reader) error {
	byteBuf, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("read all data for multi-store: %w", err)
	}

	buf := bytes.NewReader(byteBuf)
	for name, store := range ms.stores {
		buf.Seek(0, io.SeekStart)
		err := store.SaveBook(id, title, format, buf)
		if err != nil {
			return fmt.Errorf("save book to store '%s': %w", name, err)
		}
	}
	return nil
}
