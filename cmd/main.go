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

func main() {
	router := routing.New()
	router.Post("/adapted/<id>", adapter.Adapt(func(ctx *routing.Context, args struct {
		ID       string `path:"id"`
		Brand    string `header:"brand,optional"`
		Qparam   string `query:"qparam,required"`
		JSONBody Foo
	}) error {
		fmt.Printf("ID: '%s', Brand: '%s'\n", args.ID, args.Brand)
		fmt.Printf("Query: '%s', Body: '%#v'\n", args.Qparam, args.JSONBody)

		body, _ := json.Marshal(args)
		ctx.SetBody(body)

		return nil
	}))

	router.Post("/not-adapted/<id>", func(ctx *routing.Context) error {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			fmt.Println("deu ruim")
			return err
		}

		brand := string(ctx.Request.Header.Peek("brand"))
		if brand == "" {
			fmt.Println("deu ruim de novo")
			return fmt.Errorf("ta faltando a brand!")
		}

		fmt.Printf("ID: '%d', Brand: '%s'\n", id, brand)
		return nil
	})

	port := "8765"
	// Serve Start
	fmt.Println("listening-and-serve", "server listening at:", port)
	if err := fasthttp.ListenAndServe(":"+port, router.HandleRequest); err != nil {
		fmt.Println("listening-and-serve", err.Error())
	}
}
