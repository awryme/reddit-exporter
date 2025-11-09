package redditclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JsonKind string

const (
	KindListing JsonKind = "Listing"
	KindPost    JsonKind = "t3"
)

type JsonListing struct {
	Kind JsonKind
	Data struct {
		Children []JsonPost
	}
}

type JsonPost struct {
	Kind JsonKind
	Data struct {
		Title     string
		Selftext  string
		Selfhtml  string `json:"selftext_html"`
		Id        string
		Thumbnail string
		Url       string
	}
}

func jsonGetPost(httpClient *http.Client, url, token string) ([]JsonListing, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "reddit-exporter/cli/v1.1")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get response for url '%s': %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code for url '%s': %d (%s)", url, res.StatusCode, res.Status)
	}

	return jsonDecode(res.Body)
}

func jsonDecode(data io.Reader) ([]JsonListing, error) {
	postResps := make([]JsonListing, 0, 1)
	err := json.NewDecoder(data).Decode(&postResps)
	if err != nil {
		return nil, fmt.Errorf("decode response body from json: %w", err)
	}
	return postResps, nil
}
