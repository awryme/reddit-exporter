package urlparser

import (
	"fmt"
	"net/url"
	"strings"
)

func SplitNewLine(raw string) ([]*url.URL, error) {
	rawURLs := strings.Split(raw, "\n")
	return ParseUrls(rawURLs)
}

func ParseUrls(rawURLs []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(rawURLs))
	for _, urlstr := range rawURLs {
		urlstr = strings.TrimSpace(urlstr)
		if urlstr == "" {
			continue
		}

		u, err := url.Parse(urlstr)
		if err != nil {
			return nil, fmt.Errorf("parse url: %w", err)
		}
		urls = append(urls, u)
	}

	return urls, nil
}
