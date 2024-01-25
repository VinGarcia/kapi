package fasthttp_routing

import (
	routing "github.com/jackwhelpton/fasthttp-routing/v2"
	"github.com/vingarcia/kapi"
)

type any = interface{}

type Adapter struct {
	ctx *routing.Context
}

// Must implement the RequestAdapter interface:
var _ kapi.RequestAdapter = Adapter{}

func New(ctx *routing.Context) Adapter {
	return Adapter{
		ctx: ctx,
	}
}

func (a Adapter) NewHTTPError(statusCode int, msg string) error {
	return routing.NewHTTPError(statusCode, msg)
}

func (a Adapter) GetBody() []byte {
	return a.ctx.PostBody()
}

func (a Adapter) GetPathParam(paramName string) string {
	return a.ctx.Param(paramName)
}

func (a Adapter) GetHeaderParam(paramName string) string {
	return string(a.ctx.Request.Header.Peek(paramName))
}

func (a Adapter) GetQueryParam(paramName string) string {
	return string(a.ctx.Request.URI().QueryArgs().Peek(paramName))
}

func (a Adapter) GetContextValue(contextKey string) any {
	return a.ctx.UserValue(contextKey)
}

func (a Adapter) SetContextValue(contextKey string, value any) {
	a.ctx.SetUserValue(contextKey, value)
}
