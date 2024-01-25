package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	adapter "github.com/vingarcia/kapi/adapters/fiberV2"
)

type Foo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MyType struct {
	Value string
}

func main() {
	app := fiber.New()

	middleware := func(ctx *fiber.Ctx) error {
		fmt.Println("inside the middleware")
		ctx.Locals("my_type", MyType{
			Value: "foo",
		})
		return ctx.Next()
	}

	app.Post("/adapted/:id", middleware, adapter.Adapt(func(ctx *fiber.Ctx, args struct {
		ID     uint64 `path:"id"`
		Brand  string `header:"brand,optional"`
		Qparam string `query:"qparam,required"`
		MyType MyType `context:"my_type"`
		Body   Foo    `content-type:"application/json"`
	}) error {
		jsonResp, _ := json.Marshal(map[string]interface{}{
			"ID":        args.ID,
			"Brand":     args.Brand,
			"Query":     args.Qparam,
			"UserValue": args.MyType,
			"Body":      args.Body,
		})
		fmt.Println(string(jsonResp))
		ctx.Send(jsonResp)

		return nil
	}))

	// This route does exactly the same as the route above
	// but without using the library:
	app.Post("/not-adapted/:id", middleware, func(ctx *fiber.Ctx) error {
		fmt.Println("here we are on the not-adapted route")
		id, err := strconv.Atoi(ctx.Params("id"))
		if err != nil {
			fmt.Println("id is invalid")
			return err
		}

		brand := ctx.Get("brand")
		if brand == "" {
			fmt.Println("deu ruim de novo")
			return fmt.Errorf("brand is missing")
		}

		qparam := ctx.Query("qparam")
		if qparam == "" {
			return fmt.Errorf("qparam is missing")
		}

		myType, ok := ctx.Context().Value("my_type").(MyType)
		if !ok {
			return fmt.Errorf("missing required user value `my_type`")
		}

		var body Foo
		err = json.Unmarshal(ctx.Body(), &body)
		if err != nil {
			fmt.Println("error unmarshalling body as JSON")
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
		ctx.Send(jsonResp)

		return nil
	})

	port := "8765"
	// Serve Start
	fmt.Println("listening-and-serve", "server listening at:", port)
	if err := app.Listen(":" + port); err != nil {
		fmt.Println("error-serving", err.Error())
	}
}
