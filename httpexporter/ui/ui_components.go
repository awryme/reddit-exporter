package ui

import (
	"strings"

	"github.com/awryme/reddit-exporter/httpexporter/internal/routes"
	"github.com/awryme/reddit-exporter/httpexporter/ui/css"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	c "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func component(id string) Node {
	return Group{
		ID(id),
		hx.SwapOOB("true"),
	}
}

func IndexPage(books []BookInfo) Node {
	return c.HTML5(c.HTML5Props{
		Title:       "Reddit exporter",
		Description: "reddit exporter service",
		Language:    "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href(routes.FmtStatic("matcha.css"))),
			Script(Type("module"), Src(routes.FmtStatic("htmx.min.js"))),
			Meta(Name("htmx-config"), Content(`{"allowNestedOobSwaps": false, "defaultSwapStyle": "none"}`)),
		},

		Body: []Node{
			Nav(
				Text("Reddit exporter"),
			),
			statusBar(),
			bookInput(),
			bookList(books),
		},
	})
}

func statusBar(text ...string) Node {
	return Div(
		component("error_notification"),
		If(len(text) > 0,
			Group{
				css.BgDanger().Danger(),
				Text("Status:"),
				Span(Text(strings.Join(text, " "))),
			},
		),
	)
}

func bookInput() Node {
	return Div(
		component("books_input"),
		H1(Text("Add posts")),
		Textarea(
			Name(exportUrlsName),
			Style("width: 80%"),
			Rows("3"),
		),
		Button(
			Text("Add"),
			Type("submit"),
			hx.Post(routes.UiExport),
			hx.Include("previous textarea"),
		),
	)
}

func bookList(books []BookInfo) Node {
	bookElem := func(book BookInfo) Node {
		filename := book.Title + "." + book.Format
		return Div(
			A(
				Href(routes.FmtDownload(book.ID, filename)),
				Target("_blank"),
				Download(filename),
				Text(filename),
			),
		)
	}

	return Div(
		component("book_list"),
		H1(
			Text("Books"),
		),
		Div(
			css.Flex().Column(),
			Map(books, bookElem),
		),
	)
}
