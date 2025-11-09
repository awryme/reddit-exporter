package routes

import (
	"fmt"
)

// routes
const (
	IndexPage = "/"
	Static    = "/static"
	Download  = "/download"

	UiExport = "/ui/v1/export"
)

func FmtStatic(file string) string {
	return fmt.Sprintf("%s/%s", Static, file)
}

func FmtDownload(id string, filename string) string {
	return fmt.Sprintf("%s/%s/%s", Download, id, filename)
}
