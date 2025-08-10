package main

import (
	"fmt"
	"log/slog"
	"net/netip"
	"os"

	"github.com/alecthomas/kong"
	"github.com/awryme/reddit-exporter/httpexporter"
	"github.com/awryme/reddit-exporter/redditexporter"
	"github.com/awryme/reddit-exporter/redditexporter/epubencoder"
	"github.com/awryme/reddit-exporter/redditexporter/redditclient"
	"github.com/awryme/slogf"
)

type App struct {
	Port     int    `default:"8080"`
	Dir      string `help:"dir to store books" default:".data/exporter-server/books/"`
	BasicDir string `help:"dir to store a flat basic list of books"`

	ClientID     string `required:"" help:"reddit app client_id"`
	ClientSecret string `required:"" help:"reddit app client_secret"`
}

func (app *App) Run() error {
	log := slogf.DefaultHandler(os.Stdout)
	logf := slogf.New(log)

	listen := netip.MustParseAddrPort(fmt.Sprintf("0.0.0.0:%d", app.Port))

	client := redditclient.New(log,
		app.ClientID,
		app.ClientSecret,
		redditclient.NewMemoryTokenStore(),
	)

	fsStore, err := NewFsBookStore(app.Dir)
	if err != nil {
		return fmt.Errorf("create book filestore: %w", err)
	}
	var store redditexporter.BookStore = fsStore

	if app.BasicDir != "" {
		basicFsStore, err := redditexporter.NewBasicFsBookStore(app.BasicDir)
		if err != nil {
			return fmt.Errorf("init basic fs books store")
		}
		store = redditexporter.NewMultiStore(map[string]redditexporter.BookStore{
			"http_fs":  fsStore,
			"basic_fs": basicFsStore,
		})
		logf("using basic fs store", slog.String("dir", app.BasicDir))
	}

	exporter := redditexporter.New(client, epubencoder.New(), store)

	logf("running", slog.String("addr", listen.String()))
	svc := httpexporter.New(
		listen,
		fsStore,
		exporter,
	)
	return svc.Run()
}

func main() {
	ctx := kong.Parse(&App{}, kong.DefaultEnvars(""))

	ctx.FatalIfErrorf(ctx.Run())
}
