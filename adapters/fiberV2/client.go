package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vingarcia/kapi"
)

func Handler(fn any) func(c *fiber.Ctx) error {
	return kapi.NewHandlerFactory(newAdapter, fn)
}

func newAdapter(c *fiber.Ctx) kapi.RequestAdapter {
	return Adapter{ctx: c}
}
