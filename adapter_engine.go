package kapi

import (
	"encoding/json"
	"errors"
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
		const source = "path param"

		param := request.GetPathParam(key)
		if param == "" {
			// Path params are always required, that's why we won't
			// check the Default and Required fields here
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"%s '%s' is empty", source, key,
			))
		}

		v, err := decodeType(info.Kind, source, key, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		inputStruct.Elem().Field(info.Idx).Set(v)
	}
	for key, info := range funcInfo.headerParams {
		const source = "header param"

		param := request.GetHeaderParam(key)
		if param == "" {
			param = info.Default
		}
		if param == "" {
			if info.Required {
				return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required %s '%s' is empty", source, key,
				))
			}

			continue
		}

		v, err := decodeType(info.Kind, source, key, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		inputStruct.Elem().Field(info.Idx).Set(v)
	}
	for key, info := range funcInfo.queryParams {
		const source = "query param"

		param := request.GetQueryParam(key)
		if param == "" {
			param = info.Default
		}
		if param == "" {
			if info.Required {
				return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required %s '%s' is empty", source, key,
				))
			}

			continue
		}

		v, err := decodeType(info.Kind, source, key, param)
		if err != nil {
			return reflect.Value{}, request.NewHTTPError(http.StatusBadRequest, err.Error())
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

// decodeType converts the raw string value `v` received from `source`
// (e.g. "path param", "header param", "query param") under the name
// `name` into a reflect.Value of the requested `kind`.
//
// On failure it returns a *decodeError instead of the raw strconv
// error, so callers get a message that is safe to show a user as-is.
func decodeType(kind reflect.Kind, source string, name string, v string) (reflect.Value, error) {
	switch kind {
	case reflect.Int:
		i, err := strconv.Atoi(v)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(i), nil
	case reflect.Int8:
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(int8(i)), nil
	case reflect.Int16:
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(int16(i)), nil
	case reflect.Int32:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(int32(i)), nil
	case reflect.Int64:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(i), nil

	case reflect.Uint:
		// Parsed at strconv.IntSize (the platform's uint width, rather
		// than strconv.Atoi+conversion) so a negative input is rejected
		// instead of silently wrapping into a huge positive number, and
		// so a value that overflows a 32-bit platform's uint is caught
		// here instead of being truncated by the uint(i) conversion below.
		i, err := strconv.ParseUint(v, 10, strconv.IntSize)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(uint(i)), nil
	case reflect.Uint8:
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(uint8(i)), nil
	case reflect.Uint16:
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(uint16(i)), nil
	case reflect.Uint32:
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(uint32(i)), nil
	case reflect.Uint64:
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return reflect.Value{}, newDecodeError(source, name, kind, v, err)
		}
		return reflect.ValueOf(i), nil
	}

	return reflect.ValueOf(v), nil
}

// decodeError is returned by decodeType when a request value can't be
// converted to a field's target type. Error() renders a clean message
// naming the source, param, expected type and offending value — never
// the raw strconv wording (e.g. "invalid syntax") — for the HTTP
// response body. Unwrap() keeps the original strconv error reachable
// via errors.Is/errors.As for tests and any future caller that needs
// to branch on the underlying failure instead of the rendered string.
type decodeError struct {
	msg   string
	cause error
}

func (e *decodeError) Error() string {
	return e.msg
}

func (e *decodeError) Unwrap() error {
	return e.cause
}

func newDecodeError(source string, name string, kind reflect.Kind, value string, cause error) error {
	explanation := fmt.Sprintf("must be a valid %s", kind)

	var numErr *strconv.NumError
	if errors.As(cause, &numErr) && errors.Is(numErr.Err, strconv.ErrRange) {
		explanation = fmt.Sprintf("is out of range for type %s", kind)
	}

	return &decodeError{
		msg: fmt.Sprintf(
			"%s %q %s, got %q", source, name, explanation, value,
		),
		cause: cause,
	}
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
