package adapter

import (
	"testing"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

type Foo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var err error
var pathParam string
var headerParam string

var weight = 10

func BenchmarkAdapter(b *testing.B) {
	adapted := Adapt(func(ctx *routing.Context, args struct {
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

	ctx := buildFakeContext(mockedArgs{
		PathParam: "fake-path-param",
	})
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

	t.Run("testing happy paths", func(t *testing.T) {
		var returnValue interface{}
		tests := []struct {
			desc          string
			ctx           *routing.Context
			fn            interface{}
			expectedValue interface{}
		}{
			{
				desc: "should parse 1 param from path correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam: "fake-path-param",
				}),
				fn: func(ctx *routing.Context, args struct {
					P string `path:"path-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-path-param",
			},

			{
				desc: "should parse 1 param from the header correctly",
				ctx: buildFakeContext(mockedArgs{
					HeaderParam: "fake-header-param",
				}),
				fn: func(ctx *routing.Context, args struct {
					P string `header:"header-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-header-param",
			},

			{
				desc: "should parse 1 param from query correctly",
				ctx: buildFakeContext(mockedArgs{
					QueryParam: "fake-query-param",
				}),
				fn: func(ctx *routing.Context, args struct {
					P string `query:"query-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-query-param",
			},

			{
				desc: "should parse a JSONBody correctly",
				ctx: buildFakeContext(mockedArgs{
					Body: `{"id":32,"name":"John Doe"}`,
				}),
				fn: func(ctx *routing.Context, args struct {
					JSONBody Foo
				}) error {
					returnValue = args.JSONBody.Name
					return nil
				},
				expectedValue: "John Doe",
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				ctx := test.ctx

				err := Adapt(test.fn)(ctx)
				if err != nil {
					t.Fatalf("unexpected error received: %s", err.Error())
				}
				if returnValue != test.expectedValue {
					t.Fatalf("expected param was not received, got %s", returnValue)
				}
			})
		}
	})

	t.Run("should report error when path param is empty", func(t *testing.T) {
		ctx := buildFakeContext(mockedArgs{})

		var p interface{}
		err := Adapt(func(ctx *routing.Context, args struct {
			P string `path:"path-param"`
		}) error {
			p = "reached this line"
			return nil
		})(ctx)
		if ErrorCode(err) != MissingRequiredParamError {
			t.Fatalf("expected MissingRequiredParamError bug got: %T", err)
		}
		if p == "reached this line" {
			t.Fatalf("the callback should not have been executed!")
		}
	})
}

type mockedArgs struct {
	PathParam   string
	HeaderParam string
	QueryParam  string
	Body        string
}

func buildFakeContext(args mockedArgs) *routing.Context {
	ctx := &routing.Context{
		RequestCtx: &fasthttp.RequestCtx{},
	}
	ctx.SetParam("path-param", args.PathParam)

	if args.HeaderParam != "" {
		ctx.Request.Header.Set("header-param", args.HeaderParam)
	}

	if args.QueryParam != "" {
		ctx.Request.URI().QueryArgs().Set("query-param", args.QueryParam)
	}

	if args.Body != "" {
		ctx.Request.SetBody([]byte(args.Body))
	}
	return ctx
}
