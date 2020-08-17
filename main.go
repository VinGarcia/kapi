package main

import (
	"fmt"
	"log"
	"reflect"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"
)

func main() {
	router := routing.New()
	router.Get("/<foobar>/ping", adapt(func(ctx *routing.Context, args struct {
		Foobar string `path:"foobar"`
		Brand  string `header:"brand"`
	}) error {
		fmt.Printf("Foobar: '%s', Brand: '%s'\n", args.Foobar, args.Brand)
		return nil
	}))

	port := "8765"
	// Serve Start
	fmt.Println("listening-and-serve", "server listening at:", port)
	if err := fasthttp.ListenAndServe(":"+port, router.HandleRequest); err != nil {
		fmt.Println("listening-and-serve", err.Error())
	}
}

var ctxType = reflect.TypeOf(&routing.Context{})

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

	argsType := t.In(1)
	if argsType.Kind() != reflect.Struct {
		log.Fatal("second argument must a struct!")
	}

	pathParams, headerParams := getTagNames(argsType)
	return func(ctx *routing.Context) error {
		args := reflect.New(argsType)
		for key, idx := range pathParams {
			args.Elem().Field(idx).Set(reflect.ValueOf(ctx.Param(key)))
		}
		for key, idx := range headerParams {
			args.Elem().Field(idx).Set(reflect.ValueOf(string(ctx.Request.Header.Peek(key))))
		}

		err, _ := v.Call([]reflect.Value{reflect.ValueOf(ctx), args.Elem()})[0].Interface().(error)
		return err
	}
}

// This function collects only the names
// that will be used from the type
// this should save several calls to `Field(i).Tag.Get("foo")`
// which might improve the performance by a lot.
func getTagNames(t reflect.Type) (map[string]int, map[string]int) {
	pathParams := map[string]int{}
	headerParams := map[string]int{}
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
	return pathParams, headerParams
}
