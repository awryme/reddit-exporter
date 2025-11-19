package imagestore

import "io"

var NoOpImageStore = noOpImageStore(0)

type noOpImageStore int

func (noOpImageStore) SaveImage(id, name string, data io.Reader) error {
	return nil
}
