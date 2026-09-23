package fasthttp_routing

import (
	"net/http"
	"testing"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"

	tt "github.com/vingarcia/kapi/internal/testtools"
)

type routingFoo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func buildFakeContext() *routing.Context {
	return routing.NewContext(&fasthttp.RequestCtx{})
}

// TestAdapt exercises Adapt end-to-end against a real *routing.Context, so
// it only needs to prove the wiring (Adapter -> engine -> back into the
// handler) is correct. Exhaustive coverage of every tag/type combination
// lives in the framework-agnostic tests in the root kapi package.
func TestAdapt(t *testing.T) {
	t.Run("should parse path, header, query and body params", func(t *testing.T) {
		ctx := buildFakeContext()
		ctx.SetParam("path-param", "42")
		ctx.Request.Header.Set("header-param", "fake-header-param")
		ctx.Request.URI().QueryArgs().Set("query-param", "99")
		ctx.Request.SetBody([]byte(`{"id":32,"name":"John Doe"}`))

		var got struct {
			P    int
			H    string
			Q    int
			Body routingFoo
		}
		err := Adapt(func(ctx *routing.Context, args struct {
			P    int    `path:"path-param"`
			H    string `header:"header-param"`
			Q    int    `query:"query-param"`
			Body routingFoo
		}) error {
			got.P, got.H, got.Q, got.Body = args.P, args.H, args.Q, args.Body
			return nil
		})(ctx)

		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, 42, got.P)
		tt.AssertEqual(t, "fake-header-param", got.H)
		tt.AssertEqual(t, 99, got.Q)
		tt.AssertEqual(t, routingFoo{ID: 32, Name: "John Doe"}, got.Body)
	})

	t.Run("should return a 400 routing.HTTPError when a required path param is missing", func(t *testing.T) {
		ctx := buildFakeContext()

		reached := false
		err := Adapt(func(ctx *routing.Context, args struct {
			P string `path:"path-param"`
		}) error {
			reached = true
			return nil
		})(ctx)

		httpErr, ok := err.(routing.HTTPError)
		if !ok {
			t.Fatalf("expected a routing.HTTPError, got %T: %v", err, err)
		}
		tt.AssertEqual(t, http.StatusBadRequest, httpErr.StatusCode())

		if reached {
			t.Fatal("the handler should not have been called")
		}
	})
}
