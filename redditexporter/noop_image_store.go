package redditexporter

import "io"

var NoOpImageStore = noOpImageStore(0)

type noOpImageStore int

func (noOpImageStore) SaveImage(id, name, ext string, data io.Reader) error {
	return nil
}
