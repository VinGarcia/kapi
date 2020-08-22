package main

import (
	"testing"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

var err error
var pathParam string
var headerParam string

var weight = 10

func BenchmarkAdapter(b *testing.B) {
	adapted := adapt(func(ctx *routing.Context, args struct {
		PathParam   string `path:"path-param"`
		HeaderParam string `header:"header-param"`
	}) error {
		pathParam = args.PathParam
		headerParam = args.HeaderParam
		for i := 0; i < weight; i++ {
			headerParam = headerParam + "0"
		}
		return nil
	})

	notAdapted := func(ctx *routing.Context) error {
		pathParam = ctx.Param("path-param")
		headerParam = string(ctx.Request.Header.Peek("header-param"))
		for i := 0; i < weight; i++ {
			headerParam = headerParam + "0"
		}
		return nil
	}

	ctx := buildContext()
	b.Run("adapted handler", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = adapted(ctx)
		}
	})

	b.Run("not adapted handler", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = notAdapted(ctx)
		}
	})
}

func TestAdapt(t *testing.T) {
	t.Run("should parse 1 param from path correctly", func(t *testing.T) {
		ctx := buildContext()

		var p string
		err := adapt(func(ctx *routing.Context, args struct {
			P string `path:"path-param"`
		}) error {
			p = args.P
			return nil
		})(ctx)
		if err != nil {
			t.Fatalf("unexpected error received: %s", err.Error())
		}
		if p != "fake-path-param" {
			t.Fatalf("expected path param was not received, got %s", p)
		}
	})
}

func buildContext() *routing.Context {
	ctx := &routing.Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	ctx.SetParam("path-param", "fake-path-param")
	ctx.Request.Header.Set("header-param", "fake-header-param")
	return ctx
}
