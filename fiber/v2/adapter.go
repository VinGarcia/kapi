package fiber

import (
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/vingarcia/go-adapter"
)

// This constant is used for validating if the
// first arguments of the input function matches
// the arguments we are expecting:
var argTypes = []reflect.Type{
	reflect.TypeOf(&fiber.Ctx{}),
}

// Adapt was created to simplify the parsing and validation
// of the request arguments.
//
// The input argument must be a function callback whose first
// argument is a *fiber.Ctx and the second is a struct
// where each attribute contains a special Tag describing
// from where it should be parsed, e.g.:
//
//   func MyAdaptedHandler(ctx *fiber.Ctx, args struct{
//     PathArgument   int          `path:"my_path_arg"`
//     QueryArgument  uint64       `query:"my_query_arg"`
//     HeaderArgument string       `header:"my_header_arg"`
//     ContextValue   MyCustomType `context:"my_context_value"`
//     Body           MyCustomBody `content-type:"application/json"`
//   }) error {
//
//     // ... handle request ...
//
//     return nil
//   }
//
// Note: all attributes in the input struct must be public or the adapter will panic
func Adapt(fn interface{}) func(ctx *fiber.Ctx) error {
	t := reflect.TypeOf(fn)
	v := reflect.ValueOf(fn)

	// The slow steps that heavly relie on reflection are done here
	// only once during startup in order to affect as little
	// as possible the performance later on.
	fnInfo := adapter.DecodeHandlerFunction(t, argTypes)
	return func(ctx *fiber.Ctx) error {
		// This part uses cached information from `fnInfo` and uses
		// reflection only to fill the struct making it more performatic:
		inputStructPtr, err := adapter.UnmarshalRequestAsStruct(NewDialect(ctx), fnInfo)
		if err != nil {
			return err
		}

		// Here we pass the arguments to the user defined handler function in the order
		// we expect to receive them, i.e. `func(ctx *fiber.Ctx, args MyStruct) error`:
		err, _ = v.Call([]reflect.Value{reflect.ValueOf(ctx), inputStructPtr.Elem()})[0].Interface().(error)
		return err
	}
}
