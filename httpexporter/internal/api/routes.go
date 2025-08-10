package api

import (
	"fmt"
)

// routes
const (
	RouteIndexPage = "/"
	RouteStatic    = "/static"
	RouteDownload  = "/download"

	RouteUiExport = "/ui/v1/export"
)

func FmtRouteStatic(file string) string {
	return fmt.Sprintf("%s/%s", RouteStatic, file)
}

func FmtRouteDownload(id string, filename string) string {
	return fmt.Sprintf("%s/%s/%s", RouteDownload, id, filename)
}
