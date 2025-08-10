package ui

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/awryme/reddit-exporter/httpexporter/internal/api"
	"github.com/awryme/reddit-exporter/httpexporter/ui/static"
	"github.com/awryme/reddit-exporter/pkg/xhttp/render"
	"github.com/go-chi/chi/v5"
)

type BookStore interface {
	ListBooks() ([]api.BookInfo, error)
	DownloadBook(id string, w io.Writer) error
	GetSize(id string) (int64, error)
}

type Exporter interface {
	ExportURL(u ...*url.URL) ([]string, error)
}

type UI struct {
	exporter Exporter
	store    BookStore
}

func New(exporter Exporter, store BookStore) *UI {
	return &UI{exporter, store}
}

func (ui *UI) Handle(router chi.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("got", r.Method, r.URL.String())
			next.ServeHTTP(w, r)
		})
	})

	router.Mount(static.Handler())

	router.Method(ui.indexHandler())
	router.Method(ui.exportHandler())
	router.Method(ui.downloadHandler())
}

type HandleParams struct {
	Method  string
	Route   string
	Handler http.HandlerFunc
}

func (ui *UI) indexHandler() (string, string, http.HandlerFunc) {
	return http.MethodGet, api.RouteIndexPage, func(w http.ResponseWriter, r *http.Request) {
		ctx := render.New(w, r)

		books, err := ui.store.ListBooks()
		if ctx.Error(err, "list book") {
			return
		}

		ctx.Render(IndexPage(books))
	}
}

func (ui *UI) exportHandler() (string, string, http.HandlerFunc) {
	return http.MethodPost, api.RouteUiExport, handleExportUrls(ui.exporter, ui.store)
}

func (ui *UI) downloadHandler() (string, string, http.HandlerFunc) {
	route := api.FmtRouteDownload("{id}", "*")
	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := render.New(w, r)
		id := chi.URLParam(r, "id")

		// todo: check size, or remove GetSize()?
		// size, err := ui.store.GetSize(id)
		// if ctx.Error(err, "get size") {
		// 	return
		// }
		// w.Header().Set("Content-Length", fmt.Sprint(size))
		w.Header().Set("Content-Type", "application/epub+zip")
		err := ui.store.DownloadBook(id, w)
		if ctx.Error(err, "download book") {
			return
		}
	}

	return http.MethodGet, route, handler
}
