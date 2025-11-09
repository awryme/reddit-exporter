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
	books map[string]StoredBook
}

func NewMemoryBookStore() *MemoryBookStore {
	return &MemoryBookStore{
		books: make(map[string]StoredBook),
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

	store.books[id] = StoredBook{
		Title:  title,
		Format: format,
		Data:   buf,
	}

	return nil
}

func (store *MemoryBookStore) GetBook(id string) (StoredBook, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	book, ok := store.books[id]
	return book, ok
}

func (store *MemoryBookStore) DeleteBook(ids ...string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for _, id := range ids {
		delete(store.books, id)
	}
}

type StoredImage struct {
	Name string
	Ext  string
	Data *bytes.Buffer
}

type MemoryImageStore struct {
	lock   sync.Mutex
	images map[string]StoredImage
}

func NewMemoryImageStore() *MemoryImageStore {
	return &MemoryImageStore{
		images: make(map[string]StoredImage),
	}
}

func (store *MemoryImageStore) SaveImage(id, name, ext string, data io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, data)
	if err != nil {
		return fmt.Errorf("copy book data to buf: %w", err)
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.images[id] = StoredImage{
		Name: name,
		Ext:  ext,
		Data: buf,
	}

	return nil
}

func (store *MemoryImageStore) GetImage(id string) (StoredImage, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	image, ok := store.images[id]
	return image, ok
}

func (store *MemoryImageStore) DeleteImage(ids ...string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for _, id := range ids {
		delete(store.images, id)
	}
}
