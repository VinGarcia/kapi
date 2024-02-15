package kapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var errType = reflect.TypeOf(new(error)).Elem()
var byteArrType = reflect.TypeOf([]byte{})

type DecodedHandlerFunction struct {
	structType reflect.Type

	bodyContentType string
	bodyInfo        *tagInfo

	pathParams    map[string]tagInfo
	headerParams  map[string]tagInfo
	queryParams   map[string]tagInfo
	contextValues map[string]tagInfo
}

// Adapt was created to simplify the parsing and validation
// of the request arguments.
//
// The input argument must be a function callback whose first
// argument is a *fiber.Ctx and the second is a struct
// where each attribute contains a special Tag describing
// from where it should be parsed, e.g.:
//
//	func MyAdaptedHandler(ctx *fiber.Ctx, args struct{
//	  PathArgument   int          `path:"my_path_arg"`
//	  QueryArgument  uint64       `query:"my_query_arg"`
//	  HeaderArgument string       `header:"my_header_arg"`
//	  ContextValue   MyCustomType `context:"my_context_value"`
//	  Body           MyCustomBody `content-type:"application/json"`
//	}) error {
//
//	  // ... handle request ...
//
//	  return nil
//	}
//
// Note: all attributes in the input struct must be public or the adapter will panic
func NewHandlerFactory[CtxT any](newAdapter func(CtxT) RequestAdapter, fn interface{}) func(ctx CtxT) error {
	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	var c CtxT

	// The slow steps that heavily rely on reflection
	// are done here once during startup in order to affect
	// as little as possible the performance later on.
	fnInfo := DecodeHandlerFunction(fnType, []reflect.Type{
		// These are the types of the arguments we expect the function to receive
		// before the "args struct" which should always be the last argument.
		//
		// If the input function doesn't match this list the adapter will panic at startup.
		reflect.TypeOf(c),
	})

	return func(ctx CtxT) error {
		// This part uses cached information from `fnInfo` and uses
		// reflection only to fill the struct making it more performatic:
		inputStructPtr, err := UnmarshalRequestAsStruct(newAdapter(ctx), fnInfo)
		if err != nil {
			return err
		}

		// Here we pass the arguments to the user defined handler function in the order
		// we expect to receive them, e.g.: `func(ctx *fiber.Ctx, args MyStruct) error`:
		err, _ = fnValue.Call([]reflect.Value{reflect.ValueOf(ctx), inputStructPtr.Elem()})[0].Interface().(error)
		return err
	}
}

func DecodeHandlerFunction(fnType reflect.Type, expectedArgTypes []reflect.Type) DecodedHandlerFunction {
	if len(expectedArgTypes) == 0 {
		log.Fatal("adapter code error: the expected list of args for the handler must not be an empty list!")
	}

	if fnType.Kind() != reflect.Func {
		log.Fatal("adapt's argument must be a function!")
	}

	if fnType.NumIn() != 2 {
		log.Fatal("received function must have 2 arguments!")
	}

	if fnType.In(0) != expectedArgTypes[0] {
		log.Fatalf("first argument must be of type %v!", expectedArgTypes[0])
	}

	if fnType.NumOut() != 1 {
		log.Fatal("received function must have a single return value!")
	}

	if fnType.Out(0) != errType {
		log.Fatal("first return value must be of type error")
	}

	structType := fnType.In(1)
	if structType.Kind() != reflect.Struct {
		log.Fatal("second argument must a struct!")
	}

	bodyContentType, bodyInfo := getBodyInfo(structType)
	pathParams, headerParams, queryParams, contextValues := getTagNames(structType)
	return DecodedHandlerFunction{
		structType:      structType,
		bodyContentType: bodyContentType,
		bodyInfo:        bodyInfo,
		pathParams:      pathParams,
		headerParams:    headerParams,
		queryParams:     queryParams,
		contextValues:   contextValues,
	}
}

func UnmarshalRequestAsStruct(request RequestAdapter, funcInfo DecodedHandlerFunction) (inputStruct reflect.Value, _ error) {
	inputStruct = reflect.New(funcInfo.structType)
	if funcInfo.bodyInfo != nil {
		var param reflect.Value
		switch funcInfo.bodyContentType {
		case "application/json":
			param = reflect.New(funcInfo.structType.Field(funcInfo.bodyInfo.Idx).Type)
			err := json.Unmarshal(request.GetBody(), param.Interface())
			if err != nil {
				return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"could not parse body as JSON: %s", err.Error(),
				))
			}
			// Dereference the pointer:
			param = param.Elem()

		case "application/octet-stream":
			param = reflect.ValueOf(request.GetBody())
		default:
			panic(fmt.Sprintf(
				"code error: unexpected mimetype received: '%s', for the Body field",
				funcInfo.bodyContentType,
			))
		}

		inputStruct.Elem().Field(funcInfo.bodyInfo.Idx).Set(param)
	}

	for key, info := range funcInfo.pathParams {
		param := request.GetPathParam(key)
		if param == "" {
			// Path params are always required, that's why we won't
			// check the Default and Required fields here
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"path param '%s' is empty", key,
			))
		}

		v, err := decodeType(info.Kind, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"could not convert path param to %s: %s", reflect.Int, err.Error(),
			))
		}

		inputStruct.Elem().Field(info.Idx).Set(v)
	}

	for key, info := range funcInfo.headerParams {
		param := request.GetHeaderParam(key)
		if param == "" {
			param = info.Default
		}
		if param == "" {
			if info.Required {
				return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required header param '%s' is empty", key,
				))
			}

			continue
		}

		v, err := decodeType(info.Kind, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"could not convert path param to %s: %s", reflect.Int, err.Error(),
			))
		}

		inputStruct.Elem().Field(info.Idx).Set(v)
	}

	for key, info := range funcInfo.queryParams {
		param := request.GetQueryParam(key)
		if param == "" {
			param = info.Default
		}
		if param == "" {
			if info.Required {
				return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required query param '%s' is empty", key,
				))
			}

			continue
		}

		v, err := decodeType(info.Kind, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"could not convert path param to %s: %s", reflect.Int, err.Error(),
			))
		}

		inputStruct.Elem().Field(info.Idx).Set(v)
	}

	for key, info := range funcInfo.contextValues {
		param := request.GetContextValue(key)
		if info.Required && param == nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"required user value '%s' is empty", key,
			))
		}

		paramV := reflect.ValueOf(param)
		canConvert := paramV.Type().ConvertibleTo(info.Type)
		if !canConvert {
			return reflect.Value{}, request.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf(
				"could not convert userValue %s to type %T", param, info.Type,
			))
		}

		inputStruct.Elem().Field(info.Idx).Set(paramV.Convert(info.Type))
	}

	return inputStruct, nil
}

func decodeType(kind reflect.Kind, v string) (reflect.Value, error) {
	switch kind {
	case reflect.Int:
		i, err := strconv.Atoi(v)
		return reflect.ValueOf(i), err
	case reflect.Int8:
		i, err := strconv.ParseInt(v, 10, 8)
		return reflect.ValueOf(int8(i)), err
	case reflect.Int16:
		i, err := strconv.ParseInt(v, 10, 16)
		return reflect.ValueOf(int16(i)), err
	case reflect.Int32:
		i, err := strconv.ParseInt(v, 10, 32)
		return reflect.ValueOf(int32(i)), err
	case reflect.Int64:
		i, err := strconv.ParseInt(v, 10, 64)
		return reflect.ValueOf(int64(i)), err

	case reflect.Uint:
		i, err := strconv.Atoi(v)
		return reflect.ValueOf(uint(i)), err
	case reflect.Uint8:
		i, err := strconv.ParseUint(v, 10, 8)
		return reflect.ValueOf(uint8(i)), err
	case reflect.Uint16:
		i, err := strconv.ParseUint(v, 10, 16)
		return reflect.ValueOf(uint16(i)), err
	case reflect.Uint32:
		i, err := strconv.ParseUint(v, 10, 32)
		return reflect.ValueOf(uint32(i)), err
	case reflect.Uint64:
		i, err := strconv.ParseUint(v, 10, 64)
		return reflect.ValueOf(uint64(i)), err
	}

	return reflect.ValueOf(v), nil
}

type tagInfo struct {
	Idx      int
	Required bool
	Kind     reflect.Kind
	Type     reflect.Type
	Default  string // TODO: use a reflect.Value instead for saving on the conversion time
}

func getBodyInfo(t reflect.Type) (contentType string, info *tagInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		if name == "Body" {
			opts := strings.Split(field.Tag.Get("content-type"), ",")
			contentType = opts[0]
			if field.Type == byteArrType {
				contentType = "application/octet-stream"
			}

			switch contentType {
			case "":
				contentType = "application/json"
			case "application/json":
			case "application/octet-stream":
			default:
				panic(fmt.Sprintf(
					"mimetype '%s' is not supported yet for field %s",
					contentType,
					field.Name,
				))
			}

			return contentType, &tagInfo{
				Idx:      i,
				Required: true,
				Kind:     field.Type.Kind(),
				Type:     field.Type,
			}
		}
	}

	return "", nil
}

// This function collects only the names
// that will be used from the type
// this should save several calls to `Field(i).Tag.Get("foo")`
// which might improve the performance by a lot.
func getTagNames(t reflect.Type) (
	pathParams map[string]tagInfo,
	headerParams map[string]tagInfo,
	queryParams map[string]tagInfo,
	contextValues map[string]tagInfo,
) {
	pathParams = map[string]tagInfo{}
	headerParams = map[string]tagInfo{}
	queryParams = map[string]tagInfo{}
	contextValues = map[string]tagInfo{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("path"), ",")

		key := opts[0]
		if key == "" {
			continue
		}

		pathParams[key] = tagInfo{
			Idx:      i,
			Required: true,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("header"), ",")

		key := opts[0]
		if key == "" {
			continue
		}

		required := true
		if len(opts) > 1 && opts[1] == "optional" {
			required = false
		}

		headerParams[key] = tagInfo{
			Idx:      i,
			Required: required,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
			Default:  field.Tag.Get("default"),
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("query"), ",")
		key := opts[0]
		if key == "" {
			continue
		}

		required := false
		if len(opts) > 1 && opts[1] == "required" {
			required = true
		}

		queryParams[key] = tagInfo{
			Idx:      i,
			Required: required,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
			Default:  field.Tag.Get("default"),
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("context"), ",")
		key := opts[0]
		if key == "" {
			continue
		}

		required := true
		if len(opts) > 1 && opts[1] == "optional" {
			required = false
		}

		contextValues[key] = tagInfo{
			Idx:      i,
			Required: required,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	return
}
