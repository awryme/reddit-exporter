package redditclient

import (
	"context"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"

	"github.com/awryme/reddit-exporter/pkg/xhttp"
)

// reddit domains
const (
	domainReddit       = "reddit.com"
	domainRedditWWW    = "www.reddit.com"
	domainRedditOauth  = "oauth.reddit.com"
	domainRedditImages = "i.redd.it"
)

type (
	Post = struct {
		Title string
		Html  string
	}

	ImageInfo = struct {
		Name string
		Url  string
	}

	Comment = struct {
		Images []ImageInfo
	}
)

type Client struct {
	httpClient *http.Client
	auth       *AuthService
}

func New(log slog.Handler, clientID string, clientSecret string, tokenStore TokenStore) *Client {
	auth := NewAuth(log, clientID, clientSecret, tokenStore)
	httpClient := xhttp.NewClient()
	return &Client{httpClient, auth}
}

func (cli *Client) GetPostByID(subreddit, id string) (*Post, error) {
	data, err := getListings[JsonPostData](cli, subreddit, KindPost, id)
	if err != nil {
		return nil, fmt.Errorf("get json post: %w", err)
	}
	return &Post{
		Title: data.Title,
		Html:  html.UnescapeString(data.Selfhtml),
	}, nil
}

func (cli *Client) GetCommentByID(subreddit, id string) (*Comment, error) {
	data, err := getListings[JsonCommentData](cli, subreddit, KindComment, id)
	if err != nil {
		return nil, fmt.Errorf("get json post: %w", err)
	}

	meta := data.MediaMetadata
	infos := make([]ImageInfo, 0, len(meta))
	for id, info := range meta {
		var name string
		switch info.Type {
		case "image/jpeg":
			name = fmt.Sprintf("%s.jpeg", id)
		}
		url := fmt.Sprintf("https://%s/%s", domainRedditImages, name)
		infos = append(infos, ImageInfo{
			Name: name,
			Url:  url,
		})
	}

	return &Comment{
		Images: infos,
	}, nil
}

func (cli *Client) DownloadImage(ctx context.Context, info ImageInfo, buf io.Writer) error {
	client := xhttp.NewClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, info.Url, nil)
	if err != nil {
		return fmt.Errorf("create http request for reddit image: %w", err)
	}

	req.Header.Set("User-Agent", "reddit-exporter/v1.2")
	req.Header.Set("Accept", "image/*")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send http request for reddit image: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("send http request for reddit image: bad status %d (%s)", res.StatusCode, res.Status)
	}

	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return fmt.Errorf("copy response body to buf: %w", err)
	}

	return nil
}

func getListings[Data any](cli *Client, subreddit string, kind JsonKind, id string) (Data, error) {
	var data Data
	fullID := fmt.Sprintf("%s_%s", kind, id)

	token, err := cli.auth.Auth()
	if err != nil {
		return data, fmt.Errorf("auth new token: %w", err)
	}

	url := fmt.Sprintf("https://%s/r/%s/api/info?id=%s", domainRedditOauth, subreddit, fullID)
	listing, err := jsonGetPost[Data](cli.httpClient, url, token)
	if err != nil {
		return data, fmt.Errorf("get json post: %w", err)
	}

	if listing.Kind != KindListing {
		return data, fmt.Errorf("listing kind is wrong (expected = %s, got = %s)", KindListing, listing.Kind)
	}

	for _, post := range listing.Data.Children {
		if post.Kind != kind {
			continue
		}

		return post.Data, nil
	}

	return data, fmt.Errorf("post %s not found in response", fullID)
}
