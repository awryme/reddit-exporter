package main

import (
	"github.com/alecthomas/kong"
)

const (
	tokenFileName = "token"
	credsFileName = "creds"
)

var App struct {
	Auth   AuthCmd   `cmd:"" help:"authorize reddit app and retreive token"`
	Export ExportCmd `cmd:"" help:"export reddit post as book"`
}

func main() {
	ctx := kong.Parse(
		&App,
		kong.DefaultEnvars("REDDIT_EXPORTER"),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
