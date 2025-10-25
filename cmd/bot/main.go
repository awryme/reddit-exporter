package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/awryme/reddit-exporter/redditexporter"
	"github.com/awryme/reddit-exporter/redditexporter/epubencoder"
	"github.com/awryme/reddit-exporter/redditexporter/redditclient"
	"github.com/awryme/slogf"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type App struct {
	ClientID     string `help:"reddit app client_id" required:"" `
	ClientSecret string `help:"reddit app client_secret" required:"" `
	BotToken     string `help:"tg bot token from botfather" required:"" `
	BasicDir     string `help:"dir to store books"`
}

func (app *App) Run() error {
	ctx := context.Background()

	log := slogf.DefaultHandler(os.Stdout)
	logf := slogf.New(log)

	client := redditclient.New(
		log,
		app.ClientID,
		app.ClientSecret,
		redditclient.NewMemoryTokenStore(),
	)

	memstore := NewMemoryBookStore()
	var store redditexporter.BookStore = memstore
	if app.BasicDir != "" {
		basicFsStore, err := redditexporter.NewBasicFsBookStore(app.BasicDir)
		if err != nil {
			return fmt.Errorf("init basic fs books store")
		}
		store = redditexporter.NewMultiStore(map[string]redditexporter.BookStore{
			"memory":   memstore,
			"basic_fs": basicFsStore,
		})
		logf("using basic fs store", slog.String("dir", app.BasicDir))
	}
	exp := redditexporter.New(client, epubencoder.New(), store)

	b, err := bot.New(app.BotToken,
		bot.WithDefaultHandler(handler(logf, exp, memstore)),
		bot.WithErrorsHandler(func(err error) {
			logf("internal error from bot", slogf.Error(err))
		}),
	)
	if err != nil {
		return fmt.Errorf("create new bot: %w", err)
	}

	b.Start(ctx)
	return nil
}

func main() {
	ctx := kong.Parse(&App{}, kong.DefaultEnvars(""))

	ctx.FatalIfErrorf(ctx.Run())
}

func handler(logf slogf.Logf, exp *redditexporter.Exporter, store *MemoryBookStore) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update == nil {
			logf("error: update is nil")
			return
		}
		msg := firstNonNil(update.Message, update.EditedMessage, update.BusinessMessage, update.EditedBusinessMessage)
		if msg == nil {
			logf("error: update message is nil", slog.Int64("update_id", update.ID))
			return
		}

		sendText := func(text string) {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text:   text,
			})
			if err != nil {
				logf("send response", slog.String("text", text), slogf.Error(err))
			}
		}

		lines := strings.Split(msg.Text, "\n")
		urls := make([]*url.URL, 0, len(lines))
		for _, line := range lines {
			text := strings.TrimSpace(line)
			if text == "" || !strings.HasPrefix(text, "http") {
				continue
			}
			u, err := url.Parse(text)
			if err != nil {
				sendText(fmt.Sprintf("error: cannot parse url from message: %v", err))
				return
			}
			urls = append(urls, u)
		}

		ids, err := exp.ExportURL(urls...)
		if err != nil {
			sendText(fmt.Sprintf("error: cannot export urls: %v", err))
			return
		}

		for _, id := range ids {
			book, ok := store.GetBook(id)
			if !ok {
				sendText(fmt.Sprintf("error: stored book with id %s not found", id))
				return
			}

			_, err := b.SendDocument(ctx, &bot.SendDocumentParams{
				ChatID: msg.Chat.ID,
				Document: &models.InputFileUpload{
					Filename: book.Title + "." + book.Format,
					Data:     book.Data,
				},
			})
			if err != nil {
				sendText(fmt.Sprintf("error: cannot send book with id %s: %v", id, err))
				return
			}
		}

		store.DeleteBooks(ids)

		sendText("Done. ")
	}
}

func firstNonNil[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}

	return nil
}
