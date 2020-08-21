package main

import (
	"testing"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

var err error
var foobar string
var brand string

var weight = 10

func BenchmarkAdapter(b *testing.B) {
	adapted := adapt(func(ctx *routing.Context, args struct {
		Foobar string `path:"foobar"`
		Brand  string `header:"brand"`
	}) error {
		foobar = args.Foobar
		brand = args.Brand
		for i := 0; i < weight; i++ {
			brand = brand + "0"
		}
		return nil
	})

	notAdapted := func(ctx *routing.Context) error {
		foobar = ctx.Param("foobar")
		brand = string(ctx.Request.Header.Peek("brand"))
		for i := 0; i < weight; i++ {
			brand = brand + "0"
		}
		return nil
	}

	ctx := &routing.Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	ctx.SetParam("foobar", "teste")
	ctx.Request.Header.Set("brand", "dito-teste")

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
