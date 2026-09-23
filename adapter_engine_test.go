package kapi

import (
	"net/http"
	"reflect"
	"testing"

	tt "github.com/vingarcia/kapi/internal/testtools"
)

// fakeCtx stands in for whatever request-context type a concrete
// adapter (fiber.Ctx, routing.Context, ...) would normally pass as
// the handler's first argument. The engine only checks its type,
// never its contents, so a bare struct is enough to exercise
// DecodeHandlerFunction and UnmarshalRequestAsStruct in isolation
// from any actual HTTP framework.
type fakeCtx struct{}

var fakeCtxType = reflect.TypeOf(&fakeCtx{})

// fakeHTTPError is the error type our fakeAdapter returns from
// NewHTTPError, so tests can assert on the status code without
// depending on any specific framework's error type.
type fakeHTTPError struct {
	StatusCode int
	Msg        string
}

func (e *fakeHTTPError) Error() string {
	return e.Msg
}

// fakeAdapter implements RequestAdapter using plain maps, so the
// engine tests below don't need a real fiber/fasthttp-routing context.
type fakeAdapter struct {
	pathParams    map[string]string
	headerParams  map[string]string
	queryParams   map[string]string
	contextValues map[string]any
	body          string
}

func (f fakeAdapter) NewHTTPError(statusCode int, msg string) error {
	return &fakeHTTPError{StatusCode: statusCode, Msg: msg}
}

func (f fakeAdapter) GetBody() []byte {
	return []byte(f.body)
}

func (f fakeAdapter) GetPathParam(name string) string {
	return f.pathParams[name]
}

func (f fakeAdapter) GetHeaderParam(name string) string {
	return f.headerParams[name]
}

func (f fakeAdapter) GetQueryParam(name string) string {
	return f.queryParams[name]
}

func (f fakeAdapter) GetContextValue(key string) any {
	return f.contextValues[key]
}

func (f fakeAdapter) SetContextValue(key string, value any) {
	f.contextValues[key] = value
}

// Must implement the RequestAdapter interface:
var _ RequestAdapter = fakeAdapter{}

type engineFoo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// callAdapted runs the same two steps every real adapter's `Adapt`
// function performs (DecodeHandlerFunction once, then
// UnmarshalRequestAsStruct + the call itself), so these tests cover
// the actual engine logic shared by every adapter without needing a
// real HTTP framework in this package.
func callAdapted(t *testing.T, fn interface{}, adapter RequestAdapter) error {
	t.Helper()

	fnInfo := DecodeHandlerFunction(reflect.TypeOf(fn), []reflect.Type{fakeCtxType})

	args, err := UnmarshalRequestAsStruct(adapter, fnInfo)
	if err != nil {
		return err
	}

	out := reflect.ValueOf(fn).Call([]reflect.Value{
		reflect.ValueOf(&fakeCtx{}),
		args.Elem(),
	})

	retErr, _ := out[0].Interface().(error)
	return retErr
}

func TestUnmarshalRequestAsStruct(t *testing.T) {
	t.Run("happy paths", func(t *testing.T) {
		var returnValue interface{}

		tests := []struct {
			desc          string
			adapter       fakeAdapter
			fn            interface{}
			expectedValue interface{}
		}{
			{
				desc:    "should parse 1 param from path correctly",
				adapter: fakeAdapter{pathParams: map[string]string{"path-param": "fake-path-param"}},
				fn: func(ctx *fakeCtx, args struct {
					P string `path:"path-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-path-param",
			},
			{
				desc:    "should parse 1 param from the header correctly",
				adapter: fakeAdapter{headerParams: map[string]string{"header-param": "fake-header-param"}},
				fn: func(ctx *fakeCtx, args struct {
					P string `header:"header-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-header-param",
			},
			{
				desc:    "should parse 1 param from query correctly",
				adapter: fakeAdapter{queryParams: map[string]string{"query-param": "fake-query-param"}},
				fn: func(ctx *fakeCtx, args struct {
					P string `query:"query-param"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-query-param",
			},
			{
				desc:    "should parse the Body correctly",
				adapter: fakeAdapter{body: `{"id":32,"name":"John Doe"}`},
				fn: func(ctx *fakeCtx, args struct {
					Body engineFoo
				}) error {
					returnValue = args.Body.Name
					return nil
				},
				expectedValue: "John Doe",
			},
			{
				desc:    "should use the content-type tag correctly",
				adapter: fakeAdapter{body: `{"id":32,"name":"John Doe"}`},
				fn: func(ctx *fakeCtx, args struct {
					Body engineFoo `content-type:"application/json"`
				}) error {
					returnValue = args.Body.Name
					return nil
				},
				expectedValue: "John Doe",
			},
			{
				desc:    "should parse raw bodies when the Body type is []byte",
				adapter: fakeAdapter{body: `{"id":32,"name":"John Doe"}`},
				fn: func(ctx *fakeCtx, args struct {
					Body []byte `content-type:"application/json"`
				}) error {
					returnValue = string(args.Body)
					return nil
				},
				expectedValue: `{"id":32,"name":"John Doe"}`,
			},
			{
				desc: "should parse every supported integer kind correctly",
				adapter: fakeAdapter{
					pathParams:   map[string]string{"path-param": "42"},
					headerParams: map[string]string{"header-param": "43"},
					queryParams:  map[string]string{"query-param": "44"},
				},
				fn: func(ctx *fakeCtx, args struct {
					PParam int    `path:"path-param"`
					HParam int8   `header:"header-param"`
					QParam uint64 `query:"query-param"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int    `path:"path-param"`
					HParam int8   `header:"header-param"`
					QParam uint64 `query:"query-param"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},
			{
				desc:    "should parse 1 context value correctly",
				adapter: fakeAdapter{contextValues: map[string]any{"user-value": engineFoo{Name: "foo-as-context-value"}}},
				fn: func(ctx *fakeCtx, args struct {
					MyContextValue engineFoo `context:"user-value"`
				}) error {
					returnValue = args.MyContextValue
					return nil
				},
				expectedValue: engineFoo{Name: "foo-as-context-value"},
			},
			{
				desc:    "should ignore optional integers with no errors",
				adapter: fakeAdapter{},
				fn: func(ctx *fakeCtx, args struct {
					Header int `header:"header-param,optional"`
					Query  int `query:"query-param"`
				}) error {
					returnValue = map[string]int{
						"header": args.Header,
						"query":  args.Query,
					}
					return nil
				},
				expectedValue: map[string]int{
					"header": 0,
					"query":  0,
				},
			},
			{
				desc:    "should parse default integers correctly",
				adapter: fakeAdapter{},
				fn: func(ctx *fakeCtx, args struct {
					Header int `header:"header-param" default:"43"`
					Query  int `query:"query-param" default:"44"`
				}) error {
					returnValue = map[string]int{
						"header": args.Header,
						"query":  args.Query,
					}
					return nil
				},
				expectedValue: map[string]int{
					"header": 43,
					"query":  44,
				},
			},
			{
				desc:    "should parse default string correctly",
				adapter: fakeAdapter{},
				fn: func(ctx *fakeCtx, args struct {
					Header string `header:"header-param" default:"43"`
					Query  string `query:"query-param" default:"44"`
				}) error {
					returnValue = map[string]string{
						"header": args.Header,
						"query":  args.Query,
					}
					return nil
				},
				expectedValue: map[string]string{
					"header": "43",
					"query":  "44",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				err := callAdapted(t, test.fn, test.adapter)
				tt.AssertNoErr(t, err)
				tt.AssertEqual(t, test.expectedValue, returnValue)
			})
		}
	})

	t.Run("should report a 400 error when a required path param is empty", func(t *testing.T) {
		reached := false
		err := callAdapted(t, func(ctx *fakeCtx, args struct {
			P string `path:"path-param"`
		}) error {
			reached = true
			return nil
		}, fakeAdapter{})

		httpErr, ok := err.(*fakeHTTPError)
		if !ok {
			t.Fatalf("expected a *fakeHTTPError, got %T: %v", err, err)
		}
		tt.AssertEqual(t, http.StatusBadRequest, httpErr.StatusCode)

		if reached {
			t.Fatal("the handler should not have been called")
		}
	})

	t.Run("should report a 400 error when a required header param is empty", func(t *testing.T) {
		err := callAdapted(t, func(ctx *fakeCtx, args struct {
			H string `header:"header-param"`
		}) error {
			return nil
		}, fakeAdapter{})

		httpErr, ok := err.(*fakeHTTPError)
		if !ok {
			t.Fatalf("expected a *fakeHTTPError, got %T: %v", err, err)
		}
		tt.AssertEqual(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("should report a 400 error when a required query param is empty", func(t *testing.T) {
		err := callAdapted(t, func(ctx *fakeCtx, args struct {
			Q string `query:"query-param,required"`
		}) error {
			return nil
		}, fakeAdapter{})

		httpErr, ok := err.(*fakeHTTPError)
		if !ok {
			t.Fatalf("expected a *fakeHTTPError, got %T: %v", err, err)
		}
		tt.AssertEqual(t, http.StatusBadRequest, httpErr.StatusCode)
	})

	t.Run("should report a 400 error when the body is not valid JSON", func(t *testing.T) {
		err := callAdapted(t, func(ctx *fakeCtx, args struct {
			Body engineFoo
		}) error {
			return nil
		}, fakeAdapter{body: `not-json`})

		httpErr, ok := err.(*fakeHTTPError)
		if !ok {
			t.Fatalf("expected a *fakeHTTPError, got %T: %v", err, err)
		}
		tt.AssertEqual(t, http.StatusBadRequest, httpErr.StatusCode)
	})
}

func TestDecodeHandlerFunction(t *testing.T) {
	t.Run("should panic for an unsupported Body content-type", func(t *testing.T) {
		panicPayload := tt.PanicHandler(func() {
			DecodeHandlerFunction(reflect.TypeOf(func(ctx *fakeCtx, args struct {
				Body engineFoo `content-type:"application/xml"`
			}) error {
				return nil
			}), []reflect.Type{fakeCtxType})
		})

		if panicPayload == nil {
			t.Fatal("expected DecodeHandlerFunction to panic for an unsupported content-type")
		}
	})
}
