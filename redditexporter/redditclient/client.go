package redditclient

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awryme/reddit-exporter/redditexporter/internal/api"
)

// reddit domains
const (
	domainReddit      = "reddit.com"
	domainRedditWWW   = "www.reddit.com"
	domainRedditOauth = "oauth.reddit.com"
)

type Post = api.Post

type Client struct {
	httpClient *http.Client
	auth       *AuthService
}

func New(log slog.Handler, clientID string, clientSecret string, tokenStore TokenStore) *Client {
	auth := NewAuth(log, clientID, clientSecret, tokenStore)
	httpClient := newHttpClient()
	return &Client{httpClient, auth}
}

func (cli *Client) GetPostFromURL(u *url.URL) (Post, error) {
	u, err := transformURL(u)
	if err != nil {
		return Post{}, fmt.Errorf("transform reddit url: %w", err)
	}
	token, err := cli.auth.Auth()
	if err != nil {
		return Post{}, fmt.Errorf("auth new token: %w", err)
	}

	listings, err := jsonGetPost(cli.httpClient, u, token)
	if err != nil {
		return Post{}, fmt.Errorf("get json post: %w", err)
	}
	for _, list := range listings {
		if list.Kind != KindListing {
			return Post{}, fmt.Errorf("listing kind is wrong (expected = %s, got = %s)", KindListing, list.Kind)
		}
		for _, post := range list.Data.Children {
			if post.Kind == KingPost {
				data := post.Data
				rawhtml := html.UnescapeString(data.Selfhtml)
				return Post{
					Title: data.Title,
					Html:  rawhtml,
				}, nil
			}
		}
	}
	return Post{}, fmt.Errorf("post not found in response")
}

func transformURL(u *url.URL) (*url.URL, error) {
	switch u.Hostname() {
	case domainRedditOauth:
		return u, nil
	case domainReddit, domainRedditWWW:
		return transformRedditURL(u)
	}
	return nil, fmt.Errorf("unknow domain: %s", u.Hostname())
}

func transformRedditURL(u *url.URL) (*url.URL, error) {
	path := strings.Trim(u.Path, "/")
	errUnknownUrl := fmt.Errorf("unknown path url format (path = %s)", path)

	parts := strings.Split(path, "/")

	// r/HFY/comments/1kcjsc3/oocs_into_a_wider_galaxy_part_321/
	if len(parts) < 4 {
		return nil, fmt.Errorf("url parts.len != 4: %w", errUnknownUrl)
	}
	if parts[0] != "r" {
		return nil, fmt.Errorf("url[0] != 'r': %w", errUnknownUrl)
	}

	switch parts[2] {
	case "comments":
	case "s":
		newurl, err := resolveShortUrl(u)
		if err != nil {
			return nil, fmt.Errorf("resolve short url: %w", err)
		}
		transformed, err := transformURL(newurl)
		if err != nil {
			return nil, fmt.Errorf("transform redirect url: %w", err)
		}
		return transformed, nil
	default:
		return nil, fmt.Errorf("url[2] is incorrect: %w", errUnknownUrl)
	}

	subName := parts[1]
	postID := parts[3]

	return url.Parse(fmt.Sprintf("https://%s/r/%s/comments/%s", domainRedditOauth, subName, postID))
}

func resolveShortUrl(u *url.URL) (*url.URL, error) {
	cli := &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
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
