package adapter

import (
	"encoding/json"
	"fmt"
	"net/http"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
)

// BuildJSONResponse is a oneliner helper to marshal a struct or map
// into json set the response status and the content type.
func BuildJSONResponse(ctx *routing.Context, statusCode int, body interface{}) error {
	rawJSON, err := json.Marshal(body)
	if err != nil {
		return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
			"could not marshal response body, Reason: %s, Body: %v",
			err.Error(),
			body,
		))
	}

	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.SetBody(rawJSON)
	return nil
}
