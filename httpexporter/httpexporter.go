package httpexporter

import (
	"io"
	"net/http"
	"net/netip"
	"net/url"

	"github.com/awryme/reddit-exporter/httpexporter/internal/api"
	"github.com/awryme/reddit-exporter/httpexporter/ui"
	"github.com/go-chi/chi/v5"
)

type BookInfo = api.BookInfo

type BookStore interface {
	ListBooks() ([]BookInfo, error)
	DownloadBook(id string, w io.Writer) error
	GetSize(id string) (int64, error)
}

type Exporter interface {
	ExportURL(urls ...*url.URL) ([]string, error)
}

type Service struct {
	listen   netip.AddrPort
	store    BookStore
	exporter Exporter
}

func New(listen netip.AddrPort, store BookStore, exporter Exporter) *Service {
	return &Service{listen, store, exporter}
}

func (svc *Service) Run() error {
	router := chi.NewRouter()

	ui := ui.New(svc.exporter, svc.store)
	router.Group(ui.Handle)

	srv := http.Server{
		Addr:    svc.listen.String(),
		Handler: router,
	}
	return srv.ListenAndServe()
}
