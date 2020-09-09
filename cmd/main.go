package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/valyala/fasthttp"

	adapter "github.com/vingarcia/go-adapter"
)

type Foo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MyType struct {
	Value string
}

func main() {
	router := routing.New()

	middleware := func(ctx *routing.Context) error {
		ctx.SetUserValue("my_type", MyType{
			Value: "foo",
		})
		return nil
	}

	router.Post("/adapted/<id>", middleware, adapter.Adapt(func(ctx *routing.Context, args struct {
		ID       uint64 `path:"id"`
		Brand    string `header:"brand,optional"`
		Qparam   string `query:"qparam,required"`
		MyType   MyType `uservalue:"my_type"`
		JSONBody Foo
	}) error {
		jsonResp, _ := json.Marshal(map[string]interface{}{
			"ID":        args.ID,
			"Brand":     args.Brand,
			"Query":     args.Qparam,
			"Body":      args.JSONBody,
			"UserValue": args.MyType,
		})
		fmt.Println(string(jsonResp))
		ctx.SetBody(jsonResp)

		return nil
	}))

	// This route does exactly the same as the route above
	// but without using the library:
	router.Post("/not-adapted/<id>", middleware, func(ctx *routing.Context) error {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			fmt.Println("id is invalid!")
			return err
		}

		brand := string(ctx.Request.Header.Peek("brand"))
		if brand == "" {
			fmt.Println("deu ruim de novo")
			return fmt.Errorf("brand is missing!")
		}

		qparam := string(ctx.Request.URI().QueryArgs().Peek("qparam"))
		if qparam == "" {
			return fmt.Errorf("qparam is missing!")
		}

		myType, ok := ctx.UserValue("my_type").(MyType)
		if !ok {
			return fmt.Errorf("missing required user value `my_type`!")
		}

		var body Foo
		err = json.Unmarshal(ctx.PostBody(), &body)
		if err != nil {
			fmt.Println("error unmarshalling body as JSON!")
			return err
		}

		jsonResp, _ := json.Marshal(map[string]interface{}{
			"ID":        id,
			"Brand":     brand,
			"Query":     qparam,
			"Body":      body,
			"UserValue": myType,
		})
		fmt.Println(string(jsonResp))
		ctx.SetBody(jsonResp)

		return nil
	})

	port := "8765"
	// Serve Start
	fmt.Println("listening-and-serve", "server listening at:", port)
	if err := fasthttp.ListenAndServe(":"+port, router.HandleRequest); err != nil {
		fmt.Println("error-serving", err.Error())
	}
}
