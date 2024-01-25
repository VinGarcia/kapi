package kapi

type any = interface{}

// RequestAdapter is the minimum interface required for interacting
// with an input request and return a response.
//
// This interface allows different http clients to be used
// as a backend driver for this library.
type RequestAdapter interface {
	// Returns an http error that is approriate for the framework
	NewHTTPError(statusCode int, msg string) error

	// Returns the request Body as bytes
	GetBody() []byte

	// All the following params should return the appropriate value for the
	// param named `paramName` on the path, header or query
	//
	// If no value is found it should return an empty string
	GetPathParam(paramName string) string
	GetHeaderParam(paramName string) string
	GetQueryParam(paramName string) string

	// This function should return the value as an emtpy
	// interface as it will be converted to the type
	// described on the adapter's input struct.
	GetContextValue(contextKey string) any
	SetContextValue(contextKey string, value any)
}
