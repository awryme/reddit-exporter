package redditexporter

import (
	"fmt"
	"io"
	"net/url"

	"github.com/awryme/reddit-exporter/redditexporter/internal/api"
	"github.com/awryme/reddit-exporter/redditexporter/internal/bufpool"
	"github.com/oklog/ulid/v2"
)

type Post = api.Post

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

type Exporter struct {
	client  RedditClient
	encoder BookEncoder
	store   BookStore
}

func New(client RedditClient, encoder BookEncoder, store BookStore) *Exporter {
	return &Exporter{client, encoder, store}
}

func (ex *Exporter) ExportURL(urls ...*url.URL) ([]string, error) {
	ids := make([]string, 0, len(urls))
	for _, u := range urls {
		id, err := ex.exportURL(u)
		if err != nil {
			return nil, fmt.Errorf("export url '%v': %w", u, err)
		}

		ids = append(ids, id)
	}
	return ids, nil
}

func (ex *Exporter) exportURL(u *url.URL) (string, error) {
	id := ulid.Make().String()

	post, err := ex.client.GetPostFromURL(u)
	if err != nil {
		return "", fmt.Errorf("download reddit post from %s: %w", u.String(), err)
	}

	fmt.Println("got post:", post.Title, len(post.Html))

	buf := bufpool.Get()
	defer buf.Close()

	err = ex.encoder.Encode(post, buf)
	if err != nil {
		return "", fmt.Errorf("encode post: %w", err)
	}

	format := ex.encoder.Format()
	err = ex.store.SaveBook(id, post.Title, format, buf)
	if err != nil {
		return "", fmt.Errorf("save book: %w", err)
	}
	return id, nil
}
