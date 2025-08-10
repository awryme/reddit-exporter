package static

import (
	"embed"
	"net/http"

	"github.com/awryme/reddit-exporter/httpexporter/internal/api"
)

//go:embed *
var files embed.FS

func Handler() (string, http.Handler) {
	return api.RouteStatic, http.StripPrefix(api.RouteStatic, http.FileServerFS(files))
}
