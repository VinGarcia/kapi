package main

import (
	"fmt"
	"log"
	"reflect"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

type ValidationError struct {
	error
}

func NewValidationError(msg string, args ...interface{}) ValidationError {
	return ValidationError{
		error: fmt.Errorf(msg, args...),
	}
}

func main() {
	router := routing.New()
	router.Get("/adapted/<foobar>", adapt(func(ctx *routing.Context, args struct {
		Foobar string `path:"foobar"`
		Brand  string `header:"brand"`
		Testim string `query:"testimvalue"`
	}) error {
		fmt.Printf("Foobar: '%s', Brand: '%s'\n", args.Foobar, args.Brand)
		return nil
	}))

	router.Get("/not-adapted/<foobar>", func(ctx *routing.Context) error {
		foobar := ctx.Param("foobar")
		brand := ctx.Request.Header.Peek("brand")
		fmt.Printf("Foobar: '%s', Brand: '%s'\n", foobar, brand)
		return nil
	})

	port := "8765"
	// Serve Start
	fmt.Println("listening-and-serve", "server listening at:", port)
	if err := fasthttp.ListenAndServe(":"+port, router.HandleRequest); err != nil {
		fmt.Println("listening-and-serve", err.Error())
	}
}

var ctxType = reflect.TypeOf(&routing.Context{})
var errType = reflect.TypeOf(new(error)).Elem()

func adapt(fn interface{}) func(ctx *routing.Context) error {
	t := reflect.TypeOf(fn)
	v := reflect.ValueOf(fn)

	if t.Kind() != reflect.Func {
		log.Fatal("adapt's argument must be a function!")
	}

	if t.NumIn() != 2 {
		log.Fatal("received function must have 2 arguments!")
	}

	if t.In(0) != ctxType {
		log.Fatal("first argument must be of type *routing.Context!")
	}

	if t.NumOut() != 1 {
		log.Fatal("received function must have a single return value!")
	}

	if t.Out(0) != errType {
		log.Fatal("first return value must be of type error")
	}

	argsType := t.In(1)
	if argsType.Kind() != reflect.Struct {
		log.Fatal("second argument must a struct!")
	}

	pathParams, headerParams, queryParams := getTagNames(argsType)
	return func(ctx *routing.Context) error {
		args := reflect.New(argsType)
		for key, idx := range pathParams {
			param := ctx.Param(key)
			if param == "" {
				return NewValidationError("path param '%s' is empty", key)
			}

			args.Elem().Field(idx).Set(reflect.ValueOf(param))
		}
		for key, idx := range headerParams {
			param := string(ctx.Request.Header.Peek(key))
			if param == "" {
				return NewValidationError("required header param '%s' is empty", key)
			}

			args.Elem().Field(idx).Set(reflect.ValueOf(param))
		}
		for key, idx := range queryParams {
			param := string(ctx.Request.URI().QueryArgs().Peek(key))
			if param == "" {
				return NewValidationError("required header param '%s' is empty", key)
			}

			args.Elem().Field(idx).Set(reflect.ValueOf(param))
		}

		err, _ := v.Call([]reflect.Value{reflect.ValueOf(ctx), args.Elem()})[0].Interface().(error)
		return err
	}
}

// This function collects only the names
// that will be used from the type
// this should save several calls to `Field(i).Tag.Get("foo")`
// which might improve the performance by a lot.
func getTagNames(t reflect.Type) (map[string]int, map[string]int, map[string]int) {
	pathParams := map[string]int{}
	headerParams := map[string]int{}
	queryParams := map[string]int{}
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("path")
		if key == "" {
			continue
		}
		pathParams[key] = i
	}
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("header")
		if key == "" {
			continue
		}
		headerParams[key] = i
	}
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Tag.Get("query")
		if key == "" {
			continue
		}
		queryParams[key] = i
	}
	return pathParams, headerParams, queryParams
}
