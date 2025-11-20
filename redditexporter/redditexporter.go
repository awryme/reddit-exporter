package redditexporter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/awryme/reddit-exporter/pkg/bufpool"
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
		GetPostByID(ctx context.Context, subreddit, id string) (*Post, error)
		GetCommentByID(ctx context.Context, subreddit, id string) (*Comment, error)
		DownloadImage(ctx context.Context, info ImageInfo, buf io.Writer) error
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
		return ex.exportComment(ctx, urlInfo.Subreddit, urlInfo.CommentID, resp)
	}

	return ex.exportPost(ctx, urlInfo.Subreddit, urlInfo.PostID, resp)
}

func (ex *Exporter) exportPost(ctx context.Context, subreddit, postID string, resp *Response) error {
	post, err := ex.client.GetPostByID(ctx, subreddit, postID)
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

func (ex *Exporter) exportComment(ctx context.Context, subreddit, commentID string, resp *Response) error {
	comment, err := ex.client.GetCommentByID(ctx, subreddit, commentID)
	if err != nil {
		return fmt.Errorf("get comment by id: %w", err)
	}
	if len(comment.Images) == 0 {
		return fmt.Errorf("no images url in comment")
	}

	for _, info := range comment.Images {
		// todo: add image.String func, log info, use in errorf
		buf := bytes.NewBuffer(nil)
		err := ex.client.DownloadImage(ctx, info, buf)
		if err != nil {
			return fmt.Errorf("download image (name = %s, url = %s): %w", info.Name, info.Url, err)
		}

		id := ulid.Make().String()
		err = ex.imagestore.SaveImage(id, info.Name, buf)
		if err != nil {
			return fmt.Errorf("save image: %w", err)
		}

		resp.ImageIds = append(resp.ImageIds, id)
	}

	return nil
}
