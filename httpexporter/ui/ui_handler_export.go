package ui

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/awryme/reddit-exporter/pkg/urlparser"
	"github.com/awryme/reddit-exporter/pkg/xhttp/render"
)

const exportUrlsName = "urls"

func handleExportUrls(exporter Exporter, store BookStore) http.HandlerFunc {
	exportAndList := func(urlsData string) ([]BookInfo, error) {
		urls, err := urlparser.SplitNewLine(urlsData)
		if err != nil {
			return nil, fmt.Errorf("parse urls text: %w", err)
		}

		_, err = exporter.ExportURLs(urls...)
		if err != nil {
			return nil, fmt.Errorf("export reddit url: %w", err)
		}

		books, err := store.ListBooks()
		if err != nil {
			return nil, fmt.Errorf("list stored books: %w", err)
		}
		return books, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := render.New(w, r)

		urlsData := strings.TrimSpace(r.PostFormValue(exportUrlsName))
		if urlsData == "" {
			ctx.Render(statusBar("empty urls data"))
			return
		}

		books, err := exportAndList(urlsData)
		if err != nil {
			ctx.Render(
				statusBar(fmt.Sprintf("export books: %v", err.Error())),
			)

			return
		}

		ctx.Render(
			bookList(books),
			statusBar(),
			bookInput(),
		)
	}
}
