package fiber

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	tt "github.com/vingarcia/kapi/internal/testtools"
)

type fiberFoo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TestAdapt exercises Adapt end-to-end through a real fiber.App and
// app.Test(), so it only needs to prove the wiring (Adapter -> engine ->
// back into the handler) is correct. Exhaustive coverage of every
// tag/type combination lives in the framework-agnostic tests in the
// root kapi package.
func TestAdapt(t *testing.T) {
	t.Run("should parse path, header, query and body params", func(t *testing.T) {
		app := fiber.New()

		var got struct {
			P    int
			H    string
			Q    int
			Body fiberFoo
		}
		// Fiber's route parser treats "-" as a segment delimiter in path
		// patterns (see path.go's routeDelimiter), so the placeholder name
		// registered here must not contain one; it only needs to match the
		// `path:"..."` tag below, not the URL's own literal spelling.
		app.Post("/foo/:pathParam", Adapt(func(ctx *fiber.Ctx, args struct {
			P    int    `path:"pathParam"`
			H    string `header:"header-param"`
			Q    int    `query:"query-param"`
			Body fiberFoo
		}) error {
			got.P, got.H, got.Q, got.Body = args.P, args.H, args.Q, args.Body
			return ctx.SendStatus(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodPost, "/foo/42?query-param=99", strings.NewReader(`{"id":32,"name":"John Doe"}`))
		req.Header.Set("header-param", "fake-header-param")

		resp, err := app.Test(req)
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, http.StatusOK, resp.StatusCode)

		tt.AssertEqual(t, 42, got.P)
		tt.AssertEqual(t, "fake-header-param", got.H)
		tt.AssertEqual(t, 99, got.Q)
		tt.AssertEqual(t, fiberFoo{ID: 32, Name: "John Doe"}, got.Body)
	})

	t.Run("should return a 400 response when a required path param is missing", func(t *testing.T) {
		app := fiber.New()

		reached := false
		// This route has no ":path-param" placeholder, so kapi will always
		// see it as empty and reject the request before the handler runs.
		app.Get("/foo", Adapt(func(ctx *fiber.Ctx, args struct {
			P string `path:"path-param"`
		}) error {
			reached = true
			return nil
		}))

		req := httptest.NewRequest(http.MethodGet, "/foo", nil)

		resp, err := app.Test(req)
		tt.AssertNoErr(t, err)
		tt.AssertEqual(t, http.StatusBadRequest, resp.StatusCode)

		if reached {
			t.Fatal("the handler should not have been called")
		}
	})
}
