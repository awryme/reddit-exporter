package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/awryme/reddit-exporter/redditexporter/redditclient"
	"github.com/awryme/slogf"
)

type AuthCmd struct {
	ClientID   string `help:"reddit client_id, if empty will be prompted from stdin"`
	SecretsDir string `type:"path" help:"dir to cache auth token and store creds" default:"~/.reddit-exporter/"`
}

func (cmd *AuthCmd) Run() error {
	log := slogf.DefaultHandler(os.Stdout)
	logf := slogf.New(log)

	credsfile := filepath.Join(cmd.SecretsDir, credsFileName)
	tokenfile := filepath.Join(cmd.SecretsDir, tokenFileName)

	creds, err := SaveCredsToFile(credsfile, cmd.ClientID)
	if err != nil {
		return fmt.Errorf("save cred file: %w", err)
	}

	logf("saved creds", slog.String("cred_file", credsfile))

	tokenstore := redditclient.NewFileTokenStore(tokenfile)
	auth := redditclient.NewAuth(log, creds.ClientID, creds.ClientSecret, tokenstore)
	_, err = auth.Auth()
	if err != nil {
		return fmt.Errorf("auth app on reddit: %w", err)
	}

	logf("auth successful", slog.String("cred_file", credsfile), slog.String("token_file", tokenfile))

	return nil
}
