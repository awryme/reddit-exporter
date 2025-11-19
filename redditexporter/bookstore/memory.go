package bookstore

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type MemoryStoredBook struct {
	Title  string
	Format string
	Data   *bytes.Buffer
}

type Memory struct {
	lock  sync.Mutex
	books map[string]MemoryStoredBook
}

func NewMemory() *Memory {
	return &Memory{
		books: make(map[string]MemoryStoredBook),
	}
}

func (store *Memory) SaveBook(id, title, format string, data io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, data)
	if err != nil {
		return fmt.Errorf("copy book data to buf: %w", err)
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.books[id] = MemoryStoredBook{
		Title:  title,
		Format: format,
		Data:   buf,
	}

	return nil
}

func (store *Memory) GetBook(id string) (MemoryStoredBook, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	book, ok := store.books[id]
	return book, ok
}

func (store *Memory) DeleteBook(ids ...string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for _, id := range ids {
		delete(store.books, id)
	}
}
