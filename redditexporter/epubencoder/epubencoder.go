package epubencoder

import (
	"fmt"
	"io"

	"github.com/awryme/reddit-exporter/redditexporter/internal/api"
	"github.com/go-shiori/go-epub"
)

type Post = api.Post

type Encoder struct {
}

func New() Encoder {
	return Encoder{}
}
func (enc Encoder) Format() string {
	return "epub"
}

func (enc Encoder) Encode(post Post, out io.Writer) error {
	book, err := epub.NewEpub(post.Title)
	if err != nil {
		return fmt.Errorf("create epub '%s': %w", post.Title, err)
	}
	_, err = book.AddSection(post.Html, post.Title, "main.xhtml", "")
	if err != nil {
		return fmt.Errorf("add section to epub '%s': %w", post.Title, err)
	}
	_, err = book.WriteTo(out)
	if err != nil {
		return fmt.Errorf("write epub '%s' to file: %w", post.Title, err)
	}
	return nil
}
