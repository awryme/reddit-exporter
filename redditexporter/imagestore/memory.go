package imagestore

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type StoredImage struct {
	Name string
	Data *bytes.Buffer
}

type Memory struct {
	lock   sync.Mutex
	images map[string]StoredImage
}

func NewMemory() *Memory {
	return &Memory{
		images: make(map[string]StoredImage),
	}
}

func (store *Memory) SaveImage(id, name string, data io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, data)
	if err != nil {
		return fmt.Errorf("copy book data to buf: %w", err)
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.images[id] = StoredImage{
		Name: name,
		Data: buf,
	}

	return nil
}

func (store *Memory) GetImage(id string) (StoredImage, bool) {
	store.lock.Lock()
	defer store.lock.Unlock()

	image, ok := store.images[id]
	return image, ok
}

func (store *Memory) DeleteImage(ids ...string) {
	store.lock.Lock()
	defer store.lock.Unlock()

	for _, id := range ids {
		delete(store.images, id)
	}
}
