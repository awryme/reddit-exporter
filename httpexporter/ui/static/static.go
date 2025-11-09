package static

import (
	"embed"
	"net/http"

	"github.com/awryme/reddit-exporter/httpexporter/internal/routes"
)

//go:embed *
var files embed.FS

func Handler() (string, http.Handler) {
	return routes.Static, http.StripPrefix(routes.Static, http.FileServerFS(files))
}
