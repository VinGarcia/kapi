package fasthttp_routing

import (
	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/vingarcia/go-adapter"
)

type Dialect struct {
	ctx *routing.Context
}

// Must implement the adapter.Dialect interface:
var _ adapter.Dialect = Dialect{}

func NewDialect(ctx *routing.Context) adapter.Dialect {
	return Dialect{
		ctx: ctx,
	}
}

func (d Dialect) NewHTTPError(statusCode int, msg string) error {
	return routing.NewHTTPError(statusCode, msg)
}

func (d Dialect) GetBody() []byte {
	return d.ctx.PostBody()
}

func (d Dialect) GetPathParam(paramName string) string {
	return d.ctx.Param(paramName)
}

func (d Dialect) GetHeaderParam(paramName string) string {
	return string(d.ctx.Request.Header.Peek(paramName))
}

func (d Dialect) GetQueryParam(paramName string) string {
	return string(d.ctx.Request.URI().QueryArgs().Peek(paramName))
}

func (d Dialect) GetContextValue(contextKey string) interface{} {
	return d.ctx.UserValue(contextKey)
}
