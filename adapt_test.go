package adapter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
	"gotest.tools/assert"
)

type Foo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var err error
var pathParam int
var headerParam string
var body Foo

var weight = 10

func BenchmarkAdapter(b *testing.B) {
	adapted := Adapt(func(ctx *routing.Context, args struct {
		PathParam   int    `path:"path-param"`
		HeaderParam string `header:"header-param"`
		JSONBody    Foo
	}) error {
		pathParam = args.PathParam
		headerParam = args.HeaderParam
		body = args.JSONBody

		return nil
	})

	notAdapted := func(ctx *routing.Context) (err error) {
		pathParam, err = strconv.Atoi(ctx.Param("path-param"))
		if err != nil {
			fmt.Println("deu ruim!")
			return err
		}
		headerParam = string(ctx.Request.Header.Peek("header-param"))
		if headerParam == "" {
			fmt.Println("deu ruim!")
			return fmt.Errorf("deu ruim!")
		}

		err = json.Unmarshal(ctx.PostBody(), &body)
		if err != nil {
			fmt.Println("deu ruim")
			return err
		}

		return nil
	}

	ctx := buildFakeContext(mockedArgs{
		PathParam:   "42",
		HeaderParam: "fake-header-param",
		Body:        `{"id":32,"name":"John Doe"}`,
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

			{
				desc: "should parse an integer correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam int `path:"path-param"`
					HParam int `header:"header-param"`
					QParam int `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int `path:"path-param"`
					HParam int `header:"header-param"`
					QParam int `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int8 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam int8 `path:"path-param"`
					HParam int8 `header:"header-param"`
					QParam int8 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int8 `path:"path-param"`
					HParam int8 `header:"header-param"`
					QParam int8 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int16 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam int16 `path:"path-param"`
					HParam int16 `header:"header-param"`
					QParam int16 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int16 `path:"path-param"`
					HParam int16 `header:"header-param"`
					QParam int16 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int32 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam int32 `path:"path-param"`
					HParam int32 `header:"header-param"`
					QParam int32 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int32 `path:"path-param"`
					HParam int32 `header:"header-param"`
					QParam int32 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam int64 `path:"path-param"`
					HParam int64 `header:"header-param"`
					QParam int64 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int64 `path:"path-param"`
					HParam int64 `header:"header-param"`
					QParam int64 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam uint8 `path:"path-param"`
					HParam uint8 `header:"header-param"`
					QParam uint8 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint8 `path:"path-param"`
					HParam uint8 `header:"header-param"`
					QParam uint8 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam uint16 `path:"path-param"`
					HParam uint16 `header:"header-param"`
					QParam uint16 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint16 `path:"path-param"`
					HParam uint16 `header:"header-param"`
					QParam uint16 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam uint32 `path:"path-param"`
					HParam uint32 `header:"header-param"`
					QParam uint32 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint32 `path:"path-param"`
					HParam uint32 `header:"header-param"`
					QParam uint32 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				ctx: buildFakeContext(mockedArgs{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				}),
				fn: func(ctx *routing.Context, args struct {
					PParam uint64 `path:"path-param"`
					HParam uint64 `header:"header-param"`
					QParam uint64 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint64 `path:"path-param"`
					HParam uint64 `header:"header-param"`
					QParam uint64 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse 1 user value correctly",
				ctx: buildFakeContext(mockedArgs{
					UserValue: Foo{
						Name: "foo-as-user-value",
					},
				}),
				fn: func(ctx *routing.Context, args struct {
					MyUserValue Foo `uservalue:"user-value"`
				}) error {
					returnValue = args.MyUserValue
					return nil
				},
				expectedValue: Foo{
					Name: "foo-as-user-value",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				ctx := test.ctx

				err := Adapt(test.fn)(ctx)
				if err != nil {
					t.Fatalf("unexpected error received: %s", err.Error())
				}

				assert.Equal(t, test.expectedValue, returnValue)
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

		httpErr, ok := err.(routing.HTTPError)
		if !ok || httpErr.StatusCode() != http.StatusBadRequest {
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
	UserValue   interface{}
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

	if args.UserValue != nil {
		ctx.SetUserValue("user-value", args.UserValue)
	}

	if args.Body != "" {
		ctx.Request.SetBody([]byte(args.Body))
	}

	return ctx
}
