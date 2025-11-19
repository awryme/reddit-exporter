package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awryme/reddit-exporter/bookencoding"
	"github.com/awryme/reddit-exporter/redditclient"
	"github.com/awryme/reddit-exporter/redditexporter"
	"github.com/awryme/reddit-exporter/redditexporter/bookstore"
	"github.com/awryme/reddit-exporter/redditexporter/imagestore"
	"github.com/awryme/slogf"
)

type ExportCmd struct {
	Urls []string `arg:"" help:"urls to reddit posts, can be in @file format"`

	Dir        string `help:"dir to store books and images" default:".data"`
	SecretsDir string `type:"path" help:"dir to cache auth token and store creds" default:"~/.reddit-exporter/"`
}

func (cmd *ExportCmd) Run() error {
	ctx := context.Background()

	log := slogf.DefaultHandler(os.Stdout)

	creds, err := ReadCredsFromFile(filepath.Join(cmd.SecretsDir, credsFileName))
	if err != nil {
		return fmt.Errorf("read reddit creds, run 'auth' command: %w", err)
	}

	tokenfile := filepath.Join(cmd.SecretsDir, tokenFileName)
	tokenstore := redditclient.NewFileTokenStore(tokenfile)

	bookStore, err := bookstore.NewBasicFS(filepath.Join(cmd.Dir, "books"))
	if err != nil {
		return fmt.Errorf("create book file store: %w", err)
	}

	imageStore, err := imagestore.NewBasicFS(filepath.Join(cmd.Dir, "images"))
	if err != nil {
		return fmt.Errorf("create book file store: %w", err)
	}

	exporter := redditexporter.New(
		redditclient.New(log, creds.ClientID, creds.ClientSecret, tokenstore),
		bookencoding.NewEpub(),
		bookStore,
		imageStore,
	)

	urls, err := parseUrls(cmd.Urls)
	if err != nil {
		return fmt.Errorf("parse input urls: %w", err)
	}

	_, err = exporter.ExportURLs(ctx, urls...)
	return err
}

// custom parse to interpret '@ signed files' @
func parseUrls(urlstrs []string) ([]string, error) {
	urls := make([]string, 0, len(urlstrs))
	for _, u := range urlstrs {
		parsedUrls, err := parseUrlstr(u)
		if err != nil {
			return nil, err
		}
		urls = append(urls, parsedUrls...)
	}
	return urls, nil
}

func parseUrlstr(urlstr string) ([]string, error) {
	filename, ok := strings.CutPrefix(urlstr, "@")
	if ok {
		return parseFile(filename)
	}

	return []string{urlstr}, nil
}

func parseFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file @%s: %w", filename, err)
	}
	defer file.Close()

	urls := make([]string, 0, 1)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parsedUrls, err := parseUrlstr(line)
		if err != nil {
			return nil, err
		}
		urls = append(urls, parsedUrls...)
	}
	return urls, nil
}
