package kapi

import (
	"testing"

	tt "github.com/vingarcia/kapi/internal/testtools"
)

/*
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
	adapted := Adapt(func(ctx *fiber.Ctx, args struct {
		PathParam   int    `path:"pathParam"`
		HeaderParam string `header:"headerParam"`
		Body        Foo
	}) error {
		pathParam = args.PathParam
		headerParam = args.HeaderParam
		body = args.Body

		return nil
	})

	notAdapted := func(ctx *fiber.Ctx) (err error) {
		pathParam, err = strconv.Atoi(ctx.Param("pathParam"))
		if err != nil {
			fmt.Println("deu ruim!")
			return err
		}
		headerParam = string(ctx.Request.Header.Peek("headerParam"))
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

	ctx := buildFakeContext(MockedArgs{
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
// */

type TestRequest struct {
	PathParam   string
	HeaderParam string
	QueryParam  string
	Body        string
}

func AdaptTestSuite[CtxT any](
	t *testing.T,
	runHandlerWith func(fn any, testRequest TestRequest, ctxValue any) error,
) {
	t.Run("testing happy paths", func(t *testing.T) {
		var returnValue interface{}
		tests := []struct {
			desc          string
			ctxValue      any
			testRequest   TestRequest
			fn            interface{}
			expectedValue interface{}
		}{
			{
				desc: "should parse 1 param from path correctly",
				testRequest: TestRequest{
					PathParam: "fake-path-param",
				},
				fn: func(ctx CtxT, args struct {
					P string `path:"pathParam"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-path-param",
			},

			{
				desc: "should parse 1 param from the header correctly",
				testRequest: TestRequest{
					HeaderParam: "fake-header-param",
				},
				fn: func(ctx CtxT, args struct {
					P string `header:"headerParam"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-header-param",
			},

			{
				desc: "should parse 1 param from query correctly",
				testRequest: TestRequest{
					QueryParam: "fake-query-param",
				},
				fn: func(ctx CtxT, args struct {
					P string `query:"queryParam"`
				}) error {
					returnValue = args.P
					return nil
				},
				expectedValue: "fake-query-param",
			},

			{
				desc: "should parse the Body correctly",
				testRequest: TestRequest{
					Body: `{"id":32,"name":"John Doe"}`,
				},
				fn: func(ctx CtxT, args struct {
					Body struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					}
				}) error {
					returnValue = args.Body.Name
					return nil
				},
				expectedValue: "John Doe",
			},

			{
				desc: "should use the content-type tag correctly",
				testRequest: TestRequest{
					Body: `{"id":32,"name":"John Doe"}`,
				},
				fn: func(ctx CtxT, args struct {
					Body struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					} `content-type:"application/json"`
				}) error {
					returnValue = args.Body.Name
					return nil
				},
				expectedValue: "John Doe",
			},

			{
				desc: "should parse raw bodies when the Body type is []byte",
				testRequest: TestRequest{
					Body: `{"id":32,"name":"John Doe"}`,
				},
				fn: func(ctx CtxT, args struct {
					Body []byte `content-type:"application/json"`
				}) error {
					returnValue = string(args.Body)
					return nil
				},
				expectedValue: `{"id":32,"name":"John Doe"}`,
			},

			{
				desc: "should parse an integer correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam int `path:"pathParam"`
					HParam int `header:"headerParam"`
					QParam int `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int `path:"pathParam"`
					HParam int `header:"headerParam"`
					QParam int `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int8 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam int8 `path:"pathParam"`
					HParam int8 `header:"headerParam"`
					QParam int8 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int8 `path:"pathParam"`
					HParam int8 `header:"headerParam"`
					QParam int8 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int16 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam int16 `path:"pathParam"`
					HParam int16 `header:"headerParam"`
					QParam int16 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int16 `path:"pathParam"`
					HParam int16 `header:"headerParam"`
					QParam int16 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int32 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam int32 `path:"pathParam"`
					HParam int32 `header:"headerParam"`
					QParam int32 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int32 `path:"pathParam"`
					HParam int32 `header:"headerParam"`
					QParam int32 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam int64 `path:"pathParam"`
					HParam int64 `header:"headerParam"`
					QParam int64 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam int64 `path:"pathParam"`
					HParam int64 `header:"headerParam"`
					QParam int64 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam uint8 `path:"pathParam"`
					HParam uint8 `header:"headerParam"`
					QParam uint8 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint8 `path:"pathParam"`
					HParam uint8 `header:"headerParam"`
					QParam uint8 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam uint16 `path:"pathParam"`
					HParam uint16 `header:"headerParam"`
					QParam uint16 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint16 `path:"pathParam"`
					HParam uint16 `header:"headerParam"`
					QParam uint16 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam uint32 `path:"pathParam"`
					HParam uint32 `header:"headerParam"`
					QParam uint32 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint32 `path:"pathParam"`
					HParam uint32 `header:"headerParam"`
					QParam uint32 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc: "should parse an int64 correctly",
				testRequest: TestRequest{
					PathParam:   "42",
					HeaderParam: "43",
					QueryParam:  "44",
				},
				fn: func(ctx CtxT, args struct {
					PParam uint64 `path:"pathParam"`
					HParam uint64 `header:"headerParam"`
					QParam uint64 `query:"queryParam"`
				}) error {
					returnValue = args
					return nil
				},
				expectedValue: struct {
					PParam uint64 `path:"pathParam"`
					HParam uint64 `header:"headerParam"`
					QParam uint64 `query:"queryParam"`
				}{
					PParam: 42,
					HParam: 43,
					QParam: 44,
				},
			},

			{
				desc:        "should parse 1 user value correctly",
				testRequest: TestRequest{},
				ctxValue: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{
					Name: "foo-as-user-value",
				},
				fn: func(ctx CtxT, args struct {
					MyUserValue struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					} `uservalue:"userValue"`
				}) error {
					returnValue = args.MyUserValue
					return nil
				},
				expectedValue: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{
					Name: "foo-as-user-value",
				},
			},

			{
				desc:        "should ignore optional integers with no errors",
				testRequest: TestRequest{},
				fn: func(ctx CtxT, args struct {
					Header int `header:"headerParam,optional"`
					Query  int `query:"queryParam"`
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
				desc:        "should parse default integers correctly",
				testRequest: TestRequest{},
				fn: func(ctx CtxT, args struct {
					Header int `header:"headerParam" default:"43"`
					Query  int `query:"queryParam" default:"44"`
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
				desc:        "should parse default string correctly",
				testRequest: TestRequest{},
				fn: func(ctx CtxT, args struct {
					Header string `header:"headerParam" default:"43"`
					Query  string `query:"queryParam" default:"44"`
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
				err := runHandlerWith(test.fn, test.testRequest, test.ctxValue)
				if err != nil {
					t.Fatalf("unexpected error received: %s", err.Error())
				}

				tt.AssertEqual(t, test.expectedValue, returnValue)
			})
		}
	})

	/*
		t.Run("should report error when path param is empty", func(t *testing.T) {
			server := httptest.NewServer()

			ctx := buildRequestContext(MockedArgs{})

			var p interface{}
			err := Adapt(func(ctx CtxT, args struct {
				P string `path:"pathParam"`
			}) error {
				p = "reached this line"
				return nil
			})(ctx)

			httpErr, ok := err.(*fiber.Error)
			if !ok || httpErr.Code != http.StatusBadRequest {
				t.Fatalf("expected MissingRequiredParamError bug got: %T", err)
			}

			if p == "reached this line" {
				t.Fatalf("the callback should not have been executed!")
			}
		})
	// */
}
