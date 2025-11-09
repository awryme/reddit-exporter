package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/awryme/reddit-exporter/bookencoding"
	"github.com/awryme/reddit-exporter/redditclient"
	"github.com/awryme/reddit-exporter/redditexporter"
	"github.com/awryme/slogf"
)

type ExportCmd struct {
	Urls []string `arg:"" help:"urls to reddit posts, can be in @file format"`

	Dir        string `help:"dir to store books" default:".data/books"`
	SecretsDir string `type:"path" help:"dir to cache auth token and store creds" default:"~/.reddit-exporter/"`
}

func (cmd *ExportCmd) Run() error {
	log := slogf.DefaultHandler(os.Stdout)

	creds, err := ReadCredsFromFile(filepath.Join(cmd.SecretsDir, credsFileName))
	if err != nil {
		return fmt.Errorf("read reddit creds, run 'auth' command: %w", err)
	}

	tokenfile := filepath.Join(cmd.SecretsDir, tokenFileName)
	tokenstore := redditclient.NewFileTokenStore(tokenfile)

	bookstore, err := redditexporter.NewBasicFsBookStore(cmd.Dir)
	if err != nil {
		return fmt.Errorf("create book file store: %w", err)
	}

	exporter := redditexporter.New(
		redditclient.New(log, creds.ClientID, creds.ClientSecret, tokenstore),
		bookencoding.NewEpub(),
		bookstore,
		redditexporter.NoOpImageStore,
	)

	urls, err := parseUrls(cmd.Urls)
	if err != nil {
		return fmt.Errorf("parse input urls: %w", err)
	}

	_, err = exporter.ExportURLs(urls...)
	return err
}

// custom parse to interpret '@ signed files' @
func parseUrls(urlstrs []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(urlstrs))
	for _, u := range urlstrs {
		parsedUrls, err := parseUrlstr(u)
		if err != nil {
			return nil, err
		}
		urls = append(urls, parsedUrls...)
	}
	return urls, nil
}

func parseUrlstr(urlstr string) ([]*url.URL, error) {
	filename, ok := strings.CutPrefix(urlstr, "@")
	if ok {
		return parseFile(filename)
	}

	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, fmt.Errorf("parse input url '%s': %w", urlstr, err)
	}
	return []*url.URL{u}, nil
}

func parseFile(filename string) ([]*url.URL, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("open file @%s: %w", filename, err)
	}
	defer file.Close()

	urls := make([]*url.URL, 0, 1)
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
