package render

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Ctx struct {
	w    http.ResponseWriter
	r    *http.Request
	path string
}

func New(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{w, r, r.URL.Path}
}

func (ctx *Ctx) Error(err error, msgs ...string) bool {
	if err == nil {
		return false
	}
	code := GetCode(err)
	errmsg := ctx.fmtError(err, msgs)
	http.Error(ctx.w, errmsg, code)
	return true
}

func (ctx *Ctx) fmtError(err error, msgs []string) string {
	var sb strings.Builder
	sb.WriteString(ctx.path)
	sb.WriteString(": ")

	for _, msg := range msgs {
		sb.WriteString(msg)
		sb.WriteString(": ")
	}

	sb.WriteString(err.Error())

	return sb.String()
}

func (ctx *Ctx) WR() (w http.ResponseWriter, r *http.Request) {
	return ctx.w, ctx.r
}

type Component interface {
	Render(w io.Writer) error
}

func (ctx *Ctx) Render(components ...Component) bool {
	var errgroup error
	for _, c := range components {
		err := c.Render(ctx.w)
		errgroup = errors.Join(errgroup, err)
	}
	return ctx.Error(errgroup, "render components")
}

func (ctx *Ctx) MustFormValue(name string, val *string) bool {
	// todo: parse explicitly
	*val = ctx.r.FormValue(name)
	if *val == "" {
		return ctx.Error(fmt.Errorf("empty form value for %s", name))
	}
	return false
}

func (ctx *Ctx) DecodeJsonBody(val any) bool {
	err := json.NewDecoder(ctx.r.Body).Decode(val)
	return ctx.Error(err, "decode json body")
}

func (ctx *Ctx) Query(name string) string {
	return ctx.r.URL.Query().Get(name)
}

func (ctx *Ctx) Context() context.Context {
	return ctx.r.Context()
}
