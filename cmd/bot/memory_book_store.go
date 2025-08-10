package main

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type StoredBook struct {
	Title  string
	Format string
	Data   *bytes.Buffer
}

type MemoryBookStore struct {
	lock  sync.Mutex
	files map[string]StoredBook
}

func NewMemoryBookStore() *MemoryBookStore {
	return &MemoryBookStore{
		files: make(map[string]StoredBook),
	}
}

func (store *MemoryBookStore) SaveBook(id, title, format string, data io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, data)
	if err != nil {
		return fmt.Errorf("copy book data to buf: %w", err)
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.files[id] = StoredBook{
		Title:  title,
		Format: format,
		Data:   buf,
	}

	return nil
}

func (store *MemoryBookStore) GetBook(id string) (StoredBook, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	book, ok := store.files[id]
	return book, ok
}

func (store *MemoryBookStore) DeleteBooks(ids []string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for _, id := range ids {
		delete(store.files, id)
	}
}
