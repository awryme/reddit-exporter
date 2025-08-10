package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/awryme/reddit-exporter/httpexporter"
	"github.com/awryme/reddit-exporter/pkg/jsonfile"
)

const metafileName = "meta.json"

type Meta map[string]httpexporter.BookInfo

type FsBookStore struct {
	dir      string
	metafile string
	meta     Meta
}

func NewFsBookStore(dir string) (*FsBookStore, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	metafile := filepath.Join(dir, metafileName)
	meta, err := jsonfile.Read[Meta](metafile)
	if err != nil && !errors.Is(err, jsonfile.ErrFileNotFound) {
		return nil, fmt.Errorf("unmarshal existing meta: %w", err)
	}
	if meta == nil {
		meta = make(Meta)
	}
	return &FsBookStore{
		dir:      dir,
		metafile: metafile,
		meta:     meta,
	}, nil
}

func (ms *FsBookStore) saveMeta() error {
	return jsonfile.Write(ms.metafile, ms.meta)
}

func (ms *FsBookStore) SaveBook(id, title, format string, data io.Reader) error {
	file, err := os.Create(filepath.Join(ms.dir, id))
	if err != nil {
		return fmt.Errorf("create data file: %w", err)
	}
	defer file.Close()

	n, err := io.Copy(file, data)
	if err != nil {
		return fmt.Errorf("copy data to file: %w", err)
	}
	ms.meta[id] = httpexporter.BookInfo{
		ID:     id,
		Title:  title,
		Format: format,
		Size:   n,
	}

	return ms.saveMeta()
}

func (ms *FsBookStore) ListBooks() ([]httpexporter.BookInfo, error) {
	books := make([]httpexporter.BookInfo, 0, len(ms.meta))
	for _, info := range ms.meta {
		books = append(books, info)
	}
	slices.SortFunc(books, func(a, b httpexporter.BookInfo) int {
		return strings.Compare(a.ID, b.ID)
	})
	return books, nil
}

func (ms *FsBookStore) DownloadBook(id string, w io.Writer) error {
	file, err := os.Open(filepath.Join(ms.dir, id))
	if err != nil {
		return fmt.Errorf("create data file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}

func (ms *FsBookStore) GetSize(id string) (int64, error) {
	panic("not impl")
}
