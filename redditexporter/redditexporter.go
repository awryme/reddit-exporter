package redditexporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/awryme/reddit-exporter/pkg/bufpool"
	"github.com/awryme/reddit-exporter/pkg/xhttp"
	"github.com/oklog/ulid/v2"
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

	RedditClient interface {
		GetPostByID(subreddit, id string) (*Post, error)
		GetCommentByID(subreddit, id string) (*Comment, error)
	}
)

type (
	Book = struct {
		Title string
		Html  string
	}

	BookEncoder interface {
		Encode(post *Book, output io.Writer) error
		Format() string
	}
)

type (
	BookStore interface {
		SaveBook(id, title, format string, data io.Reader) error
	}

	ImageStore interface {
		SaveImage(id, name string, data io.Reader) error
	}
)

type Exporter struct {
	client      RedditClient
	bookEncoder BookEncoder
	bookstore   BookStore
	imagestore  ImageStore
}

func New(client RedditClient, encoder BookEncoder, bookstore BookStore, imagestore ImageStore) *Exporter {
	return &Exporter{client, encoder, bookstore, imagestore}
}

type Response = struct {
	BookIds  []string
	ImageIds []string
}

func (ex *Exporter) ExportURLs(ctx context.Context, urls ...string) (*Response, error) {
	// todo: add logs for exporting: found image/post id, downloading url...

	resp := &Response{
		BookIds:  make([]string, 0, len(urls)),
		ImageIds: make([]string, 0, len(urls)),
	}

	for _, url := range urls {
		url = strings.TrimSpace(url)
		// ignore empty lines
		if url == "" {
			continue
		}

		err := ex.exportURL(ctx, url, resp)
		if err != nil {
			return resp, fmt.Errorf("export url '%v': %w", url, err)
		}
	}

	return resp, nil
}

func (ex *Exporter) exportURL(ctx context.Context, url string, resp *Response) error {
	urlInfo, err := parseUrl(url)
	if err != nil {
		return err
	}

	if urlInfo.CommentID != "" {
		return ex.exportComment(ctx, urlInfo.Subreddit, urlInfo.PostID, urlInfo.CommentID, resp)
	}

	return ex.exportPost(ctx, urlInfo.Subreddit, urlInfo.PostID, resp)
}

func (ex *Exporter) exportPost(ctx context.Context, subreddit, postID string, resp *Response) error {
	post, err := ex.client.GetPostByID(subreddit, postID)
	if err != nil {
		return fmt.Errorf("download reddit post r/%s/%s: %w", subreddit, postID, err)
	}

	buf := bufpool.Get()
	defer buf.Close()

	// todo: replace to pointer in interface
	err = ex.bookEncoder.Encode(post, buf)
	if err != nil {
		return fmt.Errorf("encode post: %w", err)
	}

	format := ex.bookEncoder.Format()

	id := ulid.Make().String()
	err = ex.bookstore.SaveBook(id, post.Title, format, buf)
	if err != nil {
		return fmt.Errorf("save book: %w", err)
	}

	resp.BookIds = append(resp.BookIds, id)
	return nil
}

func (ex *Exporter) exportComment(ctx context.Context, subreddit, postID, commentID string, resp *Response) error {
	comment, err := ex.client.GetCommentByID(subreddit, commentID)
	if err != nil {
		return fmt.Errorf("get comment by id: %w", err)
	}
	if len(comment.Images) == 0 {
		return fmt.Errorf("no images url in comment")
	}

	for _, info := range comment.Images {
		err := ex.downloadImage(ctx, info, resp)
		if err != nil {
			return fmt.Errorf("download image (name = %s, url = %s): %w", info.Name, info.Url, err)
		}
	}

	return nil
}

func (ex *Exporter) downloadImage(ctx context.Context, info ImageInfo, resp *Response) error {
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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read http response body for reddit image: bad status %d (%s)", res.StatusCode, res.Status)
	}

	id := ulid.Make().String()
	err = ex.imagestore.SaveImage(id, info.Name, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("save book: %w", err)
	}

	resp.ImageIds = append(resp.ImageIds, id)
	return nil
}
