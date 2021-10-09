package adapter

import (
	"reflect"
)

type DialectFactory func(args []reflect.Value) Dialect

// Dialect describes how to interact with different HTTP Frameworks.
// By filling the following attributes you can describe a new framework yourself.
type Dialect interface {
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
	GetContextValue(contextKey string) interface{}
}
