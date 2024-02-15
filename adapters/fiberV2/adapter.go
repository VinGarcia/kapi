package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vingarcia/kapi"
)

// Adapter implements the kapi.RequestAdapter interface
type Adapter struct {
	ctx *fiber.Ctx
}

// Must implement the kapi.RequestAdapter interface:
var _ kapi.RequestAdapter = Adapter{}

func New(ctx *fiber.Ctx) Adapter {
	return Adapter{
		ctx: ctx,
	}
}

func (a Adapter) NewHTTPError(statusCode int, msg string) error {
	return fiber.NewError(statusCode, msg)
}

func (a Adapter) GetBody() []byte {
	return a.ctx.Body()
}

func (a Adapter) GetPathParam(paramName string) string {
	return a.ctx.Params(paramName)
}

func (a Adapter) GetHeaderParam(paramName string) string {
	return a.ctx.Get(paramName)
}

func (a Adapter) GetQueryParam(paramName string) string {
	return a.ctx.Query(paramName)
}

func (a Adapter) GetContextValue(contextKey string) any {
	return a.ctx.Context().Value(contextKey)
}

func (a Adapter) SetContextValue(contextKey string, value any) {
	a.ctx.Context().Value(contextKey)
}
