package redditexporter

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type urlInfo struct {
	Subreddit string
	PostID    string
	CommentID string
}

func parseUrl(url string) (*urlInfo, error) {
	// cleanup path
	url = cleanUrl(url)

	fmtErr := func(s string, args ...any) error {
		if len(args) > 0 {
			s = fmt.Sprintf(s, args...)
		}

		return fmt.Errorf("cannot parse url '%s': %s", url, s)
	}

	const reddiUrlPrefix = "https://www.reddit.com/r/"

	_, path, ok := strings.Cut(url, reddiUrlPrefix)
	if !ok {
		return nil, fmtErr("no reddit url prefix (expected: %s)", reddiUrlPrefix)
	}

	subreddit, path, ok := strings.Cut(path, "/")
	if !ok {
		return nil, fmtErr("no subreddit in url")
	}

	urlType, path, ok := strings.Cut(path, "/")
	if !ok {
		return nil, fmtErr("no url type (/r/<sub>/<type>/id)")
	}

	// short link, resolve and parse it
	if urlType == "s" {
		resolvedUrl, err := resolveShortUrl(url)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve short url '%s': %w", url, err)
		}

		return parseUrl(resolvedUrl.String())
	}

	if urlType != "comments" {
		return nil, fmtErr("unknow url type, expected (/r/<sub>/<type>/id)")
	}

	postID, path, ok := strings.Cut(path, "/")
	if !ok {
		return nil, fmtErr("no reddit post id")
	}

	commentOrPost, commentID, _ := strings.Cut(path, "/")
	if commentOrPost == "comment" {
		return &urlInfo{
			Subreddit: subreddit,
			PostID:    postID,
			CommentID: commentID,
		}, nil
	}

	return &urlInfo{
		Subreddit: subreddit,
		PostID:    postID,
	}, nil
}

func cleanUrl(url string) string {
	url = strings.TrimSpace(url)
	url, _, _ = strings.Cut(url, "#")
	url, _, _ = strings.Cut(url, "?")
	url = strings.TrimSuffix(url, "/")

	return url
}

func resolveShortUrl(url string) (*url.URL, error) {
	cli := &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create new redirect request: %w", err)
	}

	res, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("make redirect request: %w", err)
	}
	defer res.Body.Close()

	newurl, err := res.Location()
	if err != nil {
		return nil, fmt.Errorf("resolve redirect response: %w", err)
	}

	return newurl, nil
}
