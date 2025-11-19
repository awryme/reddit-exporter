package bookencoding

import (
	"fmt"
	"io"

	"github.com/go-shiori/go-epub"
)

type (
	Book = struct {
		Title string
		Html  string
	}
)

type Epub struct{}

func NewEpub() Epub {
	return Epub{}
}

func (e Epub) Format() string {
	return "epub"
}

func (e Epub) Encode(info *Book, out io.Writer) error {
	book, err := epub.NewEpub(info.Title)
	if err != nil {
		return fmt.Errorf("create epub '%s': %w", info.Title, err)
	}
	_, err = book.AddSection(info.Html, info.Title, "main.xhtml", "")
	if err != nil {
		return fmt.Errorf("add section to epub '%s': %w", info.Title, err)
	}
	_, err = book.WriteTo(out)
	if err != nil {
		return fmt.Errorf("write epub '%s' to file: %w", info.Title, err)
	}
	return nil
}
