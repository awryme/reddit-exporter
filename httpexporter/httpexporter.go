package httpexporter

import (
	"io"
	"net/http"
	"net/netip"
	"net/url"

	"github.com/awryme/reddit-exporter/httpexporter/ui"
	"github.com/awryme/reddit-exporter/redditexporter"
	"github.com/go-chi/chi/v5"
)

type BookInfo = struct {
	ID     string
	Title  string
	Format string
	Size   int64
}

type BookStore interface {
	ListBooks() ([]BookInfo, error)
	DownloadBook(id string, w io.Writer) error
	GetSize(id string) (int64, error)
}

type Exporter interface {
	ExportURLs(urls ...*url.URL) (resp redditexporter.Response, err error)
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
