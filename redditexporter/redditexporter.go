package redditexporter

import (
	"fmt"
	"io"
	"net/url"

	"github.com/awryme/reddit-exporter/redditexporter/internal/bufpool"
	"github.com/oklog/ulid/v2"
)

type Post = struct {
	Title string
	Html  string
}

type RedditClient interface {
	GetPostFromURL(u *url.URL) (Post, error)
}

type BookEncoder interface {
	Encode(post Post, output io.Writer) error
	Format() string
}

type BookStore interface {
	SaveBook(id, title, format string, data io.Reader) error
}

type ImageStore interface {
	SaveImage(id, name, ext string, data io.Reader) error
}

type Response struct {
	BookIds  []string
	ImageIds []string
}

type Exporter struct {
	client     RedditClient
	encoder    BookEncoder
	bookstore  BookStore
	imagestore ImageStore
}

func New(client RedditClient, encoder BookEncoder, bookstore BookStore, imagestore ImageStore) *Exporter {
	return &Exporter{client, encoder, bookstore, imagestore}
}

func (ex *Exporter) ExportURLs(urls ...*url.URL) (resp Response, err error) {
	resp.BookIds = make([]string, 0, len(urls))
	resp.ImageIds = make([]string, 0, len(urls))

	for _, u := range urls {
		respType, id, err := ex.exportURL(u)
		if err != nil {
			return resp, fmt.Errorf("export url '%v': %w", u, err)
		}
		switch respType {
		case responseBook:
			resp.BookIds = append(resp.BookIds, id)
		case responseImage:
			resp.ImageIds = append(resp.ImageIds, id)
		}
	}

	return resp, nil
}

type responseType int

const (
	responseNone = iota
	responseBook
	responseImage
)

func (ex *Exporter) exportURL(u *url.URL) (responseType, string, error) {
	id := ulid.Make().String()

	post, err := ex.client.GetPostFromURL(u)
	if err != nil {
		return responseNone, "", fmt.Errorf("download reddit post from %s: %w", u.String(), err)
	}

	buf := bufpool.Get()
	defer buf.Close()

	//todo: save image

	err = ex.encoder.Encode(post, buf)
	if err != nil {
		return responseNone, "", fmt.Errorf("encode post: %w", err)
	}

	format := ex.encoder.Format()
	err = ex.bookstore.SaveBook(id, post.Title, format, buf)
	if err != nil {
		return responseBook, "", fmt.Errorf("save book: %w", err)
	}
	return responseBook, id, nil
}
