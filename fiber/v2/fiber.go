package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vingarcia/go-adapter"
)

// Dialect implements the adapter.Dialect interface
type Dialect struct {
	ctx *fiber.Ctx
}

// Must implement the adapter.Dialect interface:
var _ adapter.Dialect = Dialect{}

func NewDialect(ctx *fiber.Ctx) adapter.Dialect {
	return Dialect{
		ctx: ctx,
	}
}

func (d Dialect) NewHTTPError(statusCode int, msg string) error {
	return fiber.NewError(statusCode, msg)
}

func (d Dialect) GetBody() []byte {
	return d.ctx.Body()
}

func (d Dialect) GetPathParam(paramName string) string {
	return d.ctx.Params(paramName)
}

func (d Dialect) GetHeaderParam(paramName string) string {
	return d.ctx.Get(paramName)
}

func (d Dialect) GetQueryParam(paramName string) string {
	return d.ctx.Query(paramName)
}

func (d Dialect) GetContextValue(contextKey string) interface{} {
	return d.ctx.Context().Value(contextKey)
}
