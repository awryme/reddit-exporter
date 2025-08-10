package redditexporter

import (
	"bytes"
	"fmt"
	"io"
)

type multiStore struct {
	stores map[string]BookStore
}

func NewMultiStore(stores map[string]BookStore) BookStore {
	return &multiStore{stores}
}

func (ms *multiStore) SaveBook(id, title, format string, data io.Reader) error {
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
