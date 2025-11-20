package redditclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JsonKind string

const (
	KindListing JsonKind = "Listing"
	KindComment JsonKind = "t1"
	KindPost    JsonKind = "t3"
)

type JsonPostData struct {
	Title    string
	Selftext string
	Selfhtml string `json:"selftext_html"`
	Id       string
}

type JsonCommentData struct {
	MediaMetadata map[string]struct {
		Type string `json:"m"`
	} `json:"media_metadata"`
}

type JsonPost[Data any] struct {
	Kind JsonKind
	Data Data
}

type JsonListing[Data any] struct {
	Kind JsonKind
	Data struct {
		Children []JsonPost[Data]
	}
}

func jsonGetPost[Data any](ctx context.Context, httpClient *http.Client, url, token string) (*JsonListing[Data], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "reddit-exporter/v1.2")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response for url '%s': %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for url '%s': %d (%s)", url, res.StatusCode, res.Status)
	}

	return jsonDecode[Data](res.Body)
}

func jsonDecode[Data any](data io.Reader) (*JsonListing[Data], error) {
	var resp JsonListing[Data]
	err := json.NewDecoder(data).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("decode response body from json: %w", err)
	}
	return &resp, nil
}
